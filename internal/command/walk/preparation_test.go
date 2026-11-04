package walk

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"onboard-kit/internal/tui"
)

const examplePlan = "../../../docs/reference/examples/acme/acme.obk.yaml"

func copiedPlan(t *testing.T) (string, string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(examplePlan)
	if err != nil {
		t.Fatal(err)
	}
	plan := filepath.Join(root, "acme.obk.yaml")
	if err := os.WriteFile(plan, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return root, plan
}

func TestARootThatIsNotAFolderIsAUsageError(t *testing.T) {
	_, plan := copiedPlan(t)
	var stderr bytes.Buffer
	code := Run([]string{"--plan", plan, "--root", filepath.Join(filepath.Dir(plan), "missing")}, strings.NewReader(""), &bytes.Buffer{}, &stderr)
	if code != exitUsage || !strings.Contains(stderr.String(), "is not a folder") {
		t.Errorf("exit %d, stderr %q", code, stderr.String())
	}
	fresh := filepath.Join(filepath.Dir(plan), "fresh.obk.yaml")
	Run([]string{"--plan", fresh, "--templates", "../../../docs/reference/examples/acme/templates", "--root", plan}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if _, err := os.Stat(fresh); !os.IsNotExist(err) {
		t.Errorf("a refused root writes no new plan: %v", err)
	}
}

func TestThePortListsLoadsAndSavesThroughTheStore(t *testing.T) {
	root, plan := copiedPlan(t)
	files := newPreparationFiles(root, plan)
	people, outcome := files.People()
	if outcome != tui.PreparationDone || len(people) != 0 {
		t.Fatalf("an empty root lists nobody: %v %v", people, outcome)
	}
	loaded, outcome := files.Load("newcomer-one")
	if outcome != tui.PreparationDone || loaded.Exists || !loaded.FoldersMissing || loaded.Destination != "records/newcomer-one/preparation.obk.yaml" || loaded.Plan.ID == "" {
		t.Fatalf("a new person loads as an identity-only draft: %+v %v", loaded, outcome)
	}
	reviewed := []byte("schemaVersion: 1\nkind: preparation\naccount: newcomer-one\nplanId: \"" + loaded.Plan.ID + "\"\nrecipient:\n  name: Example newcomer\n")
	if outcome := files.Save("newcomer-one", reviewed, false); outcome != tui.PreparationFoldersMissing {
		t.Errorf("without consent the folder is not made: %v", outcome)
	}
	if outcome := files.Save("newcomer-one", reviewed, true); outcome != tui.PreparationDone {
		t.Fatalf("the save is published: %v", outcome)
	}
	written, err := os.ReadFile(filepath.Join(root, "records", "newcomer-one", "preparation.obk.yaml"))
	if err != nil || !bytes.Equal(written, reviewed) {
		t.Errorf("the reviewed bytes are the saved bytes: %q %v", written, err)
	}
	if outcome := files.Save("newcomer-one", reviewed, true); outcome != tui.PreparationStale {
		t.Errorf("the same load cannot save twice: %v", outcome)
	}
	people, _ = files.People()
	if len(people) != 1 || people[0].Account != "newcomer-one" {
		t.Errorf("the saved person is listed: %+v", people)
	}
	if _, outcome := files.Load("../escape"); outcome != tui.PreparationPathRefused {
		t.Errorf("an unsafe account is refused: %v", outcome)
	}
	if outcome := files.Save("never-loaded", reviewed, true); outcome != tui.PreparationNotSaved {
		t.Errorf("a person who was never loaded cannot be saved: %v", outcome)
	}
}

func TestTheRootDefaultsToThePlansFolder(t *testing.T) {
	for _, example := range []struct{ plan, root, want string }{
		{"acme.obk.yaml", "", "."},
		{"plans/acme.obk.yaml", "", "plans"},
		{"/shared/plans/acme.obk.yaml", "", "/shared/plans"},
		{"plans/acme.obk.yaml", "people", "people"},
	} {
		if got := preparationRootFor(example.plan, example.root); got != example.want {
			t.Errorf("plan %q with --root %q: got %q, want %q", example.plan, example.root, got, example.want)
		}
	}
}
