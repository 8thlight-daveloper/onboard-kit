package starter_test

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

func ownedDescriptorCount(t *testing.T, path string) int {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	identity := info.Sys().(*syscall.Stat_t)
	directory, err := os.Open("/dev/fd")
	if err != nil {
		t.Fatal("native owned descriptor inspection unavailable")
	}
	names, err := directory.Readdirnames(-1)
	if directory.Close() != nil {
		t.Error("descriptor inspection close failed")
	}
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, name := range names {
		fd, err := strconv.Atoi(name)
		if err != nil {
			continue
		}
		var stat syscall.Stat_t
		if syscall.Fstat(fd, &stat) == nil && stat.Dev == identity.Dev && stat.Ino == identity.Ino {
			count++
		}
	}
	return count
}

func TestDarwinOwnedDescriptorsReleasedBeforeSessionInputAndReturn(t *testing.T) {
	s := setup(t)
	put(t, s.record(), fixture(t, "existing.yaml"))
	anchor := filepath.Join(s.base, "held-record-inode")
	if err := os.Link(s.record(), anchor); err != nil {
		t.Fatal(err)
	}
	paths := []string{s.root, s.invitation, anchor}
	before := make([]int, len(paths))
	for i, path := range paths {
		before[i] = ownedDescriptorCount(t, path)
	}
	clock := 0
	prefix := "confirm jules\nbegin\n"
	in := &countedReader{reader: strings.NewReader(prefix + "exit\n"), at: len(prefix)}
	in.hook = func() {
		for i, path := range paths {
			if ownedDescriptorCount(t, path) != before[i] {
				t.Error("owned inspection handles remained open during read-only session input")
			}
		}
	}
	var stderr strings.Builder
	exit := newStarter(runtimeFacts(t, &clock)).Run(s.args(), in, io.Discard, &stderr)
	if exit != 0 || clock != 0 {
		t.Fatal("descriptor control did not reach existing-record session")
	}
	for i, path := range paths {
		if ownedDescriptorCount(t, path) != before[i] {
			t.Error("owned inspection handles leaked after Run")
		}
	}
	assertBytes(t, anchor, fixture(t, "existing.yaml"))
	assertIdle(t, s)
}
