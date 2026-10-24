package editguide

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/sys/unix"
)

type controls struct {
	Device     uint64
	Inode      uint64
	Mode       uint32
	Links      uint64
	UID        uint32
	GID        uint32
	Size       int64
	Modified   unix.Timespec
	Changed    unix.Timespec
	Flags      uint32
	Attributes map[string]string
}

type directory struct {
	file  *os.File
	name  string
	facts controls
}
type selected struct {
	path  string
	name  string
	chain []directory
	file  *os.File
	facts controls
	bytes []byte
}

func failNative() error { return errors.New("unsupported or changed selected native object") }

func openSelected(path string) (*selected, error) {
	if strings.HasSuffix(path, string(os.PathSeparator)) {
		return nil, failNative()
	}
	for _, part := range strings.Split(path, string(os.PathSeparator)) {
		if part == ".." {
			return nil, failNative()
		}
	}
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		path = cwd + string(os.PathSeparator) + path
	}
	if len(path) > 4096 || !utf8.ValidString(path) || strings.IndexByte(path, 0) >= 0 {
		return nil, failNative()
	}
	parts := strings.Split(path, string(os.PathSeparator))
	if parts[len(parts)-1] == "." || parts[len(parts)-1] == "" {
		return nil, failNative()
	}
	result := &selected{path: filepath.Clean(path), name: parts[len(parts)-1]}
	root, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	result.chain = append(result.chain, directory{file: os.NewFile(uintptr(root), "/")})
	success := false
	defer func() {
		if !success {
			result.close()
		}
	}()
	for _, part := range parts[1 : len(parts)-1] {
		if part == "" || part == "." {
			continue
		}
		parent := result.chain[len(result.chain)-1].file
		fd, err := unix.Openat(int(parent.Fd()), part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if err != nil {
			return nil, err
		}
		result.chain = append(result.chain, directory{file: os.NewFile(uintptr(fd), part), name: part})
	}
	for i := range result.chain {
		facts, err := inspect(result.chain[i].file)
		if err != nil || facts.Mode&unix.S_IFMT != unix.S_IFDIR {
			return nil, failNative()
		}
		if i == len(result.chain)-1 {
			if facts.UID != uint32(os.Geteuid()) || facts.Mode&0022 != 0 || facts.Mode&07000 != 0 {
				return nil, failNative()
			}
		} else {
			if facts.UID != 0 && facts.UID != uint32(os.Geteuid()) {
				return nil, failNative()
			}
			if facts.Mode&0022 != 0 && !(facts.UID == 0 && facts.Mode&unix.S_ISVTX != 0) {
				return nil, failNative()
			}
		}
		result.chain[i].facts = directoryFacts(facts)
	}
	if !metadataBound(result.chain, controls{}) {
		return nil, failNative()
	}
	parent := result.chain[len(result.chain)-1].file
	fd, err := unix.Openat(int(parent.Fd()), result.name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	result.file = os.NewFile(uintptr(fd), result.name)
	result.facts, err = inspect(result.file)
	if err != nil || !privateFile(result.facts) || !metadataBound(result.chain, result.facts) {
		return nil, failNative()
	}
	result.bytes, err = boundedRead(result.file)
	if err != nil {
		return nil, err
	}
	if !result.fresh() {
		return nil, failNative()
	}
	success = true
	return result, nil
}

func directoryFacts(facts controls) controls {
	facts.Size = 0
	facts.Modified = unix.Timespec{}
	facts.Changed = unix.Timespec{}
	facts.Links = 0
	return facts
}

func privateFile(facts controls) bool {
	return facts.Mode == unix.S_IFREG|0600 && facts.Links == 1 && facts.UID == uint32(os.Geteuid()) && facts.GID == uint32(os.Getegid()) && facts.Flags == 0 && supportedAttributes(facts.Attributes) && facts.Size >= 0 && facts.Size <= 1048576
}

func metadataBound(chain []directory, file controls) bool {
	total := 0
	add := func(facts controls) bool {
		for name, value := range facts.Attributes {
			if len(name)+1 > 262144-total {
				return false
			}
			total += len(name) + 1
			if len(value) > 262144-total {
				return false
			}
			total += len(value)
		}
		return true
	}
	for _, directory := range chain {
		if !add(directory.facts) {
			return false
		}
	}
	return add(file)
}

func boundedRead(file *os.File) ([]byte, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(file, 1048577))
	if err != nil || len(data) > 1048576 {
		return nil, failNative()
	}
	return data, nil
}

