package editor_test

import (
	"encoding/base64"
	"os"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

func nativeACL(t *testing.T, path string) string {
	t.Helper()
	value := "absent"
	for _, name := range []string{"system.posix_acl_access", "system.posix_acl_default"} {
		size, err := syscall.Getxattr(path, name, nil)
		if err == syscall.ENODATA {
			continue
		}
		if err != nil || size > 65536 {
			t.Fatalf("native ACL inspection unavailable: %v", err)
		}
		body := make([]byte, size)
		n, err := syscall.Getxattr(path, name, body)
		if err != nil || n != size {
			t.Fatal("native ACL changed")
		}
		value += ":" + name + ":" + base64.StdEncoding.EncodeToString(body)
	}
	return value
}
func nativeFlags(t *testing.T, path string, _ os.FileInfo) uint64 {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	flags, err := unix.IoctlGetInt(int(file.Fd()), unix.FS_IOC_GETFLAGS)
	if err != nil {
		t.Fatalf("native persistent flag inspection unavailable: %v", err)
	}
	return uint64(uint32(flags))
}
