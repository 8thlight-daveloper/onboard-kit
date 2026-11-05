package tui

import (
	"reflect"
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

func onPreparation(t *testing.T, port *fakePreparations) App {
	t.Helper()
	app, _ := startPreparing(t, port)
	return press(app, typed("t"))
}

func withSamLee(t *testing.T) *fakePreparations {
	t.Helper()
	port := newFakePreparations(acmePlan(t))
	return port.holding(t, examplePreparation(port.plan))
}

func addPerson(app App, account string) App {
	return press(app, typed("a"), typed(account), enter)
}

func focusedField(t *testing.T, app App, label string) App {
	t.Helper()
	form := app.preparation.form
	for index, field := range form.fields {
		if field.label == label {
			app.preparation.form = form.focused(index)
			return app
		}
	}
	t.Fatalf("the form has no field %q", label)
	return app
}

func pickNamed(t *testing.T, app App, label string, name string) App {
	t.Helper()
	app = press(focusedField(t, app, label), enter)
	chooser := app.preparation.form.chooser
	for row := 0; row < chooser.rows(); row++ {
		if strings.HasSuffix(chooser.rowText(row), " "+name) {
			app.preparation.form.chooser.cursor = row
			return press(app, enter)
		}
	}
	t.Fatalf("the %s list has no %q:\n%s", label, name, plainView(app))
	return app
}

func TestThePreparationTabNeedsAPortAndTheOnboarder(t *testing.T) {
	t.Parallel()
	port := withSamLee(t)
	app, _ := startPreparing(t, port)
	if !strings.Contains(headerOf(app), "Contacts c  Preparation t") {
		t.Errorf("the strip names Preparation and its key: %q", headerOf(app))
	}
	plain, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	if strings.Contains(headerOf(resize(plain, 120, 30)), "Preparation") || press(plain, typed("t")).screen == screenPreparation {
		t.Error("without a port there is no Preparation tab")
	}
	onboardee, _ := startWith(t, blankPlan(t), model.AudienceOnboardee)
	onboardee = resize(onboardee.WithPreparation(port), 120, 30).openOverview()
	if strings.Contains(headerOf(onboardee), "Preparation") || press(onboardee, typed("t")).screen == screenPreparation {
		t.Errorf("the onboardee has no Preparation tab: %q", headerOf(onboardee))
	}
	if port.loads != 0 {
		t.Errorf("nothing was loaded for the onboardee, got %d loads", port.loads)
	}
}

func TestTOpensPreparationFromEveryScreenAndEscReturns(t *testing.T) {
	t.Parallel()
	app, _ := startPreparing(t, withSamLee(t))
	answering := moveTo(t, app, "plan.team_state")
	for name, from := range map[string]struct {
		app  App
		back screen
	}{
		"answer": {answering, screenAnswer}, "overview": {app.openOverview(), screenOverview}, "review": {app, screenReview},
		"sections": {press(app, typed("s")), screenReview}, "contacts": {press(app, typed("c")), screenReview},
	} {
		opened := press(from.app, typed("t"))
		if opened.screen != screenPreparation || !strings.Contains(headerOf(opened), "[Preparation]") {
			t.Errorf("%s: t opened %v with the header %q", name, opened.screen, headerOf(opened))
		}
		if got := press(opened, esc).screen; got != from.back {
			t.Errorf("%s: esc returned to %v, want %v", name, got, from.back)
		}
	}
}

func TestTheListShowsThePeopleAndAWayToAddOne(t *testing.T) {
	t.Parallel()
	app := onPreparation(t, withSamLee(t))
	view := plainView(app)
	for _, want := range []string{"People with a preparation", "> sam-lee  records/sam-lee/preparation.obk.yaml", "Add a person", "Notices"} {
		if !strings.Contains(view, want) {
			t.Errorf("the list lacks %q:\n%s", want, view)
		}
	}
	empty := onPreparation(t, newFakePreparations(acmePlan(t)))
	if !strings.Contains(plainView(empty), "No one has a preparation here yet.") || !strings.Contains(plainView(empty), "> Add a person") {
		t.Errorf("an empty list offers to add a person:\n%s", plainView(empty))
	}
}

func TestAddingAPersonOpensANewDraftAndWritesNothing(t *testing.T) {
	t.Parallel()
	port := newFakePreparations(acmePlan(t))
	app := addPerson(onPreparation(t, port), "alex-doe")
	if app.preparation.mode != preparationPerson || app.preparation.person.loaded.Account != "alex-doe" {
		t.Fatalf("adding opened %v for %q", app.preparation.mode, app.preparation.person.loaded.Account)
	}
	view := plainView(app)
	for _, want := range []string{"alex-doe", "Saves to records/alex-doe/preparation.obk.yaml", newNotice[:20], "Saving creates the folder"} {
		if !strings.Contains(view, want) {
			t.Errorf("the new person lacks %q:\n%s", want, view)
		}
	}
	if len(port.saves) != 0 {
		t.Errorf("adding a person wrote %d times", len(port.saves))
	}
}

func TestAnAccountTheLocatorRefusesStaysInTheBox(t *testing.T) {
	t.Parallel()
	app := addPerson(onPreparation(t, newFakePreparations(acmePlan(t))), "../escape")
	if app.preparation.mode != preparationAdding || !strings.Contains(plainView(app), "! problem: Use letters, digits") {
		t.Errorf("a refused account keeps the box open with a problem:\n%s", plainView(app))
	}
	empty := press(onPreparation(t, newFakePreparations(acmePlan(t))), typed("a"), enter)
	if !strings.Contains(plainView(empty), "An account is needed.") {
		t.Errorf("an empty account is refused:\n%s", plainView(empty))
	}
}

func TestAddingAPersonWhoExistsOpensTheirPreparation(t *testing.T) {
	t.Parallel()
	app := addPerson(onPreparation(t, withSamLee(t)), "sam-lee")
	if !app.preparation.person.loaded.Exists || !strings.Contains(plainView(app), "already has a preparation") {
		t.Errorf("an existing account opens with a warning:\n%s", plainView(app))
	}
}

func TestTypingIntoTheNameNeedsNoExtraStep(t *testing.T) {
	t.Parallel()
	app := press(addPerson(onPreparation(t, newFakePreparations(acmePlan(t))), "alex-doe"), enter)
	if app.preparation.mode != preparationEditing || !app.typingNow() {
		t.Fatalf("enter on the recipient opens its form with the name ready for typing, got %v", app.preparation.mode)
	}
	app = press(app, typed("Alex Doe"))
	if got := app.preparation.form.recipient().Name; got != "Alex Doe" {
		t.Errorf("the first keys go into the name, got %q", got)
	}
}

func TestTheTeamAndPeopleArePickedFromTheDirectory(t *testing.T) {
	t.Parallel()
	app := press(addPerson(onPreparation(t, newFakePreparations(acmePlan(t))), "alex-doe"), enter)
	app = pickNamed(t, app, "Team", "Design team")
	app = pickNamed(t, app, "Manager", "Sam Okafor")
	app = pickNamed(t, app, "Onboarding owner", "Alex Rivera")
	app = pickNamed(t, app, "Buddy", "IT service desk")
	app = pickNamed(t, app, "Work location", "hybrid")
	want := model.PreparationRecipient{Team: "design", Manager: "sam-okafor", OnboardingOwner: "alex-rivera", Buddy: "it-help", WorkLocation: "hybrid"}
	if got := app.preparation.form.recipient(); !reflect.DeepEqual(got, want) {
		t.Errorf("the picks store ids: got %+v, want %+v", got, want)
	}
	if !strings.Contains(plainView(app), "Team              Design team") {
		t.Errorf("the form shows names, not ids:\n%s", plainView(app))
	}
	cleared := pickNamed(t, app, "Team", "No team")
	if cleared.preparation.form.recipient().Team != "" {
		t.Error("No team clears the team")
	}
}

func TestEscOnAPickListKeepsTheValue(t *testing.T) {
	t.Parallel()
	app := press(addPerson(onPreparation(t, withSamLee(t)), "sam-lee"), enter)
	app = press(focusedField(t, app, "Manager"), enter, down, esc)
	if app.preparation.form.recipient().Manager != "sam-okafor" || app.preparation.mode != preparationEditing {
		t.Errorf("esc closes the list and keeps the manager: %+v", app.preparation.form.recipient())
	}
}

func TestSeveralChangesStageIntoOneReviewedSave(t *testing.T) {
	t.Parallel()
	port := newFakePreparations(acmePlan(t))
	app := press(addPerson(onPreparation(t, port), "alex-doe"), enter, typed("Alex Doe"), ctrlS)
	app = press(app, typed("a"), typed("laptop"))
	app = pickNamed(t, app, "Section", "Access")
	app = pickNamed(t, app, "Kind", "equipment")
	app = pickNamed(t, app, "Status", "requested")
	app = press(app, ctrlS)
	if app.preparation.mode != preparationPerson || len(app.preparation.person.arrangements()) != 1 {
		t.Fatalf("the arrangement is staged:\n%s", plainView(app))
	}
	if !strings.Contains(plainView(app), "Not saved. 2 changes are staged.") {
		t.Errorf("the notice counts the staged changes:\n%s", plainView(app))
	}
	app = press(app, typed("G"), enter)
	if app.preparation.mode != preparationReviewing {
		t.Fatalf("the last row opens the review, got %v", app.preparation.mode)
	}
	view := plainView(app)
	for _, want := range []string{"Save the preparation for alex-doe?", `Name: "Alex Doe", was not set`, "Arrangement laptop: added", "Section: Access", "Nothing is written until you press y.", "First steps guide"} {
		if !strings.Contains(view, want) {
			t.Errorf("the review lacks %q:\n%s", want, view)
		}
	}
	if len(port.saves) != 0 {
		t.Fatalf("nothing is written before y, got %d saves", len(port.saves))
	}
	app = press(app, typed("y"))
	if len(port.saves) != 1 || !port.saves[0].createFolders {
		t.Fatalf("y saves once, creating the folder: %+v", port.saves)
	}
	saved := port.saved(t, "alex-doe")
	if saved.Recipient.Name != "Alex Doe" || len(saved.Arrangements) != 1 || saved.Arrangements[0].Kind != "equipment" {
		t.Errorf("the saved preparation holds both changes: %+v", saved)
	}
	if app.preparation.mode != preparationPerson || app.preparation.person.staged() || !strings.Contains(plainView(app), "saved records/alex-doe/preparation.obk.yaml") {
		t.Errorf("after the save the person is reloaded with nothing staged:\n%s", plainView(app))
	}
}

func TestTheReviewShowsTheRealGuideAndTheBlockers(t *testing.T) {
	t.Parallel()
	port := withSamLee(t)
	blank := blankPlan(t)
	port.plan = blank
	port.files = map[string][]byte{}
	port.holding(t, model.Preparation{SchemaVersion: 1, Kind: "preparation", Account: "sam-lee", PlanID: blank.ID})
	app, _ := startWith(t, blank, model.AudienceOnboarder)
	app = press(resize(app.WithPreparation(port), 120, 40), esc, typed("t"), enter, typed("G"), enter)
	loaded, _ := port.Load("sam-lee")
	guide := reviewOf(loaded, loaded.Preparation).FirstSteps
	view := plainView(app)
	if guide == "" || !strings.Contains(view, strings.Split(guide, "\n")[0]) {
		t.Errorf("the review shows the reviewer's guide %q:\n%s", strings.Split(guide, "\n")[0], view)
	}
	if !strings.Contains(view, "Blockers in the plan") || !strings.Contains(view, "required: What is the organization's name?") {
		t.Errorf("the review lists the onboarder's blockers:\n%s", view)
	}
	if !strings.Contains(view, "No changes. Saving writes nothing.") {
		t.Errorf("an unchanged preparation says so:\n%s", view)
	}
	app = press(app, typed("y"))
	if !strings.Contains(plainView(app), "Nothing changed, so nothing was written.") || len(port.saves) != 1 || string(port.saves[0].reviewed) != string(loaded.Bytes) {
		t.Errorf("y on no changes passes the held bytes and reports no write: %+v\n%s", port.saves, plainView(app))
	}
}

func TestAnArrangementIsReplacedWholeAndKeepsItsID(t *testing.T) {
	t.Parallel()
	port := withSamLee(t)
	app := press(onPreparation(t, port), enter, down, down, enter)
	if !strings.Contains(plainView(app), "Arrangement first-pairing") || !strings.Contains(plainView(app), "Id: first-pairing, which never changes") {
		t.Fatalf("enter on an arrangement opens it:\n%s", plainView(app))
	}
	app = pickNamed(t, app, "Status", "arranged")
	app = press(app, ctrlS)
	held := app.preparation.person.arrangements()
	if len(held) != 2 || held[1].ID != "first-pairing" || held[1].Status != "arranged" || held[1].Reference != "A first ticket" || !reflect.DeepEqual(held[1].Contacts, []string{"sam-okafor"}) {
		t.Errorf("the arrangement is replaced in place with every other field kept: %+v", held)
	}
	review := press(app, typed("G"), enter)
	if got := review.preparation.review.changes; len(got) != 1 || got[0].text != "Arrangement first-pairing, Status: arranged, was pending" {
		t.Errorf("the review names the changed field:\n%s", plainView(review))
	}
}

func TestHelpersArePickedInOrder(t *testing.T) {
	t.Parallel()
	app := press(onPreparation(t, withSamLee(t)), enter, down, enter)
	app = press(focusedField(t, app, "Helpers"), enter)
	app = press(app, space, down, down, space, enter)
	if got := app.preparation.form.arrangementValue().Contacts; !reflect.DeepEqual(got, []string{"alex-rivera", "it-help"}) {
		t.Errorf("the helpers are the picks in order, got %v", got)
	}
	untouched := press(onPreparation(t, withSamLee(t)), enter, down, down, enter)
	if got := untouched.preparation.form.arrangementValue().Contacts; !reflect.DeepEqual(got, []string{"sam-okafor"}) {
		t.Errorf("helpers that were not picked again stay as held, got %v", got)
	}
}

func TestAnEditorRefusalKeepsTheFormOpen(t *testing.T) {
	t.Parallel()
	port := withSamLee(t)
	app := press(onPreparation(t, port), enter, down, enter)
	app = pickNamed(t, app, "Owner", "Nobody")
	app = pickNamed(t, app, "Status", "arranged")
	app = press(app, ctrlS)
	if app.preparation.mode != preparationEditing || !strings.Contains(plainView(app), "! problem: Owner: Assign an owner") {
		t.Errorf("the editor's problem shows under the form:\n%s", plainView(app))
	}
	if !strings.Contains(plainView(app), "not staged") || app.preparation.person.staged() {
		t.Error("a refusal stages nothing")
	}
}

func TestANewArrangementNeedsAnIDThatIsNotTaken(t *testing.T) {
	t.Parallel()
	app := press(onPreparation(t, withSamLee(t)), enter, typed("a"), ctrlS)
	if !strings.Contains(plainView(app), "Id: An id is needed") {
		t.Errorf("an empty id is refused:\n%s", plainView(app))
	}
	taken := press(onPreparation(t, withSamLee(t)), enter, typed("a"), typed("laptop"), ctrlS)
	if !strings.Contains(plainView(taken), "Id: laptop is taken") || len(taken.preparation.person.arrangements()) != 2 {
		t.Errorf("a taken id is refused and nothing is replaced:\n%s", plainView(taken))
	}
}

func TestRemovingAnArrangementIsConfirmedAndStaged(t *testing.T) {
	t.Parallel()
	port := withSamLee(t)
	app := press(onPreparation(t, port), enter, down, typed("x"))
	if !strings.Contains(plainView(app), "Remove the arrangement laptop?") {
		t.Fatalf("x asks first:\n%s", plainView(app))
	}
	kept := press(app, enter)
	if len(kept.preparation.person.arrangements()) != 2 || !strings.Contains(plainView(kept), "nothing removed") {
		t.Errorf("any key but y keeps it:\n%s", plainView(kept))
	}
	removed := press(app, typed("y"))
	if got := removed.preparation.person.arrangements(); len(got) != 1 || got[0].ID != "first-pairing" {
		t.Errorf("y stages the removal and keeps the order: %+v", got)
	}
	if len(port.saves) != 0 {
		t.Error("a staged removal writes nothing")
	}
	review := press(removed, typed("G"), enter)
	if !strings.Contains(plainView(review), "Arrangement laptop: removed") {
		t.Errorf("the review lists the removal:\n%s", plainView(review))
	}
}

func TestLeavingAPersonWithStagedChangesAsksFirst(t *testing.T) {
	t.Parallel()
	port := withSamLee(t)
	app := press(onPreparation(t, port), enter, enter, typed("!"), ctrlS, esc)
	if app.preparation.mode != preparationDropping || !strings.Contains(plainView(app), "Drop the changes for sam-lee?") {
		t.Fatalf("esc with staged changes asks:\n%s", plainView(app))
	}
	if kept := press(app, esc); kept.preparation.mode != preparationPerson || !kept.preparation.person.staged() {
		t.Error("any key but y keeps the changes")
	}
	dropped := press(app, typed("y"))
	if dropped.preparation.mode != preparationListing || len(port.saves) != 0 {
		t.Errorf("y drops them and writes nothing: %v, %d saves", dropped.preparation.mode, len(port.saves))
	}
}

func TestStagedChangesSurviveAVisitToAnotherScreen(t *testing.T) {
	t.Parallel()
	app := press(onPreparation(t, withSamLee(t)), enter, enter, typed("!"), ctrlS)
	back := press(app, typed("o"), typed("t"))
	if back.preparation.mode != preparationPerson || back.preparation.person.draft.Recipient.Name != "Sam Lee!" {
		t.Errorf("the staged draft is kept: %+v", back.preparation.person.draft.Recipient)
	}
}

func TestAStaleSaveReappliesTheChangesAndAsksAgain(t *testing.T) {
	t.Parallel()
	port := withSamLee(t)
	app := press(onPreparation(t, port), enter, enter, typed("!"), ctrlS, typed("G"), enter)
	port.saveOutcomes = []PreparationOutcome{PreparationStale}
	app = press(app, typed("y"))
	if app.preparation.mode != preparationReviewing || !strings.Contains(plainView(app), "Nothing was saved, because the files changed on disk.") {
		t.Fatalf("a stale save stays on the review and says why:\n%s", plainView(app))
	}
	if app.preparation.person.draft.Recipient.Name != "Sam Lee!" {
		t.Error("the change is applied to the fresh load")
	}
	app = press(app, typed("y"))
	if port.saved(t, "sam-lee").Recipient.Name != "Sam Lee!" {
		t.Error("the second y saves")
	}
}

func TestARefusedSaveSaysWhyAndStaysOnTheReview(t *testing.T) {
	t.Parallel()
	for outcome, want := range map[PreparationOutcome]string{
		PreparationVolumeNotPrivate:  "cannot keep the file private",
		PreparationNotSaved:          "could not be saved",
		PreparationNotSavedLeftovers: "was not saved",
	} {
		port := withSamLee(t)
		app := press(onPreparation(t, port), enter, enter, typed("!"), ctrlS, typed("G"), enter)
		port.saveOutcomes = []PreparationOutcome{outcome}
		app = press(app, typed("y"))
		if app.preparation.mode != preparationReviewing || !strings.Contains(plainView(app), want) {
			t.Errorf("%v: the status says %q:\n%s", outcome, want, plainView(app))
		}
	}
	port := withSamLee(t)
	app := press(onPreparation(t, port), enter, enter, typed("!"), ctrlS, typed("G"), enter)
	port.saveOutcomes = []PreparationOutcome{PreparationSavedLeftovers}
	if saved := press(app, typed("y")); saved.preparation.mode != preparationPerson || !strings.Contains(plainView(saved), "Saved, but a temporary file") {
		t.Errorf("a save that left a temporary file says it saved:\n%s", plainView(saved))
	}
}

func TestPreparationValuesShowControlCharactersLiterally(t *testing.T) {
	t.Parallel()
	port := newFakePreparations(acmePlan(t))
	held := examplePreparation(port.plan)
	held.Recipient.Name = "Sam\x1b[31m Lee"
	app := press(onPreparation(t, port.holding(t, held)), enter)
	if strings.Contains(app.View(), "\x1b[31m Lee") || !strings.Contains(plainView(app), `Sam\x1b[31m Lee`) {
		t.Errorf("an escape in a name is shown, not sent:\n%q", plainView(app))
	}
}

func TestAnUntouchedTextFieldKeepsItsHeldValue(t *testing.T) {
	t.Parallel()
	port := newFakePreparations(acmePlan(t))
	held := examplePreparation(port.plan)
	held.Recipient.TimeZone = "Europe/Berlin\nsecond line"
	app := press(onPreparation(t, port.holding(t, held)), enter, enter)
	if got := app.preparation.form.recipient().TimeZone; got != held.Recipient.TimeZone {
		t.Errorf("an untouched multiline value is kept exactly, got %q", got)
	}
}

func TestANarrowHeaderClipsTheTabsButKeepsTheCurrentOne(t *testing.T) {
	t.Parallel()
	app, _ := startPreparing(t, withSamLee(t))
	reviewing := headerOf(resize(app, 60, 24))
	if reviewing != "Building: ACME Corp plan Answer Overview [Review] …" {
		t.Errorf("at 60 columns the strip is cut after the tabs that fit: %q", reviewing)
	}
	preparing := headerOf(resize(press(app, typed("t")), 60, 24))
	if preparing != "Answer Overview Review Sections Contacts [Preparation]" {
		t.Errorf("the current tab stays on screen, and the title goes first: %q", preparing)
	}
	if hints := strings.Join(press(app, typed("t")).hintTexts(), " "); !strings.Contains(hints, "answer overview") {
		t.Errorf("with no keys in the header the hint line carries them: %q", hints)
	}
}