func (held *selected) close() {
	if held.file != nil {
		held.file.Close()
		held.file = nil
	}
	for i := len(held.chain) - 1; i >= 0; i-- {
		if held.chain[i].file != nil {
			held.chain[i].file.Close()
			held.chain[i].file = nil
		}
	}
}

func (held *selected) fresh() bool {
	for i, directory := range held.chain {
		facts, err := inspect(directory.file)
		if err != nil || !reflect.DeepEqual(directoryFacts(facts), directory.facts) {
			return false
		}
		var named unix.Stat_t
		if i == 0 {
			err = unix.Stat("/", &named)
		} else {
			err = unix.Fstatat(int(held.chain[i-1].file.Fd()), directory.name, &named, unix.AT_SYMLINK_NOFOLLOW)
		}
		if err != nil || uint64(named.Dev) != facts.Device || uint64(named.Ino) != facts.Inode || uint32(named.Mode)&unix.S_IFMT != unix.S_IFDIR {
			return false
		}
	}
	parent := held.chain[len(held.chain)-1].file
	fd, err := unix.Openat(int(parent.Fd()), held.name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return false
	}
	current := os.NewFile(uintptr(fd), held.name)
	defer current.Close()
	facts, err := inspect(current)
	if err != nil || !reflect.DeepEqual(facts, held.facts) {
		return false
	}
	data, err := boundedRead(current)
	if err != nil || !bytes.Equal(data, held.bytes) {
		return false
	}
	after, err := inspect(current)
	if err != nil || !reflect.DeepEqual(after, held.facts) {
		return false
	}
	heldFacts, err := inspect(held.file)
	return err == nil && reflect.DeepEqual(heldFacts, held.facts)
}

