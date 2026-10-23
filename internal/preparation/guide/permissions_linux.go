package guide

import (
	"errors"

	"golang.org/x/sys/unix"
)

func aclAbsent(fd int) bool {
	for _, name := range []string{"system.posix_acl_access", "system.posix_acl_default"} {
		_, err := unix.Fgetxattr(fd, name, nil)
		if !errors.Is(err, unix.ENODATA) && !errors.Is(err, unix.ENOTSUP) {
			return false
		}
	}
	return true
}
