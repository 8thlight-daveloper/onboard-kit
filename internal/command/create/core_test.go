package create

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"onboard-kit/internal/catalog/reader"
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/authoringvalidator"
	"onboard-kit/internal/plan/validator"
	"onboard-kit/internal/question/finder"
)

func TestTheCorePlanHoldsEveryCoreSectionTheCatalogCarries(t *testing.T) {
	plan, problems := CorePlan("plan-one", "development", "2026-11-04")
	if len(problems) > 0 {
		t.Fatalf("problems: %v", problems)
	}
	var ids []string
	for _, section := range plan.Sections {
		if section.Source != nil {
			ids = append(ids, section.ID)
		}
	}
	var want []string
	for _, staged := range coreSet {
		if !slices.Contains(notInCatalogYet, staged.source) {
			want = append(want, staged.instance)
		}
	}
	slices.Sort(want)
	if !slices.Equal(ids, want) {
		t.Fatalf("sections %v, want %v", ids, want)
	}
	if plan.ID != "plan-one" || plan.Provenance.TemplateSet != "development" || plan.Provenance.GeneratedOn != "2026-11-04" {
		t.Fatalf("identity %q, provenance %+v", plan.ID, plan.Provenance)
	}
}

var notInCatalogYet = []string{"standup-attendance", "codebase-tutor", "ai-tool-policy"}

func TestOnlyTheThreeEntriesOf0550AreLeftOutUntilTheCatalogHasThem(t *testing.T) {
	available := map[string]bool{}
	for _, summary := range reader.New().List() {
		available[summary.ID] = true
	}
	var missing []string
	for _, staged := range coreSet {
		if !available[staged.source] {
			missing = append(missing, staged.source)
		}
	}
	if !slices.Equal(missing, notInCatalogYet) {
		t.Fatalf("left out %v, want %v. A misspelled core id is left out silently, so fix the list or this test", missing, notInCatalogYet)
	}
}

func TestPairingIsStagedTwiceWithDistinctIds(t *testing.T) {
	plan, _ := CorePlan("plan-one", "development", "2026-11-04")
	var pairing []string
	for _, section := range plan.Sections {
		if section.Source != nil && section.Source.Template == "ticket-pairing" {
			pairing = append(pairing, section.ID)
		}
	}
	if !slices.Equal(pairing, []string{"ticket-pairing-driver", "ticket-pairing-navigator"}) {
		t.Fatalf("got %v", pairing)
	}
}

func TestWritingANewFileNeverReplacesOne(t *testing.T) {
	folder := t.TempDir()
	path := filepath.Join(folder, "acme.obk.yaml")
	if err := WriteNew(path, []byte("first\n")); err != nil {
		t.Fatal(err)
	}
	if err := WriteNew(path, []byte("second\n")); !errors.Is(err, ErrExists) {
		t.Fatalf("got %v", err)
	}
	if data, _ := os.ReadFile(path); string(data) != "first\n" {
		t.Fatalf("the file holds %q", data)
	}
	link := filepath.Join(folder, "link.obk.yaml")
	if err := os.Symlink(filepath.Join(folder, "nowhere"), link); err != nil {
		t.Fatal(err)
	}
	if err := WriteNew(link, []byte("through a link\n")); !errors.Is(err, ErrExists) {
		t.Fatalf("a dangling link was written through: %v", err)
	}
	entries, _ := os.ReadDir(folder)
	if len(entries) != 2 {
		t.Fatalf("a temporary file was left: %v", entries)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("mode %v", info.Mode().Perm())
	}
}

func TestTheCorePlanAsksOnlyTheOrganizationAtPlanScopeAndThenWhereThePlanIsKept(t *testing.T) {
	plan, problems := CorePlan("plan-one", "development", "2026-11-04")
	if len(problems) > 0 {
		t.Fatalf("problems: %v", problems)
	}
	questions := finder.New().Questions(plan, model.AudienceOnboarder)
	if len(questions) < 4 {
		t.Fatalf("only %d questions", len(questions))
	}
	var planScope []string
	for _, question := range questions {
		if question.Scope == "plan" {
			planScope = append(planScope, question.Name)
		}
	}
	if !slices.Equal(planScope, []string{"organization"}) || questions[0].Name != "organization" {
		t.Fatalf("plan scope asks %v, first question %s.%s", planScope, questions[0].Scope, questions[0].Name)
	}
	for index, want := range []string{"plan_repository", "plan_file"} {
		got := questions[index+1]
		if got.Scope != "plan-source" || got.Name != want || got.Required {
			t.Fatalf("question %d is %s.%s required=%v, want plan-source.%s optional", index+2, got.Scope, got.Name, got.Required, want)
		}
	}
	if plan.Sections[0].ID != "plan-source" || questions[3].Scope == "plan-source" || questions[3].Scope == "plan" {
		t.Fatalf("first section %q, fourth question %+v", plan.Sections[0].ID, questions[3])
	}
	for _, variable := range plan.Variables {
		if variable.Name == "team_name" {
			t.Fatal("a team name belongs to a team plan, not to this one")
		}
		if variable.Name == "organization" && !variable.Required {
			t.Fatal("the organization is required")
		}
	}
}

func TestTheCorePlanPassesTheAuthoringAndPlanChecks(t *testing.T) {
	plan, _ := CorePlan("plan-one", "development", "2026-11-04")
	for _, problem := range authoringvalidator.New().Validate(plan) {
		if problem.Severity == model.SeverityError {
			t.Errorf("authoring: %+v", problem)
		}
	}
	for _, problem := range validator.New().Validate(plan) {
		if problem.Severity == model.SeverityError {
			t.Errorf("plan: %+v", problem)
		}
	}
}
