package editguide

import (
	"encoding/binary"
	"unsafe"

	"golang.org/x/sys/unix"
)

const provenanceName = "com.apple.provenance"

func supportedAttribute(name string) bool { return name == provenanceName }

func nativeFlags(fd int, stat *unix.Stat_t) (uint32, bool) { return stat.Flags, true }

func timestamps(stat *unix.Stat_t) (int64, int64, int64, int64) {
	return int64(stat.Mtim.Sec), int64(stat.Mtim.Nsec), int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec)
}

func accessControlPresent(fd int) (bool, bool) {
	request := unix.Attrlist{Bitmapcount: 5, Commonattr: unix.ATTR_CMN_EXTENDED_SECURITY}
	header := make([]byte, 12)
	if !readAttributeList(fd, &request, header) {
		return false, false
	}
	size := int(binary.LittleEndian.Uint32(header[:4]))
	if size < 12 || size > maxRepresentationBytes+12 {
		return false, false
	}
	data := make([]byte, size)
	if !readAttributeList(fd, &request, data) || int(binary.LittleEndian.Uint32(data[:4])) != size {
		return false, false
	}
	offset := int(int32(binary.LittleEndian.Uint32(data[4:8]))) + 4
	length := int(binary.LittleEndian.Uint32(data[8:12]))
	if length == 0 {
		return false, true
	}
	if offset < 12 || length > maxRepresentationBytes || offset > size || length > size-offset {
		return false, false
	}
	acl := data[offset : offset+length]
	empty := len(acl) >= 44 && binary.LittleEndian.Uint32(acl[:4]) == 0x012cc16d && binary.LittleEndian.Uint32(acl[36:40]) == 0xffffffff
	return !empty, true
}

func readAttributeList(fd int, request *unix.Attrlist, buffer []byte) bool {
	_, _, errno := unix.Syscall6(unix.SYS_FGETATTRLIST, uintptr(fd), uintptr(unsafe.Pointer(request)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), unix.FSOPT_REPORT_FULLSIZE, 0)
	return errno == 0
}
