package runner

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
	"golang.org/x/sys/unix"
)

var errUnsafePath = errors.New("unsafe native path")
var errMissingPath = errors.New("missing native path")

func localParts(path string) ([]string, error) {
	if path == "" || strings.IndexByte(path, 0) >= 0 || !filepath.IsLocal(path) {
		return nil, errUnsafePath
	}
	var parts []string
	for _, part := range strings.Split(path, string(filepath.Separator)) {
		if part == ".." {
			return nil, errUnsafePath
		}
		if part != "" && part != "." {
			parts = append(parts, part)
		}
	}
	return parts, nil
}

func ownedNative(file *os.File, directory bool) bool {
	var info unix.Stat_t
	if unix.Fstat(int(file.Fd()), &info) != nil || info.Uid != uint32(os.Geteuid()) {
		return false
	}
	kind := uint32(unix.S_IFREG)
	if directory {
		kind = unix.S_IFDIR
	}
	return uint32(info.Mode)&unix.S_IFMT == kind
}

func openSandbox(folder string) (*os.File, error) {
	descriptor, err := unix.Open(folder, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, errUnsafePath
	}
	file := os.NewFile(uintptr(descriptor), folder)
	if !ownedNative(file, true) {
		file.Close()
		return nil, errUnsafePath
	}
	return file, nil
}

