package store_test

import (
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"unsafe"

	contract "onboard-kit/contracts/preparation/store"
)

const (
	filesecAttribute  = 0x00400000
	filesecMagic      = 0x012cc16d
	filesecHeaderSize = 44
	filesecNoACL      = ^uint32(0)
)

type nativeAttributes struct {
	BitmapCount uint16
	Reserved    uint16
	Common      uint32
	Volume      uint32
	Directory   uint32
	File        uint32
	Fork        uint32
}

type nativeFilesec struct {
	Present bool
	Count   uint32
	Flags   uint32
	Payload []byte
}

func checkNativeACLCall(t *testing.T, operation string, failure syscall.Errno) {
	t.Helper()
	if failure == syscall.ENOTSUP || failure == syscall.ENOSYS {
		t.Skipf("Darwin native %s explicitly unsupported: errno %d", operation, failure)
	}
	if failure != 0 {
		t.Fatalf("Darwin native %s failed: errno %d", operation, failure)
	}
}

func readNativeFilesec(t *testing.T, path string) nativeFilesec {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal("cannot open owned ACL fixture")
	}
	defer func() {
		if file.Close() != nil {
			t.Error("cannot close owned ACL fixture")
		}
	}()
	attributes := nativeAttributes{BitmapCount: 5, Common: filesecAttribute}
	buffer := make([]byte, 4096)
	_, _, failure := syscall.Syscall6(syscall.SYS_FGETATTRLIST, file.Fd(), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0, 0)
	runtime.KeepAlive(file)
	runtime.KeepAlive(attributes)
	runtime.KeepAlive(buffer)
	checkNativeACLCall(t, "fgetattrlist", failure)
	used := uint64(binary.LittleEndian.Uint32(buffer[:4]))
	if used < 12 || used > uint64(len(buffer)) {
		t.Fatal("invalid native attribute response size")
	}
	length := uint64(binary.LittleEndian.Uint32(buffer[8:12]))
	if length == 0 {
		return nativeFilesec{}
	}
	start := int64(4) + int64(int32(binary.LittleEndian.Uint32(buffer[4:8])))
	if start < 12 || uint64(start) > used || length > used-uint64(start) || length < filesecHeaderSize {
		t.Fatal("invalid native filesec bounds")
	}
	payload := append([]byte(nil), buffer[start:int64(uint64(start)+length)]...)
	if binary.LittleEndian.Uint32(payload[:4]) != filesecMagic {
		t.Fatal("invalid native filesec magic")
	}
	count := binary.LittleEndian.Uint32(payload[36:40])
	return nativeFilesec{Present: count != filesecNoACL, Count: count, Flags: binary.LittleEndian.Uint32(payload[40:44]), Payload: payload}
}

func establishExplicitEmptyACL(t *testing.T, path string) nativeFilesec {
	t.Helper()
	before := readNativeFilesec(t, path)
	if before.Present {
		t.Fatal("owned fixture unexpectedly already has an ACL; cannot prove no-ACL to empty-ACL transition")
	}
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal("cannot open owned file for native empty ACL fixture")
	}
	defer func() {
		if file.Close() != nil {
			t.Error("cannot close owned native ACL writer")
		}
	}()
	attributes := nativeAttributes{BitmapCount: 5, Common: filesecAttribute}
	buffer := make([]byte, 8+filesecHeaderSize)
	binary.LittleEndian.PutUint32(buffer[:4], 8)
	binary.LittleEndian.PutUint32(buffer[4:8], filesecHeaderSize)
	if len(before.Payload) >= filesecHeaderSize {
		copy(buffer[8:], before.Payload[:filesecHeaderSize])
	}
	binary.LittleEndian.PutUint32(buffer[8:12], filesecMagic)
	binary.LittleEndian.PutUint32(buffer[44:48], 0)
	_, _, failure := syscall.Syscall6(syscall.SYS_FSETATTRLIST, file.Fd(), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0, 0)
	runtime.KeepAlive(file)
	runtime.KeepAlive(attributes)
	runtime.KeepAlive(buffer)
	checkNativeACLCall(t, "fsetattrlist", failure)
	after := readNativeFilesec(t, path)
	if !after.Present || after.Count != 0 || len(after.Payload) < filesecHeaderSize {
		t.Fatal("native API did not establish an explicit zero-entry ACL distinct from no ACL")
	}
	return after
}

