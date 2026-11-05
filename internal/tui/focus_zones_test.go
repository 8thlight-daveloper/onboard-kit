package tui

import (
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

func hintLine(app App) string {
	rows := viewLines(app)
	return rows[len(rows)-1]
}

func statusLine(app App) string {
	rows := viewLines(app)
	return strings.TrimSpace(rows[len(rows)-2])
}

func typedGlob(t *testing.T) App {
	t.Helper()
	app, _ := startAnswering(t)
	return press(app, typed("Glob"))
}

func answerOf(app App, id string) model.Value {
	return app.questions[app.indexOf(id)].Answer
}

func TestEscapeStopsTypingAndKeepsTheText(t *testing.T) {
	t.Parallel()
	blurred := press(typedGlob(t), esc)
	if blurred.screen != screenAnswer || !blurred.entry.onOptions() || blurred.typingNow() {
		t.Fatalf("esc should stop typing and stay on the question, screen %v", blurred.screen)
	}
	view := plainView(blurred)
	for _, want := range []string{"  Not saved yet: Glob", "> Edit the answer", "Stopped typing. Your text is kept, not saved yet."} {
		if !strings.Contains(view, want) {
			t.Errorf("the blurred screen lacks %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "> Glob") {
		t.Error("the field loses its prompt mark while it is not typed in")
	}
	if !strings.HasPrefix(strings.TrimSpace(hintLine(blurred)), "Not typing.") {
		t.Errorf("the hint line says the person is not typing: %q", hintLine(blurred))
	}
	if blurred.entry.input.Value() != "Glob" {
		t.Errorf("the draft is kept, got %q", blurred.entry.input.Value())
	}
	if !strings.Contains(headerOf(blurred), "Overview o") {
		t.Errorf("the tab keys come back when typing stops: %q", headerOf(blurred))
	}
	if again := press(blurred, esc); again.screen != screenOverview {
		t.Errorf("a second esc opens the overview, got %v", again.screen)
	}
}

func TestEscapeOnAnUnchangedAnswerSaysItIsKept(t *testing.T) {
	t.Parallel()
	app, _ := startAnswering(t)
	if got := statusLine(press(app, esc)); got != "Stopped typing. The answer is kept as it was." {
		t.Errorf("status: %q", got)
	}
	if !strings.Contains(plainView(press(app, esc)), "  Nothing typed yet") {
		t.Error("an empty field still shows where it is")
	}
	if !strings.Contains(plainView(press(app, typed("Globex"), enter, ctrlP, esc)), "  Saved: Globex") {
		t.Error("a saved answer says so when typing stops")
	}
}

func TestTheBlurIsDrawnFaintWithABarOnTheChosenRow(t *testing.T) {
	withColour(t)
	typing := resize(typedGlob(t), 120, 30)
	expectBright(t, typing.View(), "Glob")
	expectFaint(t, typing.View(), "deferred, answer later")
	blurred := press(typing, esc)
	view := blurred.View()
	expectFaint(t, view, "Glob")
	expectBright(t, view, "deferred, answer later")
	if !strings.Contains(sgrBefore(t, view, "Edit the answer"), "7") {
		t.Errorf("the chosen row is drawn as a bar, got %q", sgrBefore(t, view, "Edit the answer"))
	}
}

func TestUpAndEnterOnEditReturnToTyping(t *testing.T) {
	t.Parallel()
	blurred := press(typedGlob(t), esc)
	for name, app := range map[string]App{"up": press(blurred, up), "enter": press(blurred, enter)} {
		if !app.typingNow() || app.entry.input.Value() != "Glob" {
			t.Errorf("%s: typing again with the draft, typing %v value %q", name, app.typingNow(), app.entry.input.Value())
		}
	}
	resumed := press(blurred, enter, typed("al"), enter)
	if got := resumed.plan.Answers.Plan["organization"]; got.Text != "Global" {
		t.Errorf("typing goes on from the draft, got %+v", got)
	}
}

func TestDownLeavesTheFieldAndWalksTheOptions(t *testing.T) {
	t.Parallel()
	app := typedGlob(t)
	walked := press(app, down)
	if !walked.entry.onOptions() || walked.entry.choices[walked.entry.choice].action != choiceEdit {
		t.Fatal("down leaves the field for the first row, Edit the answer")
	}
	labels := []string{}
	for range walked.entry.choices {
		labels = append(labels, walked.entry.choices[walked.entry.choice].label)
		walked = press(walked, down)
	}
	if strings.Join(labels, "|") != "Edit the answer|deferred, answer later|skip, change nothing|clear the answer" {
		t.Errorf("the rows in order: %v", labels)
	}
	if back := press(app, down, down, up, up); !back.typingNow() {
		t.Error("up from the first row returns to typing")
	}
}

func TestTheOptionsDoWhatTheySay(t *testing.T) {
	t.Parallel()
	app, recorded := startAnswering(t)
	deferred := press(app, down, down, enter)
	if got := answerOf(deferred, "plan.organization"); got.Kind != model.ValueKindDeferred || deferred.cursor != 1 {
		t.Errorf("defer saves and moves on, got %+v at %d", got, deferred.cursor)
	}
	skipped := press(app, down, down, down, enter)
	if got := answerOf(skipped, "plan.organization"); got.Kind != model.ValueKindNone || skipped.cursor != 1 {
		t.Errorf("skip changes nothing and moves on, got %+v at %d", got, skipped.cursor)
	}
	answered := press(app, typed("Globex"), enter, ctrlP)
	cleared := press(answered, down, down, down, down, enter)
	if got := answerOf(cleared, "plan.organization"); got.Kind != model.ValueKindNone || cleared.cursor != 0 || !cleared.typingNow() || cleared.entry.input.Value() != "" {
		t.Errorf("clear empties the answer and the box and stays to type, got %+v at %d", got, cleared.cursor)
	}
	if len(recorded.saved) != 3 {
		t.Errorf("defer, the answer, and clear each save once, got %d", len(recorded.saved))
	}
}

func TestLettersWorkWhenNotTyping(t *testing.T) {
	t.Parallel()
	blurred := press(typedGlob(t), esc)
	onSecond := press(startFake(suggestedBackend()), enter)
	cases := map[string]struct {
		app  App
		pass func(App) bool
	}{
		"o":     {press(blurred, typed("o")), func(app App) bool { return app.screen == screenOverview }},
		"r":     {press(blurred, typed("r")), func(app App) bool { return app.screen == screenReview }},
		"s":     {press(blurred, typed("s")), func(app App) bool { return app.screen == screenSections }},
		"c":     {press(blurred, typed("c")), func(app App) bool { return app.screen == screenContacts }},
		"d":     {press(blurred, typed("d")), func(app App) bool { return answerOf(app, "plan.organization").Kind == model.ValueKindDeferred }},
		"n":     {press(blurred, typed("n")), func(app App) bool { return app.cursor == 1 }},
		"right": {press(blurred, right), func(app App) bool { return app.cursor == 1 }},
		"p":     {press(onSecond, esc, typed("p")), func(app App) bool { return app.cursor == 0 }},
		"left":  {press(onSecond, esc, left), func(app App) bool { return app.cursor == 0 }},
		"q":     {press(blurred, typed("q")), func(app App) bool { return app.Finished() }},
		"?":     {press(blurred, typed("?")), func(app App) bool { return app.helpOpen }},
	}
	for name, each := range cases {
		if !each.pass(each.app) {
			t.Errorf("%s did not work while not typing", name)
		}
	}
}

func TestArrowsMoveTheCaretAndLettersAreTextWhileTyping(t *testing.T) {
	t.Parallel()
	app := press(typedGlob(t), left, typed("X"), right, typed("o"))
	if got := app.entry.input.Value(); got != "GloXbo" {
		t.Errorf("left and right move the caret, got %q", got)
	}
	if app.screen != screenAnswer || app.cursor != 0 {
		t.Error("letters and arrows stay in the field")
	}
}

func TestCtrlPGoesBackWhileTypingAndShiftTabMovesAPage(t *testing.T) {
	t.Parallel()
	app, _ := startAnswering(t)
	second := press(app, enter)
	if back := press(second, ctrlP); back.cursor != 0 {
		t.Errorf("ctrl+p goes back one question, got %d", back.cursor)
	}
	if page := press(second, shiftTab); page.screen == screenAnswer {
		t.Error("shift+tab moves a page, not a question")
	}
}

func TestUpFromTheFieldReachesTheTabStrip(t *testing.T) {
	t.Parallel()
	strip := press(typedGlob(t), up)
	if !strip.stripFocused || strip.typingNow() {
		t.Fatal("up from the field focuses the tab strip")
	}
	if !strings.Contains(headerOf(strip), "> [Answer]") {
		t.Errorf("the strip starts on the current tab: %q", headerOf(strip))
	}
	if !strings.Contains(hintLine(strip), "Not typing.  left/right choose a tab") {
		t.Errorf("the hint line says how to use the strip: %q", hintLine(strip))
	}
	chosen := press(strip, right, right)
	if !strings.Contains(headerOf(chosen), "> Review") {
		t.Errorf("right chooses the next tab: %q", headerOf(chosen))
	}
	if opened := press(chosen, enter); opened.screen != screenReview || opened.stripFocused {
		t.Errorf("enter opens the chosen tab, got %v", opened.screen)
	}
	for name, key := range map[string]App{"down": press(strip, down), "esc": press(strip, esc)} {
		if !key.typingNow() || key.entry.input.Value() != "Glob" {
			t.Errorf("%s returns to typing with the draft", name)
		}
	}
	if same := press(strip, enter); same.screen != screenAnswer || !same.typingNow() {
		t.Error("enter on the current tab returns to the field")
	}
	if letter := press(strip, typed("o")); letter.screen != screenOverview {
		t.Error("tab letters work in the strip")
	}
}

func TestUpFromTheFirstChoiceReachesTheTabStrip(t *testing.T) {
	t.Parallel()
	app, _ := startAnswering(t)
	choosing := moveTo(t, app, "plan.team_state")
	strip := press(choosing, up)
	if !strip.stripFocused {
		t.Fatal("up from the first choice focuses the tab strip")
	}
	if strings.Contains(plainView(strip), "> established") {
		t.Error("the choice loses its mark while the strip has focus")
	}
	if back := press(strip, down); back.stripFocused || back.entry.choice != 0 {
		t.Error("down returns to the list")
	}
}

func TestTabAndShiftTabMovePagesOnEveryScreen(t *testing.T) {
	t.Parallel()
	app := startInviting(t, newFakeInvitations(t))
	order := []screen{screenAnswer, screenOverview, screenReview, screenSections, screenContacts, screenPreparation, screenInvitation}
	current := press(app, typed("w"))
	for step := 1; step <= len(order); step++ {
		current = press(current, tab)
		if want := order[step%len(order)]; current.screen != want {
			t.Fatalf("tab %d: got %v, want %v", step, current.screen, want)
		}
	}
	for step := len(order) - 1; step >= 0; step-- {
		current = press(current, shiftTab)
		if want := order[step]; current.screen != want {
			t.Fatalf("shift+tab to %v, got %v", want, current.screen)
		}
	}
}

func TestTheNewcomerTabsAlsoMoveWithTab(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, blankPlan(t), model.AudienceOnboardee)
	if got := press(app, tab).screen; got != screenOverview {
		t.Errorf("tab from Your questions: %v", got)
	}
	if got := press(app, tab, tab, tab).screen; got != screenAnswer {
		t.Errorf("tab wraps from Review to Your questions: %v", got)
	}
	if got := press(app, shiftTab).screen; got != screenReview {
		t.Errorf("shift+tab wraps from Your questions to Review: %v", got)
	}
}

func TestTabMovesFieldsInsideAForm(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	form := press(app, typed("a"), tab)
	if form.screen != screenContacts || form.contacts.mode != contactsEditing || form.contacts.form.focus != 1 {
		t.Errorf("tab moves to the next field of the form, focus %d", form.contacts.form.focus)
	}
	if back := press(form, shiftTab); back.contacts.form.focus != 0 {
		t.Error("shift+tab moves to the previous field")
	}
}

func TestADraftSurvivesAPageVisit(t *testing.T) {
	t.Parallel()
	away := press(typedGlob(t), tab)
	if away.screen != screenOverview || statusLine(away) != "Your text on question 1 is kept, not saved yet." {
		t.Fatalf("tab leaves for the overview and says the text is kept, got %v %q", away.screen, statusLine(away))
	}
	back := press(away, shiftTab)
	if back.screen != screenAnswer || back.entry.input.Value() != "Glob" || !back.typingNow() {
		t.Errorf("the draft is back on return, got %q", back.entry.input.Value())
	}
	if escaped := press(typedGlob(t), esc, esc, typed("w")); escaped.entry.input.Value() != "Glob" {
		t.Errorf("a draft also survives esc to the overview, got %q", escaped.entry.input.Value())
	}
}

func TestLeavingTheQuestionDropsTheDraftAndSaysSo(t *testing.T) {
	t.Parallel()
	for name, app := range map[string]App{
		"defer":  press(typedGlob(t), ctrlD),
		"skip":   press(typedGlob(t), esc, typed("n")),
		"back":   press(startFake(suggestedBackend()), enter, typed("draft"), ctrlP),
		"letter": press(typedGlob(t), esc, typed("d")),
	} {
		if !strings.Contains(statusLine(app), draftDroppedText) {
			t.Errorf("%s: the status says the draft was dropped, got %q", name, statusLine(app))
		}
	}
	deferred := press(typedGlob(t), ctrlD)
	if got := answerOf(deferred, "plan.organization"); got.Kind != model.ValueKindDeferred {
		t.Errorf("defer saves the deferral, not the draft, got %+v", got)
	}
	if returned := press(deferred, ctrlP); returned.entry.input.Value() != "" {
		t.Errorf("the dropped draft does not come back, got %q", returned.entry.input.Value())
	}
	if saved := press(typedGlob(t), enter); strings.Contains(statusLine(saved), draftDroppedText) {
		t.Error("a saved answer is not a dropped draft")
	}
}

func TestClearEmptiesADraftOnAnUnansweredQuestion(t *testing.T) {
	t.Parallel()
	cleared := press(typedGlob(t), esc, typed("x"))
	if cleared.entry.input.Value() != "" || !cleared.typingNow() || cleared.cursor != 0 {
		t.Errorf("x empties the box and returns to typing, got %q", cleared.entry.input.Value())
	}
}

func TestASuggestionIsAnOptionThatFillsTheBox(t *testing.T) {
	t.Parallel()
	app := startFake(suggestedBackend())
	if !strings.Contains(plainView(app), "Use the suggestion: Ask the team lead") {
		t.Fatalf("each suggestion is a row:\n%s", plainView(app))
	}
	filled := press(app, down, down, down, enter)
	if !filled.typingNow() || filled.entry.input.Value() != "Ask the team lead" {
		t.Fatalf("enter on a suggestion puts it in the box and returns to typing, got %q", filled.entry.input.Value())
	}
	if statusLine(filled) != suggestionText {
		t.Errorf("status: %q", statusLine(filled))
	}
	if replaced := press(app, typed("mine"), down, down, enter); statusLine(replaced) != suggestionReplacedText || replaced.entry.input.Value() == "mine" {
		t.Errorf("a suggestion over a draft says it replaced it: %q", statusLine(replaced))
	}
	if saved := press(filled, enter); saved.cursor != 1 {
		t.Error("enter then saves it and moves on")
	}
	if strings.Contains(plainView(app), "Tab tries a suggestion") || strings.Contains(hintLine(app), "suggest") {
		t.Error("no key is needed for a suggestion")
	}
}

func TestTheReviewPanesSwitchWithLeftAndRight(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	app = resize(app, 100, 30)
	document := press(app, right)
	if !document.review.focusDocument || document.screen != screenReview {
		t.Fatal("right focuses the document")
	}
	if press(document, left).review.focusDocument {
		t.Error("left focuses the items")
	}
	if !strings.Contains(hintLine(app), "left/right items/document") {
		t.Errorf("the hint line names the pane keys: %q", hintLine(app))
	}
}

func TestThePreparationReviewPanesSwitchWithLeftAndRight(t *testing.T) {
	t.Parallel()
	app := onPreparation(t, withSamLee(t))
	app = press(app, enter, typed("G"), enter)
	if app.preparation.mode != preparationReviewing {
		t.Fatalf("expected the review, got mode %v", app.preparation.mode)
	}
	if guide := press(app, right); !guide.preparation.review.focusGuide {
		t.Error("right focuses the guide")
	}
	if changes := press(app, right, left); changes.preparation.review.focusGuide {
		t.Error("left focuses the changes")
	}
}

func TestHelpNamesTheNewKeys(t *testing.T) {
	t.Parallel()
	app := typedGlob(t)
	for name, each := range map[string]struct {
		app  App
		want string
	}{
		"typing":   {press(app, functionF), "stop typing, and choose from the list under the box"},
		"blurred":  {press(app, esc, typed("?")), "do what the chosen row says"},
		"strip":    {press(app, up, typed("?")), "choose a tab"},
		"overview": {press(app, esc, esc, typed("?")), "next or previous page"},
	} {
		if got := strings.Join(strings.Fields(plainView(each.app)), " "); !strings.Contains(got, each.want) {
			t.Errorf("%s help lacks %q:\n%s", name, each.want, plainView(each.app))
		}
	}
}

func TestUpFromTheFirstFieldOfAContactsFormReachesTheTabStripAndTheFormSurvives(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	form := press(app, typed("a"), typed("Dana"))
	strip := press(form, up)
	if !strip.stripFocused || strip.typingNow() {
		t.Fatal("up from the first field focuses the tab strip")
	}
	if back := press(strip, down, typed(" Lee")); back.contacts.form.name() != "Dana Lee" {
		t.Errorf("down returns to the field, got %q", back.contacts.form.name())
	}
	visited := press(strip, typed("o"), typed("c"), typed(" Lee"))
	if visited.contacts.mode != contactsEditing || visited.contacts.form.name() != "Dana Lee" {
		t.Errorf("the form survives a visit to another tab, got mode %v name %q", visited.contacts.mode, visited.contacts.form.name())
	}
	if below := press(form, tab, up); below.stripFocused || below.contacts.form.focus != 0 {
		t.Error("up from a later field moves to the field above")
	}
}

func TestUpFromTheFirstFieldOfAPreparationFormReachesTheTabStrip(t *testing.T) {
	t.Parallel()
	app := onPreparation(t, withSamLee(t))
	form := press(app, enter, enter)
	if form.preparation.mode != preparationEditing {
		t.Fatalf("expected the recipient form, got mode %v", form.preparation.mode)
	}
	strip := press(form, up)
	if !strip.stripFocused {
		t.Fatal("up from the first field focuses the tab strip")
	}
	visited := press(strip, typed("o"), typed("t"))
	if visited.preparation.mode != preparationEditing || !visited.typingNow() {
		t.Errorf("the form is still open and typing works after a visit, mode %v", visited.preparation.mode)
	}
}
