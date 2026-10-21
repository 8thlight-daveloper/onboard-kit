package tui

import (
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

func pickingApp(question model.Question, answer model.Value) (App, *fakeBackend) {
	backend := newFakeBackend(question, textQuestion("plan.next", "plan", "The next question", ""))
	if answer.Kind != model.ValueKindNone {
		backend.answers[question.ID] = answer
	}
	app := resize(startFakeIn(backend, model.Plan{}, model.AudienceOnboarder), 120, 30)
	return app.openQuestion(0), backend
}

func contactList(many bool) model.Question {
	return directoryQuestion("plan.contact", "plan", "Who answers first?", model.VariableTypeContact, many, people()...)
}

func texts(ids ...string) model.Value {
	return model.Value{Kind: model.ValueKindList, List: ids}
}

func id(text string) model.Value { return model.Value{Kind: model.ValueKindText, Text: text} }

func TestASingleDirectoryQuestionIsPickedByNameAndStoresTheId(t *testing.T) {
	t.Parallel()
	app, backend := pickingApp(contactList(false), model.Value{})
	view := plainView(app)
	for _, want := range []string{"> Alex Rivera", "Sam Okafor", "IT service desk", "deferred, answer later", "clear the answer", "enter choose"} {
		if !strings.Contains(view, want) {
			t.Errorf("missing %q:\n%s", want, view)
		}
	}
	for _, hidden := range []string{"alex-rivera", "sam-okafor", "it-help"} {
		if strings.Contains(strings.Split(view, "About this question")[0], hidden) {
			t.Errorf("the pick list must not show the id %q", hidden)
		}
	}
	app = press(app, down, enter)
	if got := backend.answers["plan.contact"]; !sameValue(got, id("sam-okafor")) {
		t.Fatalf("expected the id of the second contact, got %+v", got)
	}
	if app.cursor != 1 {
		t.Fatalf("expected to move on, cursor %d", app.cursor)
	}
}

func TestTheSingleListStartsOnTheCurrentAnswerAndMarksIt(t *testing.T) {
	t.Parallel()
	app, backend := pickingApp(contactList(false), id("sam-okafor"))
	if !strings.Contains(plainView(app), "> Sam Okafor (current)") {
		t.Fatalf("the cursor should start on the current answer:\n%s", plainView(app))
	}
	press(app, enter)
	if len(backend.submitted) != 0 {
		t.Fatal("enter on the current answer saves nothing")
	}
}

func TestAMultipleDirectoryQuestionKeepsTheOrderOfThePicks(t *testing.T) {
	t.Parallel()
	app, backend := pickingApp(contactList(true), model.Value{})
	app = press(app, down, down, space, typed("k"), typed("k"), space)
	view := plainView(app)
	for _, want := range []string{"[2] Alex Rivera", "[ ] Sam Okafor", "[1] IT service desk", "Order: IT service desk, then Alex Rivera", "space pick", "enter confirm"} {
		if !strings.Contains(view, want) {
			t.Errorf("missing %q:\n%s", want, view)
		}
	}
	app = press(app, enter)
	got := backend.answers["plan.contact"]
	if got.Kind != model.ValueKindList || strings.Join(got.List, "|") != "it-help|alex-rivera" {
		t.Fatalf("expected the ids in the order picked, got %+v", got)
	}
	if app.cursor != 1 {
		t.Fatalf("expected to move on, cursor %d", app.cursor)
	}
}

func TestUnpickingRemovesAPickAndRenumbersTheRest(t *testing.T) {
	t.Parallel()
	app, _ := pickingApp(contactList(true), model.Value{})
	app = press(app, space, down, space, typed("k"), space)
	view := plainView(app)
	if !strings.Contains(view, "[ ] Alex Rivera") || !strings.Contains(view, "[1] Sam Okafor") {
		t.Fatalf("unpicking should renumber:\n%s", view)
	}
	if strings.Contains(view, "Order:") {
		t.Error("one pick needs no order line")
	}
}

func TestAMultipleListStartsFromTheCurrentPicksInTheirOrder(t *testing.T) {
	t.Parallel()
	app, backend := pickingApp(contactList(true), texts("sam-okafor", "gone", "alex-rivera"))
	view := plainView(app)
	if !strings.Contains(view, "[1] Sam Okafor") || !strings.Contains(view, "[2] Alex Rivera") {
		t.Fatalf("the current picks should show in order:\n%s", view)
	}
	app = press(app, enter)
	if got := backend.answers["plan.contact"]; strings.Join(got.List, "|") != "sam-okafor|alex-rivera" {
		t.Fatalf("a stale id is dropped when the picks are confirmed, got %+v", got)
	}
}

func TestConfirmingNoPicksMovesOnOrClearsAnExistingAnswer(t *testing.T) {
	t.Parallel()
	app, backend := pickingApp(contactList(true), model.Value{})
	app = press(app, enter)
	if len(backend.submitted) != 0 || app.cursor != 1 {
		t.Fatalf("nothing picked and nothing answered moves on, saved %v, cursor %d", backend.submitted, app.cursor)
	}
	app, backend = pickingApp(contactList(true), texts("sam-okafor"))
	app = press(app, space, enter)
	if got := backend.answers["plan.contact"]; got.Kind != model.ValueKindNone || app.cursor != 1 {
		t.Fatalf("unpicking everything and confirming clears the answer, got %+v", got)
	}
}

func TestTheOtherKeysStillWorkOnAPickList(t *testing.T) {
	t.Parallel()
	for _, many := range []bool{false, true} {
		app, backend := pickingApp(contactList(many), model.Value{})
		deferred := press(app, typed("d"))
		if backend.answers["plan.contact"].Kind != model.ValueKindDeferred || deferred.cursor != 1 {
			t.Errorf("d defers and moves on (many %v)", many)
		}
		if press(app, typed("n")).cursor != 1 {
			t.Errorf("n skips (many %v)", many)
		}
		if press(app, typed("q")).Finished() != true {
			t.Errorf("q quits (many %v)", many)
		}
	}
	app, backend := pickingApp(contactList(true), texts("sam-okafor"))
	press(app, typed("x"))
	if backend.answers["plan.contact"].Kind != model.ValueKindNone {
		t.Error("x clears")
	}
}

func TestSameNamedEntriesAreToldApartByTheirId(t *testing.T) {
	t.Parallel()
	question := directoryQuestion("plan.contact", "plan", "Who?", model.VariableTypeContact, false,
		model.Choice{Value: "sam-1", Label: "Sam Okafor"}, model.Choice{Value: "sam-2", Label: "Sam Okafor"}, model.Choice{Value: "alex", Label: "Alex Rivera"})
	app, _ := pickingApp(question, id("sam-2"))
	view := plainView(app)
	for _, want := range []string{"Sam Okafor (sam-1)", "Sam Okafor (sam-2) (current)", "Alex Rivera"} {
		if !strings.Contains(view, want) {
			t.Errorf("missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "Alex Rivera (alex)") {
		t.Error("a unique name needs no id")
	}
}

func TestAnEmptyDirectoryListSaysSoForTheOnboarder(t *testing.T) {
	t.Parallel()
	for variableType, noun := range map[model.VariableType]string{model.VariableTypeContact: "contacts", model.VariableTypeRole: "roles", model.VariableTypeTeam: "teams"} {
		for _, many := range []bool{false, true} {
			app, backend := pickingApp(directoryQuestion("plan.contact", "plan", "Who?", variableType, many), model.Value{})
			view := strings.Join(strings.Fields(plainView(resize(app, 80, 30))), " ")
			want := "No " + noun + " to pick from yet. Add them on the Contacts screen, press c."
			if !strings.Contains(view, want) {
				t.Errorf("%s: missing %q:\n%s", noun, want, view)
			}
			app = press(app, enter)
			if app.cursor != 1 || len(backend.submitted) != 0 {
				t.Errorf("%s: enter on an empty list moves on and saves nothing", noun)
			}
		}
	}
}

func TestAnEmptyDirectoryListTellsTheOnboardeeToAskTheOnboarder(t *testing.T) {
	t.Parallel()
	question := directoryQuestion("plan.team", "plan", "Which team are you joining?", model.VariableTypeTeam, false)
	question.Audience = model.AudienceOnboardee
	app := resize(startFakeIn(newFakeBackend(question), model.Plan{}, model.AudienceOnboardee), 120, 30)
	if !strings.Contains(strings.Join(strings.Fields(plainView(resize(app, 80, 30))), " "), "Ask your onboarder to add this.") {
		t.Fatalf("the onboardee is told who adds it:\n%s", plainView(app))
	}
	if strings.Contains(plainView(app), "directory") {
		t.Error("the onboardee needs no word about the directory")
	}
}

func TestAnEmptyListStillAllowsDeferring(t *testing.T) {
	t.Parallel()
	app, backend := pickingApp(directoryQuestion("plan.contact", "plan", "Who?", model.VariableTypeContact, false), model.Value{})
	app = press(app, typed("d"))
	if backend.answers["plan.contact"].Kind != model.ValueKindDeferred {
		t.Fatal("a person who must wait for the directory can defer")
	}
}

func TestANamedIdShowsItsNameOnTheOverviewAndNotTheId(t *testing.T) {
	t.Parallel()
	app, _ := pickingApp(contactList(true), texts("sam-okafor", "alex-rivera"))
	app = app.openOverview()
	view := plainView(app)
	if !strings.Contains(view, "Sam Okafor; Alex Rivera") || !strings.Contains(view, "[x] *") {
		t.Fatalf("the overview shows names in order:\n%s", view)
	}
	for _, hidden := range []string{"sam-okafor", "alex-rivera"} {
		if strings.Contains(view, hidden) {
			t.Errorf("the overview shows the id %q", hidden)
		}
	}
}

func TestAnAnswerNamingAnIdOutsideTheDirectoryIsNotAnswered(t *testing.T) {
	t.Parallel()
	for _, answer := range []model.Value{id("gone"), texts("gone", "also-gone")} {
		app, _ := pickingApp(contactList(false), answer)
		if answerState(app.questions[0]) != "unanswered" {
			t.Fatalf("%+v is stale and counts as not answered", answer)
		}
		view := plainView(app)
		for _, want := range []string{"not answered yet", "0 of 2 answered"} {
			if !strings.Contains(view, want) {
				t.Errorf("walk lacks %q:\n%s", want, view)
			}
		}
		if strings.Contains(view, "gone") || strings.Contains(view, "(current)") {
			t.Errorf("a stale id is not shown:\n%s", view)
		}
		overview := plainView(app.openOverview())
		if !strings.Contains(overview, "[ ] *") || strings.Contains(overview, "gone") {
			t.Errorf("the overview shows a stale answer as not answered:\n%s", overview)
		}
	}
}

func TestADeferredDirectoryQuestionStaysDeferred(t *testing.T) {
	t.Parallel()
	app, _ := pickingApp(contactList(false), model.Value{Kind: model.ValueKindDeferred})
	if answerState(app.questions[0]) != "deferred" || !strings.Contains(plainView(app), "deferred") {
		t.Fatal("a deferred answer is not stale")
	}
}

func TestAPartlyStaleListShowsOnlyTheNamesItCanResolve(t *testing.T) {
	t.Parallel()
	app, _ := pickingApp(contactList(true), texts("gone", "sam-okafor"))
	if answerState(app.questions[0]) != "settled" {
		t.Fatal("one known id is enough to be answered")
	}
	if got := describeAnswer(app.questions[0], app.questions[0].Answer); got != "Sam Okafor" {
		t.Fatalf("got %q", got)
	}
}

func TestSuggestionsOnADirectoryQuestionShowNames(t *testing.T) {
	t.Parallel()
	question := contactList(false)
	question.Suggestions = []model.Suggestion{{Value: id("sam-okafor"), Source: model.SuggestionSourceDefault}, {Value: id("gone"), Source: model.SuggestionSourceGiven}}
	app, _ := pickingApp(question, model.Value{})
	view := plainView(app)
	if !strings.Contains(view, "Suggestion (default): Sam Okafor") || strings.Contains(view, "gone") {
		t.Fatalf("suggestions show names and skip stale ids:\n%s", view)
	}
}

func TestALongPickListScrollsAroundTheCursorAtEveryHeight(t *testing.T) {
	t.Parallel()
	var many []model.Choice
	for index := 0; index < 40; index++ {
		many = append(many, model.Choice{Value: "id" + strings.Repeat("x", index), Label: "Person " + strings.Repeat("x", index)})
	}
	app, _ := pickingApp(directoryQuestion("plan.contact", "plan", "Who?", model.VariableTypeContact, true, many...), model.Value{})
	for _, height := range []int{12, 16, 24, 40} {
		shown := resize(app, 120, height)
		for range 25 {
			shown = press(shown, down)
		}
		view := plainView(shown)
		if !strings.Contains(view, "> [ ] Person "+strings.Repeat("x", 25)) {
			t.Errorf("height %d: the cursor row must stay on screen:\n%s", height, view)
		}
	}
	tall := plainView(resize(app, 120, 24))
	if !strings.Contains(tall, "Showing 1 to ") || !strings.Contains(tall, " of 40") {
		t.Errorf("a cut list says how much of it is shown:\n%s", tall)
	}
}

func TestPickingFromARealDirectoryStoresTheIdInThePlan(t *testing.T) {
	t.Parallel()
	app, recorded := startWalking(t)
	app = moveTo(t, app, "access.onboarding_owner_role")
	view := plainView(app)
	for _, want := range []string{"> Engineering manager", "IT contact", "Platform team lead"} {
		if !strings.Contains(view, want) {
			t.Errorf("missing %q:\n%s", want, view)
		}
	}
	app = press(app, down, enter)
	if got := app.plan.Answers.Sections["access"]["onboarding_owner_role"]; !sameValue(got, id("it-contact")) {
		t.Fatalf("expected the role id, got %+v", got)
	}
	if len(recorded.saved) != 1 {
		t.Fatalf("expected one save, got %d", len(recorded.saved))
	}
	if !strings.Contains(plainView(press(app, esc)), "IT contact") {
		t.Error("the overview shows the role's name")
	}
}

func TestAnEmptyRealDirectoryAsksTheOnboarderToAddEntriesFirst(t *testing.T) {
	t.Parallel()
	plan := blankPlan(t)
	plan.Directory = model.Directory{}
	app, _ := startWith(t, plan, model.AudienceOnboarder)
	app = moveTo(t, app, "access.onboarding_owner_role")
	if !strings.Contains(strings.Join(strings.Fields(plainView(resize(app, 80, 30))), " "), "No roles to pick from yet. Add them on the Contacts screen, press c.") {
		t.Fatalf("got:\n%s", plainView(app))
	}
}
