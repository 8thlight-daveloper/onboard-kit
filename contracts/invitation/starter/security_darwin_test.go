package starter_test

import (
	"encoding/binary"
	"os"
	"runtime"
	"syscall"
	"testing"
	"unsafe"
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

func nativeACL(t *testing.T, path string) (bool, string, bool, string) {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if file.Close() != nil {
			t.Error("native fixture descriptor close failed")
		}
	}()
	attrs := nativeAttributes{BitmapCount: 5, Common: 0x00400000}
	buffer := make([]byte, 65548)
	_, _, failure := syscall.Syscall6(syscall.SYS_FGETATTRLIST, file.Fd(), uintptr(unsafe.Pointer(&attrs)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0, 0)
	runtime.KeepAlive(file)
	runtime.KeepAlive(attrs)
	runtime.KeepAlive(buffer)
	if failure == syscall.ENOTSUP || failure == syscall.ENOSYS {
		t.Skipf("native ACL fixture inspection unsupported: errno %d", failure)
	}
	if failure != 0 {
		t.Fatalf("native ACL fixture inspection failed: errno %d", failure)
	}
	used := uint64(binary.LittleEndian.Uint32(buffer[:4]))
	if used < 12 || used > uint64(len(buffer)) {
		t.Fatal("native ACL fixture response length")
	}
	size := uint64(binary.LittleEndian.Uint32(buffer[8:12]))
	if size == 0 {
		return false, "", false, ""
	}
	start := int64(4) + int64(int32(binary.LittleEndian.Uint32(buffer[4:8])))
	if start < 12 || uint64(start) > used || size > used-uint64(start) || size < 44 {
		t.Fatal("native ACL fixture bounds")
	}
	payload := buffer[start : uint64(start)+size]
	if binary.LittleEndian.Uint32(payload[:4]) != 0x012cc16d {
		t.Fatal("native ACL fixture magic")
	}
	if binary.LittleEndian.Uint32(payload[36:40]) == ^uint32(0) {
		return false, "", false, ""
	}
	return true, encoded(payload), false, ""
}
