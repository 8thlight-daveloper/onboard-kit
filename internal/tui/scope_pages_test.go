package tui

import (
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

func planPageBackend() *fakeBackend {
	repository := textQuestion("plan.plan_repository", "plan", "Where is the plan's repository?", "https://example.test/acme/onboarding.git")
	repository.Required = false
	file := textQuestion("plan.plan_file", "plan", "Which file in it is the plan?", "plans/platform.obk.yaml")
	file.Required = false
	return newFakeBackend(
		textQuestion("plan.organization", "plan", "What is the organization's name?", "ACME Corp"),
		textQuestion("plan.team_name", "plan", "What is the team's name?", "Platform"),
		repository,
		file,
		directoryQuestion("access.owner", "access", "Who owns access?", model.VariableTypeContact, false, people()...),
		textQuestion("access.request_url", "access", "Where do people request access?", ""),
		textQuestion("tools.editor", "tools", "Which editor is standard?", ""),
	)
}

func onPlanPage() (App, *fakeBackend) {
	backend := planPageBackend()
	return resize(startFake(backend), 120, 40), backend
}

func focusedID(app App) string { return app.questions[app.cursor].ID }

func TestThePlanPageShowsItsFourRowsTogether(t *testing.T) {
	t.Parallel()
	app, _ := onPlanPage()
	view := plainView(app)
	for _, want := range []string{
		"Plan",
		"0 of 4 answered on this page",
		"Question 1 of 7, required, not answered yet",
		"* 1. What is the organization's name?",
		"* 2. What is the team's name?",
		"  3. Where is the plan's repository?",
		"  4. Which file in it is the plan?",
		"e.g. https://example.test/acme/onboarding.git",
		"Required, not answered yet",
		"Not answered yet",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("the plan page lacks %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "Who owns access?") {
		t.Error("a question of another scope is on another page")
	}
}

func TestEnterAnswersThePlanPageInOrderAndThenOpensTheNextPage(t *testing.T) {
	t.Parallel()
	app, backend := onPlanPage()
	app = press(app, typed("Globex"), enter, typed("Platform"), enter)
	if focusedID(app) != "plan.plan_repository" || !app.typingNow() {
		t.Fatalf("enter moves to the next row and types there, got %s", focusedID(app))
	}
	if view := plainView(app); !strings.Contains(view, "Saved: Globex") || !strings.Contains(view, "Saved: Platform") || !strings.Contains(view, "2 of 4 answered on this page") {
		t.Errorf("the answered rows show their answers:\n%s", view)
	}
	app = press(app, typed("https://example.test/globex/onboarding.git"), enter, typed("plans/core.obk.yaml"), enter)
	if focusedID(app) != "access.owner" || app.screen != screenAnswer {
		t.Fatalf("enter after the last row opens the first row of the next page, got %s", focusedID(app))
	}
	if strings.Join(backend.submitted, " ") != "plan.organization plan.team_name plan.plan_repository plan.plan_file" {
		t.Errorf("each row is saved in order, got %v", backend.submitted)
	}
	if view := plainView(app); !strings.Contains(view, "Who owns access?") || strings.Contains(view, "What is the team's name?") {
		t.Errorf("the next page shows only its own rows:\n%s", view)
	}
}

func TestUpAndDownWalkOneLinearOrder(t *testing.T) {
	t.Parallel()
	app, _ := onPlanPage()
	options := press(app, down)
	if focusedID(options) != "plan.organization" || !options.entry.onOptions() || options.entry.choice != 0 {
		t.Fatal("down from the first row's field reaches that row's options")
	}
	lastOption := press(options, down, down, down)
	if lastOption.entry.choices[lastOption.entry.choice].label != clearLabel {
		t.Fatalf("the options end with clear, got %q", lastOption.entry.choices[lastOption.entry.choice].label)
	}
	second := press(lastOption, down)
	if focusedID(second) != "plan.team_name" || !second.typingNow() {
		t.Fatalf("down after the last option reaches the next row's field, got %s", focusedID(second))
	}
	backUp := press(second, up)
	if focusedID(backUp) != "plan.organization" || !backUp.entry.onOptions() || backUp.entry.choices[backUp.entry.choice].label != clearLabel {
		t.Fatalf("up from a row's field reaches the last option of the row above, got %s", focusedID(backUp))
	}
	field := press(backUp, up, up, up, up)
	if focusedID(field) != "plan.organization" || !field.typingNow() {
		t.Fatal("up through the options returns to the first row's field")
	}
	strip := press(field, up)
	if !strip.stripFocused {
		t.Fatal("up from the first row's field reaches the tab strip")
	}
	if returned := press(strip, down); returned.stripFocused || !returned.typingNow() || focusedID(returned) != "plan.organization" {
		t.Error("down from the tab strip returns to the first row")
	}
	last := press(app, typed("n"), down, down, down, down, down, down, down, down, down, down, down, down, down, down, down, down, down, down, down, down)
	if focusedID(last) != "plan.plan_file" || !last.entry.onOptions() || last.screen != screenAnswer {
		t.Errorf("down stops at the end of the page, got %s on %v", focusedID(last), last.screen)
	}
}

func TestAChoiceRowIsPartOfTheLinearOrder(t *testing.T) {
	t.Parallel()
	app, _ := onPlanPage()
	owner := moveTo(t, app, "access.owner")
	if strip := press(owner, up); !strip.stripFocused {
		t.Error("up from the first choice of the first row reaches the tab strip")
	}
	below := press(owner, down, down, down, down, down)
	if focusedID(below) != "access.request_url" || !below.typingNow() {
		t.Fatalf("down past the last choice reaches the next row's field, got %s", focusedID(below))
	}
	above := press(below, up)
	if focusedID(above) != "access.owner" || above.entry.choice != len(above.entry.choices)-1 {
		t.Errorf("up from the field reaches the last choice of the row above, got %s at %d", focusedID(above), above.entry.choice)
	}
	acme, _ := startAnswering(t)
	teamState := moveTo(t, acme, "plan.team_state")
	if up := press(teamState, up); focusedID(up) != "plan.organization" || !up.entry.onOptions() {
		t.Errorf("up from the first choice of a later row reaches the row above, got %s", focusedID(up))
	}
	view := plainView(teamState)
	if !strings.Contains(view, "> established") || !strings.Contains(view, "* 1. What is the organization's name?") {
		t.Errorf("the focused choice row shows its list under the row, beside the other rows:\n%s", view)
	}
}

func TestPAndNMoveBetweenRowsWhenNotTyping(t *testing.T) {
	t.Parallel()
	app, _ := onPlanPage()
	next := press(app, esc, typed("n"))
	if focusedID(next) != "plan.team_name" || next.screen != screenAnswer {
		t.Fatalf("n moves to the next row, got %s", focusedID(next))
	}
	if back := press(next, esc, typed("p")); focusedID(back) != "plan.organization" {
		t.Errorf("p moves to the previous row, got %s", focusedID(back))
	}
	if typing := press(app, typed("n")); typing.entry.input.Value() != "n" || focusedID(typing) != "plan.organization" {
		t.Error("n is text while typing")
	}
	if skipped := press(app, ctrlN); focusedID(skipped) != "plan.team_name" {
		t.Error("ctrl+n moves to the next row while typing")
	}
}

func TestEscapeBlursTheRowAndASecondEscapeLeaves(t *testing.T) {
	t.Parallel()
	app, _ := onPlanPage()
	second := press(app, enter, typed("Plat"))
	blurred := press(second, esc)
	if !blurred.entry.onOptions() || focusedID(blurred) != "plan.team_name" {
		t.Fatal("esc stops typing on the same row")
	}
	view := plainView(blurred)
	for _, want := range []string{"  Not saved yet: Plat", "> Edit the answer", "Stopped typing."} {
		if !strings.Contains(view, want) {
			t.Errorf("the blurred row lacks %q:\n%s", want, view)
		}
	}
	if left := press(blurred, esc); left.screen != screenOverview {
		t.Errorf("a second esc opens the overview, got %v", left.screen)
	}
	if page := press(second, tab); page.screen != screenOverview {
		t.Errorf("tab moves to the next tab, not the next row, got %v", page.screen)
	}
}

func TestDeferSkipAndClearActOnTheFocusedRow(t *testing.T) {
	t.Parallel()
	app, backend := onPlanPage()
	deferred := press(app, enter, esc, typed("d"))
	if backend.answers["plan.team_name"].Kind != model.ValueKindDeferred || focusedID(deferred) != "plan.plan_repository" {
		t.Fatalf("d defers the focused row and moves to the next row, got %s", focusedID(deferred))
	}
	if _, touched := backend.answers["plan.organization"]; touched {
		t.Error("d leaves the other rows alone")
	}
	if !strings.Contains(plainView(deferred), "  "+deferredText) {
		t.Errorf("a deferred row says so:\n%s", plainView(deferred))
	}
	cleared := press(deferred, ctrlP, esc, typed("x"))
	if backend.answers["plan.team_name"].Kind != model.ValueKindNone || focusedID(cleared) != "plan.team_name" {
		t.Errorf("x clears the focused row and stays, got %s", focusedID(cleared))
	}
	skipped := press(cleared, esc, down, down, enter)
	if focusedID(skipped) != "plan.plan_repository" || backend.answers["plan.team_name"].Kind != model.ValueKindNone {
		t.Errorf("skip moves to the next row and changes nothing, got %s", focusedID(skipped))
	}
}

func TestADraftStaysInItsRowWhileThePersonIsOnThePage(t *testing.T) {
	t.Parallel()
	app, _ := onPlanPage()
	drafted := press(app, typed("Glob"), down, down, down, down, down, typed("Plat"))
	if focusedID(drafted) != "plan.team_name" {
		t.Fatalf("expected the second row, got %s", focusedID(drafted))
	}
	if view := plainView(drafted); !strings.Contains(view, "  Not saved yet: Glob") {
		t.Errorf("the row above shows its draft, dimmed:\n%s", view)
	}
	returned := press(drafted, up, up, up, up, up)
	if focusedID(returned) != "plan.organization" || !returned.typingNow() || returned.entry.input.Value() != "Glob" {
		t.Fatalf("the first row's draft is back in its field, got %q", returned.entry.input.Value())
	}
	if !strings.Contains(plainView(returned), "  Not saved yet: Plat") {
		t.Error("the second row keeps its draft too")
	}
	for name, moved := range map[string]App{
		"n and p":          press(drafted, esc, typed("n"), typed("p")),
		"enter on the row": press(returned, typed("al"), enter),
		"overview same":    press(drafted, esc, esc, typed("w")),
	} {
		if !strings.Contains(plainView(moved), "Plat") {
			t.Errorf("%s: the second row's draft is kept on the page:\n%s", name, plainView(moved))
		}
	}
	saved := press(returned, typed("al"), enter)
	if focusedID(saved) != "plan.team_name" || saved.entry.input.Value() != "Plat" {
		t.Errorf("enter saves the row and the next row resumes its draft, got %q", saved.entry.input.Value())
	}
}

func TestLeavingAPageDropsItsDraftsAndATabVisitKeepsThem(t *testing.T) {
	t.Parallel()
	app, backend := onPlanPage()
	drafted := press(app, typed("Glob"), down, down, down, down, down, typed("Plat"), ctrlN, ctrlN)
	if focusedID(drafted) != "plan.plan_file" {
		t.Fatalf("expected the fourth row, got %s", focusedID(drafted))
	}
	left := press(drafted, typed("plans/a.yaml"), enter)
	if focusedID(left) != "access.owner" {
		t.Fatalf("enter on the last row opens the next page, got %s", focusedID(left))
	}
	if got := statusLine(left); got != "saved. warning: your unsaved text on questions 1 and 2 was dropped." {
		t.Errorf("leaving the page names the dropped drafts, got %q", got)
	}
	if back := press(left, ctrlN, esc, typed("p"), typed("p")); focusedID(back) != "plan.plan_file" || strings.Contains(plainView(back), "Not saved yet") {
		t.Error("a dropped draft does not come back")
	}
	if _, saved := backend.answers["plan.organization"]; saved {
		t.Error("a draft is never saved by leaving")
	}
	visited := press(app, typed("Glob"), down, down, down, down, down, typed("Plat"), tab)
	if visited.screen != screenOverview || statusLine(visited) != "Your text on questions 1 and 2 is kept, not saved yet." {
		t.Fatalf("a tab visit keeps the drafts and says so, got %v %q", visited.screen, statusLine(visited))
	}
	returned := press(visited, shiftTab)
	if returned.entry.input.Value() != "Plat" || !strings.Contains(plainView(returned), "Not saved yet: Glob") {
		t.Errorf("both drafts are back after the visit:\n%s", plainView(returned))
	}
	jumped := press(visited, typed("G"), enter)
	if focusedID(jumped) != "tools.editor" || statusLine(jumped) != "warning: your unsaved text on questions 1 and 2 was dropped." {
		t.Errorf("opening a question of another page drops the drafts, got %s %q", focusedID(jumped), statusLine(jumped))
	}
}

func TestDeferOnADraftedRowDropsOnlyThatRowsDraft(t *testing.T) {
	t.Parallel()
	app, backend := onPlanPage()
	deferred := press(app, typed("Glob"), down, down, down, down, down, typed("Plat"), ctrlD)
	if backend.answers["plan.team_name"].Kind != model.ValueKindDeferred {
		t.Fatal("ctrl+d defers the second row")
	}
	if got := statusLine(deferred); got != "saved. warning: your unsaved text on question 2 was dropped." {
		t.Errorf("status %q", got)
	}
	if !strings.Contains(plainView(deferred), "Not saved yet: Glob") {
		t.Error("the first row keeps its draft")
	}
}

func conditionalBackend() *fakeBackend {
	uses := textQuestion("plan.uses_repository", "plan", "Is the plan kept in a repository?", "")
	uses.Type = model.VariableTypeBoolean
	backend := newFakeBackend(
		uses,
		textQuestion("plan.plan_repository", "plan", "Where is the plan's repository?", ""),
		textQuestion("plan.plan_file", "plan", "Which file in it is the plan?", ""),
		textQuestion("tools.editor", "tools", "Which editor is standard?", ""),
	)
	backend.askIf = map[string]func(map[string]model.Value) bool{
		"plan.plan_repository": func(answers map[string]model.Value) bool {
			return answers["plan.uses_repository"].Kind == model.ValueKindFlag && answers["plan.uses_repository"].Flag
		},
	}
	return backend
}

func TestARowThatBecomesActiveAppearsInItsPlace(t *testing.T) {
	t.Parallel()
	app := resize(startFake(conditionalBackend()), 120, 40)
	if strings.Contains(plainView(app), "Where is the plan's repository?") || !strings.Contains(plainView(app), "of 2 answered on this page") {
		t.Fatalf("the follow-up is hidden until it applies:\n%s", plainView(app))
	}
	drafted := press(app, down, down, down, down, typed("plans/a.yaml"), up, up, up, up)
	if focusedID(drafted) != "plan.uses_repository" {
		t.Fatalf("expected the first row, got %s", focusedID(drafted))
	}
	answered := press(drafted, enter)
	if focusedID(answered) != "plan.plan_repository" || !answered.typingNow() {
		t.Fatalf("the new row appears after the answered one and takes focus as the next row, got %s", focusedID(answered))
	}
	view := plainView(answered)
	if !strings.Contains(view, "1 of 3 answered on this page") || !strings.Contains(view, "Not saved yet: plans/a.yaml") {
		t.Errorf("the page shows the new row and keeps the draft below it:\n%s", view)
	}
	hidden := press(answered, ctrlP, down, enter)
	if focusedID(hidden) != "plan.plan_file" || hidden.entry.input.Value() != "plans/a.yaml" {
		t.Errorf("answering no hides the follow-up and the next row is the file, with its draft, got %s %q", focusedID(hidden), hidden.entry.input.Value())
	}
	if strings.Contains(plainView(hidden), "Where is the plan's repository?") {
		t.Error("the hidden row leaves the page")
	}
}

func TestARowThatDisappearsAboveDoesNotMoveTheFocus(t *testing.T) {
	t.Parallel()
	backend := conditionalBackend()
	backend.answers["plan.uses_repository"] = model.Value{Kind: model.ValueKindFlag, Flag: true}
	backend.askIf["plan.uses_repository"] = func(answers map[string]model.Value) bool {
		return answers["plan.plan_file"].Text != "hide"
	}
	app := resize(startFake(backend), 120, 40).openQuestion(2)
	answered := press(app, typed("hide"), esc, typed("p"))
	if focusedID(answered) != "plan.plan_repository" {
		t.Fatalf("expected the repository row, got %s", focusedID(answered))
	}
	saved := press(app, typed("hide"), enter)
	if focusedID(saved) != "tools.editor" {
		t.Errorf("enter on the last row still opens the next page after a row above went away, got %s", focusedID(saved))
	}
	cleared := press(moveTo(t, saved, "plan.plan_file"), typed("x"), ctrlP)
	if focusedID(cleared) != "plan.plan_repository" {
		t.Errorf("p still moves one row up, got %s", focusedID(cleared))
	}
}

func TestTheNewcomerSeesThePlanRowsOnOnePage(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, blankPlan(t), model.AudienceOnboardee)
	view := plainView(app)
	for _, want := range []string{"0 of 2 answered on this page", "Which team are you joining?", "What kind of work will you do?"} {
		if !strings.Contains(view, want) {
			t.Errorf("the newcomer's plan page lacks %q:\n%s", want, view)
		}
	}
	second := press(app, enter)
	if focusedID(second) != "plan.discipline" || second.plan.Answers.Plan["team"].Text == "" {
		t.Fatalf("enter picks and moves to the next row, got %s", focusedID(second))
	}
	if !strings.Contains(plainView(second), "Saved: Platform team") {
		t.Errorf("the answered row shows the name:\n%s", plainView(second))
	}
	if next := press(second, enter); next.questions[next.cursor].Scope == "plan" {
		t.Error("enter on the last row opens the next page")
	}
	if up := press(second, up); focusedID(up) != "plan.team" {
		t.Error("up from the first choice of the second row reaches the row above")
	}
}

func TestTheHintsAndHelpNameTheRows(t *testing.T) {
	t.Parallel()
	app, _ := onPlanPage()
	if !strings.Contains(hintLine(app), "up tabs") {
		t.Errorf("on the first row up opens the tabs: %q", hintLine(app))
	}
	second := press(app, enter)
	if !strings.Contains(hintLine(second), "up row above") {
		t.Errorf("on a later row up reaches the row above: %q", hintLine(second))
	}
	for name, help := range map[string]App{
		"typing":   press(second, functionF),
		"blurred":  press(second, esc, typed("?")),
		"choosing": press(moveTo(t, app, "access.owner"), typed("?")),
	} {
		if text := strings.Join(strings.Fields(plainView(help)), " "); !strings.Contains(text, "row") {
			t.Errorf("%s help names the rows:\n%s", name, plainView(help))
		}
	}
}

func TestALongPageScrollsToTheFocusedRowAtEveryWidth(t *testing.T) {
	t.Parallel()
	for name, app := range map[string]App{
		"field":   longPage().openQuestion(9),
		"choice":  longPage().openQuestion(7),
		"options": press(longPage().openQuestion(6), esc, down, down, down),
	} {
		for _, width := range []int{60, 80, 100, 120, 140} {
			for _, height := range []int{12, 16, 24, 40} {
				view := plainView(resize(app, width, height))
				if !strings.Contains(view, "> ") {
					t.Errorf("%s at %dx%d: the focused row is not on screen:\n%s", name, width, height, view)
				}
				if height >= 16 && !strings.Contains(view, "more questions above") {
					t.Errorf("%s at %dx%d: the page says rows are above:\n%s", name, width, height, view)
				}
			}
		}
	}
}

func TestARowWithNothingToPickIsStillAStopOnTheWay(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, withoutDirectory(blankPlan(t)), model.AudienceOnboarder)
	empty := moveTo(t, app, "access.onboarding_owner_role")
	if below := press(empty, down); focusedID(below) != "access.access_request_url" || !below.typingNow() {
		t.Errorf("down from a row with nothing to pick reaches the next row, got %s", focusedID(below))
	}
	if strip := press(empty, up); !strip.stripFocused {
		t.Error("up from the page's first row reaches the tab strip even with nothing to pick")
	}
	if back := press(empty, down, up); focusedID(back) != "access.onboarding_owner_role" {
		t.Errorf("up from the next row's field returns to the row with nothing to pick, got %s", focusedID(back))
	}
}
