package guide

import (
	"errors"

	"golang.org/x/sys/unix"
)

func permissionsSafe(parent int, name string) bool {
	fd, err := unix.Openat(parent, name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	if err != nil {
		return false
	}
	defer unix.Close(fd)
	for _, attribute := range []string{"system.posix_acl_access", "system.posix_acl_default"} {
		size, err := unix.Fgetxattr(fd, attribute, nil)
		if !errors.Is(err, unix.ENODATA) && !errors.Is(err, unix.ENOTSUP) && !(err == nil && size == 0) {
			return false
		}
	}
	return true
}
