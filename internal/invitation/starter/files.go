package starter

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/sys/unix"
)

var (
	errContext    = errors.New("context")
	errSource     = errors.New("source")
	errLimit      = errors.New("limit")
	errInput      = errors.New("input")
	errInputLimit = errors.New("input-limit")
	errOutput     = errors.New("output")
)

type security struct {
	UID            string `json:"uid"`
	GID            string `json:"gid"`
	Mode           string `json:"mode"`
	ACLFormat      string `json:"aclFormat"`
	AccessPresent  bool   `json:"accessPresent"`
	AccessBase64   string `json:"accessBase64"`
	DefaultPresent bool   `json:"defaultPresent"`
	DefaultBase64  string `json:"defaultBase64"`
}
type objectGuard struct {
	Kind     string   `json:"kind"`
	Device   string   `json:"device"`
	Inode    string   `json:"inode"`
	Security security `json:"security"`
}

func guard(fd int) (objectGuard, error) {
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return objectGuard{}, errContext
	}
	kind := ""
	switch stat.Mode & unix.S_IFMT {
	case unix.S_IFDIR:
		kind = "directory"
	case unix.S_IFREG:
		kind = "regular"
	default:
		return objectGuard{}, errSource
	}
	format, ap, a, dp, d, err := nativeACL(fd)
	if err != nil {
		return objectGuard{}, errContext
	}
	return objectGuard{Kind: kind, Device: strconv.FormatUint(uint64(stat.Dev), 10), Inode: strconv.FormatUint(uint64(stat.Ino), 10), Security: security{UID: strconv.FormatUint(uint64(stat.Uid), 10), GID: strconv.FormatUint(uint64(stat.Gid), 10), Mode: fmt.Sprintf("%04o", stat.Mode&07777), ACLFormat: format, AccessPresent: ap, AccessBase64: base64.StdEncoding.EncodeToString(a), DefaultPresent: dp, DefaultBase64: base64.StdEncoding.EncodeToString(d)}}, nil
}
func sameObject(a, b objectGuard) bool {
	return a.Device == b.Device && a.Inode == b.Inode && a.Kind == b.Kind
}
func private(g objectGuard) bool {
	mode := "0600"
	if g.Kind == "directory" {
		mode = "0700"
	}
	if g.Security.Mode != mode || g.Security.UID != strconv.Itoa(os.Geteuid()) || g.Security.AccessPresent || g.Security.DefaultPresent {
		return false
	}
	groups, err := os.Getgroups()
	if err != nil {
		return false
	}
	group, err := strconv.Atoi(g.Security.GID)
	if err != nil {
		return false
	}
	if group == os.Getegid() {
		return true
	}
	for _, v := range groups {
		if group == v {
			return true
		}
	}
	return false
}
func newPrivate(g objectGuard) bool {
	return private(g) && g.Security.GID == strconv.Itoa(os.Getegid())
}

type directory struct {
	file     *os.File
	path     string
	original objectGuard
}

func (d *directory) fd() int      { return int(d.file.Fd()) }
func (d *directory) close() error { return d.file.Close() }
func dirAt(parent int, name, path string) (*directory, error) {
	fd, err := unix.Openat(parent, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), path)
	g, err := guard(fd)
	if err != nil {
		f.Close()
		return nil, err
	}
	return &directory{file: f, path: path, original: g}, nil
}

type chain struct {
	dirs     []*directory
	missing  []string
	absolute string
}

