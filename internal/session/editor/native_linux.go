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
	result.Flags = uint32(flags)
	return result, true
}
func replaceableFlags(flags uint32) bool { return flags == 0 }
func setNativeFlags(fd int, flags uint32) bool {
	actual, err := unix.IoctlGetInt(fd, unix.FS_IOC_GETFLAGS)
	return err == nil && uint32(actual) == flags
}
