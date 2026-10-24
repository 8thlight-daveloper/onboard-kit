package editor

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"

	"golang.org/x/sys/unix"
)

func (c *context) publish(data []byte) (saved bool, clean bool) {
	if len(data) > 1048576 || c.fresh() != "" {
		return false, false
	}
	parent := int(c.record.parent.Fd())
	name := c.stageName()
	err := unix.Mkdirat(parent, name, 0700)
	created := err == nil
	if err != nil && err != unix.EEXIST {
		return false, false
	}
	fd, err := unix.Openat(parent, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return false, false
	}
	stage := os.NewFile(uintptr(fd), name)
	defer func() {
		if stage.Close() != nil {
			clean = false
		}
	}()
	stageGuard, ok := guard(stage)
	if !ok || !c.private(stageGuard, 0700) || created && !initialOwned(stageGuard, 0700) {
		return false, false
	}
	names, ok := entries(stage, 0)
	if !ok || len(names) != 0 {
		return false, false
	}
	fd, err = unix.Openat(int(stage.Fd()), "record.new", unix.O_RDWR|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if err != nil {
		return false, false
	}
	payload := os.NewFile(uintptr(fd), "record.new")
	payloadGuard, ok := guard(payload)
	if !ok || !c.private(payloadGuard, 0600) || payloadGuard.GID != uint32(os.Getegid()) {
		_ = payload.Close()
		return false, false
	}
	published := false
	defer func() {
		if !published {
			var stat unix.Stat_t
			if unix.Fstatat(int(stage.Fd()), "record.new", &stat, unix.AT_SYMLINK_NOFOLLOW) == nil && uint64(stat.Dev) == payloadGuard.Device && uint64(stat.Ino) == payloadGuard.Inode {
				_ = unix.Unlinkat(int(stage.Fd()), "record.new", 0)
			}
		}
		if payload != nil {
			_ = payload.Close()
		}
	}()
	if n, err := payload.Write(data); err != nil || n != len(data) {
		return false, false
	}
	if !copyNativeAttributes(int(payload.Fd()), c.record.guard) {
		return false, false
	}
	body, ok := boundedBody(payload, 1048576)
	if !ok || !bytes.Equal(body, data) {
		return false, false
	}
	finalGuard, ok := guard(payload)
	if !ok || finalGuard.UID != c.record.guard.UID || finalGuard.GID != c.record.guard.GID || finalGuard.Mode != c.record.guard.Mode || !reflect.DeepEqual(finalGuard.Attributes, c.record.guard.Attributes) {
		return false, false
	}
	if !c.publicationFresh() {
		return false, false
	}
	actualStage, ok := guard(stage)
	if !ok || !reflect.DeepEqual(actualStage, stageGuard) {
		return false, false
	}
	var stat unix.Stat_t
	if unix.Fstatat(parent, name, &stat, unix.AT_SYMLINK_NOFOLLOW) != nil || uint64(stat.Dev) != stageGuard.Device || uint64(stat.Ino) != stageGuard.Inode {
		return false, false
	}
	if unix.Fstatat(int(stage.Fd()), "record.new", &stat, unix.AT_SYMLINK_NOFOLLOW) != nil || uint64(stat.Dev) != payloadGuard.Device || uint64(stat.Ino) != payloadGuard.Inode {
		return false, false
	}
	if unix.Renameat(int(stage.Fd()), "record.new", parent, c.record.name) != nil {
		return false, false
	}
	published = true
	if unix.Fstatat(parent, c.record.name, &stat, unix.AT_SYMLINK_NOFOLLOW) != nil || uint64(stat.Dev) != payloadGuard.Device || uint64(stat.Ino) != payloadGuard.Inode {
		return true, false
	}
	if payload.Close() != nil {
		return true, false
	}
	payload = nil
	names, ok = entries(stage, 0)
	if !ok || len(names) != 0 {
		return true, false
	}
	return true, true
}

func (c *context) publicationFresh() bool {
	if !c.pending() || !c.absencesFresh() {
		return false
	}
	probe := &context{guards: map[string]nativeGuard{}, groups: c.groups}
	defer probe.close()
	for path, expected := range c.guards {
		parent, ok := probe.absolute(filepath.Dir(path))
		if !ok {
			return false
		}
		var file *os.File
		if path == "/" {
			file = parent
		} else {
			fd, err := unix.Openat(int(parent.Fd()), filepath.Base(path), unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
			if err != nil {
				return false
			}
			file = os.NewFile(uintptr(fd), path)
			probe.files = append(probe.files, file)
		}
		actual, ok := guard(file)
		if !ok || !reflect.DeepEqual(actual, expected) {
			return false
		}
	}
	for _, snapshot := range []snapshot{c.plan, c.preparation, c.record} {
		var stat unix.Stat_t
		if unix.Fstatat(int(snapshot.parent.Fd()), snapshot.name, &stat, unix.AT_SYMLINK_NOFOLLOW) != nil || uint64(stat.Dev) != snapshot.guard.Device || uint64(stat.Ino) != snapshot.guard.Inode {
			return false
		}
		body, ok := boundedBody(snapshot.file, 1048576)
		if !ok || !bytes.Equal(body, snapshot.body) {
			return false
		}
	}
	for path, expected := range c.metadata {
		parent, ok := probe.absolute(filepath.Dir(path))
		if !ok {
			return false
		}
		selected, ok := probe.participant(parent, filepath.Base(path))
		if !ok {
			return false
		}
		body, ok := boundedBody(selected.file, 2097152)
		if !ok || !bytes.Equal(body, expected) {
			return false
		}
	}
	sessions, ok := c.directory(c.home, "sessions", c.home.Name()+"/sessions")
	if !ok {
		return false
	}
	keys, ok := entries(sessions, 128)
	if !ok || len(keys) != len(c.metadata) {
		return false
	}
	for _, key := range keys {
		if _, ok := c.metadata[c.home.Name()+"/sessions/"+key+"/session.json"]; !ok {
			return false
		}
	}
	return true
}
