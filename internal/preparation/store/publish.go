package store

import (
	"errors"
	"os"
	"strconv"

	contract "onboard-kit/contracts/preparation/store"

	"golang.org/x/sys/unix"
)

const (
	privateFileMode      = 0o600
	privateDirectoryMode = 0o700
	otherAccessBits      = 0o077
	temporaryAttempts    = 256
)

type publication struct {
	parent    int
	temporary string
	created   bool
	committed bool
}

func (p *publication) finish(code contract.Code) contract.SaveResult {
	err := unix.Unlinkat(p.parent, p.temporary, 0)
	if err != nil && !errors.Is(err, unix.ENOENT) {
		code = contract.CodeCleanupRefused
		if p.committed {
			code = contract.CodeCleanupSaved
		}
	}
	return contract.SaveResult{Code: code, Committed: code == contract.CodeSaved || code == contract.CodeCleanupSaved, CreatedDirectories: p.created}
}

func createDirectories(root int, parts []string) (int, bool, error) {
	descriptor, err := unix.Openat(root, ".", directoryFlags, 0)
	if err != nil {
		return -1, false, err
	}
	created := false
	for _, part := range parts {
		next, err := openChildDirectory(descriptor, part)
		if errors.Is(err, unix.ENOENT) {
			var made bool
			next, made, err = makeChildDirectory(descriptor, part)
			created = created || made
		}
		unix.Close(descriptor)
		if err != nil {
			return -1, created, err
		}
		descriptor = next
	}
	return descriptor, created, nil
}

func makeChildDirectory(parent int, name string) (int, bool, error) {
	if !aclAbsent(parent) {
		return -1, false, unix.EACCES
	}
	if err := unix.Mkdirat(parent, name, privateDirectoryMode); err != nil {
		return -1, false, err
	}
	child, err := openChildDirectory(parent, name)
	if err != nil {
		return -1, true, err
	}
	var stat unix.Stat_t
	if unix.Fstat(child, &stat) != nil || uint32(stat.Mode)&otherAccessBits != 0 || !aclAbsent(child) {
		unix.Close(child)
		return -1, true, unix.EACCES
	}
	return child, true, nil
}

func createTemporary(parent int) (string, int, error) {
	for attempt := 1; attempt <= temporaryAttempts; attempt++ {
		name := ".preparation-store-" + strconv.Itoa(attempt) + ".tmp"
		descriptor, err := unix.Openat(parent, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, privateFileMode)
		if err == nil {
			return name, descriptor, nil
		}
		if !errors.Is(err, unix.EEXIST) {
			return "", -1, err
		}
	}
	return "", -1, unix.EEXIST
}

func existingAccessKept(parent int, name string, held fileSnapshot) bool {
	if !held.exists {
		return true
	}
	opened, err := openRegular(parent, name)
	if err != nil {
		return false
	}
	kept := identityOf(opened.stat) == held.facts.identity && aclAbsent(opened.descriptor)
	return unix.Close(opened.descriptor) == nil && kept
}

func fillTemporary(descriptor int, held fileSnapshot, data []byte) bool {
	mode, ok := temporaryMode(descriptor, held)
	if !ok || unix.Fchmod(descriptor, mode) != nil {
		unix.Close(descriptor)
		return false
	}
	file := os.NewFile(uintptr(descriptor), "owned-publication")
	written, writeErr := file.Write(data)
	closeErr := file.Close()
	return writeErr == nil && closeErr == nil && written == len(data)
}

func temporaryMode(descriptor int, held fileSnapshot) (uint32, bool) {
	var stat unix.Stat_t
	if unix.Fstat(descriptor, &stat) != nil || uint32(stat.Mode)&unix.S_IFMT != unix.S_IFREG || uint32(stat.Mode)&otherAccessBits != 0 || !aclAbsent(descriptor) {
		return 0, false
	}
	mode := uint32(privateFileMode)
	if held.exists {
		if uint32(stat.Uid) != held.facts.owner || uint32(stat.Gid) != held.facts.group {
			return 0, false
		}
		mode &= held.facts.mode
	}
	return mode, true
}

func publish(state *loadedState, data []byte) contract.SaveResult {
	root, ok := openUnchangedRoot(state)
	if !ok {
		return contract.SaveResult{Code: contract.CodeStale}
	}
	defer unix.Close(root)
	if !aclAnswerable(root) {
		return contract.SaveResult{Code: contract.CodeVolume}
	}
	directories, name := splitDestination(state.destination)
	parent, created, err := createDirectories(root, directories)
	if err != nil {
		return contract.SaveResult{Code: contract.CodeCommit, CreatedDirectories: created}
	}
	defer unix.Close(parent)
	if !aclAnswerable(parent) {
		return contract.SaveResult{Code: contract.CodeVolume, CreatedDirectories: created}
	}
	if !aclAbsent(parent) || !existingAccessKept(parent, name, state.preparation) {
		return contract.SaveResult{Code: contract.CodeCommit, CreatedDirectories: created}
	}
	temporary, descriptor, err := createTemporary(parent)
	if err != nil {
		return contract.SaveResult{Code: contract.CodeCommit, CreatedDirectories: created}
	}
	attempt := &publication{parent: parent, temporary: temporary, created: created}
	if !fillTemporary(descriptor, state.preparation, data) {
		return attempt.finish(contract.CodeCommit)
	}
	if !unchangedSinceLoad(state) {
		return attempt.finish(contract.CodeStale)
	}
	if state.preparation.exists {
		err = unix.Renameat(parent, temporary, parent, name)
	} else {
		err = unix.Linkat(parent, temporary, parent, name, 0)
	}
	if errors.Is(err, unix.EEXIST) {
		return attempt.finish(contract.CodeStale)
	}
	if err != nil {
		return attempt.finish(contract.CodeCommit)
	}
	attempt.committed = true
	return attempt.finish(contract.CodeSaved)
}