func openChain(path string) (*chain, error) {
	if !filepath.IsAbs(path) || !safePath(path) {
		return nil, errContext
	}
	root, err := dirAt(unix.AT_FDCWD, "/", "/")
	if err != nil {
		return nil, errContext
	}
	c := &chain{dirs: []*directory{root}, absolute: filepath.Clean(path)}
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if p == "" || p == "." {
			continue
		}
		last := c.dirs[len(c.dirs)-1]
		next, err := dirAt(last.fd(), p, filepath.Join(last.path, p))
		if errors.Is(err, unix.ENOENT) {
			for _, v := range parts[i:] {
				if v != "" && v != "." {
					c.missing = append(c.missing, v)
				}
			}
			return c, nil
		}
		if err != nil {
			c.close()
			return nil, errContext
		}
		c.dirs = append(c.dirs, next)
	}
	return c, nil
}
func (c *chain) last() *directory { return c.dirs[len(c.dirs)-1] }
func (c *chain) close() error {
	var result error
	for i := len(c.dirs) - 1; i >= 0; i-- {
		if err := c.dirs[i].close(); err != nil {
			result = err
		}
	}
	c.dirs = nil
	return result
}
func (c *chain) recheck() error {
	fresh, err := openChain(c.absolute)
	if err != nil {
		return err
	}
	defer fresh.close()
	if len(c.missing) != len(fresh.missing) || len(c.dirs) != len(fresh.dirs) {
		return errContext
	}
	for i, d := range c.dirs {
		if !reflect.DeepEqual(d.original, fresh.dirs[i].original) {
			return errContext
		}
	}
	return nil
}
func safePath(path string) bool {
	if path == "" || len(path) > 4096 || !utf8.ValidString(path) || strings.ContainsRune(path, 0) {
		return false
	}
	for _, p := range strings.Split(path, "/") {
		if p == ".." {
			return false
		}
	}
	return true
}
func relative(path string) bool {
	if !safePath(path) || filepath.IsAbs(path) || filepath.Clean(path) != path || path == "." {
		return false
	}
	return true
}
func (c *chain) bootstrap() error {
	for _, name := range c.missing {
		p := c.last()
		if err := unix.Mkdirat(p.fd(), name, 0700); err != nil {
			return errContext
		}
		d, err := dirAt(p.fd(), name, filepath.Join(p.path, name))
		if err != nil {
			return errContext
		}
		c.dirs = append(c.dirs, d)
		if err := unix.Fchown(d.fd(), os.Geteuid(), os.Getegid()); err != nil {
			return errContext
		}
		g, err := guard(d.fd())
		if err != nil || !newPrivate(g) {
			return errContext
		}
		d.original = g
	}
	c.missing = nil
	return nil
}

type parentGuard struct {
	Ordinal string       `json:"ordinal"`
	Present bool         `json:"present"`
	Guard   *objectGuard `json:"guard"`
}
type snapshot struct {
	present bool
	data    []byte
	guard   *objectGuard
	parents []parentGuard
}

func readAt(parent int, name string, maximum int, requirePrivate bool) (snapshot, error) {
	fd, err := unix.Openat(parent, name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	if errors.Is(err, unix.ENOENT) {
		return snapshot{}, nil
	}
	if err != nil {
		return snapshot{}, errSource
	}
	file := os.NewFile(uintptr(fd), name)
	defer file.Close()
	g, err := guard(fd)
	if err != nil {
		return snapshot{}, err
	}
	if g.Kind != "regular" {
		return snapshot{}, errSource
	}
	if requirePrivate && !private(g) {
		return snapshot{}, errContext
	}
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return snapshot{}, errSource
	}
	if stat.Size > int64(maximum) {
		return snapshot{}, errLimit
	}
	data, err := io.ReadAll(io.LimitReader(file, int64(maximum)+1))
	if err != nil {
		return snapshot{}, errSource
	}
	if len(data) > maximum {
		return snapshot{}, errLimit
	}
	after, err := guard(fd)
	if err != nil || !reflect.DeepEqual(g, after) {
		return snapshot{}, errContext
	}
	var named unix.Stat_t
	if err := unix.Fstatat(parent, name, &named, unix.AT_SYMLINK_NOFOLLOW); err != nil || strconv.FormatUint(uint64(named.Dev), 10) != g.Device || strconv.FormatUint(uint64(named.Ino), 10) != g.Inode {
		return snapshot{}, errContext
	}
	return snapshot{present: true, data: data, guard: &g}, nil
}
func readTarget(base *directory, path string, maximum int, requirePrivate bool) (snapshot, error) {
	if !relative(path) {
		return snapshot{}, errContext
	}
	parts := strings.Split(path, "/")
	current := base
	var opened []*directory
	defer func() {
		for _, d := range opened {
			d.close()
		}
	}()
	parents := make([]parentGuard, 0, len(parts)-1)
	missing := false
	for i, p := range parts[:len(parts)-1] {
		entry := parentGuard{Ordinal: strconv.Itoa(i)}
		if !missing {
			d, err := dirAt(current.fd(), p, filepath.Join(current.path, p))
			if errors.Is(err, unix.ENOENT) {
				missing = true
			} else if err != nil {
				return snapshot{}, errContext
			} else {
				opened = append(opened, d)
				current = d
				entry.Present = true
				g := d.original
				entry.Guard = &g
				if requirePrivate && !private(g) {
					return snapshot{}, errContext
				}
			}
		}
		parents = append(parents, entry)
	}
	if missing {
		return snapshot{parents: parents}, nil
	}
	s, err := readAt(current.fd(), parts[len(parts)-1], maximum, requirePrivate)
	s.parents = parents
	return s, err
}
func snapshotEqual(a, b snapshot) bool {
	return a.present == b.present && bytes.Equal(a.data, b.data) && reflect.DeepEqual(a.guard, b.guard) && reflect.DeepEqual(a.parents, b.parents)
}

