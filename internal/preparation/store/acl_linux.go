package store

import (
	"errors"

	"golang.org/x/sys/unix"
)

func aclAbsent(descriptor int) bool {
	for _, name := range []string{"system.posix_acl_access", "system.posix_acl_default"} {
		_, err := unix.Fgetxattr(descriptor, name, nil)
		if !errors.Is(err, unix.ENODATA) && !errors.Is(err, unix.ENOTSUP) {
			return false
		}
	}
	return true
}

func aclAnswerable(descriptor int) bool { return true }
