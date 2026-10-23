package invite

import (
	"strings"

	"golang.org/x/sys/unix"
)

type destination struct {
	parent int
	leaf   string
}

func openParent(path string) (destination, bool) {
	if path == "" || strings.IndexByte(path, 0) >= 0 || strings.HasSuffix(path, "/") {
		return destination{}, false
	}
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part == ".." {
			return destination{}, false
		}
	}
	root := "."
	if strings.HasPrefix(path, "/") {
		root = "/"
	}
	parent, err := unix.Open(root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return destination{}, false
	}
	for _, part := range parts[:len(parts)-1] {
		if part == "" || part == "." {
			continue
		}
		next, err := unix.Openat(parent, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		unix.Close(parent)
		if err != nil {
			return destination{}, false
		}
		parent = next
	}
	return destination{parent: parent, leaf: parts[len(parts)-1]}, true
}

func (path destination) close() {
	unix.Close(path.parent)
}

func openDestination(path string) (destination, string) {
	destination, ok := openParent(path)
	if !ok {
		return destination, "output-path"
	}
	var info unix.Stat_t
	err := unix.Fstatat(destination.parent, destination.leaf, &info, unix.AT_SYMLINK_NOFOLLOW)
	if err == nil {
		destination.close()
		return destination, "output-exists"
	}
	if err != unix.ENOENT {
		destination.close()
		return destination, "output-path"
	}
	return destination, ""
}
