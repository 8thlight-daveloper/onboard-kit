package editor

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/sys/unix"
)

func (c *context) registeredRoot(path string) (*os.File, *absence, bool) {
	base, ok := c.absolute("/")
	if !ok {
		return nil, nil, false
	}
	parts := []string{}
	for _, part := range strings.Split(path, "/") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return c.walkDirectories(base, parts)
}

func (c *context) walkDirectories(base *os.File, parts []string) (*os.File, *absence, bool) {
	current := base
	for i, part := range parts {
		var stat unix.Stat_t
		err := unix.Fstatat(int(current.Fd()), part, &stat, unix.AT_SYMLINK_NOFOLLOW)
		if err == unix.ENOENT {
			return nil, &absence{ParentPath: current.Name(), Parent: c.guards[current.Name()], Remaining: append([]string(nil), parts[i:]...)}, true
		}
		if err != nil || stat.Mode&unix.S_IFMT != unix.S_IFDIR {
			return nil, nil, false
		}
		next, ok := c.directory(current, part, filepath.Join(current.Name(), part))
		if !ok {
			return nil, nil, false
		}
		current = next
	}
	return current, nil, true
}

func (c *context) optionalParticipant(base *os.File, relative string) (snapshot, *absence, bool) {
	if !relativePath(relative) {
		return snapshot{}, nil, false
	}
	parts := strings.Split(relative, "/")
	parent, missing, ok := c.walkDirectories(base, parts[:len(parts)-1])
	if !ok {
		return snapshot{}, nil, false
	}
	if missing != nil {
		missing.Remaining = append(missing.Remaining, parts[len(parts)-1])
		return snapshot{}, missing, true
	}
	name := parts[len(parts)-1]
	var stat unix.Stat_t
	err := unix.Fstatat(int(parent.Fd()), name, &stat, unix.AT_SYMLINK_NOFOLLOW)
	if err == unix.ENOENT {
		return snapshot{}, &absence{ParentPath: parent.Name(), Parent: c.guards[parent.Name()], Remaining: []string{name}}, true
	}
	if err != nil || stat.Mode&unix.S_IFMT != unix.S_IFREG {
		return snapshot{}, nil, false
	}
	participant, ok := c.participant(parent, name)
	return participant, nil, ok
}

func sameNative(a, b nativeGuard) bool { return a.Device == b.Device && a.Inode == b.Inode }
func disjointNames(a, b []string, profile int) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] == b[i] {
			continue
		}
		if !comparisonName(a[i]) || !comparisonName(b[i]) {
			return false
		}
		if profile == 1 {
			return true
		}
		if profile != 2 {
			return false
		}
		if !strings.EqualFold(a[i], b[i]) {
			return true
		}
	}
	return false
}

func comparisonName(s string) bool {
	if s == "" || s[0] == '.' || s[len(s)-1] == '.' {
		return false
	}
	for _, c := range []byte(s) {
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-' || c == '.' {
			continue
		}
		return false
	}
	return true
}

func (c *context) comparisonProfile(missing absence) int {
	probe := &context{guards: map[string]nativeGuard{}, groups: c.groups}
	defer probe.close()
	parent, ok := probe.absolute(missing.ParentPath)
	if !ok || !reflect.DeepEqual(probe.guards[parent.Name()], missing.Parent) {
		return 0
	}
	return nativeComparisonProfile(int(parent.Fd()))
}

func (c *context) disjointAbsences(participants []snapshot) bool {
	for i, missing := range c.absences {
		if len(missing.Remaining) == 0 {
			return false
		}
		for j, other := range c.absences[:i] {
			if sameNative(missing.Parent, other.Parent) {
				profile := c.comparisonProfile(missing)
				if profile == 0 || !disjointNames(missing.Remaining, other.Remaining, profile) {
					return false
				}
				c.absences[i].ComparisonProfile = profile
				c.absences[j].ComparisonProfile = profile
			}
		}
		missingPath := filepath.Join(append([]string{missing.ParentPath}, missing.Remaining...)...)
		for _, participant := range participants {
			if sameNative(missing.Parent, c.guards[participant.parent.Name()]) && missing.Remaining[0] == participant.name {
				return false
			}
			if participant.path == missingPath || strings.HasPrefix(participant.path, missingPath+string(filepath.Separator)) || strings.HasPrefix(missingPath, participant.path+string(filepath.Separator)) {
				return false
			}
		}
	}
	return true
}

func (c *context) absencesFresh() bool {
	probe := &context{guards: map[string]nativeGuard{}, groups: c.groups}
	defer probe.close()
	for _, missing := range c.absences {
		if len(missing.Remaining) == 0 {
			return false
		}
		parent, ok := probe.absolute(missing.ParentPath)
		if !ok || !reflect.DeepEqual(probe.guards[parent.Name()], missing.Parent) {
			return false
		}
		if missing.ComparisonProfile != 0 && nativeComparisonProfile(int(parent.Fd())) != missing.ComparisonProfile {
			return false
		}
		var stat unix.Stat_t
		if unix.Fstatat(int(parent.Fd()), missing.Remaining[0], &stat, unix.AT_SYMLINK_NOFOLLOW) != unix.ENOENT {
			return false
		}
	}
	return true
}
