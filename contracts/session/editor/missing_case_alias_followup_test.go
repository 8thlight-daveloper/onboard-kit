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
	"strconv"
	"strings"
	"syscall"
	"testing"
)

func missingAccountContexts(t *testing.T, firstAccount, secondAccount string, aliases bool) (nativeSession, otherContext, otherContext) {
	t.Helper()
	selected := session(t, planBody, recordBody)
	first, second := addMissingAccountContexts(t, selected, firstAccount, secondAccount, aliases)
	return selected, first, second
}

func addMissingAccountContexts(t *testing.T, selected nativeSession, firstAccount, secondAccount string, aliases bool) (otherContext, otherContext) {
	t.Helper()
	first := registeredOtherContext(t, selected, "private-other-plan", firstAccount, true)
	firstParent := filepath.Dir(first.record)
	secondParent := filepath.Join(first.root, "records", secondAccount)
	firstInfo := otherInfo(t, firstParent)
	secondInfo, err := os.Lstat(secondParent)
	if aliases && errors.Is(err, syscall.ENOENT) {
		t.Skip("missing-account-case-alias profile unavailable: opposite-case account path does not name the created native directory")
	}
	if !aliases {
		if !errors.Is(err, syscall.ENOENT) {
			t.Fatalf("distinct account fixture did not start with its own absent second directory: %v", err)
		}
		if err := os.Mkdir(secondParent, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.Chown(secondParent, -1, os.Getegid()); err != nil {
			t.Fatal(err)
		}
		secondInfo, err = os.Lstat(secondParent)
	}
	if err != nil {
		t.Fatalf("cannot inspect actual second account path: %v", err)
	}
	if !firstInfo.IsDir() || !secondInfo.IsDir() || firstInfo.Mode()&os.ModeSymlink != 0 || secondInfo.Mode()&os.ModeSymlink != 0 || os.SameFile(firstInfo, secondInfo) != aliases {
		t.Fatal("actual nonsymlink account-directory alias relation differs from the requested fixture profile")
	}
	firstObject := nativeObjectAt(t, firstParent)
	secondObject := nativeObjectAt(t, secondParent)
	if (firstObject.Device == secondObject.Device && firstObject.Inode == secondObject.Inode) != aliases {
		t.Fatal("actual account device/inode relation was not independently established")
	}
	for _, path := range []string{firstParent, secondParent} {
		object := nativeObjectAt(t, path)
		if object.UID != uint64(os.Geteuid()) || object.GID != uint64(os.Getegid()) || object.Mode.Perm() != 0700 || object.ACL != "absent" {
			t.Fatal("actual account directory is not an owned private native fixture")
		}
		entries, err := os.ReadDir(path)
		if err != nil || len(entries) != 0 {
			t.Fatal("the test-owned account directory is not actually empty")
		}
	}
	second := registeredSharedRootContext(t, selected, first, secondAccount)
	if !os.SameFile(firstInfo, otherInfo(t, firstParent)) || !os.SameFile(secondInfo, otherInfo(t, secondParent)) {
		t.Fatal("actual account directories changed before the owned empty-directory removal")
	}
	if err := os.Remove(firstParent); err != nil {
		t.Fatal(err)
	}
	if !aliases {
		if err := os.Remove(secondParent); err != nil {
			t.Fatal(err)
		}
	}
	missingAccountParticipants(t, first, second)
	return first, second
}

func registeredSharedRootContext(t *testing.T, selected nativeSession, first otherContext, account string) otherContext {
	t.Helper()
	second := otherContext{root: first.root, planID: first.planID, account: account}
	second.key = sessionKey(second.planID, second.account, second.root)
	if second.key == first.key || second.key == selected.key {
		t.Fatal("two different exact identities did not produce distinct registry keys")
	}
	registryDirectory := filepath.Join(selected.home, "sessions", second.key)
	second.plan = filepath.Join(registryDirectory, "plan.obk.yaml")
	second.metadata = filepath.Join(registryDirectory, "session.json")
	recordRelative := "records/" + account + "/record.obk.yaml"
	preparationRelative := "records/" + account + "/preparation.obk.yaml"
	second.record = filepath.Join(second.root, filepath.FromSlash(recordRelative))
	second.preparation = filepath.Join(second.root, filepath.FromSlash(preparationRelative))
	body := strings.Replace(planBody, "id: plan-one", "id: "+second.planID, 1) + "# PRIVATE_UNRELATED_PLAN_BODY\n"
	digest := sha256.Sum256([]byte(body))
	root := nativeObjectAt(t, second.root)
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
	}{1, second.key, base64.StdEncoding.EncodeToString([]byte(second.planID)), second.account, second.root, strconv.FormatUint(root.Device, 10), strconv.FormatUint(root.Inode, 10), recordRelative, preparationRelative, hex.EncodeToString(digest[:]), strings.Repeat("b", 64)}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	put(t, second.plan, body)
	put(t, second.metadata, string(append(encoded, '\n')))
	if err := os.Chown(registryDirectory, -1, os.Getegid()); err != nil {
		t.Fatal(err)
	}
	if nativeACL(t, registryDirectory) != "absent" || os.SameFile(otherInfo(t, registryDirectory), otherInfo(t, filepath.Dir(first.metadata))) {
		t.Fatal("the second strict registration is not a distinct private native namespace")
	}
	return second
}

