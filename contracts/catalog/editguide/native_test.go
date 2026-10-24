package editguide_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

func TestUnsafeNativeProfilesRefuseWithoutInputOrControlRepair(t *testing.T) {
	for _, kind := range []string{"public-file", "hardlink", "writable-parent", "symlink"} {
		root, target := setup(t, 0755)
		selected := target
		switch kind {
		case "public-file":
			if err := os.Chmod(target, 0644); err != nil {
				t.Fatal(err)
			}
		case "hardlink":
			if err := os.Link(target, filepath.Join(root, "second-name.yaml")); err != nil {
				t.Fatal(err)
			}
		case "writable-parent":
			if err := os.Chmod(root, 0775); err != nil {
				t.Fatal(err)
			}
		case "symlink":
			selected = filepath.Join(root, "alias.yaml")
			if err := os.Symlink(target, selected); err != nil {
				t.Fatal(err)
			}
		}
		before, err := os.Stat(target)
		if err != nil {
			t.Fatal(err)
		}
		parentBefore, err := os.Stat(root)
		if err != nil {
			t.Fatal(err)
		}
		input := &countedInput{source: []byte("review\nsave\n")}
		var out, stderr bytes.Buffer
		if newCommand().Run([]string{"--plan=" + selected}, input, &out, &stderr) != 1 || input.position != 0 || out.Len() != 0 {
			t.Error("unsafe native profile crossed startup/input")
		}
		report(t, stderr.Bytes(), "refused", "edit.path", "Select a safe single-link effective-owned0600 plan in an owned directory without group/other write; unsupported controls require separate review.")
		after, err := os.Stat(target)
		if err != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() {
			t.Error("original security profile was silently repaired")
		}
		parentAfter, err := os.Stat(root)
		if err != nil || parentAfter.Mode() != parentBefore.Mode() {
			t.Error("unrelated parent permissions changed")
		}
		unchanged(t, target)
		personalPreserved(t, root)
	}
}

func TestCooperativeParentLockBusyRefusesWithoutReplacement(t *testing.T) {
	root, target := setup(t, 0755)
	parent, err := os.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Close()
	if err := unix.Flock(int(parent.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		t.Skipf("named native fixture gap: parent advisory lock precondition unavailable: %v", err)
	}
	defer unix.Flock(int(parent.Fd()), unix.LOCK_UN)
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--plan=" + target}, strings.NewReader("retire\nresource-old\nreview\nsave\n"), &out, &stderr) != 1 {
		t.Error("busy cooperative parent lock did not refuse")
	}
	report(t, stderr.Bytes(), "refused", "edit.busy", "Another cooperating plan edit is saving; reload and review again.")
	unchanged(t, target)
	personalPreserved(t, root)
}

func TestSelectedParentGroupNeedNotEqualEffectiveGroup(t *testing.T) {
	groups, err := os.Getgroups()
	if err != nil {
		t.Fatal(err)
	}
	other := -1
	for _, group := range groups {
		if group != os.Getegid() {
			other = group
			break
		}
	}
	if other < 0 {
		t.Skip("named native fixture gap: no other permitted group for independently owned parent")
	}
	root, target := setup(t, 0755)
	if err := os.Chown(root, -1, other); err != nil {
		t.Skipf("named native fixture gap: owned parent group change unavailable: %v", err)
	}
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--plan=" + target}, strings.NewReader("retire\nresource-old\nreview\nsave\n"), &out, &stderr) != 0 {
		t.Fatalf("different retained parent group was unnecessarily refused: %q", stderr.String())
	}
	data, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(data, fixture(t, "retired.yaml")) {
		t.Error("supported retained-group parent did not save exact reviewed bytes")
	}
	personalPreserved(t, root)
	info, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	native, ok := info.Sys().(*syscall.Stat_t)
	if !ok || native.Gid != uint32(other) || info.Mode().Perm() != 0755 {
		t.Error("parent retained group or mode was changed to qualify")
	}
}

func TestObservedSymlinkSubstitutionCannotOverwriteSelectedNamespaceOrRecord(t *testing.T) {
	root, target := setup(t, 0755)
	input := &countedInput{source: []byte("retire\nresource-old\nreview\nsave\n"), hookAt: len("retire\nresource-old\nreview\n"), hook: func() {
		if err := os.Remove(target); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("record.yaml", target); err != nil {
			t.Fatal(err)
		}
	}}
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--plan=" + target}, input, &out, &stderr) != 1 {
		t.Error("observed namespace substitution accepted")
	}
	report(t, stderr.Bytes(), "refused", "edit.stale", "The selected plan context changed; reload and review again.")
	info, err := os.Lstat(target)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Error("independent observed symlink namespace was replaced")
	}
	personalPreserved(t, root)
}
