package starter_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func completedFrozenSession(t *testing.T, output string) string {
	t.Helper()
	opening := "First opened at: \"" + string(openedAt) + "\"\n"
	ending := "Session: read only. Choose exit:\n"
	if strings.Count(output, opening) != 1 || strings.Count(output, "Read-only onboarding session.\n") != 1 || !strings.HasSuffix(output, ending) {
		t.Fatal("one complete frozen session with the supplied event and final prompt required")
	}
	_, session, found := strings.Cut(output, opening)
	if !found || !strings.HasPrefix(session, "Read-only onboarding session.\n") {
		t.Fatal("session must immediately follow its frozen event")
	}
	return strings.TrimSuffix(session, ending)
}

func TestFrozenSessionPreservesCompleteStepAndChecks(t *testing.T) {
	s := setup(t)
	sentinel := filepath.Join(s.base, "display-must-not-execute")
	plan := strings.NewReplacer(
		"__OWNED_DIRECTORY__", `"work/\"quoted\"\tback\\slash\u001b\u202e"`,
		"__OWNED_SENTINEL__", sentinel,
	).Replace(string(fixture(t, "full.plan.yaml")))
	plan = strings.Replace(plan, "    steps:\n", "      - name: helper_ref\n        type: contact\n    steps:\n", 1)
	plan = strings.Replace(plan, "        directory: ", "        owner: Display owner\n        escalation: [$helper_ref, $helper_ref]\n        directory: ", 1)
	plan = strings.Replace(plan, "          - ask: Did the person choose to verify later?", `          - ask: "First question?\nSecond \"quoted\" question\twith back\\slash\u001b\u202e\n"`, 1)
	planBytes := []byte(plan)
	prep := fixture(t, "full.preparation.yaml")
	put(t, s.invitation, zipBytes(t, map[string][]byte{
		"plan.obk.yaml":                      planBytes,
		"records/jules/record.obk.yaml":      seedFor(planBytes),
		"records/jules/preparation.obk.yaml": prep,
	}))
	clock := 0
	got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\nexit\n")
	wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	if clock != 1 {
		t.Fatal("new complete session must sample exactly one event")
	}
	session := completedFrozenSession(t, got.out)
	_, steps, found := strings.Cut(session, "Steps: 1\n")
	if !found || strings.Count(session, "Steps: 1\n") != 1 {
		t.Fatal("one held step required in the completed session")
	}
	steps, _, found = strings.Cut(steps, "Activity: 2\n")
	if !found || strings.Count(session, "Questions: 3\n") != 1 {
		t.Fatal("step/check block must end before the next activity and retain all three questions")
	}
	_, questions, found := strings.Cut(session, "Questions: 3\n")
	if !found {
		t.Fatal("full-plan question list missing")
	}
	questions, _, found = strings.Cut(questions, "Blockers: ")
	if !found || strings.Count(questions, "Question: ") != 3 {
		t.Fatal("three questions must precede the blocker list")
	}
	for _, literal := range []string{
		"Question: 1\nID: \"setup.flavor\"\nScope: \"setup\"\nName: \"flavor\"\nDescription: \"Which flavor?\"\nExample: \"\"\nType: \"choice\"\nAudience: \"onboardee\"\n",
		"Question: 2\nID: \"setup.tools\"\nScope: \"setup\"\nName: \"tools\"\nDescription: \"Which tools?\"\nExample: \"\"\nType: \"command\"\nAudience: \"onboardee\"\n",
		"Question: 3\nID: \"setup.helper_ref\"\nScope: \"setup\"\nName: \"helper_ref\"\nDescription: \"\"\nExample: \"\"\nType: \"contact\"\nAudience: \"onboarder\"\n",
	} {
		_, remaining, present := strings.Cut(questions, literal)
		if !present {
			t.Fatal("full-plan question identity, metadata, audience or declaration order differs")
		}
		questions = remaining
	}
	want := `Step: 1
ID: "observe"
Name: "Only display"
Owner: "Display owner"
Timing: "days-8-to-30"
Open: "https://company.example/instructions"
Directory: "work/\"quoted\"\tback\\slash\u001b\u202e"
Notes:
  | "line one"
  | "line two"
Run: 2
  - "touch"
  - @SENTINEL@
Escalation: 2
  - "$helper_ref"
  - "$helper_ref"
Environment: 2
  - Key: "A_MODE" Value: "local"
  - Key: "Z_MODE" Value: "safe"
Checks: 3
Check: 1
Run: 2
  - "touch"
  - @SENTINEL@
Ask:
  | ""
Environment question: ""
Check: 2
Run: 0
Ask:
  | "First question?"
  | "Second \"quoted\" question\twith back\\slash\u001b\u202e"
  | ""
Environment question: ""
Check: 3
Run: 0
Ask:
  | ""
Environment question: "NOT_READ_BY_STARTER"
`
	want = strings.ReplaceAll(want, "@SENTINEL@", fmt.Sprintf("%q", sentinel))
	if steps != want {
		t.Errorf("complete literal step/check projection differs\nwant:\n%s\ngot:\n%s", want, steps)
	}
	if strings.ContainsAny(session, "\x1b\u202e") {
		t.Error("raw authored terminal controls entered the frozen session")
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Error("display executed a held run/check")
	}
	assertBytes(t, s.prep(), prep)
	assertBytes(t, filepath.Join(s.registration(), "plan.obk.yaml"), planBytes)
	assertIdle(t, s)
}

func TestFrozenSessionStaleSupportCannotContradictReview(t *testing.T) {
	s := setup(t)
	stale := fixture(t, "stale.preparation.yaml")
	put(t, s.prep(), stale)
	before := snapshot(t, s.root)
	clock := 0
	got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\nexit\n")
	wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	if clock != 1 {
		t.Fatal("new record must receive exactly one event despite unavailable support")
	}
	session := completedFrozenSession(t, got.out)
	if session != string(fixture(t, "stale-session.txt")) {
		t.Errorf("entire frozen stale-support session differs\n%s", session)
	}
	if strings.Count(session, "Support available: false\n") != 1 || strings.Contains(session, "Support available: true\n") {
		t.Error("completed session must have one false support claim and no contradictory true claim")
	}
	assertBytes(t, s.prep(), stale)
	if snapshot(t, s.root)["records/jules/preparation.obk.yaml"] != before["records/jules/preparation.obk.yaml"] {
		t.Error("unavailable retained preparation bytes/inode/controls changed")
	}
	assertBytes(t, s.record(), fixture(t, "begun.yaml"))
	assertIdle(t, s)
}
