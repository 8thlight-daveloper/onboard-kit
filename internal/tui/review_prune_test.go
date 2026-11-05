package tui

import (
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

const secretLookingText = "sk-live-4242424242424242424242"

func planWithLeftOvers(t *testing.T) model.Plan {
	t.Helper()
	plan := acmePlan(t)
	plan.Answers.Plan["old_tool"] = model.Value{Kind: model.ValueKindText, Text: "Trello"}
	plan.Answers.Plan["old_token"] = model.Value{Kind: model.ValueKindText, Text: secretLookingText}
	plan.Answers.Sections = map[string]map[string]model.Value{}
	plan.Answers.Sections["retired"] = map[string]model.Value{"flag": {Kind: model.ValueKindFlag, Flag: true}}
	return plan
}

func planWithManyLeftOvers(t *testing.T, count int) model.Plan {
	t.Helper()
	plan := acmePlan(t)
	for index := 0; index < count; index++ {
		plan.Answers.Plan["old_"+strings.Repeat("x", index+1)] = model.Value{Kind: model.ValueKindText, Text: "v"}
	}
	return plan
}

func startReviewing(t *testing.T, plan model.Plan, audience model.Audience) (App, *recorder) {
	t.Helper()
	app, recorded := startWith(t, plan, audience)
	if app.screen != screenReview {
		t.Fatalf("expected the review, got %v", app.screen)
	}
	return app, recorded
}

func TestTheReviewSaysQuietlyHowManyAnswersAreLeftOver(t *testing.T) {
	withColour(t)
	app, _ := startReviewing(t, planWithLeftOvers(t), model.AudienceOnboarder)
	view := plainView(app)
	if !strings.Contains(view, "3 answers are left over") || !strings.Contains(view, "Press h") {
		t.Fatalf("the review should say 3 answers are left over and name h:\n%s", view)
	}
	expectFaint(t, app.View(), "3 answers are left over")
	if !strings.Contains(hintText(app.hints(true)), "h left over") {
		t.Fatal("the hint line should offer h while something is left over")
	}
}

func TestTheNoticeIsSingularForOneAnswer(t *testing.T) {
	plan := acmePlan(t)
	plan.Answers.Plan["old_tool"] = model.Value{Kind: model.ValueKindText, Text: "Trello"}
	app, _ := startReviewing(t, plan, model.AudienceOnboarder)
	if !strings.Contains(plainView(resize(app, 80, 24)), "1 answer is left over from a question that is no longer asked") {
		t.Fatalf("expected the singular notice:\n%s", plainView(app))
	}
}

func TestTheReviewHasNoNoticeAndNoHintWhenNothingIsLeftOver(t *testing.T) {
	app, _ := startReviewing(t, acmePlan(t), model.AudienceOnboarder)
	if strings.Contains(plainView(app), "left over") || strings.Contains(hintText(app.hints(true)), "left over") {
		t.Fatalf("nothing is left over, so nothing should be said:\n%s", plainView(app))
	}
}

func TestHWithNothingLeftOverSaysSoAndChangesNothing(t *testing.T) {
	app, recorded := startReviewing(t, acmePlan(t), model.AudienceOnboarder)
	app = press(app, typed("h"))
	if app.status.text != nothingLeftOverNotice || app.review.pruning.asking || len(recorded.saved) != 0 {
		t.Fatalf("expected a plain message and no change, got %q", app.status.text)
	}
}

func TestHShowsTheDryRunByIdAndShortValuesAndChangesNothing(t *testing.T) {
	plan := planWithLeftOvers(t)
	app, recorded := startReviewing(t, plan, model.AudienceOnboarder)
	app = press(app, typed("h"))
	view := plainView(app)
	for _, want := range []string{"Remove 3 left over answers?", "plan.old_token", "plan.old_tool", "retired.flag", `value: "Trello"`, "value: yes", "Nothing is removed until you press y."} {
		if !strings.Contains(view, want) {
			t.Fatalf("the confirmation should contain %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, secretLookingText) || strings.Contains(view, "4242424242") {
		t.Fatalf("a long value must be cut to a few characters:\n%s", view)
	}
	if !strings.Contains(view, `"sk-l…" (30 characters)`) {
		t.Fatalf("a long value shows its first characters and its length:\n%s", view)
	}
	if len(app.plan.Answers.Plan) != len(plan.Answers.Plan) || len(recorded.saved) != 0 {
		t.Fatal("the dry run must change and save nothing")
	}
	if got := hintText(app.hints(true)); !strings.Contains(got, "y remove") || !strings.Contains(got, "esc keep") {
		t.Fatalf("the hint line should offer y and esc, got %q", got)
	}
	if strings.Contains(view, "your answers") {
		t.Fatal("left over data is not the person's answers")
	}
}

func TestYRemovesTheLeftOverAnswersSavesOnceAndTheNoticeGoes(t *testing.T) {
	app, recorded := startReviewing(t, planWithLeftOvers(t), model.AudienceOnboarder)
	app = press(app, typed("h"), typed("y"))
	if app.review.pruning.asking || app.review.leftOver != 0 {
		t.Fatal("the confirmation should close and nothing should be left over")
	}
	if _, found := app.plan.Answers.Plan["old_tool"]; found {
		t.Fatal("the answer should be gone from the plan")
	}
	if _, found := app.plan.Answers.Sections["retired"]; found {
		t.Fatal("an emptied section map should go")
	}
	if len(recorded.saved) != 1 || len(recorded.saved[0].Answers.Plan) != len(app.plan.Answers.Plan) {
		t.Fatalf("expected one save of the pruned plan, got %d", len(recorded.saved))
	}
	if app.status.text != "removed 3 left over answers. saved" {
		t.Fatalf("unexpected result message %q", app.status.text)
	}
	if strings.Contains(plainView(app), "left over") && !strings.Contains(plainView(app), "removed 3 left over answers") {
		t.Fatal("the notice should be gone")
	}
	if app.screen != screenReview {
		t.Fatalf("the person stays on the review, got %v", app.screen)
	}
}

func TestEveryKeyButYKeepsTheLeftOverAnswers(t *testing.T) {
	for _, key := range []string{"enter", "esc", "n", "x", "q", "p", "j"} {
		app, recorded := startReviewing(t, planWithLeftOvers(t), model.AudienceOnboarder)
		pressed := typed(key)
		switch key {
		case "enter":
			pressed = enter
		case "esc":
			pressed = esc
		}
		app = press(app, typed("h"), pressed)
		if app.review.pruning.asking || app.review.leftOver != 3 || len(recorded.saved) != 0 || app.status.text != "nothing removed" {
			t.Fatalf("%s should keep everything, got status %q", key, app.status.text)
		}
		if app.finished {
			t.Fatalf("%s should cancel and not quit", key)
		}
	}
}

func TestQuestionMarkOpensHelpOverTheConfirmationAndKeepsIt(t *testing.T) {
	app, _ := startReviewing(t, planWithLeftOvers(t), model.AudienceOnboarder)
	app = press(app, typed("h"), typed("?"))
	if !app.helpOpen || !strings.Contains(plainView(app), "keep them and go back") {
		t.Fatal("help should open over the confirmation")
	}
	app = press(app, typed("z"))
	if !app.review.pruning.asking {
		t.Fatal("closing help should return to the confirmation")
	}
}

func TestTheNewcomerPrunesTheirOwnPlanWithTheSameWords(t *testing.T) {
	app, recorded := startReviewing(t, planWithLeftOvers(t), model.AudienceOnboardee)
	if !strings.Contains(plainView(app), "3 answers are left over") {
		t.Fatalf("the newcomer should see the notice:\n%s", plainView(app))
	}
	app = press(app, typed("h"))
	if !strings.Contains(plainView(app), "questions the plan no longer asks") {
		t.Fatalf("the confirmation should say where they came from:\n%s", plainView(app))
	}
	app = press(app, typed("y"))
	if len(recorded.saved) != 1 || app.review.leftOver != 0 {
		t.Fatal("the newcomer's plan should be pruned and saved")
	}
}

func TestALongListIsCutWithACount(t *testing.T) {
	app, _ := startReviewing(t, planWithManyLeftOvers(t, 9), model.AudienceOnboarder)
	app = resize(press(app, typed("h")), 140, 40)
	if !strings.Contains(plainView(app), "and 5 more") {
		t.Fatalf("expected a cut list:\n%s", plainView(app))
	}
}

func TestTheValuesAreShownShort(t *testing.T) {
	cases := []struct {
		want  string
		value model.Value
	}{
		{`"Trello"`, model.Value{Kind: model.ValueKindText, Text: "Trello"}},
		{`"12345678901234567890"`, model.Value{Kind: model.ValueKindText, Text: "12345678901234567890"}},
		{`"1234…" (21 characters)`, model.Value{Kind: model.ValueKindText, Text: "123456789012345678901"}},
		{"empty", model.Value{Kind: model.ValueKindText}},
		{"yes", model.Value{Kind: model.ValueKindFlag, Flag: true}},
		{"no", model.Value{Kind: model.ValueKindFlag}},
		{"deferred", model.Value{Kind: model.ValueKindDeferred}},
		{"no value", model.Value{}},
		{`"a, b"`, model.Value{Kind: model.ValueKindList, List: []string{"a", "b"}}},
		{`"alex…" (3 items)`, model.Value{Kind: model.ValueKindList, List: []string{"alex-rivera", "sam-okafor", "it-help"}}},
		{`"line one line two"`, model.Value{Kind: model.ValueKindText, Text: "line one\nline two"}},
		{`"éééé…" (22 characters)`, model.Value{Kind: model.ValueKindText, Text: strings.Repeat("é", 22)}},
	}
	for _, each := range cases {
		if got := shortValue(each.value); got != each.want {
			t.Errorf("shortValue(%v) = %q, want %q", each.value, got, each.want)
		}
	}
}

func TestHIsTextWhileTypingAndDoesNothingOnTheOtherScreens(t *testing.T) {
	app, recorded := startWith(t, planWithLeftOvers(t), model.AudienceOnboarder)
	answering := moveTo(t, press(app, esc), "plan.team_state")
	typing := press(moveTo(t, app, "access.repo_url"), typed("h"))
	if typing.review.pruning.asking || typing.screen != screenAnswer {
		t.Fatal("h is text while typing")
	}
	overview := press(answering.openOverview(), typed("h"))
	if overview.screen != screenOverview || overview.review.pruning.asking {
		t.Fatal("h does nothing on the overview")
	}
	if len(recorded.saved) != 0 {
		t.Fatal("nothing should be saved")
	}
}

func TestTheConfirmationFitsEveryWindow(t *testing.T) {
	asking := press(startReviewingFrom(t, planWithLeftOvers(t)), typed("h"))
	crowded := press(startReviewingFrom(t, planWithManyLeftOvers(t, 9)), typed("h"))
	noticed := startReviewingFrom(t, planWithLeftOvers(t))
	for name, app := range map[string]App{"asking": asking, "crowded": crowded, "noticed": noticed} {
		for _, width := range widths {
			for _, height := range heights {
				sized := resize(app, width, height)
				for _, row := range viewLines(sized) {
					if cellWidth(row) > width {
						t.Fatalf("%s at %dx%d has a row of %d cells: %q", name, width, height, cellWidth(row), row)
					}
				}
			}
		}
	}
}

func startReviewingFrom(t *testing.T, plan model.Plan) App {
	t.Helper()
	app, _ := startReviewing(t, plan, model.AudienceOnboarder)
	return app
}

func TestPruningWithTheFakeBackendUsesThePlanTheBackendReturns(t *testing.T) {
	plan := model.Plan{Answers: model.Answers{Plan: map[string]model.Value{"gone": {Kind: model.ValueKindText, Text: "x"}}}}
	app := startFakeIn(newFakeBackend(), plan, model.AudienceOnboarder)
	app = press(app.openReview(), typed("h"))
	if len(app.review.pruning.ids) != 1 {
		t.Fatalf("expected one left over answer, got %v", app.review.pruning.ids)
	}
	app = press(app, typed("y"))
	if len(app.plan.Answers.Plan) != 0 {
		t.Fatal("the backend's plan should replace the old one")
	}
}
