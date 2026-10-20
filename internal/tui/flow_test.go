package tui

import (
	"strconv"
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

func TestAPlanWithNoAnswersOpensOnTheFirstQuestion(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	if app.screen != screenWalk || app.cursor != 0 {
		t.Fatalf("expected the first question, got screen %v cursor %d", app.screen, app.cursor)
	}
	if !strings.Contains(plainView(app), "1 of ") {
		t.Fatal("the position should be shown")
	}
}

func TestAPlanWithAnswersOpensOnTheReview(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	if app.screen != screenReview {
		t.Fatalf("expected the review, got %v", app.screen)
	}
}

func TestEnterAcceptsAndMovesToTheNextQuestionInOrder(t *testing.T) {
	t.Parallel()
	app, recorded := startWalking(t)
	total := len(app.questions)
	app = press(app, typed("Globex"), enter)
	if got := app.plan.Answers.Plan["organization"]; got.Text != "Globex" {
		t.Fatalf("answer not stored: %+v", got)
	}
	if app.cursor != 1 || app.screen != screenWalk {
		t.Fatalf("expected the second question, got %d on %v", app.cursor, app.screen)
	}
	if len(recorded.saved) != 1 {
		t.Fatalf("expected one save, got %d", len(recorded.saved))
	}
	if !strings.Contains(plainView(app), "2 of "+strconv.Itoa(total)) {
		t.Fatalf("position text missing:\n%s", plainView(app))
	}
}

func TestBackGoesToThePreviousQuestion(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	app = press(app, typed("Globex"), enter)
	app = press(app, shiftTab)
	if app.cursor != 0 {
		t.Fatalf("expected the first question, got %d", app.cursor)
	}
	app = press(app, enter)
	app = moveTo(t, app, "plan.team_state")
	app = press(app, typed("p"))
	if app.cursor != app.indexOf("plan.team_state")-1 {
		t.Fatalf("p should go back, cursor %d", app.cursor)
	}
}

func TestDeferringMovesOn(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	app = moveTo(t, app, "plan.team_state")
	index := app.cursor
	app = press(app, typed("d"))
	if got := app.plan.Answers.Plan["team_state"]; got.Kind != model.ValueKindDeferred {
		t.Fatalf("expected deferred, got %+v", got)
	}
	if app.cursor != index+1 {
		t.Fatalf("expected to move on, cursor %d", app.cursor)
	}
}

func TestClearingStaysOnTheQuestion(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	app = moveTo(t, app, "plan.team_state")
	app = press(app, typed("d"), shiftTab, typed("x"))
	if _, there := app.plan.Answers.Plan["team_state"]; there {
		t.Fatal("the answer should be cleared")
	}
	if app.questions[app.cursor].ID != "plan.team_state" {
		t.Fatal("clearing should stay on the question")
	}
}

func TestFinishingTheLastQuestionLandsOnTheReview(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	app = moveTo(t, app, app.questions[len(app.questions)-1].ID)
	app = press(app, tab, enter)
	if app.screen != screenReview {
		t.Fatalf("expected the review, got %v", app.screen)
	}
}

func TestReviewListsBlockersAndJumpsToAQuestion(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	app = press(app, esc)
	if app.screen != screenOverview {
		t.Fatalf("escape from a fresh walk goes to the overview, got %v", app.screen)
	}
	app = press(app, typed("r"))
	if app.review.counts(reviewBlocker) == 0 {
		t.Fatal("a blank plan has required questions waiting")
	}
	if !strings.Contains(plainView(app), "Blockers") {
		t.Fatal("the review should name the blockers")
	}
	target := app.review.items[0].questionID
	app = press(app, enter)
	if app.screen != screenWalk || app.questions[app.cursor].ID != target {
		t.Fatalf("expected to open %s, got screen %v", target, app.screen)
	}
	app = press(app, esc)
	if app.screen != screenReview {
		t.Fatalf("escape should return to the review, got %v", app.screen)
	}
}

func TestOverviewJumpsToAnyQuestionAndEscapeReturns(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	app = press(moveTo(t, app, "plan.team_state"), typed("o"))
	if app.screen != screenOverview {
		t.Fatalf("expected the overview, got %v", app.screen)
	}
	app.overviewCursor = 0
	app = press(app, down, down, enter)
	if app.screen != screenWalk || app.cursor != 2 {
		t.Fatalf("expected the third question, got %d on %v", app.cursor, app.screen)
	}
	app = press(app, esc)
	if app.screen != screenOverview {
		t.Fatalf("escape should return to the overview, got %v", app.screen)
	}
}

func TestAChoiceIsPickedFromItsOptions(t *testing.T) {
	t.Parallel()
	app, recorded := startWalking(t)
	app = moveTo(t, app, "plan.team_state")
	labels := make([]string, 0, len(app.entry.choices))
	for _, option := range app.entry.choices {
		labels = append(labels, option.label)
	}
	if strings.Join(labels, "|") != "established|first-consultants|deferred, answer later|clear the answer" {
		t.Fatalf("unexpected choices %q", labels)
	}
	app = press(app, down, enter)
	if got := app.plan.Answers.Plan["team_state"]; got.Text != "first-consultants" {
		t.Fatalf("expected first-consultants, got %+v", got)
	}
	if len(recorded.saved) != 1 {
		t.Fatalf("expected one save, got %d", len(recorded.saved))
	}
}

func TestASecretIsRefusedAndTheQuestionStaysWithTheProblemShown(t *testing.T) {
	t.Parallel()
	app, recorded := startWalking(t)
	app = moveTo(t, app, "access.repo_url")
	app = press(app, typed("https://user:hunter2@git.example.test/r.git"), enter)
	if app.questions[app.cursor].ID != "access.repo_url" || len(app.entry.problems) == 0 {
		t.Fatalf("expected to stay with a problem, got problems %v", app.entry.problems)
	}
	if _, there := app.plan.Answers.Sections["access"]["repo_url"]; there {
		t.Fatal("the answer must not be stored")
	}
	if len(recorded.saved) != 0 {
		t.Fatal("a refused answer must not be saved")
	}
	if !strings.Contains(plainView(app), "problem:") {
		t.Fatal("the problem should be shown")
	}
}

func TestEnterOnAnUnchangedAnswerMovesOnWithoutSaving(t *testing.T) {
	t.Parallel()
	app, recorded := startWith(t, acmePlan(t), model.AudienceOnboarder)
	app = app.openQuestion(0)
	app = press(app, enter)
	if len(recorded.saved) != 0 || app.cursor != 1 {
		t.Fatalf("expected no save and the next question, got %d saves, cursor %d", len(recorded.saved), app.cursor)
	}
}

func TestSectionsAreSwitchedOffAndOn(t *testing.T) {
	t.Parallel()
	app, recorded := startWalking(t)
	app = press(moveTo(t, app, "plan.team_state"), typed("s"))
	if app.screen != screenSections {
		t.Fatalf("expected the sections screen, got %v", app.screen)
	}
	app = press(app, space)
	if isOn(app.plan.Sections[0]) {
		t.Fatal("the first section should be off")
	}
	app = press(app, space)
	if !isOn(app.plan.Sections[0]) {
		t.Fatal("the first section should be on again")
	}
	if len(recorded.saved) != 2 {
		t.Fatalf("expected two saves, got %d", len(recorded.saved))
	}
	app = press(app, esc)
	if app.screen != screenWalk {
		t.Fatalf("escape should return to the walk, got %v", app.screen)
	}
}

func TestQuittingIsOfferedOnListScreensAndNotWhileTyping(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	typing := press(app, typed("q"))
	if typing.Finished() {
		t.Fatal("q is text while typing")
	}
	if !press(app, esc, typed("q")).Finished() {
		t.Fatal("q should quit from the overview")
	}
	if !press(app, esc, typed("r"), typed("q")).Finished() {
		t.Fatal("q should quit from the review")
	}
}

func TestTheOnboardeeSeesOnlyTheirQuestionsAndNoSectionSwitches(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, blankPlan(t), model.AudienceOnboardee)
	if len(app.questions) == 0 {
		t.Fatal("the onboardee should have questions")
	}
	for _, question := range app.questions {
		if question.Audience != model.AudienceOnboardee {
			t.Errorf("%s is for %s", question.ID, question.Audience)
		}
	}
	app = moveTo(t, app, "plan.discipline")
	for _, screenKeys := range [][]string{{}, {"o"}, {"r"}} {
		shown := app
		for _, key := range screenKeys {
			shown = press(shown, typed(key))
		}
		before := shown.screen
		shown = press(shown, typed("s"))
		if shown.screen == screenSections || shown.screen != before {
			t.Fatalf("the onboardee has no section switches (from %v)", before)
		}
		if strings.Contains(plainView(shown), "sections") {
			t.Error("the onboardee hint should not offer sections")
		}
	}
}