func hasNativeReadGrant(t *testing.T, state nativeFilesec, requiredFlags uint32) bool {
	t.Helper()
	if !state.Present {
		return false
	}
	for index := uint32(0); index < state.Count; index++ {
		start := uint64(filesecHeaderSize) + uint64(index)*24
		if start+24 > uint64(len(state.Payload)) {
			t.Fatal("invalid native ACE bounds")
		}
		flags := binary.LittleEndian.Uint32(state.Payload[start+16 : start+20])
		rights := binary.LittleEndian.Uint32(state.Payload[start+20 : start+24])
		if flags&0xf == 1 && flags&requiredFlags == requiredFlags && rights&(1<<1) != 0 {
			return true
		}
	}
	return false
}

func accessEntries(t *testing.T, path string) []string {
	t.Helper()
	value, err := exec.Command("/bin/ls", "-lde", path).CombinedOutput()
	if err != nil {
		t.Fatalf("observe disposable file ACL: %v: %s", err, value)
	}
	var entries []string
	for _, line := range strings.Split(string(value), "\n") {
		line = strings.TrimSpace(line)
		if len(line) > 0 && line[0] >= '0' && line[0] <= '9' && strings.Contains(line, ": ") {
			entries = append(entries, line)
		}
	}
	return entries
}

type aclOutcome struct {
	result contract.SaveResult
	before os.FileInfo
	after  os.FileInfo
	held   map[string]string
}

func saveUnderACL(t *testing.T, box sandbox, reviewed string) aclOutcome {
	t.Helper()
	before, err := os.Stat(box.Prep)
	if err != nil {
		t.Fatal(err)
	}
	loaded := box.load(t)
	held := snapshot(t, box.Root)
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, reviewed)})
	after, err := os.Stat(box.Prep)
	if err != nil {
		t.Fatal("preparation disappeared")
	}
	return aclOutcome{result: result, before: before, after: after, held: held}
}

func (o aclOutcome) expect(t *testing.T, box sandbox, unchanged bool, savedBytes string) {
	t.Helper()
	want := o.held
	saved := false
	switch {
	case unchanged:
		wantCode(t, o.result.Code, contract.CodeUnchanged)
	case o.result.Code == contract.CodeSaved:
		saved = true
		want[destination] = string(fixture(t, savedBytes))
	default:
		wantCode(t, o.result.Code, contract.CodeCommit)
		if o.result.Committed {
			t.Fatal("refusal reports a commit")
		}
	}
	if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, want) {
		t.Fatalf("unexpected files %#v; want %#v", actual, want)
	}
	if replaced := !os.SameFile(o.before, o.after); replaced != saved {
		t.Fatalf("inode replaced %v; saved %v", replaced, saved)
	}
}

func TestDarwinExistingAccessControlIsPreservedOrRefused(t *testing.T) {
	for _, unchanged := range []bool{true, false} {
		box := newSandbox(t, "plan.yaml", "identity.yaml")
		value, err := exec.Command("/bin/chmod", "+a", "everyone deny execute", box.Prep).CombinedOutput()
		if err != nil {
			t.Fatalf("add ACL only to disposable preparation: %v: %s", err, value)
		}
		entries := accessEntries(t, box.Prep)
		if !strings.Contains(strings.Join(entries, "\n"), "everyone deny execute") {
			t.Fatalf("fixture ACL not observed: %q", entries)
		}
		reviewed := "named.yaml"
		if unchanged {
			reviewed = "identity.yaml"
		}
		outcome := saveUnderACL(t, box, reviewed)
		outcome.expect(t, box, unchanged, "named.yaml")
		if actual := accessEntries(t, box.Prep); !reflect.DeepEqual(actual, entries) {
			t.Fatalf("existing ACL lost or changed: %q want %q", actual, entries)
		}
	}
}

