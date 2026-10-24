package starter_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type namespaceGapInput struct {
	source *strings.Reader
	reads  int
}

func (r *namespaceGapInput) Read(p []byte) (int, error) {
	r.reads++
	return r.source.Read(p)
}

func namespaceGapRegistration(t *testing.T, s sandbox, account string) (string, string) {
	t.Helper()
	parent := filepath.Join(s.root, "records", account)
	if err := os.MkdirAll(parent, 0700); err != nil {
		t.Fatal(err)
	}
	registration := filepath.Join(s.home, "sessions", sessionKey("plan-a", account, s.root))
	put(t, filepath.Join(registration, "plan.obk.yaml"), fixture(t, "plan.yaml"))
	put(t, filepath.Join(registration, "session.json"), contextMetadata(t, s, account, rawPlanHash, recoveryToken))
	for _, name := range []string{"record.obk.yaml", "preparation.obk.yaml"} {
		if _, err := os.Lstat(filepath.Join(parent, name)); !os.IsNotExist(err) {
			t.Fatalf("old registered final participant must be absent: %v", err)
		}
	}
	return parent, registration
}

func TestNativeAccountParentAliasRefusesWithBothRegisteredFinalFilesMissing(t *testing.T) {
	s := setup(t)
	oldParent, _ := namespaceGapRegistration(t, s, "Jules")
	newParent := filepath.Join(s.root, "records", "jules")
	oldInfo, err := os.Lstat(oldParent)
	if err != nil {
		t.Fatal(err)
	}
	newInfo, err := os.Lstat(newParent)
	if os.IsNotExist(err) {
		t.Skip("native Jules/jules account-parent alias unavailable on this case-sensitive volume; missing-final-files collision remains a named gap")
	}
	if err != nil {
		t.Fatal(err)
	}
	oldGuard, newGuard := objectGuard(t, oldParent), objectGuard(t, newParent)
	if !oldInfo.IsDir() || !newInfo.IsDir() || oldInfo.Mode()&os.ModeSymlink != 0 || newInfo.Mode()&os.ModeSymlink != 0 || !os.SameFile(oldInfo, newInfo) || oldGuard.Device != newGuard.Device || oldGuard.Inode != newGuard.Inode {
		t.Fatal("case-distinct participant parents must be actual native directory aliases with equal device/inode")
	}
	if sessionKey("plan-a", "Jules", s.root) == sessionKey("plan-a", "jules", s.root) {
		t.Fatal("case-distinct accounts must have independently distinct registration keys")
	}
	for _, name := range []string{"record.obk.yaml", "preparation.obk.yaml"} {
		if _, err := os.Lstat(filepath.Join(newParent, name)); !os.IsNotExist(err) {
			t.Fatalf("incoming final participant must also be absent through its case alias: %v", err)
		}
	}
	before := snapshot(t, s.base)
	clock := 0
	input := &namespaceGapInput{source: strings.NewReader("confirm jules\nbegin\nexit\nUNUSED\n")}
	remaining := input.source.Len()
	var stdout, stderr bytes.Buffer
	exit := newStarter(runtimeFacts(t, &clock)).Run(s.args(), input, &stdout, &stderr)
	got := streamResult(t, exit, stdout.String(), stderr.Bytes())
	wantReport(t, got, 1, "refused", "start.context", "Select a safe supported root, state home, and unchanged session context.")
	if got.report.Transaction != "" || stdout.Len() != 0 {
		t.Error("native different-account namespace collision must refuse before identity/review or a transaction token")
	}
	if clock != 0 || input.reads != 0 || input.source.Len() != remaining {
		t.Errorf("native collision must consume no consent or clock: clock=%d reads=%d remaining=%d", clock, input.reads, input.source.Len())
	}
	unchanged(t, s, before)
	if objectGuard(t, oldParent) != oldGuard || objectGuard(t, newParent) != newGuard {
		t.Error("native account-parent controls changed during refusal")
	}
}

func TestDistinctNativeAccountParentsPermitAnIndependentPrivateOpening(t *testing.T) {
	for _, account := range []string{"Jules", "other-person"} {
		t.Run(account, func(t *testing.T) {
			s := setup(t)
			oldParent, oldRegistration := namespaceGapRegistration(t, s, account)
			newParent := filepath.Join(s.root, "records", "jules")
			if info, err := os.Lstat(newParent); err == nil {
				oldInfo, statErr := os.Lstat(oldParent)
				if statErr != nil {
					t.Fatal(statErr)
				}
				if account == "Jules" && os.SameFile(oldInfo, info) {
					t.Skip("case-distinct disjoint-directory countercontrol unavailable on this case-insensitive volume; other-person countercontrol remains independent")
				}
				t.Fatal("fresh incoming account parent unexpectedly exists")
			} else if !os.IsNotExist(err) {
				t.Fatal(err)
			}
			if err := os.Mkdir(newParent, 0700); err != nil {
				t.Fatal(err)
			}
			oldInfo, err := os.Lstat(oldParent)
			if err != nil {
				t.Fatal(err)
			}
			newInfo, err := os.Lstat(newParent)
			if err != nil {
				t.Fatal(err)
			}
			oldGuard, newGuard := objectGuard(t, oldParent), objectGuard(t, newParent)
			if !oldInfo.IsDir() || !newInfo.IsDir() || oldInfo.Mode()&os.ModeSymlink != 0 || newInfo.Mode()&os.ModeSymlink != 0 || os.SameFile(oldInfo, newInfo) || oldGuard.Device == newGuard.Device && oldGuard.Inode == newGuard.Inode {
				t.Fatal("countercontrol must prove distinct actual account-parent directories")
			}
			oldParentBefore := snapshot(t, oldParent)
			oldRegistrationBefore := snapshot(t, oldRegistration)
			oldRegistrationGuard := objectGuard(t, oldRegistration)
			archiveBefore := snapshot(t, s.base)["invitation.zip"]
			clock := 0
			got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\nexit\n")
			wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
			if clock != 1 || strings.Count(got.out, "Other matching contexts: 0\n") != 1 {
				t.Error("disjoint different-account context must remain independent and create only its own first-use event")
			}
			for _, forbidden := range []string{oldParent, oldRegistration, "Account: \"" + account + "\"\n", "Other matching contexts: 1\n", "This opening registers an independent local context;", "Begin creates an independent record and session here;"} {
				if strings.Contains(got.out, forbidden) {
					t.Errorf("different-account metadata leaked into the incoming review: %q", forbidden)
				}
			}
			assertBytes(t, s.record(), fixture(t, "begun.yaml"))
			assertBytes(t, s.prep(), fixture(t, "preparation.yaml"))
			assertBytes(t, filepath.Join(s.registration(), "plan.obk.yaml"), fixture(t, "plan.yaml"))
			assertBytes(t, filepath.Join(s.registration(), "session.json"), contextMetadata(t, s, "jules", rawPlanHash, got.report.Transaction))
			contextPreserved(t, oldParent, oldParentBefore, oldGuard)
			contextPreserved(t, oldRegistration, oldRegistrationBefore, oldRegistrationGuard)
			if snapshot(t, s.base)["invitation.zip"] != archiveBefore {
				t.Error("independent incoming opening changed its archive input")
			}
			for _, name := range []string{"record.obk.yaml", "preparation.obk.yaml"} {
				if _, err := os.Lstat(filepath.Join(oldParent, name)); !os.IsNotExist(err) {
					t.Fatalf("old missing participant was recreated for a different account: %v", err)
				}
			}
			assertIdle(t, s)
		})
	}
}
