package tui

import (
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

func vacantRolesApp(t *testing.T, audience model.Audience) App {
	t.Helper()
	plan := acmePlan(t)
	for index := range plan.Directory.Roles {
		plan.Directory.Roles[index].Holders = nil
	}
	app, _ := startWith(t, plan, audience)
	return resize(app, 140, 40)
}

func firstProblem(t *testing.T, app App) App {
	t.Helper()
	for index, item := range app.review.items {
		if item.kind == reviewProblem && item.hasEntry {
			app.review.cursor = index
			return app
		}
	}
	t.Fatal("no directory problem in the review")
	return app
}

func TestAVacantRoleProblemNamesTheRole(t *testing.T) {
	t.Parallel()
	app := firstProblem(t, vacantRolesApp(t, model.AudienceOnboarder))
	item, _ := app.review.selected()
	want := entryName(app.plan.Directory, item.entry) + ": This role has no holder."
	if item.text != want {
		t.Errorf("problem text %q, want %q", item.text, want)
	}
}

func TestEnterOnADirectoryProblemOpensThatEntryInContacts(t *testing.T) {
	t.Parallel()
	app := firstProblem(t, vacantRolesApp(t, model.AudienceOnboarder))
	item, _ := app.review.selected()
	opened := press(app, enter)
	if opened.screen != screenContacts || opened.contacts.kind != kindRole {
		t.Fatalf("screen %v kind %v, want Contacts on roles", opened.screen, opened.contacts.kind)
	}
	selected, _ := opened.selectedEntry()
	if selected.id != item.entry.id {
		t.Errorf("selected %q, want %q", selected.id, item.entry.id)
	}
	if press(opened, esc).screen != screenReview {
		t.Error("esc from Contacts does not return to the review")
	}
}

func TestEnterOnADirectoryProblemExplainsWhenTheOnboardeeCannotEdit(t *testing.T) {
	t.Parallel()
	app := firstProblem(t, vacantRolesApp(t, model.AudienceOnboardee))
	if opened := press(app, enter); opened.screen != screenReview || !strings.Contains(opened.status.text, "onboarder") {
		t.Errorf("screen %v status %q", opened.screen, opened.status.text)
	}
}

func TestEnterOnAQuestionProblemStillOpensTheQuestion(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, blankPlan(t), model.AudienceOnboarder)
	app = resize(app, 140, 40)
	question := app.questions[0].ID
	app.review.items = []reviewItem{{kind: reviewProblem, text: "x", questionID: question}}
	app.review.cursor = 0
	if opened := press(app, enter); opened.screen != screenAnswer {
		t.Errorf("screen %v, want the Answer screen", opened.screen)
	}
}
