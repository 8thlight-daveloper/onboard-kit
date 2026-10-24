package starter

import "golang.org/x/sys/unix"

func nativeACL(fd int) (string, bool, []byte, bool, []byte, error) {
	read := func(name string) (bool, []byte, error) {
		n, err := unix.Fgetxattr(fd, name, nil)
		if err == unix.ENODATA {
			return false, nil, nil
		}
		if err != nil {
			return false, nil, err
		}
		if n > 65536 {
			return false, nil, errContext
		}
		data := make([]byte, n)
		n, err = unix.Fgetxattr(fd, name, data)
		if err != nil {
			return false, nil, err
		}
		return true, data[:n], nil
	}
	a, b, err := read("system.posix_acl_access")
	if err != nil {
		return "", false, nil, false, nil, err
	}
	c, d, err := read("system.posix_acl_default")
	return "linux-posix-v1", a, b, c, d, err
}
func publishAbsent(from int, name string, to int, target string) error {
	return unix.Renameat2(from, name, to, target, unix.RENAME_NOREPLACE)
}
