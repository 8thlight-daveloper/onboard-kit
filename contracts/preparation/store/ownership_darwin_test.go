package store_test

import (
	"bytes"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	contract "onboard-kit/contracts/preparation/store"
)

func groupOf(t *testing.T, path string) int {
	t.Helper()
	var stat syscall.Stat_t
	if err := syscall.Stat(path, &stat); err != nil {
		t.Fatal(err)
	}
	return int(stat.Gid)
}

func otherGroup(t *testing.T, than int) int {
	t.Helper()
	groups, err := os.Getgroups()
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range groups {
		if group != than {
			return group
		}
	}
	t.Skipf("the current user belongs to no group other than %d, so a group mismatch cannot be set up", than)
	return -1
}

func TestExistingPreparationWithAnotherGroupIsNeverSavedWithTheWrongGroup(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name   string
		target func(box sandbox) string
	}{
		{name: "preparation-group-differs-from-its-directory", target: func(box sandbox) string { return box.Prep }},
		{name: "directory-group-differs-from-the-preparation", target: func(box sandbox) string { return filepath.Dir(box.Prep) }},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			box := newSandbox(t, "plan.yaml", "identity.yaml")
			other := otherGroup(t, groupOf(t, box.Prep))
			if err := os.Chown(c.target(box), -1, other); err != nil {
				t.Skipf("cannot change the group to %d: %v", other, err)
			}
			loaded := box.load(t)
			held := snapshot(t, box.Base)
			group, before := groupOf(t, box.Prep), identityOf(t, box.Prep)
			result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "named.yaml")})
			wantCode(t, result.Code, contract.CodeCommit)
			if result.Committed || result.CreatedDirectories {
				t.Fatalf("refusal reports effects %#v", result)
			}
			if after := snapshot(t, box.Base); !sameMap(after, held) {
				t.Fatal("refused save changed the filesystem")
			}
			if !bytes.Equal(readBytes(t, box.Prep), fixture(t, "identity.yaml")) || groupOf(t, box.Prep) != group || identityOf(t, box.Prep) != before {
				t.Fatal("refused save changed the held file")
			}
		})
	}
}
