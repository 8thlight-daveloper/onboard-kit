package editguide

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

const maxTemporaryNames = 256

func (s *selection) publish(data []byte, unchanged bool) string {
	if err := unix.Flock(s.parent, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		if err == unix.EWOULDBLOCK {
			return "busy"
		}
		return "commit"
	}
	if !s.current() {
		return "stale"
	}
	if unchanged {
		return "unchanged"
	}
	return s.replace(data)
}

func (s *selection) current() bool {
	var stat unix.Stat_t
	if unix.Fstat(s.parent, &stat) != nil || directoryFactsOf(&stat) != s.parentFacts {
		return false
	}
	fd, err := unix.Openat(s.parent, s.name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return false
	}
	facts, raw, code := inspectOriginal(fd)
	closed := unix.Close(fd) == nil
	return code == "" && closed && facts.same(s.original) && bytes.Equal(raw, s.raw) && s.sameName(facts)
}

type temporary struct {
	parent int
	name   string
	fd     int
	device uint64
	inode  uint64
}

func (s *selection) createTemporary() (temporary, bool) {
	for i := 0; i < maxTemporaryNames; i++ {
		name := fmt.Sprintf(".onboard-kit-edit-%d", i)
		fd, err := unix.Openat(s.parent, name, unix.O_RDWR|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
		if err == unix.EEXIST {
			continue
		}
		if err != nil {
			return temporary{}, false
		}
		var stat unix.Stat_t
		if unix.Fstat(fd, &stat) != nil {
			unix.Close(fd)
			unix.Unlinkat(s.parent, name, 0)
			return temporary{}, false
		}
		return temporary{parent: s.parent, name: name, fd: fd, device: uint64(stat.Dev), inode: stat.Ino}, true
	}
	return temporary{}, false
}

func (t *temporary) remove() bool {
	var stat unix.Stat_t
	err := unix.Fstatat(t.parent, t.name, &stat, unix.AT_SYMLINK_NOFOLLOW)
	if err == unix.ENOENT {
		return true
	}
	if err != nil || uint64(stat.Dev) != t.device || stat.Ino != t.inode {
		return false
	}
	err = unix.Unlinkat(t.parent, t.name, 0)
	return err == nil || err == unix.ENOENT
}

func (t *temporary) closeDescriptor() bool {
	if t.fd < 0 {
		return true
	}
	err := unix.Close(t.fd)
	t.fd = -1
	return err == nil
}

func (s *selection) replace(data []byte) string {
	t, ok := s.createTemporary()
	if !ok {
		return "commit"
	}
	abandon := func(code string) string {
		t.closeDescriptor()
		if !t.remove() {
			return "cleanup-refused"
		}
		return code
	}
	if !s.establishControls(t.fd) || !s.holds(t.fd, 0, nil) {
		return abandon("commit")
	}
	if !writeFully(t.fd, data) || !s.holds(t.fd, len(data), data) {
		return abandon("commit")
	}
	if !t.closeDescriptor() || !s.holdsNamed(t, data) {
		return abandon("commit")
	}
	if !s.current() {
		return abandon("stale")
	}
	if unix.Renameat(s.parent, t.name, s.parent, s.name) != nil {
		return abandon("commit")
	}
	return "written"
}

func (s *selection) establishControls(fd int) bool {
	var stat unix.Stat_t
	if unix.Fstat(fd, &stat) != nil {
		return false
	}
	if stat.Gid != uint32(os.Getegid()) && unix.Fchown(fd, -1, os.Getegid()) != nil {
		return false
	}
	if unix.Fchmod(fd, 0600) != nil {
		return false
	}
	for name, value := range s.original.attributes {
		if unix.Fsetxattr(fd, name, value, 0) != nil {
			return false
		}
	}
	return true
}

func (s *selection) holds(fd int, size int, data []byte) bool {
	facts, ok := captureFile(fd)
	if !ok || !supportedProfile(facts) || facts.core.Size != int64(size) || !sameAttributes(facts.attributes, s.original.attributes) {
		return false
	}
	if data == nil {
		return true
	}
	body, ok := readBounded(fd, len(data))
	return ok && bytes.Equal(body, data)
}

func (s *selection) holdsNamed(t temporary, data []byte) bool {
	fd, err := unix.Openat(t.parent, t.name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return false
	}
	var stat unix.Stat_t
	same := unix.Fstat(fd, &stat) == nil && uint64(stat.Dev) == t.device && stat.Ino == t.inode
	held := same && s.holds(fd, len(data), data)
	return unix.Close(fd) == nil && held
}

func writeFully(fd int, data []byte) bool {
	for len(data) > 0 {
		n, err := unix.Write(fd, data)
		if err == unix.EINTR {
			continue
		}
		if err != nil || n <= 0 {
			return false
		}
		data = data[n:]
	}
	return true
}