func missingAccountParticipants(t *testing.T, first, second otherContext) {
	t.Helper()
	for _, path := range []string{filepath.Dir(first.record), filepath.Dir(second.record), first.record, first.preparation, second.record, second.preparation} {
		provedENOENT(t, path)
	}
}

func missingAccountAmbiguityRefuses(t *testing.T, selected nativeSession, first, second otherContext) {
	t.Helper()
	before := snapshot(t, selected.base)
	clock := &clockFact{value: editEvent}
	got := run(t, selected, clock, confirm+"answer personal.choice text \"org-a\"\nsave\nUNREAD_TAIL\n")
	report(t, got, "edit.context", 1, false)
	if got.consumed != 0 || clock.calls != 0 || got.out != "" {
		t.Fatal("ambiguous missing native destinations disclosed identity, consumed input or sampled an event")
	}
	unchanged(t, selected, before)
	missingAccountParticipants(t, first, second)
	otherNotDisclosed(t, first, got)
	otherNotDisclosed(t, second, got)
}

func TestMissingUnrelatedCaseAliasedDestinationsRefuseBeforeIdentity(t *testing.T) {
	selected, first, second := missingAccountContexts(t, "Jules", "jules", true)
	missingAccountAmbiguityRefuses(t, selected, first, second)
}

func TestMissingUnrelatedCaseAliasedDestinationsRefuseWithReverseRegistrationOrder(t *testing.T) {
	selected, first, second := missingAccountContexts(t, "jules", "Jules", true)
	missingAccountAmbiguityRefuses(t, selected, first, second)
}

func TestDistinctMissingAccountsUnderOneNativeParentAllowSelectedSave(t *testing.T) {
	selected, first, second := missingAccountContexts(t, "first-person", "second-person", false)
	before, held := snapshot(t, selected.base), nativeObjectAt(t, selected.record)
	clock := &clockFact{value: editEvent}
	prefix := confirm + "answer personal.choice text \"org-a\"\nsave\n"
	got := run(t, selected, clock, prefix+"UNREAD_TAIL\n")
	report(t, got, "edit.saved", 0, true)
	if got.consumed != len(prefix) || clock.calls != 0 {
		t.Fatal("distinct absent accounts changed selected consent or sampled an answer event")
	}
	savedRecord(t, selected, held, ownedRecord)
	companions(t, selected, before)
	onlySelectedRecordChanged(t, selected, before)
	missingAccountParticipants(t, first, second)
	otherNotDisclosed(t, first, got)
	otherNotDisclosed(t, second, got)
}
