package editguide_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

const editProvenanceName = "com.apple.provenance"

func editProvenanceOpen(t *testing.T, path string, writable bool) *os.File {
	t.Helper()
	flags := unix.O_RDONLY | unix.O_NOFOLLOW | unix.O_CLOEXEC
	if writable {
		flags = unix.O_RDWR | unix.O_NOFOLLOW | unix.O_CLOEXEC
	}
	fd, err := unix.Open(path, flags, 0)
	if err != nil {
		t.Fatal(err)
	}
	file := os.NewFile(uintptr(fd), path)
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Error("owned native fixture descriptor close failed")
		}
	})
	return file
}

func editProvenanceAttrs(t *testing.T, file *os.File) map[string]string {
	t.Helper()
	fd := int(file.Fd())
	n, err := unix.Flistxattr(fd, nil)
	if err == unix.ENOTSUP || err == unix.ENOSYS {
		t.Skip("named native fixture gap: Darwin descriptor xattr enumeration unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	if n < 0 || n > 65536 {
		t.Fatal("owned xattr name representation exceeds fixture bound")
	}
	buffer := make([]byte, n)
	used, err := unix.Flistxattr(fd, buffer)
	if err != nil || used != n {
		t.Fatal("owned xattr enumeration changed during bounded capture")
	}
	result := make(map[string]string)
	total := n
	for len(buffer) > 0 {
		end := bytes.IndexByte(buffer, 0)
		if end <= 0 {
			t.Fatal("invalid native xattr name framing")
		}
		name := string(buffer[:end])
		buffer = buffer[end+1:]
		if _, exists := result[name]; exists {
			t.Fatal("duplicate native xattr name")
		}
		size, e := unix.Fgetxattr(fd, name, nil)
		if e != nil {
			t.Fatal(e)
		}
		if size < 0 || size > 65536 || total > 262144-size {
			t.Fatal("owned xattr value exceeds fixture bound")
		}
		total += size
		body := make([]byte, size)
		actual, e := unix.Fgetxattr(fd, name, body)
		if e != nil || actual != size {
			t.Fatal("owned xattr value changed during bounded capture")
		}
		result[name] = string(body)
	}
	return result
}

func editProvenanceFact(t *testing.T, file *os.File) map[string]string {
	t.Helper()
	attrs := editProvenanceAttrs(t, file)
	if len(attrs) != 1 {
		t.Skip("named native fixture gap: selected owned file does not have exactly the ambient Darwin provenance attribute")
	}
	if _, ok := attrs[editProvenanceName]; !ok {
		t.Skip("named native fixture gap: ambient Darwin provenance is absent")
	}
	return attrs
}

func editProvenanceSetExact(t *testing.T, file *os.File, name, value string) {
	t.Helper()
	if err := unix.Fsetxattr(int(file.Fd()), name, []byte(value), 0); err != nil {
		t.Skipf("named native fixture gap: owned descriptor xattr assignment unavailable: %v", err)
	}
	if editProvenanceAttrs(t, file)[name] != value {
		t.Skip("named native fixture gap: native xattr assignment did not establish the exact requested bytes")
	}
}

type editProvenanceGuard struct {
	Device uint64
	Inode  uint64
	Mode   uint32
	UID    uint32
	GID    uint32
	Links  uint64
	Flags  uint32
	Attrs  map[string]string
}

func editProvenanceGuardAt(t *testing.T, path string) editProvenanceGuard {
	t.Helper()
	file := editProvenanceOpen(t, path, false)
	info, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatal("Darwin native stat unavailable")
	}
	return editProvenanceGuard{Device: uint64(stat.Dev), Inode: uint64(stat.Ino), Mode: uint32(stat.Mode), UID: stat.Uid, GID: stat.Gid, Links: uint64(stat.Nlink), Flags: stat.Flags, Attrs: editProvenanceAttrs(t, file)}
}

type editProvenanceOutput struct{ buffer bytes.Buffer }

func (writer *editProvenanceOutput) Write(body []byte) (int, error) {
	if len(body) > 9*1048576-writer.buffer.Len() {
		return 0, errors.New("owned fixture output cap")
	}
	return writer.buffer.Write(body)
}

func editProvenanceContextUnchanged(t *testing.T, root string, before editProvenanceGuard) {
	t.Helper()
	if !reflect.DeepEqual(before, editProvenanceGuardAt(t, root)) {
		t.Error("selected parent native controls changed")
	}
	personalPreserved(t, root)
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 3 {
		t.Error("owned directory contains unexpected temporary or sidecar")
	}
}