type lockSet struct{ dirs []*directory }

func lockDirectories(dirs ...*directory) (*lockSet, error) {
	values := append([]*directory(nil), dirs...)
	sort.Slice(values, func(i, j int) bool {
		a, _ := strconv.ParseUint(values[i].original.Device, 10, 64)
		b, _ := strconv.ParseUint(values[j].original.Device, 10, 64)
		if a != b {
			return a < b
		}
		a, _ = strconv.ParseUint(values[i].original.Inode, 10, 64)
		b, _ = strconv.ParseUint(values[j].original.Inode, 10, 64)
		return a < b
	})
	result := &lockSet{}
	for _, d := range values {
		if len(result.dirs) > 0 && sameObject(d.original, result.dirs[len(result.dirs)-1].original) {
			continue
		}
		if err := unix.Flock(d.fd(), unix.LOCK_EX|unix.LOCK_NB); err != nil {
			result.release()
			return nil, errContext
		}
		result.dirs = append(result.dirs, d)
	}
	return result, nil
}
func (l *lockSet) release() error {
	if l == nil {
		return nil
	}
	var result error
	for _, d := range l.dirs {
		if err := unix.Flock(d.fd(), unix.LOCK_UN); err != nil {
			result = err
		}
	}
	l.dirs = nil
	return result
}
func entries(d *directory, maximum int) ([]os.DirEntry, error) {
	fd, err := unix.Openat(d.fd(), ".", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, errContext
	}
	f := os.NewFile(uintptr(fd), d.path)
	defer f.Close()
	values, err := f.ReadDir(maximum + 1)
	if err != nil && err != io.EOF {
		return nil, errContext
	}
	if len(values) > maximum {
		return nil, errLimit
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Name() < values[j].Name() })
	return values, nil
}
func createFile(parent *directory, name string) (*os.File, objectGuard, error) {
	fd, err := unix.Openat(parent.fd(), name, unix.O_RDWR|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if err != nil {
		return nil, objectGuard{}, errContext
	}
	f := os.NewFile(uintptr(fd), name)
	if err := unix.Fchown(fd, os.Geteuid(), os.Getegid()); err != nil {
		f.Close()
		return nil, objectGuard{}, errContext
	}
	g, err := guard(fd)
	if err != nil || !newPrivate(g) {
		f.Close()
		return nil, objectGuard{}, errContext
	}
	return f, g, nil
}
func createDirectory(parent *directory, name string) (*directory, error) {
	if err := unix.Mkdirat(parent.fd(), name, 0700); err != nil {
		return nil, errContext
	}
	d, err := dirAt(parent.fd(), name, filepath.Join(parent.path, name))
	if err != nil {
		return nil, err
	}
	if err := unix.Fchown(d.fd(), os.Geteuid(), os.Getegid()); err != nil {
		d.close()
		return nil, errContext
	}
	d.original, err = guard(d.fd())
	if err != nil || !newPrivate(d.original) {
		d.close()
		return nil, errContext
	}
	return d, nil
}
func applySecurity(fd int, s security) error {
	if s.AccessPresent || s.DefaultPresent {
		return errContext
	}
	uid, err := strconv.ParseUint(s.UID, 10, 32)
	if err != nil {
		return errContext
	}
	gid, err := strconv.ParseUint(s.GID, 10, 32)
	if err != nil {
		return errContext
	}
	mode, err := strconv.ParseUint(s.Mode, 8, 16)
	if err != nil {
		return errContext
	}
	if err := unix.Fchown(fd, int(uid), int(gid)); err != nil {
		return errContext
	}
	if err := unix.Fchmod(fd, uint32(mode)); err != nil {
		return errContext
	}
	g, err := guard(fd)
	if err != nil || !reflect.DeepEqual(g.Security, s) {
		return errContext
	}
	return nil
}
