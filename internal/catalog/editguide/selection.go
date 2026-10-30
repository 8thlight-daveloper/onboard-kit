package editguide

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

type selection struct {
	path        string
	name        string
	parent      int
	parentFacts directoryFacts
	original    fileFacts
	raw         []byte
}

type directoryFacts struct {
	device, inode uint64
	mode          uint32
	uid, gid      uint32
}

func directoryFactsOf(stat *unix.Stat_t) directoryFacts {
	return directoryFacts{device: uint64(stat.Dev), inode: stat.Ino, mode: uint32(stat.Mode), uid: stat.Uid, gid: stat.Gid}
}

func anchor(path string) (string, string) {
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part == ".." {
			return "", "path"
		}
	}
	last := parts[len(parts)-1]
	if last == "" || last == "." {
		return "", "path"
	}
	if !filepath.IsAbs(path) {
		current, err := unix.Getwd()
		if err != nil || !filepath.IsAbs(current) {
			return "", "path"
		}
		path = current + "/" + path
	}
	if len(path) > maxPathBytes {
		return "", "usage"
	}
	return filepath.Clean(path), ""
}

func selectPlan(path string) (*selection, string) {
	anchored, code := anchor(path)
	if code != "" {
		return nil, code
	}
	parent, ok := openParent(anchored)
	if !ok {
		return nil, "path"
	}
	chosen := &selection{path: anchored, name: filepath.Base(anchored), parent: parent}
	if code := chosen.load(); code != "" {
		chosen.close()
		return nil, code
	}
	return chosen, ""
}

func openParent(anchored string) (int, bool) {
	current, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, false
	}
	for _, part := range strings.Split(filepath.Dir(anchored), "/") {
		if part == "" {
			continue
		}
		next, err := unix.Openat(current, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		unix.Close(current)
		if err != nil {
			return -1, false
		}
		current = next
	}
	return current, true
}

func (s *selection) load() string {
	var stat unix.Stat_t
	if unix.Fstat(s.parent, &stat) != nil || stat.Mode&unix.S_IFMT != unix.S_IFDIR || stat.Uid != uint32(os.Geteuid()) || stat.Mode&(unix.S_IWGRP|unix.S_IWOTH) != 0 {
		return "path"
	}
	s.parentFacts = directoryFactsOf(&stat)
	fd, err := unix.Openat(s.parent, s.name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return "path"
	}
	facts, raw, code := inspectOriginal(fd)
	if unix.Close(fd) != nil && code == "" {
		code = "path"
	}
	if code != "" {
		return code
	}
	if !s.sameName(facts) {
		return "path"
	}
	s.original, s.raw = facts, raw
	return ""
}

func (s *selection) sameName(facts fileFacts) bool {
	var stat unix.Stat_t
	if unix.Fstatat(s.parent, s.name, &stat, unix.AT_SYMLINK_NOFOLLOW) != nil {
		return false
	}
	return uint64(stat.Dev) == facts.core.Device && stat.Ino == facts.core.Inode
}

func (s *selection) close() error {
	unlockErr := unix.Flock(s.parent, unix.LOCK_UN)
	closeErr := unix.Close(s.parent)
	if closeErr != nil {
		return closeErr
	}
	if unlockErr != nil && unlockErr != unix.EBADF {
		return unlockErr
	}
	return nil
}
