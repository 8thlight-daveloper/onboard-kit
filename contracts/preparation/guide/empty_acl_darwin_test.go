package guide_test

import (
	"encoding/binary"
	"os"
	"reflect"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"unsafe"

	"golang.org/x/sys/unix"
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
	if failure == unix.ENOTSUP || failure == unix.ENOSYS {
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
	_, _, failure := unix.Syscall6(unix.SYS_FGETATTRLIST, file.Fd(), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0, 0)
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
	_, _, failure := unix.Syscall6(unix.SYS_FSETATTRLIST, file.Fd(), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0, 0)
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

func TestDarwinExplicitEmptyACLIsNotNoACL(t *testing.T) {
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" {
		t.Skip("native filesec fixture specifies little-endian Darwin arm64 and amd64")
	}
	for _, changed := range []bool{false, true} {
		t.Run(map[bool]string{false: "noop", true: "changed-save"}[changed], func(t *testing.T) {
			box := newSandbox(t, "identity.yaml")
			held := establishExplicitEmptyACL(t, box.Prep)
			before := snapshot(t, box.Root)
			info, err := os.Stat(box.Prep)
			if err != nil {
				t.Fatal("cannot observe owned preparation metadata")
			}
			input := "review\nsave\n"
			if changed {
				input = "recipient\n1\nchange\nExample newcomer\napply\nreview\nsave\n"
			}
			var stdout, stderr strings.Builder
			exit := newCommand().Run(box.args(), strings.NewReader(input), &stdout, &stderr)
			want := before
			saved := false
			if !changed {
				if exit != 0 {
					t.Fatalf("empty ACL no-op exit %d", exit)
				}
				assertReport(t, []byte(stderr.String()), "unchanged", "guide.unchanged")
			} else if exit == 0 {
				saved = true
				assertReport(t, []byte(stderr.String()), "saved", "guide.saved")
				want["records/newcomer/preparation.obk.yaml"] = string(fixture(t, "named.yaml"))
			} else {
				if exit != 1 {
					t.Fatalf("empty ACL changed-save exit %d", exit)
				}
				assertReport(t, []byte(stderr.String()), "refused", "guide.commit")
			}
			if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, want) {
				t.Fatal("empty ACL save changed unexpected bytes or paths")
			}
			after, err := os.Stat(box.Prep)
			if err != nil {
				t.Fatal("owned preparation disappeared")
			}
			if after.Mode().Perm() != info.Mode().Perm() {
				t.Fatal("empty ACL save changed existing mode")
			}
			replaced := !os.SameFile(info, after)
			if replaced != saved {
				t.Fatal("successful replacement must replace pathname inode; refusal/no-op must preserve original inode")
			}
			actual := readNativeFilesec(t, box.Prep)
			if !actual.Present || actual.Count != 0 || actual.Flags != held.Flags || !reflect.DeepEqual(actual.Payload, held.Payload) {
				t.Fatal("explicit zero-entry ACL was changed or silently replaced by absence")
			}
		})
	}
}
