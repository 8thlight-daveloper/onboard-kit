package invite

import (
	"io"
	"os"

	"golang.org/x/sys/unix"
)

const sourceLimit = 1048576

func readSource(path string) ([]byte, bool) {
	source, ok := openParent(path)
	if !ok {
		return nil, false
	}
	defer source.close()
	var info unix.Stat_t
	if err := unix.Fstatat(source.parent, source.leaf, &info, unix.AT_SYMLINK_NOFOLLOW); err != nil || info.Mode&unix.S_IFMT != unix.S_IFREG {
		return nil, false
	}
	fd, err := unix.Openat(source.parent, source.leaf, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, false
	}
	file := os.NewFile(uintptr(fd), "invitation source")
	if err := unix.Fstat(fd, &info); err != nil || info.Mode&unix.S_IFMT != unix.S_IFREG {
		file.Close()
		return nil, false
	}
	data, readError := io.ReadAll(io.LimitReader(file, sourceLimit+1))
	closeError := file.Close()
	if readError != nil || closeError != nil || len(data) > sourceLimit {
		return nil, false
	}
	return data, true
}
