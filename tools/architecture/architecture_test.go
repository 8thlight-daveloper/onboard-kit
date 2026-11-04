package main

import (
	"bytes"
	"strings"
	"testing"
)

func fixtureRules() rules {
	var loaded rules
	loaded.Naming.DepthOneExceptions = []string{"model", "tui"}
	loaded.Kinds = map[string]kindRule{
		"model":   {Imports: []string{}, Contract: "none"},
		"core":    {Imports: []string{"model", "core"}, BannedCategories: []string{"filesystem", "terminal"}, Contract: "required"},
		"driver":  {Imports: []string{"model", "core", "driver"}, BannedCategories: []string{"terminal"}, Contract: "required"},
		"screen":  {Imports: []string{"model", "core"}, Contract: "none"},
		"command": {Imports: []string{"model", "core", "driver", "screen", "command"}, Contract: "none"},
	}
	loaded.Components = map[string]componentRule{
		"model":         {Kind: "model"},
		"plan/decoder":  {Kind: "core", Imports: []string{"model"}},
		"plan/editor":   {Kind: "core", Imports: []string{"plan/decoder", "model"}},
		"plan/store":    {Kind: "driver", Imports: []string{"plan/decoder", "model"}},
		"tui":           {Kind: "screen", Imports: []string{"plan/editor", "model"}},
		"command/walk":  {Kind: "command"},
		"plan/migrator": {Kind: "core", Status: "deferred"},
	}
	return loaded
}

func fixtureGraph() map[string]packageImports {
	return map[string]packageImports{
		"model":        {},
		"plan/decoder": {components: []string{"model"}},
		"plan/editor":  {components: []string{"model", "plan/decoder"}},
		"plan/store":   {components: []string{"model", "plan/decoder"}, categories: map[string][]string{"filesystem": {"os"}}},
		"tui":          {components: []string{"model", "plan/editor"}, categories: map[string][]string{"terminal": {"bubbletea"}}},
		"command/walk": {components: []string{"tui"}},
	}
}

func always(string) bool { return true }

