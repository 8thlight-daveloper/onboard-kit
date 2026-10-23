package guide_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

type borrowedInput struct {
	Reader io.Reader
	Closed bool
}

func (r *borrowedInput) Read(value []byte) (int, error) { return r.Reader.Read(value) }
func (r *borrowedInput) Close() error                   { r.Closed = true; return nil }

type borrowedOutput struct {
	bytes.Buffer
	Closed bool
}

func (w *borrowedOutput) Close() error { w.Closed = true; return nil }

func originalInodeDescriptors(t *testing.T, identity unix.Stat_t) int {
	t.Helper()
	directory, err := os.Open("/dev/fd")
	if err != nil {
		t.Fatal("cannot observe native descriptor names")
	}
	names, readErr := directory.Readdirnames(-1)
	closeErr := directory.Close()
	if readErr != nil || closeErr != nil {
		t.Fatal("cannot finish native descriptor-name observation")
	}
	count := 0
	for _, name := range names {
		descriptor, err := strconv.Atoi(name)
		if err != nil {
			continue
		}
		var current unix.Stat_t
		if err := unix.Fstat(descriptor, &current); errors.Is(err, unix.EBADF) {
			continue
		} else if err != nil {
			t.Fatal("cannot identify observed native descriptor")
		}
		if current.Dev == identity.Dev && current.Ino == identity.Ino {
			count++
		}
	}
	return count
}

func TestOwnedInspectionDescriptorReleasedWithoutClosingBorrowedStreams(t *testing.T) {
	box := newSandbox(t, "identity.yaml")
	anchor := filepath.Join(box.Root, "owned-original-preparation.yaml")
	if err := os.Link(box.Prep, anchor); err != nil {
		t.Fatal("cannot anchor owned original preparation inode")
	}
	before := snapshot(t, box.Root)
	var original unix.Stat_t
	if err := unix.Stat(box.Prep, &original); err != nil {
		t.Fatal("cannot identify original owned preparation inode")
	}
	if count := originalInodeDescriptors(t, original); count != 0 {
		t.Fatalf("fixture unexpectedly holds %d descriptors to original inode", count)
	}
	input := &borrowedInput{Reader: strings.NewReader("recipient\n1\nchange\nExample newcomer\napply\nreview\nsave\n")}
	var stdout, stderr borrowedOutput
	if exit := newCommand().Run(box.args(), input, &stdout, &stderr); exit != 0 {
		t.Fatalf("descriptor lifecycle save exit %d", exit)
	}
	assertReport(t, stderr.Bytes(), "saved", "guide.saved")
	if input.Closed || stdout.Closed || stderr.Closed {
		t.Fatal("driver closed a caller-owned stream")
	}
	if count := originalInodeDescriptors(t, original); count != 0 {
		t.Fatalf("Run returned retaining %d owned descriptors to original preparation inode", count)
	}
	if held, err := os.ReadFile(anchor); err != nil || !bytes.Equal(held, fixture(t, "identity.yaml")) {
		t.Fatal("owned hard-link anchor lost original preparation bytes")
	}
	var current unix.Stat_t
	if err := unix.Stat(box.Prep, &current); err != nil {
		t.Fatal("committed preparation disappeared")
	}
	if current.Dev == original.Dev && current.Ino == original.Ino {
		t.Fatal("save must replace selected pathname inode")
	}
	before["records/newcomer/preparation.obk.yaml"] = string(fixture(t, "named.yaml"))
	if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
		t.Fatal("descriptor lifecycle save changed unexpected bytes or left publication paths")
	}
}
