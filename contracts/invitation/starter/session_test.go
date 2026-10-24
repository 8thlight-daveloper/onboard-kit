package starter_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func seedFor(plan []byte) []byte {
	return []byte(fmt.Sprintf("schemaVersion: 1\nkind: record\naccount: jules\nplan:\n  id: plan-a\n  hash: \"%s\"\n", digest(plan)))
}

func TestPurposefulFullSessionAndSupportRemainReadOnly(t *testing.T) {
	s := setup(t)
	sentinel := filepath.Join(s.base, "instruction-must-not-run")
	plan := []byte(strings.NewReplacer("__OWNED_DIRECTORY__", s.root, "__OWNED_SENTINEL__", sentinel).Replace(string(fixture(t, "full.plan.yaml"))))
	prep := fixture(t, "full.preparation.yaml")
	put(t, s.invitation, zipBytes(t, map[string][]byte{"plan.obk.yaml": plan, "records/jules/record.obk.yaml": seedFor(plan), "records/jules/preparation.obk.yaml": prep}))
	clock := 0
	got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\nexit\n")
	wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	if clock != 1 {
		t.Error("new full session event count differs")
	}
	for _, literal := range []string{
		"Activities: 2\n", "ID: \"follow-up\"\n", "Waiting on: 1\n  - \"setup\"\n", "Timing: \"before-day-one\"\n",
		"Steps: 1\nStep: 1\nID: \"observe\"\nName: \"Only display\"\n", "Open: \"https://company.example/instructions\"\n",
		"Notes:\n  | \"line one\"\n  | \"line two\"\n", "Environment: 2\n  - Key: \"A_MODE\" Value: \"local\"\n  - Key: \"Z_MODE\" Value: \"safe\"\n", "Checks: 3\n", "Environment question: \"NOT_READ_BY_STARTER\"\n",
		"Body:\n  | \"Café \\u001b[2J \\u202e \\u200d\"\n  | \"Do not execute these instructions.\"\n",
		"Default suggestion: text \"blue\"\n", "Default suggestion: list 2\n  - \"echo\"\n  - \"hello\"\n", "Options: 2\n  - \"blue\"\n  - \"green\"\n", "Answer: none\n",
		"Name: \" Jules \"\nCompany email: \"jules@company.example\"\nPlanned start date (goal): \"2026-11-02\"\n", "Team ID: \"platform\"\nManager contact ID: \"helper\"\nOnboarding owner contact ID: \"helper\"\nBuddy contact ID: \"helper\"\nWork location: \"remote\"\nTime zone: \"Europe/Paris\"\n",
		"Arrangements: 1\nArrangement: 1\nID: \"pairing\"\nSection: \"setup\"\nKind: \"support\"\nOwner contact ID: \"helper\"\nReference: \"Pair with the assigned teammate\"\nTarget date (goal): \"2026-11-03\"\nStatus: \"arranged\"\nHelper contact IDs: 2\n  - \"helper\"\n  - \"helper\"\nNote:\n  | \"line one\"\n  | \"line two\"\n",
	} {
		if !strings.Contains(got.out, literal) {
			t.Errorf("purposeful session field missing: %.70s", literal)
		}
	}
	if strings.ContainsAny(got.out, "\x1b\u202e\u200d") || strings.Contains(got.out, "NEVER_DISPLAY_THIS_UNKNOWN_PERSONAL_NOTE") {
		t.Error("terminal controls or private unknown field leaked")
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Error("display executed a command/check")
	}
	want := strings.Replace(string(fixture(t, "begun.yaml")), rawPlanHash, digest(plan), 1)
	assertBytes(t, s.record(), []byte(want))
	assertBytes(t, s.prep(), prep)
	assertIdle(t, s)
}

func TestRetainedStalePreparationIsNotReplacedOrCalledCurrentSupport(t *testing.T) {
	s := setup(t)
	stale := fixture(t, "stale.preparation.yaml")
	put(t, s.prep(), stale)
	before := snapshot(t, s.root)
	clock := 0
	got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\nexit\n")
	wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	assertBytes(t, s.prep(), stale)
	if snapshot(t, s.root)["records/jules/preparation.obk.yaml"] != before["records/jules/preparation.obk.yaml"] {
		t.Error("stale preparation inode/controls changed")
	}
	if !strings.Contains(got.out, "Support available: false\n") || !strings.Contains(got.out, "Current support is unavailable; ask the onboarding owner to repair preparation.\n") || !strings.Contains(got.out, "Role: \"preparation\" Action: \"retain\"") {
		t.Error("stale support falsely replaced/projected")
	}
	if strings.Contains(got.out, "NEVER_DISPLAY_THIS_STALE_UNKNOWN_PERSONAL_NOTE") {
		t.Error("unknown stale personal field leaked")
	}
}

func TestDifferentLocalPlanNeedsSeparateExplicitChoice(t *testing.T) {
	for _, input := range []string{"confirm jules\ncancel\n", "confirm jules\nuse-incoming\ncancel\n", "confirm jules\nbegin\ncancel\n"} {
		s := setup(t)
		initialOpening(t, s)
		put(t, filepath.Join(s.registration(), "plan.obk.yaml"), []byte("malformed: [\n"))
		before := snapshot(t, s.base)
		clock := 0
		got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), input)
		wantReport(t, got, 0, "cancelled", "start.cancelled", "Opening cancelled before begin; no files were changed.")
		if !strings.Contains(got.out, "Local plan bytes differ from the invitation. No recency or merge is assumed.\nPlan copy: choose use-incoming or cancel:\n") {
			t.Error("replacement decision was omitted")
		}
		if strings.Contains(input, "use-incoming") && (!strings.Contains(got.out, "Previous comparison: \"decode\"\n") || !strings.Contains(got.out, "Previous comparison problems: ")) {
			t.Error("unavailable previous comparison was hidden")
		}
		if clock != 0 {
			t.Error("choice sampled event")
		}
		unchanged(t, s, before)
	}
}

type onceReportFailure struct {
	bytes.Buffer
	failed bool
}

func (w *onceReportFailure) Write(p []byte) (int, error) {
	n, err := w.Buffer.Write(p)
	if !w.failed && bytes.Contains(w.Bytes(), []byte(`"code":"start.opened"`)) {
		w.failed = true
		return n, errors.New("invented report failure after full accepted count")
	}
	return n, err
}

func TestFinalReportFallbackPreservesCompletionAndToken(t *testing.T) {
	s := setup(t)
	clock := 0
	stderr := &onceReportFailure{}
	exit := newStarter(runtimeFacts(t, &clock)).Run(s.args(), strings.NewReader("confirm jules\nbegin\nexit\n"), io.Discard, stderr)
	if exit != 1 || clock != 1 || !stderr.failed {
		t.Fatal("report writer failure not reached")
	}
	lines := strings.Split(strings.TrimSuffix(stderr.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatal("exactly one fallback attempt required")
	}
	var original, fallback report
	if err := json.Unmarshal([]byte(lines[0]), &original); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &fallback); err != nil {
		t.Fatal(err)
	}
	if original.Code != "start.opened" || fallback.Status != "opened" || fallback.Code != "start.report-opened" || fallback.Message != "Opening is complete, but its final report could not be written." || original.Transaction != fallback.Transaction {
		t.Error("fallback completion/token truth differs")
	}
	assertBytes(t, s.record(), fixture(t, "begun.yaml"))
	assertIdle(t, s)
}
