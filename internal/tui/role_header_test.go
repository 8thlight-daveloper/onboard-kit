package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

var roleHeaderWidths = []int{60, 80, 100, 120}

func withOrganization(t *testing.T, name string) model.Plan {
	t.Helper()
	plan := acmePlan(t)
	answers := map[string]model.Value{}
	for id, value := range plan.Answers.Plan {
		answers[id] = value
	}
	answers["organization"] = model.Value{Kind: model.ValueKindText, Text: name}
	plan.Answers.Plan = answers
	return plan
}

func withAllTabs(t *testing.T, plan model.Plan) App {
	t.Helper()
	app, _ := startWith(t, plan, model.AudienceOnboarder)
	return app.WithPreparation(withSamLee(t)).WithInvitation(newFakeInvitations(t))
}

func TestTheOnboarderHeaderNamesThePlanOfTheOrganization(t *testing.T) {
	t.Parallel()
	answered, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	if got := headerOf(resize(answered, 140, 30)); !strings.HasPrefix(got, "Building: ACME Corp plan  ") {
		t.Errorf("with the organization answered the header is %q", got)
	}
	unanswered, _ := startAnswering(t)
	if got := headerOf(resize(unanswered, 140, 30)); !strings.HasPrefix(got, "Building an onboarding plan  ") {
		t.Errorf("with no organization the header is %q", got)
	}
}

func TestTheNewcomerHeaderNamesTheirOnboardingAtTheOrganization(t *testing.T) {
	t.Parallel()
	answered, _ := startWith(t, acmePlan(t), model.AudienceOnboardee)
	if got := headerOf(resize(answered, 140, 30)); !strings.HasPrefix(got, "Your onboarding at ACME Corp  ") {
		t.Errorf("with the organization answered the header is %q", got)
	}
	unanswered, _ := startWith(t, blankPlan(t), model.AudienceOnboardee)
	if got := headerOf(resize(unanswered, 140, 30)); !strings.HasPrefix(got, "Your onboarding  ") {
		t.Errorf("with no organization the header is %q", got)
	}
}

func TestAnsweringTheOrganizationRenamesTheHeaderAtOnce(t *testing.T) {
	t.Parallel()
	app, _ := startAnswering(t)
	answered := press(app, typed("Globex"), enter)
	if got := headerOf(resize(answered, 140, 30)); !strings.HasPrefix(got, "Building: Globex plan  ") {
		t.Errorf("after answering the header is %q", got)
	}
	deferred := press(moveTo(t, answered, "plan.organization"), tea.KeyMsg{Type: tea.KeyCtrlD})
	if got := headerOf(resize(deferred, 140, 30)); !strings.HasPrefix(got, "Building an onboarding plan  ") {
		t.Errorf("a deferred organization is not a name, header %q", got)
	}
}

func TestABlankOrganizationIsNotAName(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, withOrganization(t, "  \n "), model.AudienceOnboardee)
	if got := headerOf(resize(app, 140, 30)); !strings.HasPrefix(got, "Your onboarding  ") {
		t.Errorf("header %q", got)
	}
}

func TestALongOrInvisibleOrganizationNameIsCutAndShownAsText(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, withOrganization(t, "Acme\x1b[31m Industrial\nManufacturing Corporation of America"), model.AudienceOnboarder)
	title := app.headerTitle()
	if title != `Building: Acme\x1b[31m Industrial… plan` {
		t.Errorf("title %q", title)
	}
	if cellWidth(title) != len("Building: ")+organizationTitleCells+len(" plan") {
		t.Errorf("the organization takes at most %d cells: %q", organizationTitleCells, title)
	}
}

func TestTheTabsNameEachRole(t *testing.T) {
	t.Parallel()
	onboarder := resize(withAllTabs(t, acmePlan(t)), 200, 30)
	if got := headerOf(onboarder); !strings.Contains(got, "Answer w  Overview o  [Review]  Sections s  Contacts c  Preparation t  Invitation i") {
		t.Errorf("the onboarder's tabs are %q", got)
	}
	newcomer, _ := startWith(t, acmePlan(t), model.AudienceOnboardee)
	newcomer = resize(newcomer, 200, 30)
	if got := headerOf(newcomer); !strings.Contains(got, "Your onboarding at ACME Corp  Your questions w  Overview o  [Review]  ") {
		t.Errorf("the newcomer's tabs are %q", got)
	}
	overview := press(newcomer, typed("o"))
	if hints := viewLines(overview)[len(viewLines(overview))-1]; !strings.Contains(hints, "esc your questions") {
		t.Errorf("the overview names the newcomer's question screen: %q", hints)
	}
	if press(overview, typed("w")).screen != screenAnswer {
		t.Error("w opens the newcomer's questions")
	}
}

func roleHeaderApps(t *testing.T) map[string]App {
	t.Helper()
	newcomer, _ := startWith(t, acmePlan(t), model.AudienceOnboardee)
	unnamedNewcomer, _ := startWith(t, blankPlan(t), model.AudienceOnboardee)
	longNewcomer, _ := startWith(t, withOrganization(t, "Acme Industrial Manufacturing Corporation of America"), model.AudienceOnboardee)
	return map[string]App{
		"onboarder":                  withAllTabs(t, acmePlan(t)),
		"onboarder with no name":     withAllTabs(t, blankPlan(t)),
		"onboarder with a long name": withAllTabs(t, withOrganization(t, "Acme Industrial Manufacturing Corporation of America")),
		"newcomer":                   newcomer,
		"newcomer with no name":      unnamedNewcomer,
		"newcomer with a long name":  longNewcomer,
	}
}

func TestEveryTabStaysInTheHeaderAtNarrowWidths(t *testing.T) {
	t.Parallel()
	for name, app := range roleHeaderApps(t) {
		for _, tab := range app.tabs() {
			opened := press(app.openOverview(), typed(tab.key))
			if opened.screen != tab.screen {
				t.Fatalf("%s: %s did not open %s", name, tab.key, tab.name)
			}
			for _, width := range roleHeaderWidths {
				header := headerOf(resize(opened, width, 30))
				if cellWidth(header) > width-2 {
					t.Errorf("%s, %s, %d columns: the header is %d wide: %q", name, tab.name, width, cellWidth(header), header)
				}
				if !strings.Contains(header, "["+tab.name+"]") {
					t.Errorf("%s, %s, %d columns: the current tab is not shown: %q", name, tab.name, width, header)
				}
			}
		}
	}
}

func TestTheTitleStaysWhereTheTabsFitBesideIt(t *testing.T) {
	t.Parallel()
	for name, app := range roleHeaderApps(t) {
		title := app.headerTitle()
		for _, tab := range app.tabs()[:3] {
			opened := press(app.openOverview(), typed(tab.key))
			for _, width := range []int{80, 100, 120} {
				if header := headerOf(resize(opened, width, 30)); !strings.HasPrefix(header, title) {
					t.Errorf("%s, %s, %d columns: the title %q is gone: %q", name, tab.name, width, title, header)
				}
			}
		}
	}
}

func TestNoScreenSaysWalk(t *testing.T) {
	t.Parallel()
	for name, build := range goldenScreens(t) {
		for _, size := range goldenSizes {
			app := build(size)
			for _, shown := range []App{app, withHelpOpen(app)} {
				if view := strings.ToLower(plainView(shown)); strings.Contains(view, "walk") {
					t.Errorf("%s at %d columns says walk:\n%s", name, size.width, plainView(shown))
				}
			}
		}
	}
}

func withHelpOpen(app App) App {
	app.helpOpen = true
	return app
}
