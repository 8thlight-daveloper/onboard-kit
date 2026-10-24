package editor

import (
	"bytes"
	"os"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/sys/unix"
)

type nativeAttributes struct {
	AccessPresent, DefaultPresent bool
	Access, Default               []byte
	Flags                         uint32
	Xattrs                        map[string][]byte
}

func readNativeAttributes(fd int) (nativeAttributes, bool) {
	result, ok := platformAttributes(fd)
	if !ok {
		return nativeAttributes{}, false
	}
	names, ok := attributeNames(fd)
	if !ok {
		return nativeAttributes{}, false
	}
	result.Xattrs = map[string][]byte{}
	total := 0
	for _, name := range names {
		size, err := unix.Fgetxattr(fd, name, nil)
		if err != nil || size < 0 || size > 65536 {
			return nativeAttributes{}, false
		}
		total += size
		if total > 1048576 {
			return nativeAttributes{}, false
		}
		data := make([]byte, size)
		n, err := unix.Fgetxattr(fd, name, data)
		if err != nil || n != size {
			return nativeAttributes{}, false
		}
		result.Xattrs[name] = data
	}
	next, ok := attributeNames(fd)
	if !ok || !reflect.DeepEqual(names, next) {
		return nativeAttributes{}, false
	}
	for _, name := range names {
		data := make([]byte, len(result.Xattrs[name]))
		n, err := unix.Fgetxattr(fd, name, data)
		if err != nil || n != len(data) || !bytes.Equal(data, result.Xattrs[name]) {
			return nativeAttributes{}, false
		}
	}
	final, ok := platformAttributes(fd)
	if !ok || result.AccessPresent != final.AccessPresent || result.DefaultPresent != final.DefaultPresent || !bytes.Equal(result.Access, final.Access) || !bytes.Equal(result.Default, final.Default) || result.Flags != final.Flags {
		return nativeAttributes{}, false
	}
	return result, true
}
func attributeNames(fd int) ([]string, bool) {
	size, err := unix.Flistxattr(fd, nil)
	if err != nil || size < 0 || size > 65536 {
		return nil, false
	}
	if size == 0 {
		return []string{}, true
	}
	data := make([]byte, size)
	n, err := unix.Flistxattr(fd, data)
	if err != nil || n != size || data[len(data)-1] != 0 {
		return nil, false
	}
	names := strings.Split(string(data[:len(data)-1]), "\x00")
	if len(names) > 256 {
		return nil, false
	}
	sort.Strings(names)
	for i, name := range names {
		if name == "" || i > 0 && names[i-1] == name {
			return nil, false
		}
	}
	return names, true
}
func copyNativeAttributes(fd int, original nativeGuard) bool {
	if original.Attributes.AccessPresent || original.Attributes.DefaultPresent || !replaceableFlags(original.Attributes.Flags) {
		return false
	}
	attrs, ok := readNativeAttributes(fd)
	if !ok || attrs.AccessPresent || attrs.DefaultPresent {
		return false
	}
	for name := range attrs.Xattrs {
		if _, exists := original.Attributes.Xattrs[name]; !exists {
			if unix.Fremovexattr(fd, name) != nil {
				return false
			}
		}
	}
	for name, data := range original.Attributes.Xattrs {
		if held, exists := attrs.Xattrs[name]; exists && bytes.Equal(held, data) {
			continue
		}
		if unix.Fsetxattr(fd, name, data, 0) != nil {
			return false
		}
	}
	var stat unix.Stat_t
	if unix.Fstat(fd, &stat) != nil {
		return false
	}
	if stat.Uid != original.UID || stat.Gid != original.GID {
		if unix.Fchown(fd, int(original.UID), int(original.GID)) != nil {
			return false
		}
	}
	if unix.Fchmod(fd, original.Mode) != nil || !setNativeFlags(fd, original.Attributes.Flags) {
		return false
	}
	return true
}
func initialOwned(g nativeGuard, mode uint32) bool {
	return g.UID == uint32(os.Geteuid()) && g.GID == uint32(os.Getegid()) && g.Mode == mode && !g.Attributes.AccessPresent && !g.Attributes.DefaultPresent
}