func TestTheOnboardeeAnswersTheirOwnChoice(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, blankPlan(t), model.AudienceOnboardee)
	app = moveTo(t, app, "plan.discipline")
	app = press(app, down, enter)
	if got := app.plan.Answers.Plan["discipline"]; got.Kind != model.ValueKindText || got.Text == "" {
		t.Fatalf("expected an answer, got %+v", got)
	}
}

func TestHelpOpensAndAnyKeyClosesIt(t *testing.T) {
	t.Parallel()
	app, _ := startWalking(t)
	app = press(app, esc, typed("?"))
	if !strings.Contains(plainView(app), "Keys") {
		t.Fatal("help should list the keys")
	}
	app = press(app, typed("x"))
	if app.helpOpen {
		t.Fatal("any key closes help")
	}
}

func TestWordsAreSplitAndJoinedWithQuotes(t *testing.T) {
	t.Parallel()
	cases := map[string][]string{
		"make test":             {"make", "test"},
		`git commit -m "a b c"`: {"git", "commit", "-m", "a b c"},
		`run '' two`:            {"run", "", "two"},
		"  spaced   out  ":      {"spaced", "out"},
	}
	for text, want := range cases {
		got := splitWords(text)
		if strings.Join(got, "|") != strings.Join(want, "|") {
			t.Errorf("splitWords(%q) = %q, want %q", text, got, want)
		}
		again := splitWords(joinWords(got))
		if strings.Join(again, "|") != strings.Join(want, "|") {
			t.Errorf("joinWords round trip of %q gave %q", text, again)
		}
	}
}

func TestPeopleAreSeparatedBySemicolons(t *testing.T) {
	t.Parallel()
	question := model.Question{Type: model.VariableTypePerson, Multiple: true}
	value := parseEntry(question, "Ann Lee, ann@x.test, @ann; Bo Ray, bo@x.test, @bo ;")
	if len(value.List) != 2 || value.List[1] != "Bo Ray, bo@x.test, @bo" {
		t.Fatalf("unexpected people %q", value.List)
	}
	if parseEntry(question, "   ").Kind != model.ValueKindNone {
		t.Fatal("blank text means no answer")
	}
}
