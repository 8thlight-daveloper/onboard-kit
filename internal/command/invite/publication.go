package invite

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

func (path destination) temporary() (*os.File, string) {
	for attempt := range 128 {
		name := ".onboard-kit-invitation-" + strconv.Itoa(attempt) + ".tmp"
		if strings.EqualFold(name, path.leaf) {
			continue
		}
		fd, err := unix.Openat(path.parent, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0o600)
		if err == unix.EEXIST {
			continue
		}
		if err != nil {
			return nil, ""
		}
		return os.NewFile(uintptr(fd), "invitation temporary"), name
	}
	return nil, ""
}

type publication struct {
	committed bool
	code      string
}

func (path destination) publish(archive []byte) (result publication) {
	file, name := path.temporary()
	if file == nil {
		return publication{code: "publish"}
	}
	defer func() {
		if err := unix.Unlinkat(path.parent, name, 0); err != nil && err != unix.ENOENT {
			result.code = "cleanup-refused"
			if result.committed {
				result.code = "cleanup-written"
			}
		}
	}()
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return publication{code: "publish"}
	}
	n, writeError := file.Write(archive)
	closeError := file.Close()
	if writeError != nil || n != len(archive) || closeError != nil {
		return publication{code: "publish"}
	}
	if err := unix.Linkat(path.parent, name, path.parent, path.leaf, 0); err != nil {
		if err == unix.EEXIST {
			return publication{code: "output-exists"}
		}
		return publication{code: "publish"}
	}
	return publication{committed: true}
}
