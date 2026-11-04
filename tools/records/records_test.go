package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const architectureFile = "components:\n  model:\n    kind: model\n  plan/decoder:\n    kind: core\n  tui:\n    kind: screen\n"

func recordText(number string, front string, title string) string {
	return "---\n" + front + "---\n\n# " + number + ". " + title + "\n\n## Context\n"
}

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "contracts"), 0o755); err != nil {
		t.Fatal(err)
	}
	files["architecture.yaml"] = architectureFile
	for name, text := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func problemsOf(t *testing.T, files map[string]string) []string {
	t.Helper()
	got, err := check(writeTree(t, files))
	if err != nil {
		t.Fatal(err)
	}
	return got.problems
}

func wantProblem(t *testing.T, problems []string, fragment string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem, fragment) {
			return
		}
	}
	t.Fatalf("no problem contains %q in %v", fragment, problems)
}

func TestAConsistentSetOfRecordsHasNoProblems(t *testing.T) {
	t.Parallel()
	files := map[string]string{
		"docs/decisions/0010-first.md":  recordText("0010", "category: behavior\nstatus: accepted\napproval: person\naffects: [plan/decoder]\n", "First"),
		"docs/decisions/0020-second.md": recordText("0020", "category: architecture\nstatus: proposed\nbuilds-on: ['0010']\n", "Second"),
		"docs/decisions/0030-third.md":  recordText("0030", "category: operations\nstatus: accepted\napproval: agent\nnote: A local choice.\n", "Third"),
	}
	if problems := problemsOf(t, files); len(problems) != 0 {
		t.Fatalf("got %v", problems)
	}
}

func TestAFileWithoutFrontMatterIsReported(t *testing.T) {
	t.Parallel()
	problems := problemsOf(t, map[string]string{"docs/decisions/0010-first.md": "# 0010. First\n"})
	wantProblem(t, problems, "0010-first.md: no front matter")
}

func TestAnUnknownFieldIsReported(t *testing.T) {
	t.Parallel()
	problems := problemsOf(t, map[string]string{"docs/decisions/0010-first.md": recordText("0010", "category: operations\nstatus: accepted\napproval: person\nowner: someone\n", "First")})
	wantProblem(t, problems, "owner")
}

func TestAValueOutsideItsListIsReported(t *testing.T) {
	t.Parallel()
	problems := problemsOf(t, map[string]string{"docs/decisions/0010-first.md": recordText("0010", "category: design\nstatus: done\napproval: agent-judgment\n", "First")})
	wantProblem(t, problems, `category "design"`)
	wantProblem(t, problems, `status "done"`)
	wantProblem(t, problems, `approval "agent-judgment"`)
}

func TestAProposedRecordNeedsNoApproval(t *testing.T) {
	t.Parallel()
	problems := problemsOf(t, map[string]string{"docs/decisions/0010-first.md": recordText("0010", "category: operations\nstatus: proposed\n", "First")})
	if len(problems) != 0 {
		t.Fatalf("got %v", problems)
	}
}

func TestAnAcceptedRecordNeedsApproval(t *testing.T) {
	t.Parallel()
	problems := problemsOf(t, map[string]string{"docs/decisions/0010-first.md": recordText("0010", "category: operations\nstatus: accepted\n", "First")})
	wantProblem(t, problems, `approval ""`)
}

func TestTheHeadingMustCarryTheFileNumber(t *testing.T) {
	t.Parallel()
	problems := problemsOf(t, map[string]string{"docs/decisions/0010-first.md": recordText("0020", "category: operations\nstatus: proposed\n", "First")})
	wantProblem(t, problems, "the first heading must start with")
}

func TestAffectsBelongsOnProductRecordsOnly(t *testing.T) {
	t.Parallel()
	problems := problemsOf(t, map[string]string{
		"docs/decisions/0010-first.md":  recordText("0010", "category: operations\nstatus: proposed\naffects: [tui]\n", "First"),
		"docs/decisions/0020-second.md": recordText("0020", "category: behavior\nstatus: proposed\n", "Second"),
		"docs/decisions/0030-third.md":  recordText("0030", "category: ux\nstatus: proposed\naffects: []\n", "Third"),
		"docs/decisions/0040-fourth.md": recordText("0040", "category: architecture\nstatus: proposed\n", "Fourth"),
	})
	wantProblem(t, problems, "0010-first.md: a operations record has no affects field")
	wantProblem(t, problems, "0020-second.md: a behavior record needs an affects field")
	if len(problems) != 2 {
		t.Fatalf("got %v", problems)
	}
}

func TestAffectsMustNameComponentsAndNeverAll(t *testing.T) {
	t.Parallel()
	problems := problemsOf(t, map[string]string{
		"docs/decisions/0010-first.md":  recordText("0010", "category: behavior\nstatus: proposed\naffects: [plan/encoder]\n", "First"),
		"docs/decisions/0020-second.md": recordText("0020", "category: architecture\nstatus: proposed\naffects: all\n", "Second"),
	})
	wantProblem(t, problems, `affects "plan/encoder" is not a component`)
	wantProblem(t, problems, "0020-second.md: front matter")
}

func TestALinkToNoRecordOrToItselfIsReported(t *testing.T) {
	t.Parallel()
	problems := problemsOf(t, map[string]string{"docs/decisions/0010-first.md": recordText("0010", "category: operations\nstatus: proposed\nbuilds-on: ['0099', '0010']\n", "First")})
	wantProblem(t, problems, "builds-on 0099 is not a record")
	wantProblem(t, problems, "builds-on lists the record itself")
}

