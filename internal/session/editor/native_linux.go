package editor

import "golang.org/x/sys/unix"

func platformAttributes(fd int) (nativeAttributes, bool) {
	result := nativeAttributes{}
	for _, name := range []string{"system.posix_acl_access", "system.posix_acl_default"} {
		size, err := unix.Fgetxattr(fd, name, nil)
		if err == unix.ENODATA {
			continue
		}
		if err != nil || size < 0 || size > 65536 {
			return nativeAttributes{}, false
		}
		value := make([]byte, size)
		n, err := unix.Fgetxattr(fd, name, value)
		if err != nil || n != size {
			return nativeAttributes{}, false
		}
		if name == "system.posix_acl_access" {
			result.AccessPresent = true
			result.Access = value
		} else {
			result.DefaultPresent = true
			result.Default = value
		}
	}
	flags, err := unix.IoctlGetInt(fd, unix.FS_IOC_GETFLAGS)
	if err != nil {
		return nativeAttributes{}, false
	}
	result.Flags = uint32(flags) &^ filesystemManagedFlags
	return result, true
}

const filesystemManagedFlags = 0x80000

func replaceableFlags(flags uint32) bool { return flags == 0 }
func setNativeFlags(fd int, flags uint32) bool {
	actual, err := unix.IoctlGetInt(fd, unix.FS_IOC_GETFLAGS)
	return err == nil && uint32(actual)&^filesystemManagedFlags == flags
}

const casefoldFlag = 0x40000000

func nativeComparisonProfile(fd int) int {
	var system unix.Statfs_t
	if unix.Fstatfs(fd, &system) != nil {
		return 0
	}
	switch system.Type {
	case unix.EXT4_SUPER_MAGIC, unix.XFS_SUPER_MAGIC, unix.BTRFS_SUPER_MAGIC, unix.TMPFS_MAGIC:
	default:
		return 0
	}
	flags, err := unix.IoctlGetInt(fd, unix.FS_IOC_GETFLAGS)
	if err != nil || uint32(flags)&casefoldFlag != 0 {
		return 0
	}
	return 1
}
