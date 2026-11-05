package frontdoor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"onboard-kit/internal/command/create"
	"onboard-kit/internal/command/planfile"
	"onboard-kit/internal/command/registry"
	"onboard-kit/internal/tui"
)

func portIn(t *testing.T) (*port, string, string) {
	t.Helper()
	home := filepath.Join(t.TempDir(), "state")
	working := t.TempDir()
	t.Setenv("OBK_HOME", home)
	t.Chdir(working)
	return newPort("1.2.3", "plan-one", "2026-11-04"), home, working
}

func TestKnownPlansComeFromTheListAndAMovedOneShowsAsMissing(t *testing.T) {
	door, home, working := portIn(t)
	kept := copyOf(t, working, acmePlan)
	gone := filepath.Join(working, "moved.obk.yaml")
	older := time.Date(2026, 11, 1, 8, 0, 0, 0, time.UTC)
	if err := registry.Record(home, gone, older); err != nil {
		t.Fatal(err)
	}
	if err := registry.Record(home, kept, older.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	plans, problem := door.KnownPlans()
	if problem != "" || len(plans) != 2 {
		t.Fatalf("got %+v, %q", plans, problem)
	}
	first, second := plans[0], plans[1]
	if first.Path != kept || first.Name != "acme.obk.yaml" || first.Folder != working || first.Missing || !first.Opened.Equal(older.Add(time.Hour)) {
		t.Errorf("first %+v", first)
	}
	if second.Path != gone || !second.Missing {
		t.Errorf("second %+v", second)
	}
	if problem := door.Forget(gone); problem != "" {
		t.Fatal(problem)
	}
	if plans, _ := door.KnownPlans(); len(plans) != 1 || plans[0].Path != kept {
		t.Fatalf("after forgetting: %+v", plans)
	}
}

func TestADamagedListShowsNoPlansAndSaysWhy(t *testing.T) {
	door, home, _ := portIn(t)
	if err := os.MkdirAll(home, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(registry.Path(home), []byte("garbage\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	plans, problem := door.KnownPlans()
	want := "The list of plans at " + registry.Path(home) + " cannot be read: its first line is not onboard-kit plans 1. It is left as it is, and plans you open are not added to it. Remove the file to start a new list."
	if len(plans) != 0 || problem != want {
		t.Fatalf("got %+v, %q", plans, problem)
	}
	if about := door.About(); len(about.Problems) != 1 || about.Problems[0] != want {
		t.Fatalf("the About screen does not report it: %+v", about.Problems)
	}
}

func TestTheProposedNameIsTheFirstFreeOne(t *testing.T) {
	door, _, working := portIn(t)
	if got := door.ProposedPlanPath(); got != "onboarding.obk.yaml" {
		t.Fatalf("got %q", got)
	}
	writeFile(t, working, "onboarding.obk.yaml", nil)
	writeFile(t, working, "onboarding-2.obk.yaml", nil)
	if got := door.ProposedPlanPath(); got != "onboarding-3.obk.yaml" {
		t.Fatalf("got %q", got)
	}
	if got := door.PlanPath("acme.obk.yaml"); got != filepath.Join(working, "acme.obk.yaml") {
		t.Fatalf("got %q", got)
	}
	if got := door.PlanPath("/elsewhere/acme.obk.yaml"); got != "/elsewhere/acme.obk.yaml" {
		t.Fatalf("got %q", got)
	}
}

func TestCreatingWritesTheCorePlanRecordsItAndNeverReplacesAFile(t *testing.T) {
	door, home, working := portIn(t)
	path := filepath.Join(working, "acme.obk.yaml")
	created := door.CreatePlan(path)
	if created.Refusal != "" || created.Path != path {
		t.Fatalf("got %+v", created)
	}
	plan, ok := planfile.Open(path, os.Stderr)
	core, _ := create.CorePlan("plan-one", "1.2.3", "2026-11-04")
	if !ok || plan.ID != "plan-one" || plan.Provenance.TemplateSet != "1.2.3" || len(plan.Sections) == 0 || len(plan.Sections) != len(core.Sections) {
		t.Fatalf("plan %q %+v with %d sections, want the %d of the core set", plan.ID, plan.Provenance, len(plan.Sections), len(core.Sections))
	}
	entries, err := registry.Load(home)
	if err != nil || len(entries) != 1 || entries[0].Path != path {
		t.Fatalf("list %+v, %v", entries, err)
	}
	before, _ := os.ReadFile(path)
	again := door.CreatePlan(path)
	if again.Refusal != "That file already exists, and a new plan never replaces a file. Choose another name." {
		t.Fatalf("got %+v", again)
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatal("the plan was replaced")
	}
}

func TestAboutListsEveryPlaceWithTheVersionAndTemplateSet(t *testing.T) {
	door, home, _ := portIn(t)
	about := door.About()
	if about.Version != "1.2.3" || !strings.HasPrefix(about.TemplateSet, "1.2.3, ") {
		t.Fatalf("got %q %q", about.Version, about.TemplateSet)
	}
	var names []string
	byName := map[string]tui.AboutPlace{}
	for _, place := range about.Places {
		names = append(names, place.Name)
		byName[place.Name] = place
	}
	if got := strings.Join(names, ", "); got != "State home, Data folder, Plan list, Sessions, Imports, Plan folder, Records root, Downloads" {
		t.Fatalf("got %s", got)
	}
	checks := map[string]string{"State home": home, "Plan list": registry.Path(home), "Sessions": filepath.Join(home, "sessions"), "Imports": filepath.Join(home, "import-pending"), "Records root": filepath.Join(home, "onboarding"), "Data folder": home}
	for name, path := range checks {
		if byName[name].Path != path {
			t.Errorf("%s: %q, want %q", name, byName[name].Path, path)
		}
	}
	if byName["State home"].Chosen != "Chosen by the OBK_HOME setting." || byName["Downloads"].Use != tui.PlaceNotUsedYet || byName["Plan list"].Use != tui.PlaceReadAndWritten {
		t.Fatalf("got %+v", about.Places)
	}
}

func TestATypedPathMayStartAtTheHomeFolder(t *testing.T) {
	door, _, _ := portIn(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	copyOf(t, home, acmePlan)
	if checked := door.Check("~/acme.obk.yaml"); checked.Kind != tui.FilePlan || checked.Path != filepath.Join(home, "acme.obk.yaml") {
		t.Fatalf("got %+v", checked)
	}
	if got := door.PlanPath("~/new.obk.yaml"); got != filepath.Join(home, "new.obk.yaml") {
		t.Fatalf("got %q", got)
	}
}

func TestCreatingInAFolderThatDoesNotExistCreatesTheFolder(t *testing.T) {
	door, _, working := portIn(t)
	path := filepath.Join(working, "plans", "acme", "acme.obk.yaml")
	if got := door.MissingFolder(path); got != filepath.Join(working, "plans", "acme") {
		t.Fatalf("missing %q", got)
	}
	if existing := door.MissingFolder(filepath.Join(working, "acme.obk.yaml")); existing != "" {
		t.Fatalf("an existing folder is not missing: %q", existing)
	}
	created := door.CreatePlan(path)
	if created.Refusal != "" || created.Path != path {
		t.Fatalf("got %+v", created)
	}
	if _, ok := planfile.Open(path, os.Stderr); !ok {
		t.Fatal("the plan was not written")
	}
	if door.MissingFolder(path) != "" {
		t.Fatal("the folder is still missing")
	}
}

func TestCreatingBelowAFileSaysWhyInPlainWords(t *testing.T) {
	door, _, working := portIn(t)
	blocker := writeFile(t, working, "plans", []byte("a file"))
	created := door.CreatePlan(filepath.Join(blocker, "acme.obk.yaml"))
	if created.Refusal != blocker+" is a file, and a plan needs a folder to go in." {
		t.Fatalf("got %q", created.Refusal)
	}
	deeper := door.CreatePlan(filepath.Join(blocker, "inner", "acme.obk.yaml"))
	if !strings.HasPrefix(deeper.Refusal, "The folder ") || strings.Contains(deeper.Refusal, ".obk-new-") {
		t.Fatalf("got %q", deeper.Refusal)
	}
}
