package frontdoor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"onboard-kit/internal/tui"
)

var acmePlan, _ = filepath.Abs("../../../docs/reference/examples/acme/acme.obk.yaml")

func writeFile(t *testing.T, folder string, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(folder, name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func copyOf(t *testing.T, folder string, source string) string {
	t.Helper()
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	return writeFile(t, folder, filepath.Base(source), data)
}

func TestAPlanAndAnInvitationAreKnownByTheirContent(t *testing.T) {
	folder := t.TempDir()
	plan := copyOf(t, folder, acmePlan)
	invitation := writeFile(t, folder, "named-like-a-plan.obk.yaml", []byte("PK\x03\x04 rest of an archive"))
	for path, want := range map[string]tui.FileKind{plan: tui.FilePlan, invitation: tui.FileInvitation} {
		checked, problems := Check(path)
		if checked.Kind != want || checked.Path != path || checked.Refusal != "" || len(problems) != 0 {
			t.Errorf("%s: got %+v, %v", path, checked, problems)
		}
	}
}

func TestEveryOtherFileIsRefusedInOneSentenceThatNamesWhatIsExpected(t *testing.T) {
	folder := t.TempDir()
	preparation := copyOf(t, folder, "../../../acceptance/testdata/preparation/preparation.yaml")
	record := copyOf(t, folder, "../../../acceptance/testdata/preparation/record.yaml")
	notes := writeFile(t, folder, "notes.txt", []byte("not yaml: [\n"))
	list := writeFile(t, folder, "list.yaml", []byte("- a\n- b\n"))
	other := writeFile(t, folder, "other.yaml", []byte("schemaVersion: 1\nkind: \"spreadsheet\"\n"))
	cases := map[string]string{
		preparation:                        preparation + " is a preparation, but a plan or an invitation is expected.",
		record:                             record + " is a record, but a plan or an invitation is expected.",
		notes:                              notes + " is neither a plan nor an invitation.",
		list:                               list + " is neither a plan nor an invitation.",
		other:                              other + " is neither a plan nor an invitation.",
		folder:                             folder + " is a folder, but a plan or an invitation is expected.",
		filepath.Join(folder, "gone.yaml"): filepath.Join(folder, "gone.yaml") + " does not exist.",
	}
	for path, want := range cases {
		checked, problems := Check(path)
		if checked.Kind != tui.FileRefused || checked.Refusal != want || len(problems) != 0 {
			t.Errorf("%s: got %+v, %v, want %q", path, checked, problems, want)
		}
		if strings.Count(checked.Refusal, ". ") > 0 {
			t.Errorf("%s: more than one sentence: %q", path, checked.Refusal)
		}
	}
}

func TestAPlanThatCannotBeDecodedIsRefusedWithItsProblems(t *testing.T) {
	broken := writeFile(t, t.TempDir(), "broken.obk.yaml", []byte("schemaVersion: 1\nid: plan-one\nsections: 7\n"))
	checked, problems := Check(broken)
	if checked.Kind != tui.FileRefused || !strings.HasPrefix(checked.Refusal, broken+" is a plan that cannot be opened: ") || len(problems) == 0 {
		t.Fatalf("got %+v, %v", checked, problems)
	}
}

func TestARelativePathIsCheckedAndGivenBackInFull(t *testing.T) {
	folder := t.TempDir()
	copyOf(t, folder, acmePlan)
	t.Chdir(folder)
	checked, _ := Check("acme.obk.yaml")
	if checked.Kind != tui.FilePlan || checked.Path != filepath.Join(folder, "acme.obk.yaml") {
		t.Fatalf("got %+v", checked)
	}
}

func TestASectionTemplateIsNotAPlan(t *testing.T) {
	folder := t.TempDir()
	template := copyOf(t, folder, filepath.Join("..", "..", "..", "docs", "reference", "examples", "acme", "templates", "access.yaml"))
	checked, _ := Check(template)
	if checked.Kind != tui.FileRefused || !strings.Contains(checked.Refusal, "neither a plan nor an invitation") {
		t.Fatalf("got %+v", checked)
	}
}

func TestAPlanWithNoSectionsIsStillAPlan(t *testing.T) {
	folder := t.TempDir()
	empty := writeFile(t, folder, "empty.obk.yaml", []byte("schemaVersion: 1\nid: plan-one\n"))
	checked, problems := Check(empty)
	if checked.Kind != tui.FilePlan || checked.Refusal != "" || len(problems) != 0 {
		t.Fatalf("got %+v, %v", checked, problems)
	}
}
