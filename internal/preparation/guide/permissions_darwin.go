package guide

import (
	"encoding/binary"
	"unsafe"

	"golang.org/x/sys/unix"
)

func permissionsSafe(parent int, name string) bool {
	fd, err := unix.Openat(parent, name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	if err != nil {
		return false
	}
	defer unix.Close(fd)
	attributes := unix.Attrlist{Bitmapcount: unix.ATTR_BIT_MAP_COUNT, Commonattr: unix.ATTR_CMN_EXTENDED_SECURITY}
	buffer := make([]byte, 4096)
	_, _, failure := unix.Syscall6(unix.SYS_FGETATTRLIST, uintptr(fd), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0, 0)
	if failure != 0 {
		return false
	}
	size := binary.NativeEndian.Uint32(buffer[:4])
	if size < 12 || size > uint32(len(buffer)) {
		return false
	}
	length := binary.NativeEndian.Uint32(buffer[8:12])
	if length == 0 {
		return true
	}
	offset := int(int32(binary.NativeEndian.Uint32(buffer[4:8]))) + 4
	const fileSecurityBaseSize = 44
	const fileSecurityMagic = 0x012cc16d
	const fileSecurityAbsentACL = ^uint32(0)
	if offset < 12 || length != fileSecurityBaseSize || uint64(offset)+uint64(length) > uint64(size) {
		return false
	}
	if binary.NativeEndian.Uint32(buffer[offset:offset+4]) != fileSecurityMagic {
		return false
	}
	for _, value := range buffer[offset+4 : offset+36] {
		if value != 0 {
			return false
		}
	}
	count := binary.NativeEndian.Uint32(buffer[offset+36 : offset+40])
	flags := binary.NativeEndian.Uint32(buffer[offset+40 : offset+44])
	return count == fileSecurityAbsentACL && flags == 0
}
