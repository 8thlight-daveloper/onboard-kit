package creator_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestExistingAndUnsafeDestinationsRefuseBeforeReading(t *testing.T) {
	for _, kind := range []string{"existing", "directory", "dangling-symlink", "parent-symlink", "missing-parent", "dotdot"} {
		t.Run(kind, func(t *testing.T) {
			root, target := sandbox(t)
			switch kind {
			case "existing":
				if err := os.WriteFile(target, []byte("OWNER BYTES\n"), 0600); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Mkdir(target, 0700); err != nil {
					t.Fatal(err)
				}
			case "dangling-symlink":
				if err := os.Symlink(filepath.Join(root, "absent"), target); err != nil {
					t.Fatal(err)
				}
			case "parent-symlink":
				if err := os.Symlink(root, filepath.Join(root, "alias")); err != nil {
					t.Fatal(err)
				}
				target = filepath.Join(root, "alias", "new.yaml")
			case "missing-parent":
				target = filepath.Join(root, "missing", "new.yaml")
			case "dotdot":
				target = root + "/../" + filepath.Base(root) + "/new.yaml"
			}
			input := &countedInput{source: []byte("review\nsave\n")}
			var out, stderr bytes.Buffer
			if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, &out, &stderr) != 1 || input.position != 0 || out.Len() != 0 {
				t.Error("unsafe destination crossed admission")
			}
			assertReport(t, stderr.Bytes(), "refused", "create.path")
			if kind == "existing" {
				body, err := os.ReadFile(target)
				if err != nil || string(body) != "OWNER BYTES\n" {
					t.Error("existing target overwritten")
				}
			}
			if kind == "missing-parent" {
				if _, err := os.Lstat(filepath.Join(root, "missing")); !os.IsNotExist(err) {
					t.Error("preconsent directory created")
				}
			}
			preserved(t, root)
		})
	}
}

type consentReader struct {
	before   []byte
	position int
	change   func()
	changed  bool
}

func (r *consentReader) Read(p []byte) (int, error) {
	if r.position == len("review\n") && !r.changed {
		r.changed = true
		r.change()
	}
	if r.position == len(r.before) {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	p[0] = r.before[r.position]
	r.position++
	return 1, nil
}

func TestAppearingTargetPreservesOwnerBytesAfterReview(t *testing.T) {
	root, target := sandbox(t)
	input := &consentReader{before: []byte("review\nsave\n"), change: func() {
		if err := os.WriteFile(target, []byte("NEW OWNER FILE\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}}
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, &out, &stderr) != 1 {
		t.Error("stale save accepted")
	}
	assertReport(t, stderr.Bytes(), "refused", "create.stale")
	body, err := os.ReadFile(target)
	if err != nil || string(body) != "NEW OWNER FILE\n" {
		t.Error("concurrently appearing owner file changed")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Errorf("unexpected publication remnants: %v", entries)
	}
	preserved(t, root)
}

func TestReplacedParentInvalidatesReviewedDestinationWithoutWritingEitherDirectory(t *testing.T) {
	root, _ := sandbox(t)
	parent := filepath.Join(root, "selected")
	moved := filepath.Join(root, "moved-selected")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chown(parent, -1, os.Getegid()); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "owner.txt"), []byte("ORIGINAL PARENT\n"), 0600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(parent, "created.obk.yaml")
	input := &consentReader{before: []byte("review\nsave\n"), change: func() {
		if err := os.Rename(parent, moved); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(parent, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.Chown(parent, -1, os.Getegid()); err != nil {
			t.Fatal(err)
		}
	}}
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, &out, &stderr) != 1 {
		t.Error("replaced parent retained stale save authority")
	}
	assertReport(t, stderr.Bytes(), "refused", "create.stale")
	for _, directory := range []string{parent, moved} {
		if _, err := os.Lstat(filepath.Join(directory, "created.obk.yaml")); !os.IsNotExist(err) {
			t.Errorf("stale publication reached %s", directory)
		}
		entries, err := os.ReadDir(directory)
		if err != nil {
			t.Fatal(err)
		}
		want := 0
		if directory == moved {
			want = 1
		}
		if len(entries) != want {
			t.Errorf("stale parent gained product remnants: %s %v", directory, entries)
		}
	}
	body, err := os.ReadFile(filepath.Join(moved, "owner.txt"))
	if err != nil || string(body) != "ORIGINAL PARENT\n" {
		t.Error("original parent artifact changed")
	}
	preserved(t, root)
}

func TestCreatedFileHasExactPrivateNativeModeAndNoTemporaryRemnants(t *testing.T) {
	root, target := sandbox(t)
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, strings.NewReader("review\nsave\n"), &out, &stderr) != 0 {
		t.Fatalf("save failed: %q", stderr.String())
	}
	info, err := os.Lstat(target)
	if err != nil {
		t.Fatal(err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.Mode().IsRegular() || info.Mode().Perm() != 0600 || stat.Uid != uint32(os.Geteuid()) || stat.Gid != uint32(os.Getegid()) {
		t.Error("created file controls differ from private effective ownership")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Errorf("unexpected temporary remnant: %v", entries)
	}
	preserved(t, root)
}
