package store

import (
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/sys/unix"
)

const (
	directoryFlags = unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC | unix.O_NOFOLLOW
	fileFlags      = unix.O_RDONLY | unix.O_NONBLOCK | unix.O_NOFOLLOW | unix.O_CLOEXEC
)

type fileIdentity struct {
	device uint64
	inode  uint64
}

type fileFacts struct {
	identity fileIdentity
	mode     uint32
	owner    uint32
	group    uint32
}

type fileSnapshot struct {
	exists  bool
	facts   fileFacts
	content []byte
}

type openedFile struct {
	descriptor int
	stat       unix.Stat_t
}

func identityOf(stat unix.Stat_t) fileIdentity {
	return fileIdentity{device: uint64(stat.Dev), inode: uint64(stat.Ino)}
}

func factsOf(stat unix.Stat_t) fileFacts {
	return fileFacts{identity: identityOf(stat), mode: uint32(stat.Mode), owner: uint32(stat.Uid), group: uint32(stat.Gid)}
}

func sameSnapshot(left, right fileSnapshot) bool {
	if left.exists != right.exists {
		return false
	}
	return !left.exists || left.facts == right.facts && string(left.content) == string(right.content)
}

func safePath(path string) bool {
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
	if !safePath(path) {
		return -1, unix.EINVAL
	}
	start := "."
	if strings.HasPrefix(path, "/") {
		start = "/"
	}
	descriptor, err := unix.Open(start, directoryFlags, 0)
	if err != nil {
		return -1, err
	}
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." {
			continue
		}
		next, err := unix.Openat(descriptor, part, directoryFlags, 0)
		unix.Close(descriptor)
		if err != nil {
			return -1, err
		}
		descriptor = next
	}
	return descriptor, nil
}

func openChildDirectory(parent int, name string) (int, error) {
	return unix.Openat(parent, name, directoryFlags, 0)
}

func walkDirectories(start int, parts []string) (int, error) {
	descriptor, err := unix.Openat(start, ".", directoryFlags, 0)
	if err != nil {
		return -1, err
	}
	for _, part := range parts {
		next, err := openChildDirectory(descriptor, part)
		unix.Close(descriptor)
		if err != nil {
			return -1, err
		}
		descriptor = next
	}
	return descriptor, nil
}

func openRegular(parent int, name string) (openedFile, error) {
	descriptor, err := unix.Openat(parent, name, fileFlags, 0)
	if err != nil {
		return openedFile{}, err
	}
	var stat unix.Stat_t
	if err := unix.Fstat(descriptor, &stat); err != nil {
		unix.Close(descriptor)
		return openedFile{}, err
	}
	if uint32(stat.Mode)&unix.S_IFMT != unix.S_IFREG {
		unix.Close(descriptor)
		return openedFile{}, unix.EINVAL
	}
	return openedFile{descriptor: descriptor, stat: stat}, nil
}

func readBounded(opened openedFile) (fileSnapshot, error) {
	file := os.NewFile(uintptr(opened.descriptor), "selected-source")
	content, readErr := io.ReadAll(io.LimitReader(file, byteLimit+1))
	closeErr := file.Close()
	if readErr != nil {
		return fileSnapshot{}, readErr
	}
	if closeErr != nil {
		return fileSnapshot{}, closeErr
	}
	if len(content) > byteLimit || !utf8.Valid(content) {
		return fileSnapshot{}, unix.EINVAL
	}
	return fileSnapshot{exists: true, facts: factsOf(opened.stat), content: content}, nil
}

func readPlanFile(path string) (fileSnapshot, error) {
	if !safePath(path) || strings.HasSuffix(path, "/") {
		return fileSnapshot{}, unix.EINVAL
	}
	directory, name := ".", path
	if index := strings.LastIndex(path, "/"); index >= 0 {
		directory, name = path[:index], path[index+1:]
		if directory == "" {
			directory = "/"
		}
	}
	parent, err := openDirectory(directory)
	if err != nil {
		return fileSnapshot{}, err
	}
	opened, err := openRegular(parent, name)
	unix.Close(parent)
	if err != nil {
		return fileSnapshot{}, err
	}
	return readBounded(opened)
}