func TestAReplacementMustBeListedOnBothRecords(t *testing.T) {
	t.Parallel()
	older := recordText("0010", "category: operations\nstatus: accepted\napproval: person\n", "Older")
	newer := recordText("0020", "category: operations\nstatus: accepted\napproval: person\npartly-supersedes: ['0010']\nsupersedes: ['0010']\n", "Newer")
	problems := problemsOf(t, map[string]string{"docs/decisions/0010-older.md": older, "docs/decisions/0020-newer.md": newer})
	wantProblem(t, problems, "partly-supersedes 0010, but 0010-older.md does not list 0020 in partly-superseded-by")
	wantProblem(t, problems, "supersedes 0010, but 0010-older.md does not list 0020 in superseded-by")
}

func TestAReplacementListedOnBothRecordsPasses(t *testing.T) {
	t.Parallel()
	older := recordText("0010", "category: operations\nstatus: accepted\napproval: person\npartly-superseded-by: ['0020']\n", "Older")
	newer := recordText("0020", "category: operations\nstatus: accepted\napproval: person\npartly-supersedes: ['0010']\n", "Newer")
	if problems := problemsOf(t, map[string]string{"docs/decisions/0010-older.md": older, "docs/decisions/0020-newer.md": newer}); len(problems) != 0 {
		t.Fatalf("got %v", problems)
	}
}

func TestAContractThatCitesAProductRecordMustBeListedInItsAffects(t *testing.T) {
	t.Parallel()
	problems := problemsOf(t, map[string]string{
		"docs/decisions/0010-first.md":       recordText("0010", "category: behavior\nstatus: proposed\naffects: []\n", "First"),
		"contracts/plan/decoder/contract.md": "Builds on [0010](x).\n",
	})
	wantProblem(t, problems, "the contract of plan/decoder cites this record")
}

func TestAContractMayCiteArchitectureWithoutAffectsAndProcessRecords(t *testing.T) {
	t.Parallel()
	problems := problemsOf(t, map[string]string{
		"docs/decisions/0010-first.md":       recordText("0010", "category: architecture\nstatus: proposed\n", "First"),
		"docs/decisions/0020-second.md":      recordText("0020", "category: regenerative-process\nstatus: proposed\n", "Second"),
		"contracts/plan/decoder/contract.md": "See [0010](x) and docs/decisions/0020-y.md, in 2026, with 10100 and 40960.\n",
	})
	if len(problems) != 0 {
		t.Fatalf("got %v", problems)
	}
}

func TestAFileModeIsNotACitation(t *testing.T) {
	t.Parallel()
	problems := problemsOf(t, map[string]string{
		"docs/decisions/0640-first.md":       recordText("0640", "category: behavior\nstatus: proposed\naffects: []\n", "First"),
		"contracts/plan/decoder/contract.md": "A new file is 0600 and an existing 0644 or 0640 becomes 0600. See ADR 0010.\n",
	})
	if len(problems) != 0 {
		t.Fatalf("got %v", problems)
	}
}

func TestAListedComponentWhoseContractDoesNotCiteTheRecordIsOnlyANote(t *testing.T) {
	t.Parallel()
	root := writeTree(t, map[string]string{
		"docs/decisions/0010-first.md":       recordText("0010", "category: behavior\nstatus: proposed\naffects: [plan/decoder, tui]\n", "First"),
		"contracts/plan/decoder/contract.md": "No citation here.\n",
	})
	got, err := check(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.problems) != 0 {
		t.Fatalf("got %v", got.problems)
	}
	if len(got.notes) != 1 || !strings.Contains(got.notes[0], "plan/decoder") || strings.Contains(got.notes[0], "tui") {
		t.Fatalf("got %v", got.notes)
	}
}

func TestRunNamesTheProblemsAndExitsWithAnError(t *testing.T) {
	t.Chdir(writeTree(t, map[string]string{"go.mod": "module example\n", "docs/decisions/0010-first.md": "# 0010. First\n"}))
	var out bytes.Buffer
	err := run([]string{"check"}, &out)
	if err == nil || !strings.Contains(err.Error(), "no front matter") {
		t.Fatalf("got %v", err)
	}
}

func TestRunPrintsNotesOnlyWhenAsked(t *testing.T) {
	t.Chdir(writeTree(t, map[string]string{
		"go.mod":                             "module example\n",
		"docs/decisions/0010-first.md":       recordText("0010", "category: behavior\nstatus: proposed\naffects: [plan/decoder]\n", "First"),
		"contracts/plan/decoder/contract.md": "No citation here.\n",
	}))
	var quiet, loud bytes.Buffer
	if err := run([]string{"check"}, &quiet); err != nil || !strings.Contains(quiet.String(), "cite them: 1") || strings.Contains(quiet.String(), "note:") {
		t.Fatalf("quiet: %v %q", err, quiet.String())
	}
	if err := run([]string{"check", "-v"}, &loud); err != nil || !strings.Contains(loud.String(), "note: 0010-first.md") {
		t.Fatalf("loud: %v %q", err, loud.String())
	}
}

func TestRunRefusesOtherArguments(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{nil, {"fix"}, {"check", "-x"}, {"check", "-v", "extra"}} {
		if err := run(args, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "usage") {
			t.Errorf("%v: got %v", args, err)
		}
	}
}
