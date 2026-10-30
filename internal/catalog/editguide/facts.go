package editguide

import (
	"bytes"
	"os"

	"golang.org/x/sys/unix"
)

const (
	maxRepresentationBytes = 65536
	maxMetadataBytes       = 262144
)

type fileCore struct {
	Device, Inode, Links      uint64
	Mode, UID, GID, Flags     uint32
	Size                      int64
	ModifiedSec, ModifiedNsec int64
	ChangedSec, ChangedNsec   int64
	AccessControl             bool
}

type fileFacts struct {
	core       fileCore
	attributes map[string][]byte
}

func (f fileFacts) same(other fileFacts) bool {
	return f.core == other.core && sameAttributes(f.attributes, other.attributes)
}

func sameAttributes(left, right map[string][]byte) bool {
	if len(left) != len(right) {
		return false
	}
	for name, value := range left {
		other, found := right[name]
		if !found || !bytes.Equal(value, other) {
			return false
		}
	}
	return true
}

func captureFile(fd int) (fileFacts, bool) {
	var stat unix.Stat_t
	if unix.Fstat(fd, &stat) != nil {
		return fileFacts{}, false
	}
	flags, ok := nativeFlags(fd, &stat)
	if !ok {
		return fileFacts{}, false
	}
	controlled, ok := accessControlPresent(fd)
	if !ok {
		return fileFacts{}, false
	}
	attributes, ok := captureAttributes(fd)
	if !ok {
		return fileFacts{}, false
	}
	modifiedSec, modifiedNsec, changedSec, changedNsec := timestamps(&stat)
	core := fileCore{
		Device: uint64(stat.Dev), Inode: stat.Ino, Links: uint64(stat.Nlink),
		Mode: uint32(stat.Mode), UID: stat.Uid, GID: stat.Gid, Flags: flags, Size: stat.Size,
		ModifiedSec: modifiedSec, ModifiedNsec: modifiedNsec, ChangedSec: changedSec, ChangedNsec: changedNsec,
		AccessControl: controlled,
	}
	return fileFacts{core: core, attributes: attributes}, true
}

func captureAttributes(fd int) (map[string][]byte, bool) {
	size, err := unix.Flistxattr(fd, nil)
	if err != nil || size < 0 || size > maxRepresentationBytes {
		return nil, false
	}
	listing := make([]byte, size)
	if used, err := unix.Flistxattr(fd, listing); err != nil || used != size {
		return nil, false
	}
	attributes := map[string][]byte{}
	total := size
	for len(listing) > 0 {
		end := bytes.IndexByte(listing, 0)
		if end <= 0 {
			return nil, false
		}
		name := string(listing[:end])
		listing = listing[end+1:]
		if _, duplicate := attributes[name]; duplicate {
			return nil, false
		}
		valueSize, err := unix.Fgetxattr(fd, name, nil)
		if err != nil || valueSize < 0 || valueSize > maxRepresentationBytes || total > maxMetadataBytes-valueSize {
			return nil, false
		}
		total += valueSize
		value := make([]byte, valueSize)
		if used, err := unix.Fgetxattr(fd, name, value); err != nil || used != valueSize {
			return nil, false
		}
		attributes[name] = value
	}
	return attributes, true
}

func supportedProfile(facts fileFacts) bool {
	core := facts.core
	if core.Mode&unix.S_IFMT != unix.S_IFREG || core.Links != 1 || core.UID != uint32(os.Geteuid()) || core.GID != uint32(os.Getegid()) {
		return false
	}
	if core.Mode&07777 != 0600 || core.Flags != 0 || core.AccessControl {
		return false
	}
	for name := range facts.attributes {
		if !supportedAttribute(name) {
			return false
		}
	}
	return true
}

func readBounded(fd int, limit int) ([]byte, bool) {
	data := make([]byte, 0, 4096)
	buffer := make([]byte, 32768)
	for {
		n, err := unix.Pread(fd, buffer, int64(len(data)))
		if err == unix.EINTR {
			continue
		}
		if err != nil || n < 0 {
			return nil, false
		}
		if n == 0 {
			return data, true
		}
		if len(data)+n > limit {
			return nil, false
		}
		data = append(data, buffer[:n]...)
	}
}

func inspectOriginal(fd int) (fileFacts, []byte, string) {
	before, ok := captureFile(fd)
	if !ok || !supportedProfile(before) {
		return fileFacts{}, nil, "path"
	}
	if before.core.Size > maxSourceBytes {
		return fileFacts{}, nil, "path"
	}
	raw, ok := readBounded(fd, maxSourceBytes)
	if !ok {
		return fileFacts{}, nil, "path"
	}
	after, ok := captureFile(fd)
	if !ok || !before.same(after) || int64(len(raw)) != after.core.Size {
		return fileFacts{}, nil, "path"
	}
	return after, raw, ""
}
