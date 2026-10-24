package editor_test

import (
	"encoding/base64"
	"encoding/binary"
	"os"
	"runtime"
	"syscall"
	"testing"
	"unsafe"
)

type attributes struct {
	Count, Reserved                       uint16
	Common, Volume, Directory, File, Fork uint32
}

func nativeACL(t *testing.T, path string) string {
	t.Helper()
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" {
		t.Skip("actual native filesec fixture requires little-endian Darwin")
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	attrs := attributes{Count: 5, Common: 0x00400000}
	buffer := make([]byte, 65548)
	_, _, failure := syscall.Syscall6(syscall.SYS_FGETATTRLIST, file.Fd(), uintptr(unsafe.Pointer(&attrs)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0, 0)
	runtime.KeepAlive(file)
	runtime.KeepAlive(attrs)
	runtime.KeepAlive(buffer)
	if failure != 0 {
		t.Fatalf("actual native filesec inspection failed: %d", failure)
	}
	used := uint64(binary.LittleEndian.Uint32(buffer[:4]))
	size := uint64(binary.LittleEndian.Uint32(buffer[8:12]))
	if used < 12 || used > uint64(len(buffer)) {
		t.Fatal("native ACL bounds")
	}
	if size == 0 {
		return "absent"
	}
	start := int64(4) + int64(int32(binary.LittleEndian.Uint32(buffer[4:8])))
	if start < 12 || uint64(start) > used || size < 44 || size > used-uint64(start) {
		t.Fatal("native filesec extent")
	}
	body := buffer[start : uint64(start)+size]
	if binary.LittleEndian.Uint32(body[:4]) != 0x012cc16d {
		t.Fatal("native filesec magic")
	}
	if binary.LittleEndian.Uint32(body[36:40]) == ^uint32(0) {
		return "absent"
	}
	return "present:" + base64.StdEncoding.EncodeToString(body)
}
func nativeFlags(t *testing.T, _ string, info os.FileInfo) uint64 {
	t.Helper()
	return uint64(info.Sys().(*syscall.Stat_t).Flags)
}
