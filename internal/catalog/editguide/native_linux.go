package editguide

import "golang.org/x/sys/unix"

func supportedAttribute(name string) bool { return false }

func nativeFlags(fd int, stat *unix.Stat_t) (uint32, bool) {
	flags, err := unix.IoctlGetInt(fd, unix.FS_IOC_GETFLAGS)
	return uint32(flags), err == nil
}

func timestamps(stat *unix.Stat_t) (int64, int64, int64, int64) {
	return int64(stat.Mtim.Sec), int64(stat.Mtim.Nsec), int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec)
}

func accessControlPresent(fd int) (bool, bool) {
	for _, name := range []string{"system.posix_acl_access", "system.posix_acl_default"} {
		_, err := unix.Fgetxattr(fd, name, nil)
		if err == unix.ENODATA || err == unix.ENOTSUP {
			continue
		}
		return err == nil, err == nil
	}
	return false, true
}
