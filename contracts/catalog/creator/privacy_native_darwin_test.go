package creator_test

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"unsafe"

	"golang.org/x/sys/unix"
)

type creatorGapAttributes struct {
	BitmapCount uint16
	Reserved    uint16
	Common      uint32
	Volume      uint32
	Directory   uint32
	File        uint32
	Fork        uint32
}

type creatorGapGuard struct {
	Device     uint64
	Inode      uint64
	UID        uint32
	GID        uint32
	Mode       uint32
	ACLPresent bool
	ACL        string
}

func creatorGapACL(t *testing.T, path string) (bool, string) {
	t.Helper()
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" {
		t.Skip("creator privacy gap: native filesec fixture requires little-endian Darwin arm64/amd64")
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if file.Close() != nil {
			t.Error("owned ACL inspection descriptor close failed")
		}
	}()
	attributes := creatorGapAttributes{BitmapCount: 5, Common: 0x00400000}
	buffer := make([]byte, 65548)
	_, _, failure := syscall.Syscall6(syscall.SYS_FGETATTRLIST, file.Fd(), uintptr(unsafe.Pointer(&attributes)), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), 0, 0)
	runtime.KeepAlive(file)
	runtime.KeepAlive(attributes)
	runtime.KeepAlive(buffer)
	if failure == syscall.ENOTSUP || failure == syscall.ENOSYS {
		t.Skipf("creator privacy gap: native filesec ACL inspection unsupported, errno %d", failure)
	}
	if failure != 0 {
		t.Fatalf("owned ACL inspection failed: errno %d", failure)
	}
	used := uint64(binary.LittleEndian.Uint32(buffer[:4]))
	if used < 12 || used > uint64(len(buffer)) {
		t.Fatal("owned ACL response length invalid")
	}
	size := uint64(binary.LittleEndian.Uint32(buffer[8:12]))
	if size == 0 {
		return false, ""
	}
	start := int64(4) + int64(int32(binary.LittleEndian.Uint32(buffer[4:8])))
	if start < 12 || uint64(start) > used || size > used-uint64(start) || size < 44 || size > 65536 {
		t.Fatal("owned ACL response exceeds its independent bounds")
	}
	payload := buffer[start : uint64(start)+size]
	if binary.LittleEndian.Uint32(payload[:4]) != 0x012cc16d {
		t.Fatal("owned filesec ACL magic invalid")
	}
	if binary.LittleEndian.Uint32(payload[36:40]) == ^uint32(0) {
		return false, ""
	}
	return true, string(payload)
}

func creatorGapObject(t *testing.T, path string) creatorGapGuard {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() && !info.Mode().IsRegular() {
		t.Fatal("owned native fixture must be a real regular file or directory")
	}
	present, acl := creatorGapACL(t, path)
	return creatorGapGuard{Device: uint64(stat.Dev), Inode: uint64(stat.Ino), UID: stat.Uid, GID: stat.Gid, Mode: uint32(stat.Mode) & 07777, ACLPresent: present, ACL: acl}
}

func creatorGapReadGrant(t *testing.T, acl string, inherited bool) bool {
	t.Helper()
	body := []byte(acl)
	if len(body) < 44 {
		t.Fatal("owned ACL grant has no filesec header")
	}
	count := binary.LittleEndian.Uint32(body[36:40])
	if count == ^uint32(0) {
		return false
	}
	if uint64(count) > uint64(len(body)-44)/24 {
		t.Fatal("owned ACL ACE count exceeds observed payload")
	}
	for i := uint32(0); i < count; i++ {
		start := uint64(44) + uint64(i)*24
		flags := binary.LittleEndian.Uint32(body[start+16 : start+20])
		rights := binary.LittleEndian.Uint32(body[start+20 : start+24])
		if flags&0xf == 1 && rights&(1<<1) != 0 && (inherited && flags&(1<<4) != 0 || !inherited && flags&(1<<5) != 0) {
			return true
		}
	}
	return false
}

type creatorGapWatch struct{ descriptor int }

func creatorGapNamespaceWatch(t *testing.T, root string) creatorGapWatch {
	t.Helper()
	directory, err := os.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if directory.Close() != nil {
			t.Error("owned namespace descriptor close failed")
		}
	})
	descriptor, err := unix.Kqueue()
	if err == unix.ENOSYS || err == unix.ENOTSUP {
		t.Skip("creator privacy gap: native vnode notification unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if unix.Close(descriptor) != nil {
			t.Error("owned namespace queue close failed")
		}
	})
	change := unix.Kevent_t{Ident: uint64(directory.Fd()), Filter: unix.EVFILT_VNODE, Flags: unix.EV_ADD | unix.EV_CLEAR, Fflags: unix.NOTE_WRITE}
	if _, err = unix.Kevent(descriptor, []unix.Kevent_t{change}, nil, nil); err != nil {
		t.Fatal(err)
	}
	watch := creatorGapWatch{descriptor: descriptor}
	control := filepath.Join(root, "owned-vnode-capability-control")
	file, err := os.OpenFile(control, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(control); err != nil {
		t.Fatal(err)
	}
	if !watch.changed(t) {
		t.Skip("creator privacy gap: vnode queue did not observe owned create/remove control")
	}
	if watch.changed(t) {
		t.Fatal("owned namespace queue did not drain before product invocation")
	}
	return watch
}

func (watch creatorGapWatch) changed(t *testing.T) bool {
	t.Helper()
	events := make([]unix.Kevent_t, 8)
	n, err := unix.Kevent(watch.descriptor, nil, events, &unix.Timespec{})
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for _, event := range events[:n] {
		if event.Flags&unix.EV_ERROR != 0 {
			t.Fatal("owned vnode queue reported an error")
		}
		if event.Filter == unix.EVFILT_VNODE && event.Fflags&unix.NOTE_WRITE != 0 {
			changed = true
		}
	}
	return changed
}
