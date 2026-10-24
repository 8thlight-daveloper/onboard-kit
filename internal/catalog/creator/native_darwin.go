package creator

import (
	"encoding/binary"
	"unsafe"

	"golang.org/x/sys/unix"
)

func privateACL(fd int) bool {
	attributes := unix.Attrlist{Bitmapcount: 5, Commonattr: unix.ATTR_CMN_EXTENDED_SECURITY}
	var header [12]byte
	_, _, errno := unix.Syscall6(unix.SYS_FGETATTRLIST, uintptr(fd), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&header[0])), uintptr(len(header)), unix.FSOPT_REPORT_FULLSIZE, 0)
	if errno != 0 {
		return false
	}
	size := binary.LittleEndian.Uint32(header[:4])
	if size < 12 || size > 65536 {
		return false
	}
	buffer := make([]byte, int(size))
	_, _, errno = unix.Syscall6(unix.SYS_FGETATTRLIST, uintptr(fd), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), unix.FSOPT_REPORT_FULLSIZE, 0)
	if errno != 0 || binary.LittleEndian.Uint32(buffer[:4]) != size {
		return false
	}
	return binary.LittleEndian.Uint32(buffer[8:12]) == 0
}

func publishExclusive(parent int, from, to string) error {
	return unix.RenameatxNp(parent, from, parent, to, unix.RENAME_EXCL)
}
