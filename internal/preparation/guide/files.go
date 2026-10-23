package guide

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/sys/unix"
)

type snapshot struct {
	bytes  []byte
	stat   unix.Stat_t
	exists bool
}

type sources struct {
	selected    options
	root        int
	rootStat    unix.Stat_t
	plan        snapshot
	preparation snapshot
	destination string
}

func safeArgument(path string) bool {
	if path == "" || strings.ContainsRune(path, 0) {
		return false
	}
	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return false
		}
	}
	return true
}

func openDirectory(path string) (int, error) {
	if !safeArgument(path) {
		return -1, unix.EINVAL
	}
	start := "."
	if strings.HasPrefix(path, "/") {
		start = "/"
	}
	fd, err := unix.Open(start, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return -1, err
	}
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." {
			continue
		}
		next, nextErr := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		closeErr := unix.Close(fd)
		if nextErr != nil {
			return -1, nextErr
		}
		if closeErr != nil {
			unix.Close(next)
			return -1, closeErr
		}
		fd = next
	}
	return fd, nil
}

func parentAt(root int, path string, create bool) (int, string, error) {
	parts := strings.Split(path, "/")
	fd, err := unix.Openat(root, ".", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return -1, "", err
	}
	for _, part := range parts[:len(parts)-1] {
		if part == "" || part == "." || part == ".." {
			unix.Close(fd)
			return -1, "", unix.EINVAL
		}
		next, nextErr := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if errors.Is(nextErr, unix.ENOENT) && create {
			if !aclAbsent(fd) {
				unix.Close(fd)
				return -1, "", unix.EACCES
			}
			if mkdirErr := unix.Mkdirat(fd, part, 0700); mkdirErr != nil && !errors.Is(mkdirErr, unix.EEXIST) {
				unix.Close(fd)
				return -1, "", mkdirErr
			}
			next, nextErr = unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
			if nextErr == nil {
				var stat unix.Stat_t
				if unix.Fstat(next, &stat) != nil || uint32(stat.Mode)&0077 != 0 || !aclAbsent(next) {
					unix.Close(next)
					unix.Close(fd)
					return -1, "", unix.EACCES
				}
			}
		}
		closeErr := unix.Close(fd)
		if nextErr != nil {
			return -1, "", nextErr
		}
		if closeErr != nil {
			unix.Close(next)
			return -1, "", closeErr
		}
		fd = next
	}
	name := parts[len(parts)-1]
	if name == "" || name == "." || name == ".." {
		unix.Close(fd)
		return -1, "", unix.EINVAL
	}
	return fd, name, nil
}

func regularAt(root int, path string) (int, error) {
	parent, name, err := parentAt(root, path, false)
	if err != nil {
		return -1, err
	}
	fd, openErr := unix.Openat(parent, name, unix.O_RDONLY|unix.O_NONBLOCK|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	closeErr := unix.Close(parent)
	if openErr != nil {
		return -1, openErr
	}
	if closeErr != nil {
		unix.Close(fd)
		return -1, closeErr
	}
	var stat unix.Stat_t
	if err = unix.Fstat(fd, &stat); err != nil {
		unix.Close(fd)
		return -1, err
	}
	if uint32(stat.Mode)&unix.S_IFMT != unix.S_IFREG {
		unix.Close(fd)
		return -1, unix.EINVAL
	}
	return fd, nil
}

func boundedSnapshot(fd int) (snapshot, error) {
	file := os.NewFile(uintptr(fd), "selected-source")
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		file.Close()
		return snapshot{}, err
	}
	if stat.Size < 0 || stat.Size > byteLimit {
		file.Close()
		return snapshot{}, unix.EFBIG
	}
	data, err := io.ReadAll(io.LimitReader(file, byteLimit+1))
	closeErr := file.Close()
	if err != nil {
		return snapshot{}, err
	}
	if closeErr != nil {
		return snapshot{}, closeErr
	}
	if len(data) > byteLimit || !utf8.Valid(data) {
		return snapshot{}, unix.EINVAL
	}
	return snapshot{bytes: data, stat: stat, exists: true}, nil
}

func selectedPlan(path string) (snapshot, error) {
	if !safeArgument(path) || strings.HasSuffix(path, "/") {
		return snapshot{}, unix.EINVAL
	}
	index := strings.LastIndex(path, "/")
	directory, name := ".", path
	if index >= 0 {
		directory, name = path[:index], path[index+1:]
		if directory == "" {
			directory = "/"
		}
	}
	parent, err := openDirectory(directory)
	if err != nil {
		return snapshot{}, err
	}
	fd, err := regularAt(parent, name)
	closeErr := unix.Close(parent)
	if err != nil {
		return snapshot{}, err
	}
	if closeErr != nil {
		unix.Close(fd)
		return snapshot{}, closeErr
	}
	return boundedSnapshot(fd)
}