func (held *selected) commit(data []byte, unchanged bool) string {
	parent := held.chain[len(held.chain)-1].file
	err := unix.Flock(int(parent.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if err == unix.EWOULDBLOCK || err == unix.EAGAIN {
		return "busy"
	}
	if err != nil {
		return "commit"
	}
	defer unix.Flock(int(parent.Fd()), unix.LOCK_UN)
	if !held.fresh() {
		return "stale"
	}
	if unchanged {
		return "unchanged"
	}
	var randomness [16]byte
	if _, err := rand.Read(randomness[:]); err != nil {
		return "commit"
	}
	name := ".onboard-kit-edit-" + hex.EncodeToString(randomness[:])
	fd, err := unix.Openat(int(parent.Fd()), name, unix.O_RDWR|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if err != nil {
		return "commit"
	}
	temporary := os.NewFile(uintptr(fd), name)
	var initial unix.Stat_t
	statErr := unix.Fstat(fd, &initial)
	owned := controls{Device: uint64(initial.Dev), Inode: uint64(initial.Ino)}
	cleanup := func(code string) string {
		if temporary != nil {
			if temporary.Close() != nil && code == "written" {
				code = "cleanup-written"
			}
			temporary = nil
		}
		var current unix.Stat_t
		err := unix.Fstatat(int(parent.Fd()), name, &current, unix.AT_SYMLINK_NOFOLLOW)
		if err == unix.ENOENT {
			return code
		}
		if err != nil || uint64(current.Dev) != owned.Device || uint64(current.Ino) != owned.Inode || unix.Unlinkat(int(parent.Fd()), name, 0) != nil {
			if code == "written" || code == "cleanup-written" {
				return "cleanup-written"
			}
			return "cleanup-refused"
		}
		return code
	}
	if statErr != nil {
		return cleanup("commit")
	}
	if initial.Uid != uint32(os.Geteuid()) {
		return cleanup("commit")
	}
	if initial.Gid != uint32(os.Getegid()) && unix.Fchown(fd, -1, os.Getegid()) != nil {
		return cleanup("commit")
	}
	if unix.Fchmod(fd, 0600) != nil {
		return cleanup("commit")
	}
	private, err := inspect(temporary)
	if err != nil || !privateFile(private) {
		return cleanup("commit")
	}
	if establishAttributes(fd, held.facts.Attributes) != nil {
		return cleanup("commit")
	}
	private, err = inspect(temporary)
	if err != nil || !privateFile(private) || !reflect.DeepEqual(private.Attributes, held.facts.Attributes) {
		return cleanup("commit")
	}
	owned = private
	n, err := temporary.Write(data)
	if err != nil || n != len(data) {
		return cleanup("commit")
	}
	after, err := inspect(temporary)
	if err != nil || !privateFile(after) || !reflect.DeepEqual(after.Attributes, held.facts.Attributes) {
		return cleanup("commit")
	}
	actual, err := boundedRead(temporary)
	if err != nil || !bytes.Equal(actual, data) {
		return cleanup("commit")
	}
	final, err := inspect(temporary)
	if err != nil || !reflect.DeepEqual(after, final) {
		return cleanup("commit")
	}
	if temporary.Close() != nil {
		temporary = nil
		return cleanup("commit")
	}
	temporary = nil
	var named unix.Stat_t
	if unix.Fstatat(int(parent.Fd()), name, &named, unix.AT_SYMLINK_NOFOLLOW) != nil || uint64(named.Dev) != owned.Device || uint64(named.Ino) != owned.Inode {
		return cleanup("commit")
	}
	if !temporaryMatches(parent, name, final, data) {
		return cleanup("commit")
	}
	if !held.fresh() {
		return cleanup("stale")
	}
	if unix.Renameat(int(parent.Fd()), name, int(parent.Fd()), held.name) != nil {
		return cleanup("commit")
	}
	return cleanup("written")
}

func temporaryMatches(parent *os.File, name string, expected controls, data []byte) bool {
	fd, err := unix.Openat(int(parent.Fd()), name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return false
	}
	file := os.NewFile(uintptr(fd), name)
	before, err := inspect(file)
	if err != nil || !reflect.DeepEqual(before, expected) {
		file.Close()
		return false
	}
	actual, err := boundedRead(file)
	if err != nil || !bytes.Equal(actual, data) {
		file.Close()
		return false
	}
	after, err := inspect(file)
	closed := file.Close()
	return err == nil && closed == nil && reflect.DeepEqual(after, expected)
}

func attributes(fd int) (map[string]string, error) {
	length, err := unix.Flistxattr(fd, nil)
	if err != nil || length < 0 || length > 65536 {
		return nil, failNative()
	}
	names := make([]byte, length)
	actual, err := unix.Flistxattr(fd, names)
	if err != nil || actual != length {
		return nil, failNative()
	}
	result := map[string]string{}
	budget := length
	ordered := strings.Split(string(names), "\x00")
	sort.Strings(ordered)
	for _, name := range ordered {
		if name == "" {
			continue
		}
		length, err := unix.Fgetxattr(fd, name, nil)
		if err != nil || length < 0 || length > 65536 || length > 262144-budget {
			return nil, failNative()
		}
		value := make([]byte, length)
		actual, err := unix.Fgetxattr(fd, name, value)
		if err != nil || actual != length {
			return nil, failNative()
		}
		result[name] = string(value)
		budget += length
	}
	return result, nil
}
