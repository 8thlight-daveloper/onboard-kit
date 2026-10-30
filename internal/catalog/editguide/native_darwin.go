package editguide

import (
	"encoding/binary"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

func inspect(file *os.File) (controls, error) {
	fd := int(file.Fd())
	var before unix.Stat_t
	if unix.Fstat(fd, &before) != nil {
		return controls{}, failNative()
	}
	if !absentACL(fd) {
		return controls{}, failNative()
	}
	values, err := attributes(fd)
	if err != nil {
		return controls{}, err
	}
	var after unix.Stat_t
	if unix.Fstat(fd, &after) != nil || before.Dev != after.Dev || before.Ino != after.Ino || before.Mode != after.Mode || before.Uid != after.Uid || before.Gid != after.Gid || before.Flags != after.Flags || (before.Mode&unix.S_IFMT != unix.S_IFDIR && (before.Nlink != after.Nlink || before.Size != after.Size || before.Mtim != after.Mtim || before.Ctim != after.Ctim)) {
		return controls{}, failNative()
	}
	return controls{Device: uint64(after.Dev), Inode: uint64(after.Ino), Mode: uint32(after.Mode), Links: uint64(after.Nlink), UID: after.Uid, GID: after.Gid, Size: after.Size, Modified: after.Mtim, Changed: after.Ctim, Flags: after.Flags, Attributes: values}, nil
}

func absentACL(fd int) bool {
	attributes := unix.Attrlist{Bitmapcount: 5, Commonattr: unix.ATTR_CMN_EXTENDED_SECURITY}
	buffer := make([]byte, 65536+12)
	_, _, errno := unix.Syscall6(unix.SYS_FGETATTRLIST, uintptr(fd), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), unix.FSOPT_REPORT_FULLSIZE, 0)
	if errno != 0 {
		return false
	}
	order := binary.NativeEndian
	total := uint64(order.Uint32(buffer[:4]))
	if total < 12 || total > uint64(len(buffer)) {
		return false
	}
	offset := int64(int32(order.Uint32(buffer[4:8])))
	length := uint64(order.Uint32(buffer[8:12]))
	if length > 65536 {
		return false
	}
	if length == 0 {
		return true
	}
	start := int64(4) + offset
	if start < 12 || start%4 != 0 || uint64(start) > total || length > total-uint64(start) || length != 44 {
		return false
	}
	security := buffer[int(start) : int(start)+int(length)]
	return order.Uint32(security[:4]) == 0x012cc16d && order.Uint32(security[36:40]) == ^uint32(0) && order.Uint32(security[40:44]) == 0
}

func supportedAttributes(values map[string]string) bool {
	if len(values) == 0 {
		return true
	}
	if len(values) != 1 {
		return false
	}
	_, exists := values["com.apple.provenance"]
	return exists
}

func establishAttributes(fd int, values map[string]string) error {
	if !supportedAttributes(values) {
		return failNative()
	}
	if value, exists := values["com.apple.provenance"]; exists {
		return unix.Fsetxattr(fd, "com.apple.provenance", []byte(value), 0)
	}
	return nil
}