func nativeParent(folder, path string, create bool) (*os.File, string, error) {
	parts, err := localParts(path)
	if err != nil {
		return nil, "", err
	}
	parent, err := openSandbox(folder)
	if err != nil {
		return nil, "", err
	}
	if len(parts) == 0 {
		return parent, ".", nil
	}
	for _, part := range parts[:len(parts)-1] {
		descriptor, err := unix.Openat(int(parent.Fd()), part, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
		if errors.Is(err, unix.ENOENT) && create {
			if err = unix.Mkdirat(int(parent.Fd()), part, 0o755); err == nil || errors.Is(err, unix.EEXIST) {
				descriptor, err = unix.Openat(int(parent.Fd()), part, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
			}
		}
		if err != nil {
			parent.Close()
			if errors.Is(err, unix.ENOENT) {
				return nil, "", errMissingPath
			}
			return nil, "", errUnsafePath
		}
		next := os.NewFile(uintptr(descriptor), part)
		parent.Close()
		parent = next
		if !ownedNative(parent, true) {
			parent.Close()
			return nil, "", errUnsafePath
		}
	}
	return parent, parts[len(parts)-1], nil
}

func proveNative(folder, path string, allowMissing bool) error {
	parent, name, err := nativeParent(folder, path, false)
	if errors.Is(err, errMissingPath) && allowMissing {
		return nil
	}
	if err != nil {
		return err
	}
	defer parent.Close()
	var info unix.Stat_t
	err = unix.Fstatat(int(parent.Fd()), name, &info, unix.AT_SYMLINK_NOFOLLOW)
	if errors.Is(err, unix.ENOENT) && allowMissing {
		return nil
	}
	if err != nil || info.Uid != uint32(os.Geteuid()) || info.Mode&unix.S_IFMT != unix.S_IFREG {
		return errUnsafePath
	}
	return nil
}

func openNative(folder, path string, directory bool) (*os.File, error) {
	parent, name, err := nativeParent(folder, path, false)
	if err != nil {
		return nil, err
	}
	defer parent.Close()
	flags := unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK
	if directory {
		flags |= unix.O_DIRECTORY
	}
	descriptor, err := unix.Openat(int(parent.Fd()), name, flags, 0)
	if err != nil {
		var info unix.Stat_t
		if unix.Fstatat(int(parent.Fd()), name, &info, unix.AT_SYMLINK_NOFOLLOW) == nil && (info.Uid != uint32(os.Geteuid()) || info.Mode&unix.S_IFMT == unix.S_IFLNK || (info.Mode&unix.S_IFMT != unix.S_IFREG && info.Mode&unix.S_IFMT != unix.S_IFDIR)) {
			return nil, errUnsafePath
		}
		if errors.Is(err, unix.ENOENT) {
			return nil, errMissingPath
		}
		return nil, err
	}
	file := os.NewFile(uintptr(descriptor), name)
	var info unix.Stat_t
	if unix.Fstat(descriptor, &info) != nil || info.Uid != uint32(os.Geteuid()) || (info.Mode&unix.S_IFMT != unix.S_IFREG && info.Mode&unix.S_IFMT != unix.S_IFDIR) {
		file.Close()
		return nil, errUnsafePath
	}
	if !ownedNative(file, directory) {
		file.Close()
		return nil, errors.New("wrong source kind")
	}
	return file, nil
}

func placeNativeFiles(files map[string]FileSource, folder, repositoryRoot string) error {
	for path, source := range files {
		parent, name, err := nativeParent(folder, path, true)
		if err != nil {
			return err
		}
		var info unix.Stat_t
		err = unix.Fstatat(int(parent.Fd()), name, &info, unix.AT_SYMLINK_NOFOLLOW)
		if err == nil && (info.Uid != uint32(os.Geteuid()) || info.Mode&unix.S_IFMT != unix.S_IFREG) {
			parent.Close()
			return errUnsafePath
		}
		if err != nil && !errors.Is(err, unix.ENOENT) {
			parent.Close()
			return errUnsafePath
		}
		content := []byte(source.Text)
		if source.From != "" {
			origin, err := resolve(repositoryRoot, source.From)
			if err != nil {
				parent.Close()
				return err
			}
			content, err = os.ReadFile(origin)
			if err != nil {
				parent.Close()
				return err
			}
		}
		descriptor, err := unix.Openat(int(parent.Fd()), name, unix.O_WRONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CREAT, 0o644)
		parent.Close()
		if err != nil {
			return errUnsafePath
		}
		file := os.NewFile(uintptr(descriptor), name)
		if !ownedNative(file, false) {
			file.Close()
			return errUnsafePath
		}
		if err = file.Truncate(0); err == nil {
			var count int
			count, err = file.Write(content)
			if err == nil && count != len(content) {
				err = io.ErrShortWrite
			}
		}
		closeErr := file.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func captureDirectory(folder, path string) (string, error) {
	directory, err := openNative(folder, path, true)
	if err != nil {
		return "", err
	}
	defer directory.Close()
	entries, err := directory.ReadDir(2)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if len(entries) != 1 {
		return "", errors.New("directory count")
	}
	name := entries[0].Name()
	descriptor, err := unix.Openat(int(directory.Fd()), name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		var info unix.Stat_t
		if unix.Fstatat(int(directory.Fd()), name, &info, unix.AT_SYMLINK_NOFOLLOW) == nil && (info.Uid != uint32(os.Geteuid()) || (info.Mode&unix.S_IFMT != unix.S_IFREG && info.Mode&unix.S_IFMT != unix.S_IFDIR)) {
			return "", errUnsafePath
		}
		return "", errors.New("directory entry")
	}
	child := os.NewFile(uintptr(descriptor), name)
	defer child.Close()
	if !ownedNative(child, true) {
		return "", errUnsafePath
	}
	return name, nil
}

func captureYAML(folder string, selector YAMLScalarCapture) (string, error) {
	file, err := openNative(folder, selector.File, false)
	if err != nil {
		return "", err
	}
	content, readErr := io.ReadAll(io.LimitReader(file, yamlScalarByteLimit+1))
	closeErr := file.Close()
	if readErr != nil {
		return "", readErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	if int64(len(content)) > yamlScalarByteLimit {
		return "", errors.New("YAML limit")
	}
	root, err := decodeYAMLScalarDocument(content)
	if err != nil {
		return "", err
	}
	selected, present, err := selectYAML(root, selector.Path)
	if err != nil {
		return "", err
	}
	if !present || selected.Kind != yaml.ScalarNode || selected.ShortTag() != "!!str" {
		return "", errors.New("YAML string")
	}
	return selected.Value, nil
}
