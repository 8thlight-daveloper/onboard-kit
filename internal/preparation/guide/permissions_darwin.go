package guide

import (
	"encoding/binary"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func aclAbsent(fd int) bool {
	attributes := unix.Attrlist{Bitmapcount: 5, Commonattr: unix.ATTR_CMN_EXTENDED_SECURITY}
	var buffer [8192]byte
	_, _, errno := syscall.Syscall6(syscall.SYS_FGETATTRLIST, uintptr(fd), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0, 0)
	runtime.KeepAlive(&attributes)
	runtime.KeepAlive(&buffer)
	if errno != 0 {
		return false
	}
	length := int(binary.NativeEndian.Uint32(buffer[:4]))
	if length < 12 || length > len(buffer) {
		return false
	}
	offset := int(int32(binary.NativeEndian.Uint32(buffer[4:8])))
	size := int(binary.NativeEndian.Uint32(buffer[8:12]))
	if size == 0 {
		return true
	}
	start := 4 + offset
	if start < 12 || size < 44 || start > length || size > length-start {
		return false
	}
	return binary.NativeEndian.Uint32(buffer[start:start+4]) == 0x012cc16d && binary.NativeEndian.Uint32(buffer[start+36:start+40]) == 0xffffffff
}
