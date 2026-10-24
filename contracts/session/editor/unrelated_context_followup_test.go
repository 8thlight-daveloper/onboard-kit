//go:build darwin || linux

package editor_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

type otherContext struct{ root, plan, metadata, key, planID, account, record, preparation string }

func registeredOtherContext(t *testing.T, selected nativeSession, planID, account string, parents bool) otherContext {
	t.Helper()
	other := otherContext{root: filepath.Join(selected.base, "private-unrelated-root"), planID: planID, account: account}
	if err := os.Mkdir(other.root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chown(other.root, -1, os.Getegid()); err != nil {
		t.Fatal(err)
	}
	root := nativeObjectAt(t, other.root)
	if root.ACL != "absent" || os.SameFile(otherInfo(t, other.root), otherInfo(t, selected.root)) {
		t.Fatal("other real root is not a safely distinct fixture")
	}
	other.key = sessionKey(planID, account, other.root)
	directory := filepath.Join(selected.home, "sessions", other.key)
	other.plan, other.metadata = filepath.Join(directory, "plan.obk.yaml"), filepath.Join(directory, "session.json")
	recordRelative := "records/" + account + "/record.obk.yaml"
	preparationRelative := "records/" + account + "/preparation.obk.yaml"
	other.record, other.preparation = filepath.Join(other.root, filepath.FromSlash(recordRelative)), filepath.Join(other.root, filepath.FromSlash(preparationRelative))
	if parents {
		if err := os.MkdirAll(filepath.Dir(other.record), 0700); err != nil {
			t.Fatal(err)
		}
		for _, parent := range []string{filepath.Join(other.root, "records"), filepath.Dir(other.record)} {
			if err := os.Chown(parent, -1, os.Getegid()); err != nil {
				t.Fatal(err)
			}
			if nativeACL(t, parent) != "absent" {
				t.Fatal("other participant parent ACL fixture is unsafe")
			}
		}
		if os.SameFile(otherInfo(t, filepath.Dir(other.record)), otherInfo(t, filepath.Dir(selected.record))) {
			t.Fatal("missing finals do not have genuinely distinct native parents")
		}
	}
	body := strings.Replace(planBody, "id: plan-one", "id: "+planID, 1) + "# PRIVATE_UNRELATED_PLAN_BODY\n"
	digest := sha256.Sum256([]byte(body))
	metadata := struct {
		Version     int    `json:"version"`
		Key         string `json:"sessionKey"`
		PlanID      string `json:"planIdBase64"`
		Account     string `json:"account"`
		Root        string `json:"root"`
		Device      string `json:"rootDevice"`
		Inode       string `json:"rootInode"`
		Record      string `json:"record"`
		Preparation string `json:"preparation"`
		Hash        string `json:"planSHA256"`
		Transaction string `json:"transaction"`
	}{1, other.key, base64.StdEncoding.EncodeToString([]byte(planID)), account, other.root, strconv.FormatUint(root.Device, 10), strconv.FormatUint(root.Inode, 10), recordRelative, preparationRelative, hex.EncodeToString(digest[:]), strings.Repeat("b", 64)}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	put(t, other.plan, body)
	put(t, other.metadata, string(append(encoded, '\n')))
	if err := os.Chown(directory, -1, os.Getegid()); err != nil {
		t.Fatal(err)
	}
	if nativeACL(t, directory) != "absent" {
		t.Fatal("actual unrelated registry namespace has an ACL")
	}
	return other
}

func otherInfo(t *testing.T, path string) os.FileInfo {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info
}

func provedENOENT(t *testing.T, path string) {
	t.Helper()
	_, err := os.Lstat(path)
	if !errors.Is(err, syscall.ENOENT) {
		t.Fatalf("actual native ENOENT was not established: %v", err)
	}
}

func otherNotDisclosed(t *testing.T, other otherContext, got runResult) {
	t.Helper()
	private := []string{other.root, other.key, other.plan, "PRIVATE_UNRELATED_PLAN_BODY"}
	if other.planID != "plan-one" {
		private = append(private, other.planID)
	}
	if other.account != "newcomer" {
		private = append(private, other.account)
	}
	for _, value := range private {
		if strings.Contains(got.out+got.err, value) {
			t.Fatal("unrelated context was disclosed")
		}
	}
}

func onlySelectedRecordChanged(t *testing.T, selected nativeSession, before map[string]nativeObject) {
	t.Helper()
	after := snapshot(t, selected.base)
	want := make(map[string]nativeObject, len(before))
	for path, object := range before {
		want[path] = object
	}
	record, err := filepath.Rel(selected.base, selected.record)
	if err != nil {
		t.Fatal(err)
	}
	stage, err := filepath.Rel(selected.base, selected.stage())
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := before[stage]; exists {
		t.Fatal("positive fixture already had staging")
	}
	delete(want, record)
	delete(after, record)
	delete(after, stage)
	if !reflect.DeepEqual(after, want) {
		t.Fatal("unrelated registration/root/absence or selected companion native state changed")
	}
}

func TestGenuinelyMissingUnrelatedRootDoesNotBlockSelectedRecordSave(t *testing.T) {
	for _, identity := range []struct{ plan, account string }{{"plan-one", "private-account"}, {"private-other-plan", "newcomer"}} {
		t.Run(identity.plan+"/"+identity.account, func(t *testing.T) {
			selected := session(t, planBody, recordBody)
			other := registeredOtherContext(t, selected, identity.plan, identity.account, false)
			if err := os.Remove(other.root); err != nil {
				t.Fatal(err)
			}
			provedENOENT(t, other.root)
			before, held := snapshot(t, selected.base), nativeObjectAt(t, selected.record)
			clock := &clockFact{value: editEvent}
			prefix := confirm + "answer personal.choice text \"org-a\"\nsave\n"
			got := run(t, selected, clock, prefix+"UNREAD_TAIL\n")
			report(t, got, "edit.saved", 0, true)
			if clock.calls != 0 || got.consumed != len(prefix) {
				t.Fatal("unrelated absence changed explicit answer consent or event")
			}
			savedRecord(t, selected, held, ownedRecord)
			companions(t, selected, before)
			onlySelectedRecordChanged(t, selected, before)
			provedENOENT(t, other.root)
			otherNotDisclosed(t, other, got)
		})
	}
}

func TestDisjointUnrelatedParentsWithBothFinalFilesMissingAreNotNilAliases(t *testing.T) {
	for _, identity := range []struct{ plan, account string }{{"plan-one", "private-account"}, {"private-other-plan", "newcomer"}} {
		t.Run(identity.plan+"/"+identity.account, func(t *testing.T) {
			selected := session(t, planBody, recordBody)
			other := registeredOtherContext(t, selected, identity.plan, identity.account, true)
			provedENOENT(t, other.record)
			provedENOENT(t, other.preparation)
			before, held := snapshot(t, selected.base), nativeObjectAt(t, selected.record)
			clock := &clockFact{value: editEvent}
			got := run(t, selected, clock, confirm+"answer personal.choice text \"org-a\"\nsave\n")
			report(t, got, "edit.saved", 0, true)
			if clock.calls != 0 {
				t.Fatal("unrelated final absence sampled answer event")
			}
			savedRecord(t, selected, held, ownedRecord)
			companions(t, selected, before)
			onlySelectedRecordChanged(t, selected, before)
			provedENOENT(t, other.record)
			provedENOENT(t, other.preparation)
			otherNotDisclosed(t, other, got)
		})
	}
}

func TestUnsafeUnrelatedNativeNamespacesAreNotTreatedAsAbsence(t *testing.T) {
	for _, kind := range []string{"symlink-root", "actual-record-hardlink"} {
		t.Run(kind, func(t *testing.T) {
			selected := session(t, planBody, recordBody)
			other := registeredOtherContext(t, selected, "plan-one", "private-account", true)
			if kind == "symlink-root" {
				held := other.root + "-held"
				if err := os.Rename(other.root, held); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(held, other.root); err != nil {
					t.Fatal(err)
				}
				if otherInfo(t, other.root).Mode()&os.ModeSymlink == 0 {
					t.Fatal("actual unsafe root symlink was not created")
				}
				through, err := os.Stat(other.root)
				if err != nil || !os.SameFile(through, otherInfo(t, held)) {
					t.Fatal("symlink control does not reach the actual original root")
				}
			} else {
				if err := os.Link(selected.record, other.record); err != nil {
					t.Fatal(err)
				}
				if !os.SameFile(otherInfo(t, selected.record), otherInfo(t, other.record)) {
					t.Fatal("actual participant collision was not established")
				}
				provedENOENT(t, other.preparation)
			}
			before := snapshot(t, selected.base)
			clock := &clockFact{value: editEvent}
			got := run(t, selected, clock, confirm+"answer personal.choice text \"org-a\"\nsave\n")
			report(t, got, "edit.context", 1, false)
			if got.consumed != 0 || clock.calls != 0 {
				t.Fatal("unsafe unrelated namespace consumed selected consent or event")
			}
			unchanged(t, selected, before)
			otherNotDisclosed(t, other, got)
		})
	}
}

func TestSelectedAndMatchingMissingRootsStillRefuseBeforeIdentity(t *testing.T) {
	for _, kind := range []string{"selected", "matching-old"} {
		t.Run(kind, func(t *testing.T) {
			selected := session(t, planBody, recordBody)
			if kind == "selected" {
				if err := os.Rename(selected.root, filepath.Join(selected.base, "held-selected-root")); err != nil {
					t.Fatal(err)
				}
				provedENOENT(t, selected.root)
			} else {
				matching := registeredOtherContext(t, selected, "plan-one", "newcomer", false)
				if err := os.Remove(matching.root); err != nil {
					t.Fatal(err)
				}
				provedENOENT(t, matching.root)
			}
			before := snapshot(t, selected.base)
			clock := &clockFact{value: editEvent}
			got := run(t, selected, clock, confirm+"answer personal.choice text \"org-a\"\nsave\n")
			report(t, got, "edit.context", 1, false)
			if got.consumed != 0 || clock.calls != 0 {
				t.Fatal("missing selected/matching root consumed identity or event")
			}
			unchanged(t, selected, before)
		})
	}
}
