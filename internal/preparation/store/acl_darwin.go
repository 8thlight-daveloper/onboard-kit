package store

import (
	"encoding/binary"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	attributeBuffer  = 8192
	attributeBitmaps = 5
)

func aclAbsent(descriptor int) bool {
	absent, answered := aclAnswer(descriptor)
	return answered && absent
}

func aclAnswerable(descriptor int) bool {
	_, answered := aclAnswer(descriptor)
	return answered
}

func aclAnswer(descriptor int) (bool, bool) {
	attributes := unix.Attrlist{Bitmapcount: attributeBitmaps, Commonattr: unix.ATTR_CMN_EXTENDED_SECURITY}
	var buffer [attributeBuffer]byte
	_, _, failure := syscall.Syscall6(syscall.SYS_FGETATTRLIST, uintptr(descriptor), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0, 0)
	runtime.KeepAlive(&attributes)
	runtime.KeepAlive(&buffer)
	if failure != 0 {
		return false, false
	}
	return binary.NativeEndian.Uint32(buffer[8:12]) == 0, true
}
