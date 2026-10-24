//go:build darwin || linux

package main

import (
	"os"
	"syscall"
	"testing"
)

func TestVerificationWorkspaceHasEffectiveNativeOwnerBeforeCallerWrites(t *testing.T) {
	repository := t.TempDir()
	workspace, err := verificationWorkspace(repository)
	if err != nil {
		t.Fatalf("create fresh verification workspace: %v", err)
	}
	info, err := os.Lstat(workspace)
	if err != nil {
		t.Fatalf("inspect returned verification workspace: %v", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("returned workspace must be a real directory: %v", info.Mode())
	}
	t.Cleanup(func() {
		if err := os.Remove(workspace); err != nil && !os.IsNotExist(err) {
			t.Errorf("remove empty verification workspace: %v", err)
		}
	})
	entries, err := os.ReadDir(workspace)
	if err != nil {
		t.Fatalf("inspect fresh workspace contents: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("workspace already contains caller data or child objects: %d entries", len(entries))
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatal("native UID/GID ownership could not be inspected on this POSIX host")
	}
	if stat.Uid != uint32(os.Geteuid()) || stat.Gid != uint32(os.Getegid()) {
		t.Fatalf("fresh workspace native owner UID/GID %d/%d, want effective UID/GID %d/%d before caller writes", stat.Uid, stat.Gid, os.Geteuid(), os.Getegid())
	}
}