func problemsFor(loaded rules, graph map[string]packageImports, contractExists func(string) bool) []string {
	return evaluate(loaded, graph, contractExists).problems
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

func TestRulesThatTheCodeKeepsHaveNoProblems(t *testing.T) {
	t.Parallel()
	got := evaluate(fixtureRules(), fixtureGraph(), always)
	if len(got.problems) != 0 {
		t.Fatalf("got %v", got.problems)
	}
	if len(got.undeclared) != 1 || got.undeclared[0] != "command/walk" {
		t.Fatalf("undeclared %v", got.undeclared)
	}
}

func TestAnImportOutsideTheDeclaredListIsReported(t *testing.T) {
	t.Parallel()
	graph := fixtureGraph()
	graph["plan/decoder"] = packageImports{components: []string{"model", "plan/editor"}}
	problems := problemsFor(fixtureRules(), graph, always)
	wantProblem(t, problems, "plan/decoder: imports plan/editor, which its declared imports do not list")
}

func TestAnImportThatItsKindForbidsIsReportedEvenWhenListed(t *testing.T) {
	t.Parallel()
	loaded := fixtureRules()
	loaded.Components["plan/editor"] = componentRule{Kind: "core", Imports: []string{"model", "plan/store"}}
	graph := fixtureGraph()
	graph["plan/editor"] = packageImports{components: []string{"model", "plan/store"}}
	problems := problemsFor(loaded, graph, always)
	wantProblem(t, problems, "a core component may not import a driver component, and it imports plan/store")
}

func TestABannedCategoryIsReportedForItsKindOnly(t *testing.T) {
	t.Parallel()
	graph := fixtureGraph()
	graph["plan/decoder"] = packageImports{components: []string{"model"}, categories: map[string][]string{"filesystem": {"os", "io/fs"}}}
	problems := problemsFor(fixtureRules(), graph, always)
	wantProblem(t, problems, "plan/decoder: a core component may not use the filesystem category, and it imports os, io/fs")
	if len(problems) != 1 {
		t.Fatalf("the driver and the screen use their own categories: %v", problems)
	}
}

func TestAPackageWithNoComponentIsReported(t *testing.T) {
	t.Parallel()
	graph := fixtureGraph()
	graph["plan/refresher"] = packageImports{}
	wantProblem(t, problemsFor(fixtureRules(), graph, always), "plan/refresher: the code has a package for it, but architecture.yaml has no component")
}

func TestAComponentWithNoPackageIsReportedUnlessPlanned(t *testing.T) {
	t.Parallel()
	graph := fixtureGraph()
	delete(graph, "tui")
	problems := problemsFor(fixtureRules(), graph, always)
	wantProblem(t, problems, "tui: architecture.yaml lists it, and the code has no package for it")
	for _, problem := range problems {
		if strings.Contains(problem, "plan/migrator") {
			t.Fatalf("a deferred component needs no package: %v", problems)
		}
	}
}

func TestAKindThatRequiresAContractNeedsItsFolder(t *testing.T) {
	t.Parallel()
	problems := problemsFor(fixtureRules(), fixtureGraph(), func(id string) bool { return id != "plan/store" && id != "plan/editor" })
	wantProblem(t, problems, "plan/store: a driver component needs a contracts folder")
	wantProblem(t, problems, "plan/editor: a core component needs a contracts folder")
	if len(problems) != 2 {
		t.Fatalf("screens, commands and the model need no folder: %v", problems)
	}
}

func TestADeclaredImportMustNameAComponent(t *testing.T) {
	t.Parallel()
	loaded := fixtureRules()
	loaded.Components["plan/decoder"] = componentRule{Kind: "core", Imports: []string{"model", "plan/migrator", "plan/nothing"}}
	wantProblem(t, problemsFor(loaded, fixtureGraph(), always), "plan/decoder: declares an import of plan/nothing, which is not a component")
}

func TestAnIdMustBeNounSlashRoleOrAnException(t *testing.T) {
	t.Parallel()
	loaded := fixtureRules()
	loaded.Components["secrets"] = componentRule{Kind: "core"}
	graph := fixtureGraph()
	graph["secrets"] = packageImports{}
	wantProblem(t, problemsFor(loaded, graph, always), "secrets: the id is not noun/role")
}

func TestAnUndefinedKindIsReported(t *testing.T) {
	t.Parallel()
	loaded := fixtureRules()
	loaded.Components["plan/decoder"] = componentRule{Kind: "driverish"}
	wantProblem(t, problemsFor(loaded, fixtureGraph(), always), `plan/decoder: kind "driverish" is not defined`)
}

func TestAnImportCycleIsReported(t *testing.T) {
	t.Parallel()
	loaded := fixtureRules()
	graph := fixtureGraph()
	graph["plan/decoder"] = packageImports{components: []string{"model", "plan/editor"}}
	graph["plan/editor"] = packageImports{components: []string{"model", "plan/decoder"}}
	problems := problemsFor(loaded, graph, always)
	wantProblem(t, problems, "import cycle: plan/editor -> plan/decoder -> plan/editor")
}

func TestComponentOfMapsImportPathsToIds(t *testing.T) {
	t.Parallel()
	for path, want := range map[string]string{"onboard-kit/internal/plan/decoder": "plan/decoder", "onboard-kit/binary": "binary", "onboard-kit/internal/tui": "tui"} {
		if got, ok := componentOf("onboard-kit", path); !ok || got != want {
			t.Errorf("%s: got %q %v", path, got, ok)
		}
	}
	if _, ok := componentOf("onboard-kit", "onboard-kit/contracts/plan/decoder"); ok {
		t.Error("a contract package is not a component")
	}
}

func TestImportsOfSortsComponentsAndGroupsCategories(t *testing.T) {
	t.Parallel()
	got := importsOf("onboard-kit", []string{"onboard-kit/internal/plan/editor", "os", "fmt", "onboard-kit/internal/model", "path/filepath", "github.com/charmbracelet/bubbletea"})
	if strings.Join(got.components, ",") != "model,plan/editor" {
		t.Fatalf("components %v", got.components)
	}
	if strings.Join(got.categories["filesystem"], ",") != "os,path/filepath" || len(got.categories["terminal"]) != 1 || len(got.categories) != 2 {
		t.Fatalf("categories %v", got.categories)
	}
}

func TestRunRefusesOtherArguments(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{nil, {"fix"}, {"check", "-v"}} {
		if err := run(args, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "usage") {
			t.Errorf("%v: got %v", args, err)
		}
	}
}
