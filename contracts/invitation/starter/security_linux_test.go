package starter_test

import (
	"syscall"
	"testing"
)

func nativeACL(t *testing.T, path string) (bool, string, bool, string) {
	t.Helper()
	read := func(name string) (bool, string) {
		size, err := syscall.Getxattr(path, name, nil)
		if err == syscall.ENODATA {
			return false, ""
		}
		if err == syscall.ENOTSUP || err == syscall.ENOSYS {
			t.Skip("native POSIX ACL fixture inspection unsupported")
		}
		if err != nil || size > 65536 {
			t.Fatal("native POSIX ACL fixture inspection failed")
		}
		body := make([]byte, size)
		n, err := syscall.Getxattr(path, name, body)
		if err != nil || n != size {
			t.Fatal("native POSIX ACL fixture changed")
		}
		return true, encoded(body)
	}
	a, b := read("system.posix_acl_access")
	c, d := read("system.posix_acl_default")
	return a, b, c, d
}
