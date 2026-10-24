package creator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

type directoryGuard struct {
	fd   int
	name string
	stat unix.Stat_t
}

type destination struct {
	path  string
	name  string
	chain []directoryGuard
}

func admitDestination(path string) (*destination, error) {
	if strings.HasSuffix(path, "/") {
		return nil, errors.New("path")
	}
	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return nil, errors.New("path")
		}
	}
	if !filepath.IsAbs(path) {
		current, err := unix.Getwd()
		if err != nil || !filepath.IsAbs(current) {
			return nil, errors.New("path")
		}
		path = current + "/" + path
	}
	if len(path) > 4096 {
		return nil, errors.New("path")
	}
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]
	if name == "" || name == "." || name == ".." {
		return nil, errors.New("path")
	}
	d := &destination{path: filepath.Clean(path), name: name}
	root, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	d.chain = append(d.chain, directoryGuard{fd: root})
	if err = unix.Fstat(root, &d.chain[0].stat); err != nil {
		d.close()
		return nil, err
	}
	for _, part := range parts[:len(parts)-1] {
		if part == "" || part == "." {
			continue
		}
		fd, openErr := unix.Openat(d.parent(), part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if openErr != nil {
			d.close()
			return nil, openErr
		}
		guard := directoryGuard{fd: fd, name: part}
		if err = unix.Fstat(fd, &guard.stat); err != nil {
			unix.Close(fd)
			d.close()
			return nil, err
		}
		d.chain = append(d.chain, guard)
	}
	if !d.current() {
		d.close()
		return nil, errors.New("path")
	}
	return d, nil
}

func (d *destination) parent() int { return d.chain[len(d.chain)-1].fd }

func sameObject(a, b unix.Stat_t) bool {
	return a.Dev == b.Dev && a.Ino == b.Ino && a.Mode&unix.S_IFMT == b.Mode&unix.S_IFMT
}

func (d *destination) current() bool {
	var stat unix.Stat_t
	if unix.Lstat("/", &stat) != nil || !sameObject(stat, d.chain[0].stat) {
		return false
	}
	for i := 1; i < len(d.chain); i++ {
		if unix.Fstatat(d.chain[i-1].fd, d.chain[i].name, &stat, unix.AT_SYMLINK_NOFOLLOW) != nil || !sameObject(stat, d.chain[i].stat) {
			return false
		}
	}
	err := unix.Fstatat(d.parent(), d.name, &stat, unix.AT_SYMLINK_NOFOLLOW)
	return errors.Is(err, unix.ENOENT)
}

func (d *destination) close() error {
	var result error
	for i := len(d.chain) - 1; i >= 0; i-- {
		if err := unix.Close(d.chain[i].fd); err != nil {
			result = err
		}
	}
	d.chain = nil
	return result
}

func (d *destination) publish(bytes []byte) string {
	if !d.current() {
		return "stale"
	}
	fd := -1
	temporary := ""
	for i := 0; i < 256; i++ {
		name := fmt.Sprintf(".onboard-kit-create-%d", i)
		opened, err := unix.Openat(d.parent(), name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
		if errors.Is(err, unix.EEXIST) {
			continue
		}
		if err != nil {
			return "commit"
		}
		fd, temporary = opened, name
		break
	}
	if fd < 0 {
		return "commit"
	}
	cleanup := func(code string) string {
		if fd >= 0 {
			if unix.Close(fd) != nil && code != "cleanup-written" {
				code = "commit"
			}
			fd = -1
		}
		if err := unix.Unlinkat(d.parent(), temporary, 0); err != nil && !errors.Is(err, unix.ENOENT) {
			if code == "written" || code == "cleanup-written" {
				return "cleanup-written"
			}
			return "cleanup-refused"
		}
		return code
	}
	var stat unix.Stat_t
	if unix.Fstat(fd, &stat) != nil || stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Uid != uint32(os.Geteuid()) || stat.Gid != uint32(os.Getegid()) || stat.Mode&07777 != 0600 || !privateACL(fd) {
		return cleanup("commit")
	}
	n, err := unix.Write(fd, bytes)
	if err != nil || n != len(bytes) {
		return cleanup("commit")
	}
	err = unix.Close(fd)
	fd = -1
	if err != nil {
		return cleanup("commit")
	}
	if !d.current() {
		return cleanup("stale")
	}
	if err = publishExclusive(d.parent(), temporary, d.name); err != nil {
		if errors.Is(err, unix.EEXIST) {
			return cleanup("stale")
		}
		return cleanup("commit")
	}
	return cleanup("written")
}
