package editor

import (
	"encoding/binary"
	"unsafe"

	"golang.org/x/sys/unix"
)

func platformAttributes(fd int) (nativeAttributes, bool) {
	var stat unix.Stat_t
	if unix.Fstat(fd, &stat) != nil {
		return nativeAttributes{}, false
	}
	request := unix.Attrlist{Bitmapcount: 5, Commonattr: unix.ATTR_CMN_EXTENDED_SECURITY}
	header := make([]byte, 12)
	_, _, err := unix.Syscall6(unix.SYS_FGETATTRLIST, uintptr(fd), uintptr(unsafe.Pointer(&request)), uintptr(unsafe.Pointer(&header[0])), uintptr(len(header)), unix.FSOPT_REPORT_FULLSIZE, 0)
	if err != 0 {
		return nativeAttributes{}, false
	}
	size := int(binary.LittleEndian.Uint32(header[:4]))
	if size < 12 || size > 65548 {
		return nativeAttributes{}, false
	}
	data := make([]byte, size)
	_, _, err = unix.Syscall6(unix.SYS_FGETATTRLIST, uintptr(fd), uintptr(unsafe.Pointer(&request)), uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)), unix.FSOPT_REPORT_FULLSIZE, 0)
	if err != 0 || int(binary.LittleEndian.Uint32(data[:4])) != size {
		return nativeAttributes{}, false
	}
	offset := int(int32(binary.LittleEndian.Uint32(data[4:8]))) + 4
	length := int(binary.LittleEndian.Uint32(data[8:12]))
	if length == 0 {
		return nativeAttributes{Flags: stat.Flags}, true
	}
	if offset < 12 || length > 65536 || offset > size || length > size-offset {
		return nativeAttributes{}, false
	}
	acl := data[offset : offset+length]
	present := true
	if len(acl) == 0 {
		present = false
	} else if len(acl) >= 44 && binary.LittleEndian.Uint32(acl[:4]) == 0x012cc16d && binary.LittleEndian.Uint32(acl[36:40]) == 0xffffffff {
		present = false
	}
	result := nativeAttributes{AccessPresent: present, Flags: stat.Flags}
	if present {
		result.Access = append([]byte(nil), acl...)
	}
	return result, true
}
func replaceableFlags(flags uint32) bool {
	return flags&(unix.UF_IMMUTABLE|unix.UF_APPEND|unix.SF_IMMUTABLE|unix.SF_APPEND|unix.SF_RESTRICTED|unix.SF_NOUNLINK) == 0
}
func setNativeFlags(fd int, flags uint32) bool { return unix.Fchflags(fd, int(flags)) == nil }
