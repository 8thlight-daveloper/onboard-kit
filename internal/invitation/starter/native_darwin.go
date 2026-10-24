package starter

import (
	"encoding/binary"
	"unsafe"

	"golang.org/x/sys/unix"
)

func nativeACL(fd int) (string, bool, []byte, bool, []byte, error) {
	attributes := unix.Attrlist{Bitmapcount: 5, Commonattr: unix.ATTR_CMN_EXTENDED_SECURITY}
	buffer := make([]byte, 65536+64)
	_, _, errno := unix.Syscall6(unix.SYS_FGETATTRLIST, uintptr(fd), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0, 0)
	if errno != 0 {
		return "", false, nil, false, nil, errno
	}
	size := int(binary.LittleEndian.Uint32(buffer[:4]))
	if size < 12 || size > len(buffer) {
		return "", false, nil, false, nil, errContext
	}
	offset := 4 + int(int32(binary.LittleEndian.Uint32(buffer[4:8])))
	length := int(binary.LittleEndian.Uint32(buffer[8:12]))
	if length == 0 {
		return "darwin-filesec-v1", false, nil, false, nil, nil
	}
	if length > 65536 || offset < 12 || offset+length > size {
		return "", false, nil, false, nil, errContext
	}
	data := buffer[offset : offset+length]
	if len(data) < 44 {
		return "", false, nil, false, nil, errContext
	}
	if binary.LittleEndian.Uint32(data[36:40]) == 0xffffffff {
		return "darwin-filesec-v1", false, nil, false, nil, nil
	}
	return "darwin-filesec-v1", true, append([]byte(nil), data...), false, nil, nil
}
func publishAbsent(from int, name string, to int, target string) error {
	return unix.RenameatxNp(from, name, to, target, unix.RENAME_EXCL)
}
