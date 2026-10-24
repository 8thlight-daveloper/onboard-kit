package starter_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

func contextInvitation(t *testing.T, account string, plan []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, entry := range []struct {
		name string
		body []byte
	}{
		{"plan.obk.yaml", plan},
		{"records/" + account + "/preparation.obk.yaml", []byte("schemaVersion: 1\nkind: preparation\naccount: " + account + "\nplanId: plan-a\n")},
		{"records/" + account + "/record.obk.yaml", []byte("schemaVersion: 1\nkind: record\naccount: " + account + "\nplan:\n  id: plan-a\n  hash: \"" + digest(plan) + "\"\n")},
		{"first-steps.md", fixture(t, "first-steps.md")},
	} {
		f, err := w.CreateHeader(&zip.FileHeader{Name: entry.name, Method: zip.Store})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(entry.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func contextMetadata(t *testing.T, s sandbox, account, planHash, token string) []byte {
	t.Helper()
	g := objectGuard(t, s.root)
	return jsonBytes(t, metadata{
		Version: 1, SessionKey: sessionKey("plan-a", account, s.root), PlanIDBase64: "cGxhbi1h", Account: account,
		Root: s.root, RootDevice: g.Device, RootInode: g.Inode,
		Record: "records/" + account + "/record.obk.yaml", Preparation: "records/" + account + "/preparation.obk.yaml",
		PlanSHA256: planHash, Transaction: token,
	})
}

func contextPreserved(t *testing.T, path string, before map[string]object, control guard) {
	t.Helper()
	if !reflect.DeepEqual(snapshot(t, path), before) || objectGuard(t, path) != control {
		t.Error("unrelated context bytes, native identities or controls changed")
	}
}

func TestContextGapExactAccountCaseKeepsOtherContextPrivate(t *testing.T) {
	for _, missingOldRoot := range []bool{false, true} {
		t.Run(fmt.Sprint(missingOldRoot), func(t *testing.T) {
			s := setup(t)
			initialOpening(t, s)
			oldRegistration := s.registration()
			oldRegistrationBefore := snapshot(t, oldRegistration)
			oldRegistrationGuard := objectGuard(t, oldRegistration)
			oldRoot := s.root
			if missingOldRoot {
				oldRoot += "-held"
				if err := os.Rename(s.root, oldRoot); err != nil {
					t.Fatal(err)
				}
			}
			oldRootBefore := snapshot(t, oldRoot)
			oldRootGuard := objectGuard(t, oldRoot)
			second := s
			second.root = filepath.Join(s.base, "other-root")
			second.invitation = filepath.Join(s.base, "other-invitation.zip")
			if err := os.Mkdir(second.root, 0700); err != nil {
				t.Fatal(err)
			}
			archive := contextInvitation(t, "Jules", fixture(t, "plan.yaml"))
			put(t, second.invitation, archive)
			clock := 0
			facts := runtimeFacts(t, &clock)
			facts.Clock = func() model.Timestamp { clock++; return "2026-10-24T09:00:00Z" }
			got := invoke(t, newStarter(facts), second.args(), "confirm Jules\nbegin\nexit\n")
			wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
			if clock != 1 || strings.Count(got.out, "Other matching contexts: 0\n") != 1 {
				t.Error("case-distinct account was treated as matching or did not create its own event")
			}
			for _, forbidden := range []string{s.root, "Other matching contexts: 1\n", "This opening registers an independent local context;", "Begin creates an independent record and session here;"} {
				if strings.Contains(got.out, forbidden) {
					t.Errorf("case-distinct account disclosed another context: %q", forbidden)
				}
			}
			registration := filepath.Join(second.home, "sessions", sessionKey("plan-a", "Jules", second.root))
			assertBytes(t, filepath.Join(registration, "plan.obk.yaml"), fixture(t, "plan.yaml"))
			assertBytes(t, filepath.Join(registration, "session.json"), contextMetadata(t, second, "Jules", rawPlanHash, got.report.Transaction))
			assertBytes(t, filepath.Join(second.root, "records/Jules/preparation.obk.yaml"), []byte("schemaVersion: 1\nkind: preparation\naccount: Jules\nplanId: plan-a\n"))
			wantRecord := strings.Replace(string(fixture(t, "begun.yaml")), "account: jules", "account: Jules", 1)
			wantRecord = strings.Replace(wantRecord, string(openedAt), "2026-10-24T09:00:00Z", 1)
			assertBytes(t, filepath.Join(second.root, "records/Jules/record.obk.yaml"), []byte(wantRecord))
			assertBytes(t, second.invitation, archive)
			contextPreserved(t, oldRegistration, oldRegistrationBefore, oldRegistrationGuard)
			contextPreserved(t, oldRoot, oldRootBefore, oldRootGuard)
			assertIdle(t, second)
		})
	}
}

func TestContextGapNativeCaseAliasKeepsRegisteredSpellingAndKey(t *testing.T) {
	s := setup(t)
	registeredRoot := filepath.Join(s.base, "NativeCaseRoot")
	if err := os.Rename(s.root, registeredRoot); err != nil {
		t.Fatal(err)
	}
	s.root = registeredRoot
	aliasRoot := filepath.Join(s.base, "nativecaseroot")
	aliasInfo, err := os.Lstat(aliasRoot)
	if os.IsNotExist(err) {
		t.Skip("native case alias unavailable on this case-sensitive volume; symlinks are not a substitute")
	}
	if err != nil {
		t.Fatal(err)
	}
	registeredInfo, err := os.Lstat(registeredRoot)
	if err != nil {
		t.Fatal(err)
	}
	registeredGuard, aliasGuard := objectGuard(t, registeredRoot), objectGuard(t, aliasRoot)
	if aliasInfo.Mode()&os.ModeSymlink != 0 || !aliasInfo.IsDir() || !os.SameFile(registeredInfo, aliasInfo) || registeredGuard.Device != aliasGuard.Device || registeredGuard.Inode != aliasGuard.Inode {
		t.Fatal("differently cased native paths do not name the exact same directory device/inode")
	}
	if key(registeredRoot) == key(aliasRoot) {
		t.Fatal("fixture does not distinguish the two length-framed root spellings")
	}
	initialOpening(t, s)
	before := snapshot(t, s.base)
	metadataGuard := objectGuard(t, filepath.Join(s.registration(), "session.json"))
	args := s.args()
	args[3] = aliasRoot
	clock := 0
	got := invoke(t, newStarter(runtimeFacts(t, &clock)), args, "confirm jules\nbegin\nexit\n")
	wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	if clock != 0 {
		t.Error("native alias opening sampled a new event")
	}
	for _, line := range []string{
		"Role: \"local-plan\" Action: \"retain\" Path: \"" + filepath.Join(s.registration(), "plan.obk.yaml") + "\"\n",
		"Role: \"session-metadata\" Action: \"retain\" Path: \"" + filepath.Join(s.registration(), "session.json") + "\"\n",
		"Role: \"record\" Action: \"retain\" Path: \"" + s.record() + "\"\n",
		"Other matching contexts: 0\n",
		"First opened at: \"" + string(openedAt) + "\"\n",
	} {
		if !strings.Contains(got.out, line) {
			t.Errorf("native alias did not retain registered context fact %q", line)
		}
	}
	unchanged(t, s, before)
	if objectGuard(t, filepath.Join(s.registration(), "session.json")) != metadataGuard {
		t.Error("equivalent metadata controls changed through a native alias")
	}
}

const contextHistoryRecord = `schemaVersion: 1
kind: record
account: jules
startDate: "2026-10-19"
firstOpenedAt: "2026-10-20T08:00:00.123456789012+00:00"
plan:
  id: plan-a
  hash: "00d6473bd698e40caca271f61af7b87288f0c864924008bd24375d6b19b3c6a4"
answers:
  sections:
    removed:
      hidden: deferred
    setup:
      account_name: local-person
outcomes:
  - section: setup
    outcome: failed
    on: "2026-10-22T10:00:00.123456789012+00:00"
    sectionHash: legacy
    note: Held note
    differently: Try again
  - section: absent
    outcome: deferred
    on: "2026-10-21T00:00:00Z"
    sectionHash: gone
    note: Hidden activity history
notes: preserve me
`

const contextHistoryDisplay = `Recorded outcomes: 2
Recorded outcome: 1
Section: "absent"
Outcome: "deferred"
On: "2026-10-21T00:00:00Z"
Recorded definition hash: "gone"
Current definition hash: ""
Comparison: "unknown"
Note:
  | "Hidden activity history"
Differently:
  | ""
Recorded outcome: 2
Section: "setup"
Outcome: "failed"
On: "2026-10-22T10:00:00.123456789012+00:00"
Recorded definition hash: "legacy"
Current definition hash: "5631a3d34e963b37019b9ae53b12907c3bdef3d17c6642f1462784d178c3f4ad"
Comparison: "different"
Note:
  | "Held note"
Differently:
  | "Try again"
Support available: true
`

func TestContextGapChangedIncomingHashIsExplicitlyReviewedAndRegistered(t *testing.T) {
	s := setup(t)
	put(t, s.record(), []byte(contextHistoryRecord))
	clock := 0
	first := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\nexit\n")
	wantReport(t, first, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	if clock != 0 || first.report.Transaction == "" {
		t.Fatal("registering an existing record must retain its event")
	}
	assertBytes(t, s.record(), []byte(contextHistoryRecord))
	assertBytes(t, filepath.Join(s.registration(), "session.json"), contextMetadata(t, s, "jules", rawPlanHash, first.report.Transaction))
	plan := append(append([]byte(nil), fixture(t, "plan.yaml")...), []byte("# Reviewed incoming context B.\n")...)
	newHash := digest(plan)
	if newHash == rawPlanHash {
		t.Fatal("literal incoming plan change did not change its raw-byte hash")
	}
	archive := contextInvitation(t, "jules", plan)
	put(t, s.invitation, archive)
	before := snapshot(t, s.base)
	for _, input := range []string{"confirm jules\ncancel\n", "confirm jules\nuse-incoming\ncancel\n"} {
		got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), input)
		wantReport(t, got, 0, "cancelled", "start.cancelled", "Opening cancelled before begin; no files were changed.")
		if !strings.Contains(got.out, "Local plan bytes differ from the invitation. No recency or merge is assumed.\nPlan copy: choose use-incoming or cancel:\n") || clock != 0 {
			t.Error("changed invitation did not reach explicit plan choice without sampling time")
		}
		if strings.Contains(input, "use-incoming") && !strings.Contains(got.out, "Review complete. Choose begin or cancel:\n") {
			t.Error("explicit incoming choice did not reach complete review")
		}
		unchanged(t, s, before)
	}
	prepBefore := snapshot(t, s.root)["records/jules/preparation.obk.yaml"]
	prepGuard := objectGuard(t, s.prep())
	recordGuard := objectGuard(t, s.record())
	got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nuse-incoming\nbegin\nexit\n")
	wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	if clock != 0 || got.report.Transaction == "" || got.report.Transaction == first.report.Transaction {
		t.Error("registration change resampled history or lacked its own transaction")
	}
	for _, line := range []string{
		"Role: \"local-plan\" Action: \"replace\" Path: \"" + filepath.Join(s.registration(), "plan.obk.yaml") + "\"\n",
		"Role: \"session-metadata\" Action: \"replace\" Path: \"" + filepath.Join(s.registration(), "session.json") + "\"\n",
		"Role: \"record\" Action: \"reconcile\" Path: \"" + s.record() + "\"\n",
		"Previous comparison: \"available\"\n",
		"New questions: 0\nRemoved questions: 0\nAnswers to review: 0\nNew activities: 0\nRemoved activities: 0\nDifferent recorded definitions: 1\n  - \"setup\"\n",
		"First use: the existing event is preserved; an empty event stays unknown.\n",
		"First opened at: \"2026-10-20T08:00:00.123456789012+00:00\"\n",
		contextHistoryDisplay,
	} {
		if !strings.Contains(got.out, line) {
			t.Errorf("changed invitation lost exact review/history fact %q", line)
		}
	}
	assertBytes(t, filepath.Join(s.registration(), "plan.obk.yaml"), plan)
	assertBytes(t, filepath.Join(s.registration(), "session.json"), contextMetadata(t, s, "jules", newHash, got.report.Transaction))
	hashScalar := newHash
	if newHash[0] >= '0' && newHash[0] <= '9' {
		hashScalar = `"` + newHash + `"`
	}
	wantRecord := strings.Replace(contextHistoryRecord, `"`+rawPlanHash+`"`, hashScalar, 1)
	assertBytes(t, s.record(), []byte(wantRecord))
	afterRecordGuard := objectGuard(t, s.record())
	if recordGuard.Security != afterRecordGuard.Security || recordGuard.Device != afterRecordGuard.Device || recordGuard.Inode == afterRecordGuard.Inode {
		t.Error("changed record did not preserve controls through atomic pathname replacement")
	}
	if snapshot(t, s.root)["records/jules/preparation.obk.yaml"] != prepBefore || objectGuard(t, s.prep()) != prepGuard {
		t.Error("retained preparation bytes, native identity or controls changed")
	}
	assertBytes(t, s.invitation, archive)
	assertIdle(t, s)
}

func TestContextGapNestedHomePermitsDisjointParticipantsAndRefusesActualAlias(t *testing.T) {
	for _, collision := range []bool{false, true} {
		t.Run(fmt.Sprint(collision), func(t *testing.T) {
			s := setup(t)
			s.home = filepath.Join(s.root, "state")
			if collision {
				put(t, s.record(), fixture(t, "seed.yaml"))
				if err := os.Link(s.record(), s.prep()); err != nil {
					t.Fatal(err)
				}
				a, b := objectGuard(t, s.record()), objectGuard(t, s.prep())
				if a.Device != b.Device || a.Inode != b.Inode {
					t.Fatal("participant hard-link collision was not established")
				}
			}
			before := snapshot(t, s.base)
			clock := 0
			got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\nexit\n")
			if collision {
				wantReport(t, got, 1, "refused", "start.context", "Select a safe supported root, state home, and unchanged session context.")
				if clock != 0 {
					t.Error("colliding participants sampled a first event")
				}
				unchanged(t, s, before)
				return
			}
			wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
			if clock != 1 || !strings.Contains(got.out, "State home: \""+s.home+"\"\n") {
				t.Error("disjoint participants under nested root/home did not begin normally")
			}
			assertBytes(t, s.record(), fixture(t, "begun.yaml"))
			assertBytes(t, s.prep(), fixture(t, "preparation.yaml"))
			assertBytes(t, filepath.Join(s.registration(), "plan.obk.yaml"), fixture(t, "plan.yaml"))
			assertBytes(t, filepath.Join(s.registration(), "session.json"), contextMetadata(t, s, "jules", rawPlanHash, got.report.Transaction))
			assertBytes(t, s.invitation, zipBytes(t, nil))
			for _, path := range []string{s.home, s.pending(), s.registration(), filepath.Dir(s.record())} {
				assertNewPrivate(t, path, "0700")
			}
			for _, path := range []string{s.record(), s.prep(), filepath.Join(s.registration(), "plan.obk.yaml"), filepath.Join(s.registration(), "session.json")} {
				assertNewPrivate(t, path, "0600")
			}
			assertIdle(t, s)
		})
	}
}