func TestNativeProvenanceTransferControlBeforeAndAfterData(t *testing.T) {
	root, target := setup(t, 0755)
	source := editProvenanceOpen(t, target, false)
	fact := editProvenanceFact(t, source)
	path := filepath.Join(root, "owned-provenance-transfer-control")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		file.Close()
		if err := os.Remove(path); err != nil {
			t.Error(err)
		}
	})
	if err := file.Chown(-1, os.Getegid()); err != nil {
		t.Fatal(err)
	}
	editProvenanceSetExact(t, file, editProvenanceName, fact[editProvenanceName])
	if !reflect.DeepEqual(fact, editProvenanceAttrs(t, file)) {
		t.Fatal("actual replacement attribute map differs before data")
	}
	data := []byte("OWNED NATIVE TRANSFER CONTROL\n")
	if n, err := file.Write(data); err != nil || n != len(data) {
		t.Fatal("owned transfer control write failed")
	}
	if !reflect.DeepEqual(fact, editProvenanceAttrs(t, file)) {
		t.Fatal("actual replacement attribute map differs after data")
	}
	if !reflect.DeepEqual(fact, editProvenanceAttrs(t, source)) {
		t.Fatal("transfer control changed original metadata")
	}
}

func TestReviewedSavePreservesExactNativeProvenance(t *testing.T) {
	for _, distinct := range []bool{false, true} {
		t.Run(map[bool]string{false: "current-native-fact", true: "distinguishable-native-fact"}[distinct], func(t *testing.T) {
			root, target := setup(t, 0755)
			file := editProvenanceOpen(t, target, true)
			fact := editProvenanceFact(t, file)
			if distinct {
				probePath := filepath.Join(root, "owned-ambient-control")
				probe, err := os.OpenFile(probePath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
				if err != nil {
					t.Fatal(err)
				}
				ambient := editProvenanceAttrs(t, probe)
				if err := probe.Close(); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(probePath); err != nil {
					t.Fatal(err)
				}
				candidate := []byte(fact[editProvenanceName])
				if len(candidate) == 0 {
					candidate = []byte("owned-native-provenance-control")
				} else {
					candidate = append([]byte(nil), candidate...)
					candidate[len(candidate)-1] ^= 1
				}
				editProvenanceSetExact(t, file, editProvenanceName, string(candidate))
				fact = editProvenanceFact(t, file)
				if reflect.DeepEqual(ambient, fact) {
					t.Skip("named native fixture gap: selected provenance cannot be distinguished from fresh temporary provenance")
				}
			}
			before := editProvenanceGuardAt(t, target)
			parent := editProvenanceGuardAt(t, root)
			input := &countedInput{source: []byte("retire\nresource-old\nreview\nsave\nTAIL")}
			out, stderr := &editProvenanceOutput{}, &editProvenanceOutput{}
			if newCommand().Run([]string{"--plan=" + target}, input, out, stderr) != 0 {
				t.Fatal("supported native provenance save refused")
			}
			report(t, stderr.buffer.Bytes(), "written", "edit.written", "The reviewed plan edits were saved.")
			if input.position != len("retire\nresource-old\nreview\nsave\n") {
				t.Error("save read queued tail")
			}
			data, err := os.ReadFile(target)
			if err != nil || !bytes.Equal(data, fixture(t, "retired.yaml")) {
				t.Error("actual saved bytes differ from independently authored retirement")
			}
			after := editProvenanceGuardAt(t, target)
			if !reflect.DeepEqual(after.Attrs, fact) {
				t.Error("published replacement lost or changed exact native provenance")
			}
			if before.Device == after.Device && before.Inode == after.Inode {
				t.Error("changed save modified original inode in place")
			}
			if after.Mode&07777 != 0600 || after.UID != uint32(os.Geteuid()) || after.GID != uint32(os.Getegid()) || after.Links != 1 || after.Flags != 0 {
				t.Error("published replacement lost private native profile")
			}
			editProvenanceContextUnchanged(t, root, parent)
		})
	}
}

func TestCancelAndNoopRetainNativeProvenanceAndInode(t *testing.T) {
	for _, action := range []string{"cancel\nTAIL", "review\nsave\nTAIL"} {
		t.Run(strings.Split(action, "\n")[0], func(t *testing.T) {
			root, target := setup(t, 0755)
			editProvenanceFact(t, editProvenanceOpen(t, target, false))
			before := editProvenanceGuardAt(t, target)
			parent := editProvenanceGuardAt(t, root)
			out, stderr := &editProvenanceOutput{}, &editProvenanceOutput{}
			if newCommand().Run([]string{"--plan=" + target}, strings.NewReader(action), out, stderr) != 0 {
				t.Fatal("cancel or unchanged provenance profile refused")
			}
			if strings.HasPrefix(action, "cancel") {
				report(t, stderr.buffer.Bytes(), "cancelled", "edit.cancelled", "Cancelled. The selected plan was not changed.")
			} else {
				report(t, stderr.buffer.Bytes(), "unchanged", "edit.unchanged", "The reviewed proposal leaves the selected plan unchanged.")
			}
			if !reflect.DeepEqual(before, editProvenanceGuardAt(t, target)) {
				t.Error("cancel or no-op changed original inode or metadata")
			}
			unchanged(t, target)
			editProvenanceContextUnchanged(t, root, parent)
		})
	}
}

func TestUnknownNativeAttributeStillRefusesBeforeInput(t *testing.T) {
	root, target := setup(t, 0755)
	file := editProvenanceOpen(t, target, true)
	editProvenanceFact(t, file)
	editProvenanceSetExact(t, file, "org.example.owned-fixture", "UNSUPPORTED OWNED METADATA")
	before := editProvenanceGuardAt(t, target)
	parent := editProvenanceGuardAt(t, root)
	input := &countedInput{source: []byte("review\nsave\n")}
	out, stderr := &editProvenanceOutput{}, &editProvenanceOutput{}
	if newCommand().Run([]string{"--plan=" + target}, input, out, stderr) != 1 || input.position != 0 || out.buffer.Len() != 0 {
		t.Error("unknown xattr crossed initial admission")
	}
	report(t, stderr.buffer.Bytes(), "refused", "edit.path", "Select a safe single-link effective-owned0600 plan in an owned directory without group/other write; unsupported controls require separate review.")
	if !reflect.DeepEqual(before, editProvenanceGuardAt(t, target)) {
		t.Error("unsupported original metadata was stripped or repaired")
	}
	unchanged(t, target)
	editProvenanceContextUnchanged(t, root, parent)
}

func TestChangedNativeSourceAttributeRefusesStaleSave(t *testing.T) {
	for _, kind := range []string{"provenance-value", "added-unknown"} {
		t.Run(kind, func(t *testing.T) {
			root, target := setup(t, 0755)
			file := editProvenanceOpen(t, target, true)
			fact := editProvenanceFact(t, file)
			parent := editProvenanceGuardAt(t, root)
			var changed editProvenanceGuard
			reached := false
			input := &countedInput{source: []byte("retire\nresource-old\nreview\nsave\n"), hookAt: len("retire\nresource-old\nreview\n"), hook: func() {
				if kind == "added-unknown" {
					editProvenanceSetExact(t, file, "org.example.owned-fixture", "INDEPENDENT OWNER ATTRIBUTE CHANGE")
				} else {
					candidate := []byte(fact[editProvenanceName])
					if len(candidate) == 0 {
						candidate = []byte("owned-changed-provenance")
					} else {
						candidate = append([]byte(nil), candidate...)
						candidate[len(candidate)-1] ^= 1
					}
					editProvenanceSetExact(t, file, editProvenanceName, string(candidate))
				}
				changed = editProvenanceGuardAt(t, target)
				if reflect.DeepEqual(changed.Attrs, fact) {
					t.Skip("named native fixture gap: source attribute change not independently established")
				}
				reached = true
			}}
			out, stderr := &editProvenanceOutput{}, &editProvenanceOutput{}
			if newCommand().Run([]string{"--plan=" + target}, input, out, stderr) != 1 {
				t.Error("observed changed source metadata was overwritten")
			}
			if !reached {
				t.Fatal("independent native stale mutation was not reached")
			}
			report(t, stderr.buffer.Bytes(), "refused", "edit.stale", "The selected plan context changed; reload and review again.")
			if !reflect.DeepEqual(changed, editProvenanceGuardAt(t, target)) {
				t.Error("observed owner metadata change was overwritten or repaired")
			}
			unchanged(t, target)
			editProvenanceContextUnchanged(t, root, parent)
		})
	}
}