func loadSources(selected options) (*sources, string) {
	root, err := openDirectory(selected.root)
	if err != nil {
		return nil, "guide.path"
	}
	result := &sources{selected: selected, root: root}
	if unix.Fstat(root, &result.rootStat) != nil {
		result.close()
		return nil, "guide.path"
	}
	result.plan, err = selectedPlan(selected.plan)
	if err != nil {
		result.close()
		return nil, "guide.source"
	}
	return result, ""
}

func (s *sources) close() { unix.Close(s.root) }

func (s *sources) selectPreparation(path string) string {
	s.destination = path
	fd, err := regularAt(s.root, path)
	if errors.Is(err, unix.ENOENT) {
		return ""
	}
	if err != nil {
		return "guide.path"
	}
	var stat unix.Stat_t
	if unix.Fstat(fd, &stat) != nil {
		unix.Close(fd)
		return "guide.path"
	}
	if sameInode(stat, s.plan.stat) {
		unix.Close(fd)
		return "guide.path"
	}
	s.preparation, err = boundedSnapshot(fd)
	if err != nil {
		return "guide.source"
	}
	return ""
}

func sameInode(left, right unix.Stat_t) bool { return left.Dev == right.Dev && left.Ino == right.Ino }

func sameSnapshot(left, right snapshot) bool {
	return left.exists == right.exists && (!left.exists || sameInode(left.stat, right.stat) && left.stat.Mode == right.stat.Mode && left.stat.Uid == right.stat.Uid && left.stat.Gid == right.stat.Gid && bytes.Equal(left.bytes, right.bytes))
}

func (s *sources) unchanged() bool {
	root, err := openDirectory(s.selected.root)
	if err != nil {
		return false
	}
	var rootStat unix.Stat_t
	err = unix.Fstat(root, &rootStat)
	closeErr := unix.Close(root)
	if err != nil || closeErr != nil || !sameInode(rootStat, s.rootStat) {
		return false
	}
	plan, err := selectedPlan(s.selected.plan)
	if err != nil || !sameSnapshot(plan, s.plan) {
		return false
	}
	fd, err := regularAt(s.root, s.destination)
	if errors.Is(err, unix.ENOENT) {
		return !s.preparation.exists
	}
	if err != nil {
		return false
	}
	preparation, err := boundedSnapshot(fd)
	return err == nil && sameSnapshot(preparation, s.preparation) && !sameInode(preparation.stat, plan.stat)
}

func (s *sources) existingPermissions() bool {
	if !s.preparation.exists {
		return true
	}
	fd, err := regularAt(s.root, s.destination)
	if err != nil {
		return false
	}
	var stat unix.Stat_t
	safe := unix.Fstat(fd, &stat) == nil && sameInode(stat, s.preparation.stat) && aclAbsent(fd)
	return unix.Close(fd) == nil && safe
}

func (s *sources) publish(data []byte) string {
	if !s.existingPermissions() {
		return "guide.commit"
	}
	parent, name, err := parentAt(s.root, s.destination, true)
	if err != nil {
		return "guide.commit"
	}
	defer unix.Close(parent)
	if !aclAbsent(parent) {
		return "guide.commit"
	}
	var temporary string
	fd := -1
	for attempt := 1; attempt <= 256; attempt++ {
		temporary = ".preparation-guide-" + strconv.Itoa(attempt) + ".tmp"
		fd, err = unix.Openat(parent, temporary, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
		if err == nil {
			break
		}
		if !errors.Is(err, unix.EEXIST) {
			return "guide.commit"
		}
	}
	if fd < 0 {
		return "guide.commit"
	}
	committed := false
	finish := func(code string) string {
		cleanupErr := unix.Unlinkat(parent, temporary, 0)
		if cleanupErr != nil && !errors.Is(cleanupErr, unix.ENOENT) {
			if committed {
				return "guide.cleanup-saved"
			}
			return "guide.cleanup-refused"
		}
		return code
	}
	file := os.NewFile(uintptr(fd), "owned-publication")
	var stat unix.Stat_t
	if unix.Fstat(fd, &stat) != nil || uint32(stat.Mode)&unix.S_IFMT != unix.S_IFREG || uint32(stat.Mode)&0077 != 0 || !aclAbsent(fd) {
		file.Close()
		return finish("guide.commit")
	}
	mode := uint32(0600)
	if s.preparation.exists {
		if stat.Uid != s.preparation.stat.Uid || stat.Gid != s.preparation.stat.Gid {
			file.Close()
			return finish("guide.commit")
		}
		mode &= uint32(s.preparation.stat.Mode)
	}
	if unix.Fchmod(fd, mode) != nil {
		file.Close()
		return finish("guide.commit")
	}
	n, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil || n != len(data) || closeErr != nil {
		return finish("guide.commit")
	}
	if !s.unchanged() {
		return finish("guide.stale-review")
	}
	if s.preparation.exists {
		err = unix.Renameat(parent, temporary, parent, name)
	} else {
		err = unix.Linkat(parent, temporary, parent, name, 0)
	}
	if err != nil {
		if errors.Is(err, unix.EEXIST) {
			return finish("guide.stale-review")
		}
		return finish("guide.commit")
	}
	committed = true
	return finish("guide.saved")
}