func TestDarwinExplicitEmptyACLIsNotNoACL(t *testing.T) {
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" {
		t.Skip("native filesec fixture specifies little-endian Darwin arm64 and amd64")
	}
	for _, unchanged := range []bool{true, false} {
		box := newSandbox(t, "plan.yaml", "identity.yaml")
		held := establishExplicitEmptyACL(t, box.Prep)
		reviewed := "named.yaml"
		if unchanged {
			reviewed = "identity.yaml"
		}
		outcome := saveUnderACL(t, box, reviewed)
		outcome.expect(t, box, unchanged, "named.yaml")
		if outcome.after.Mode().Perm() != outcome.before.Mode().Perm() {
			t.Fatal("existing mode changed")
		}
		actual := readNativeFilesec(t, box.Prep)
		if !actual.Present || actual.Count != 0 || actual.Flags != held.Flags || !reflect.DeepEqual(actual.Payload, held.Payload) {
			t.Fatal("explicit zero-entry ACL was changed or silently replaced by absence")
		}
	}
}

func TestDarwinInheritedACLDoesNotBroadenPrivatePublication(t *testing.T) {
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" {
		t.Skip("native filesec fixture specifies little-endian Darwin arm64 and amd64")
	}
	for _, unchanged := range []bool{true, false} {
		box := newSandbox(t, "plan.yaml", "identity.yaml")
		parent := filepath.Dir(box.Prep)
		value, err := exec.Command("/bin/chmod", "+a", "everyone allow read,file_inherit,directory_inherit", parent).CombinedOutput()
		if err != nil {
			t.Fatalf("cannot add inheritable ACL to owned parent: output bytes %d", len(value))
		}
		parentState := readNativeFilesec(t, parent)
		if !hasNativeReadGrant(t, parentState, 1<<5) {
			t.Fatal("owned parent lacks native file-inheritable read grant")
		}
		if value, err = exec.Command("/bin/chmod", "-N", box.Prep).CombinedOutput(); err != nil {
			t.Fatalf("cannot remove ACL only from owned selected fixture: output bytes %d", len(value))
		}
		if readNativeFilesec(t, box.Prep).Present {
			t.Fatal("selected fixture must start with no ACL")
		}
		control := filepath.Join(parent, "owned-inheritance-control")
		writeFile(t, control, []byte("disposable inheritance control\n"))
		if !hasNativeReadGrant(t, readNativeFilesec(t, control), 1<<4) {
			t.Fatal("native file creation did not inherit the read grant; fixture does not reach the concern")
		}
		if err := os.Remove(control); err != nil {
			t.Fatal("cannot remove owned inheritance control")
		}
		reviewed := "named.yaml"
		if unchanged {
			reviewed = "identity.yaml"
		}
		outcome := saveUnderACL(t, box, reviewed)
		outcome.expect(t, box, unchanged, "named.yaml")
		if actual := readNativeFilesec(t, parent); !reflect.DeepEqual(actual.Payload, parentState.Payload) {
			t.Fatal("publication changed the parent ACL")
		}
		if outcome.after.Mode().Perm() != 0o600 {
			t.Fatal("publication lost the selected mode 0600 file")
		}
		if readNativeFilesec(t, box.Prep).Present {
			t.Fatal("publication silently granted an inherited ACL to an ACL-free preparation")
		}
	}
}

func TestDarwinNewDirectoriesAndFileInheritNoGrant(t *testing.T) {
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" {
		t.Skip("native filesec fixture specifies little-endian Darwin arm64 and amd64")
	}
	box := newSandbox(t, "plan.yaml", "")
	value, err := exec.Command("/bin/chmod", "+a", "everyone allow read,file_inherit,directory_inherit", box.Root).CombinedOutput()
	if err != nil {
		t.Fatalf("cannot add inheritable ACL to owned root: output bytes %d", len(value))
	}
	rootState := readNativeFilesec(t, box.Root)
	loaded := box.load(t)
	held := snapshot(t, box.Root)
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml"), CreateDirectories: true})
	if result.Code == contract.CodeSaved {
		for _, path := range []string{filepath.Join(box.Root, "records"), filepath.Dir(box.Prep), box.Prep} {
			if state := readNativeFilesec(t, path); hasNativeReadGrant(t, state, 0) {
				t.Fatalf("%s silently received an inherited read grant", path)
			}
		}
	} else {
		wantCode(t, result.Code, contract.CodeCommit)
		if result.Committed {
			t.Fatal("refusal reports a commit")
		}
		if after := snapshot(t, box.Root); !result.CreatedDirectories && !sameMap(after, held) {
			t.Fatal("refusal left an unreported change")
		}
	}
	if actual := readNativeFilesec(t, box.Root); !reflect.DeepEqual(actual.Payload, rootState.Payload) {
		t.Fatal("publication changed the root ACL")
	}
}
