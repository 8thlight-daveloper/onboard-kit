package creator

import (
	"errors"

	"golang.org/x/sys/unix"
)

func privateACL(fd int) bool {
	for _, name := range []string{"system.posix_acl_access", "system.posix_acl_default"} {
		size, err := unix.Fgetxattr(fd, name, nil)
		if errors.Is(err, unix.ENODATA) || errors.Is(err, unix.ENOTSUP) {
			continue
		}
		if err != nil || size < 0 || size > 65536 {
			return false
		}
		return false
	}
	return true
}

func publishExclusive(parent int, from, to string) error {
	return unix.Renameat2(parent, from, parent, to, unix.RENAME_NOREPLACE)
}
