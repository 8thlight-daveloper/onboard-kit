package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

const owner = "onboard-kit-manual-test-v1\n"

func (r runner) checked(path string, directory bool) (os.FileInfo, error) {
	relative, err := filepath.Rel(r.root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return nil, errors.New("demo path is outside the checkout")
	}
	current := r.root
	var info os.FileInfo
	for _, part := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err = os.Lstat(current)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("refusing symlink: %s", current)
		}
		if current != path && !info.IsDir() {
			return nil, fmt.Errorf("not a directory: %s", current)
		}
	}
	if directory && !info.IsDir() || !directory && !info.Mode().IsRegular() {
		return nil, fmt.Errorf("unexpected file kind: %s", path)
	}
	return info, nil
}

func (r runner) regular(path string) error {
	_, err := r.checked(path, false)
	return err
}

func (r runner) mkdir(path string) error {
	if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
		if _, err := r.checked(filepath.Dir(path), true); err != nil {
			return err
		}
		if err := os.Mkdir(path, 0700); err != nil {
			return err
		}
	}
	info, err := r.checked(path, true)
	if err == nil && info.Mode().Perm() != 0700 {
		err = fmt.Errorf("demo directory must already be private (0700), not changing it: %s", path)
	}
	return err
}

func (r runner) workspace() error {
	output := filepath.Join(r.root, "output")
	if _, err := os.Lstat(output); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(output, 0755); err != nil {
			return err
		}
	}
	if _, err := r.checked(output, true); err != nil {
		return err
	}
	_, err := os.Lstat(r.base)
	if errors.Is(err, os.ErrNotExist) {
		if err := r.mkdir(r.base); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(r.base, ".owner"), []byte(owner), 0600)
	}
	if err != nil {
		return err
	}
	if err := r.mkdir(r.base); err != nil {
		return err
	}
	marker := filepath.Join(r.base, ".owner")
	data, err := r.read(marker)
	if err != nil || string(data) != owner {
		return errors.New("refusing unrecognized manual-test workspace; preserve it and choose another location")
	}
	return nil
}

func (r runner) directory(stage string) error {
	return r.mkdir(filepath.Join(r.base, stage))
}

func (r runner) lock(stage string) (*os.File, error) {
	path := filepath.Join(r.base, ".lock-"+stage)
	if _, err := os.Lstat(path); err == nil {
		if err := r.regular(path); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	fd, err := unix.Open(path, unix.O_RDWR|unix.O_CREAT|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	var stat unix.Stat_t
	if unix.Fstat(fd, &stat) != nil || stat.Nlink != 1 || stat.Uid != uint32(os.Geteuid()) || stat.Mode&0777 != 0600 || unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB) != nil {
		file.Close()
		return nil, fmt.Errorf("stage %s is active or its lock is unsafe; close that stage before resetting it", stage)
	}
	return file, nil
}

func (r runner) reset(stage string) error {
	path := filepath.Join(r.base, stage)
	if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
		_, err := fmt.Fprintln(r.output, stage+": already at the initial state")
		return err
	}
	if _, err := r.checked(path, true); err != nil {
		return err
	}
	history := filepath.Join(r.root, "output", "manual-test-history")
	if err := r.mkdir(history); err != nil {
		return err
	}
	slot, err := os.MkdirTemp(history, time.Now().UTC().Format("20060102T150405Z")+"-")
	if err != nil {
		return err
	}
	destination := filepath.Join(slot, stage)
	if err := os.Rename(path, destination); err != nil {
		return err
	}
	_, err = fmt.Fprintf(r.output, "%s reset; recoverable data: %s\n", stage, destination)
	return err
}

func (r runner) read(path string) ([]byte, error) {
	if err := r.regular(path); err != nil {
		return nil, err
	}
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_NONBLOCK|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 1048577))
	if err == nil && len(data) > 1048576 {
		err = errors.New("demo input exceeds 1 MiB")
	}
	return data, err
}

func (r runner) snapshot(source, destination string) error {
	data, err := r.read(source)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(destination); err == nil {
		if err := r.regular(destination); err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(destination), ".snapshot-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		return errors.Join(writeErr, closeErr)
	}
	return os.Rename(file.Name(), destination)
}
