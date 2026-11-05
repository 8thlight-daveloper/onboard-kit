package tui

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

var (
	ctrlS     = tea.KeyMsg{Type: tea.KeyCtrlS}
	ctrlU     = tea.KeyMsg{Type: tea.KeyCtrlU}
	functionF = tea.KeyMsg{Type: tea.KeyF1}
)

func onContacts(t *testing.T, plan model.Plan) (App, *recorder) {
	t.Helper()
	app, recorded := startWith(t, plan, model.AudienceOnboarder)
	return press(resize(app, 120, 30).openOverview(), typed("c")), recorded
}

func rowOf(t *testing.T, app App, name string) string {
	t.Helper()
	for _, row := range viewLines(app) {
		left, _, _ := strings.Cut(row, "│")
		if strings.Contains(left, name) {
			return strings.Join(strings.Fields(left), " ")
		}
	}
	t.Fatalf("no row names %q:\n%s", name, plainView(app))
	return ""
}

func flat(app App) string { return strings.Join(strings.Fields(plainView(app)), " ") }

func contactIDs(plan model.Plan) []string {
	return idsOf(entriesOf(plan.Directory, kindContact))
}

func TestTheContactsTabIsOnTheStripAndCOpensItFromEveryScreen(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	app = resize(app, 120, 30)
	if !strings.Contains(headerOf(app), "Sections s  Contacts c") {
		t.Fatalf("the strip names Contacts and its key: %q", headerOf(app))
	}
	answering := moveTo(t, app, "plan.team_state")
	for name, from := range map[string]App{"answer": answering, "overview": app.openOverview(), "review": app, "sections": press(app, typed("s"))} {
		opened := press(from, typed("c"))
		if opened.screen != screenContacts || !strings.Contains(headerOf(opened), "[Contacts]") {
			t.Errorf("%s: c opened %v with the header %q", name, opened.screen, headerOf(opened))
		}
	}
}

func TestTheOnboardeeHasNoContactsScreen(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, blankPlan(t), model.AudienceOnboardee)
	overview := resize(app, 120, 30).openOverview()
	if strings.Contains(headerOf(overview), "Contacts") {
		t.Errorf("the onboardee strip names Contacts: %q", headerOf(overview))
	}
	if press(overview, typed("c")).screen != screenOverview {
		t.Error("c does nothing for the onboardee")
	}
	if strings.Contains(strings.Join(overview.hintTexts(), " "), "contacts") {
		t.Error("the onboardee's hints do not name Contacts")
	}
}

func (app App) hintTexts() []string {
	var texts []string
	for _, entry := range app.hints(false) {
		texts = append(texts, entry.label)
	}
	return texts
}

func TestEscapeFromContactsReturnsToTheScreenThePersonCameFrom(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	app = resize(app, 120, 30)
	answering := moveTo(t, app, "plan.team_state")
	for name, want := range map[string]struct {
		from App
		to   screen
	}{"answer": {answering, screenAnswer}, "overview": {app.openOverview(), screenOverview}, "review": {app, screenReview}} {
		if got := press(want.from, typed("c"), esc).screen; got != want.to {
			t.Errorf("%s: esc returned to %v, want %v", name, got, want.to)
		}
	}
	if press(app, typed("s"), typed("c"), esc).screen != screenReview {
		t.Error("esc from Contacts opened from Sections returns to where Sections was opened from")
	}
}

func TestTheListsShowNamesDetailsAndHoldersInOrder(t *testing.T) {
	t.Parallel()
	plan := acmePlan(t)
	plan.Directory.Roles[0].Holders = []string{"sam-okafor", "alex-rivera"}
	app, _ := onContacts(t, plan)
	if got := rowOf(t, app, "Sam Okafor"); got != "Sam Okafor sam.okafor@acme.example @sam" {
		t.Errorf("contact row: %q", got)
	}
	if !strings.Contains(plainView(app), "[Contacts 4]  Roles 3  Teams 2") {
		t.Errorf("the kinds and their counts:\n%s", plainView(app))
	}
	roles := press(app, tab)
	if got := rowOf(t, roles, "Engineering manager"); !strings.Contains(got, "Sam Okafor, Alex Rivera") {
		t.Errorf("a role names its holders in order: %q", got)
	}
	detail := plainView(roles)
	for _, want := range []string{"Holders, in order:", "1. Sam Okafor", "2. Alex Rivera"} {
		if !strings.Contains(detail, want) {
			t.Errorf("the role detail lacks %q:\n%s", want, detail)
		}
	}
	teams := press(roles, tab)
	if got := rowOf(t, teams, "Design team"); !strings.Contains(got, "1 override") {
		t.Errorf("a team counts its overrides: %q", got)
	}
	if got := rowOf(t, teams, "Platform team"); !strings.Contains(got, "no overrides") {
		t.Errorf("a team with none says so: %q", got)
	}
	if !strings.Contains(plainView(press(teams, down)), "Engineering manager: Alex Rivera") {
		t.Error("the team detail names the override and its holders")
	}
	if press(teams, tab).contacts.kind != kindContact || press(app, shiftTab).contacts.kind != kindTeam {
		t.Error("tab and shift+tab cycle the three lists")
	}
}

func TestProblemsTheValidatorFindsAreFlaggedNextToTheEntry(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, contactsWithFlags(t))
	cases := []struct {
		app  App
		name string
		want string
	}{
		{app, "Pat Lee", "! email looks wrong"},
		{app, "Front desk", "! no way to reach"},
		{press(app, tab), "Security lead", "! vacant"},
	}
	for _, each := range cases {
		if got := rowOf(t, each.app, each.name); !strings.Contains(got, each.want) {
			t.Errorf("%s: %q lacks %q", each.name, got, each.want)
		}
	}
	if got := rowOf(t, app, "Sam Okafor"); strings.Contains(got, "!") {
		t.Errorf("an entry with no problem is not flagged: %q", got)
	}
	narrow := resize(app, 80, 30)
	if !strings.Contains(flat(press(narrow, typed("G"))), "warning: No email, Slack handle, or Slack id, so nobody can reach this contact.") {
		t.Errorf("the detail explains the flag in words:\n%s", plainView(press(narrow, typed("G"))))
	}
	if !strings.Contains(flat(press(narrow, tab, typed("G"))), "warning: Nobody holds this role, so it shows as a gap in the document.") {
		t.Error("the detail explains a vacant role")
	}
}

func TestANarrowWindowStillFlagsAProblemAndShowsTheDetailBelow(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, contactsWithFlags(t))
	app = resize(press(app, typed("G")), 80, 30)
	view := plainView(app)
	for _, want := range []string{"! no way to reach", "About this contact", "Id: front-desk"} {
		if !strings.Contains(view, want) {
			t.Errorf("lacks %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "│") {
		t.Error("one column below 120")
	}
}

func TestAddingAContactMakesTheIdFromTheNameAndSavesAtOnce(t *testing.T) {
	t.Parallel()
	app, recorded := onContacts(t, acmePlan(t))
	app = press(app, typed("a"), typed("Dana O'Neil-Smith"))
	if !strings.Contains(flat(app), "Id: dana-o-neil-smith, made from the name") {
		t.Fatalf("the id is shown as it is typed:\n%s", plainView(app))
	}
	app = press(app, enter, typed("dana@acme.example"), enter, typed("@dana"), enter, typed("U0DANA"), enter)
	want := model.Contact{ID: "dana-o-neil-smith", Name: "Dana O'Neil-Smith", Email: "dana@acme.example", Slack: "@dana", SlackID: "U0DANA"}
	contacts := app.plan.Directory.Contacts
	if got := contacts[len(contacts)-1]; got != want {
		t.Fatalf("stored %+v, want %+v", got, want)
	}
	if len(recorded.saved) != 1 || !reflect.DeepEqual(recorded.saved[0], app.plan) {
		t.Fatalf("every accepted edit is saved at once, saved %d times", len(recorded.saved))
	}
	if app.contacts.mode != contactsBrowsing || !strings.Contains(plainView(app), "> Dana O'Neil-Smith") || !strings.Contains(plainView(app), "saved") {
		t.Errorf("the list returns with the new entry chosen:\n%s", plainView(app))
	}
}

func TestANameThatCollidesGetsANumberAndAnExistingIdNeverChanges(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	app = press(app, typed("a"), typed("Sam Okafor"))
	if !strings.Contains(flat(app), "Id: sam-okafor-2, made from the name, with a number because sam-okafor is taken") {
		t.Fatalf("the number is explained:\n%s", plainView(app))
	}
	app = press(app, ctrlS)
	if got := contactIDs(app.plan); !slices.Equal(got, []string{"alex-rivera", "sam-okafor", "it-help", "platform-team", "sam-okafor-2"}) {
		t.Fatalf("ids %v", got)
	}
	edited := press(app, typed("g"), down, enter)
	if !strings.Contains(plainView(edited), "Id: sam-okafor, which never changes") {
		t.Fatalf("an existing id is shown as fixed:\n%s", plainView(edited))
	}
	edited = press(edited, ctrlU, typed("Samuel Okafor"), ctrlS)
	got := edited.plan.Directory.Contacts[1]
	if got.ID != "sam-okafor" || got.Name != "Samuel Okafor" || got.Email != "sam.okafor@acme.example" || got.Slack != "@sam" {
		t.Errorf("a rename keeps the id and the other fields: %+v", got)
	}
	if !slices.Equal(edited.plan.Answers.Sections["stuck-general"]["general_contacts"].List, []string{"alex-rivera", "sam-okafor"}) {
		t.Error("answers that name the id are untouched")
	}
}

func TestTheSlugFoldsCaseAndDropsPunctuation(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]string{
		"Sam Okafor": "sam-okafor", "  IT   help desk ": "it-help-desk", "R&D (West)": "r-d-west", "42nd Street": "42nd-street", "!!!": "", "": "",
	} {
		if got := slugOf(name); got != want {
			t.Errorf("slugOf(%q) = %q, want %q", name, got, want)
		}
	}
	if got := freshID("Sam Okafor", []string{"sam-okafor", "sam-okafor-2"}); got != "sam-okafor-3" {
		t.Errorf("the next free number, got %q", got)
	}
	if got := freshID("Sam Okafor", []string{"sam-okafor-2"}); got != "sam-okafor" {
		t.Errorf("no number when the plain id is free, got %q", got)
	}
}

func TestARefusedEditShowsTheProblemInWordsAndLeavesThePlanAlone(t *testing.T) {
	t.Parallel()
	app, recorded := onContacts(t, acmePlan(t))
	before := app.plan
	empty := press(app, typed("a"), ctrlS)
	if empty.contacts.mode != contactsEditing {
		t.Fatal("the form stays open on a refusal")
	}
	view := plainView(empty)
	if !strings.Contains(view, "! problem: A name is needed.") || strings.Contains(view, "id can be made") {
		t.Errorf("a blank name gets one plain message:\n%s", view)
	}
	symbols := press(app, typed("a"), typed("!!!"), ctrlS)
	if !strings.Contains(flat(symbols), "! problem: The name needs a letter or a digit, so an id can be made from it.") {
		t.Errorf("a name with no letters:\n%s", plainView(symbols))
	}
	for _, refused := range []App{empty, symbols} {
		if !reflect.DeepEqual(refused.plan, before) || len(recorded.saved) != 0 {
			t.Fatal("a refusal changes nothing and saves nothing")
		}
		if !strings.Contains(plainView(refused), "not saved") {
			t.Error("the status says it was not saved")
		}
	}
	typedAgain := press(empty, typed("Dana"))
	if strings.Contains(plainView(typedAgain), "A name is needed") {
		t.Error("typing clears the old problem")
	}
}

func TestEscapeCancelsAFormWithoutSaving(t *testing.T) {
	t.Parallel()
	app, recorded := onContacts(t, acmePlan(t))
	before := app.plan
	app = press(app, typed("a"), typed("Dana"), esc)
	if app.contacts.mode != contactsBrowsing || !reflect.DeepEqual(app.plan, before) || len(recorded.saved) != 0 {
		t.Error("esc leaves the plan as it was")
	}
}

func TestARoleKeepsTheOrderOfItsHolders(t *testing.T) {
	t.Parallel()
	app, recorded := onContacts(t, acmePlan(t))
	app = press(app, tab, typed("a"), typed("Security lead"), enter, down, space, typed("k"), space)
	view := plainView(app)
	for _, want := range []string{"[1] Sam Okafor", "[2] Alex Rivera", "Order: Sam Okafor, then Alex Rivera"} {
		if !strings.Contains(view, want) {
			t.Errorf("lacks %q:\n%s", want, view)
		}
	}
	app = press(app, enter)
	roles := app.plan.Directory.Roles
	if got := roles[len(roles)-1]; got.ID != "security-lead" || !slices.Equal(got.Holders, []string{"sam-okafor", "alex-rivera"}) {
		t.Fatalf("stored %+v", got)
	}
	if len(recorded.saved) != 1 {
		t.Errorf("saved %d times", len(recorded.saved))
	}
	if app.contacts.kind != kindRole || !strings.Contains(plainView(app), "> Security lead") {
		t.Errorf("the new role is chosen on the roles list:\n%s", plainView(app))
	}
}

func TestARoleWithNoHolderIsSavedAndFlaggedVacant(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	app = press(app, tab, typed("a"), typed("Security lead"), ctrlS)
	if got := rowOf(t, app, "Security lead"); !strings.Contains(got, "! vacant") {
		t.Errorf("a role with nobody is flagged: %q", got)
	}
}

func TestEditingARolePrefillsItsHoldersAndAnEditCanDropThem(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	app = press(app, tab, enter)
	if !strings.Contains(plainView(app), "Engineering manager") || !strings.Contains(plainView(press(app, enter)), "[1] Sam Okafor") {
		t.Fatalf("the form starts from the stored role:\n%s", plainView(app))
	}
	app = press(app, enter, space, enter)
	if got := app.plan.Directory.Roles[0]; got.ID != "engineering-manager" || len(got.Holders) != 0 {
		t.Errorf("unpicking the only holder leaves the role vacant: %+v", got)
	}
	if !strings.Contains(rowOf(t, app, "Engineering manager"), "! vacant") {
		t.Error("and it is flagged")
	}
}

func TestATeamOverridesARoleWithHoldersOrAnExplicitVacancy(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	app = press(app, tab, tab, typed("a"), typed("Data team"), enter, down)
	view := plainView(app)
	for _, want := range []string{"Who holds each role for this team", "> Engineering manager  same as the role: Sam Okafor", "IT contact", "Platform team lead"} {
		if !strings.Contains(strings.Join(strings.Fields(view), " "), strings.Join(strings.Fields(want), " ")) {
			t.Errorf("lacks %q:\n%s", want, view)
		}
	}
	app = press(app, enter, down, space)
	if view := plainView(app); !strings.Contains(view, "(o) Vacant for this team") || !strings.Contains(view, "( ) Use the role's own holders") {
		t.Fatalf("the two explicit choices:\n%s", view)
	}
	app = press(app, enter)
	if !strings.Contains(rowOf(t, app, "Engineering manager"), "vacant for this team") {
		t.Errorf("the row says it is vacant:\n%s", plainView(app))
	}
	app = press(app, down, enter, down, down, space, down, space, enter)
	app = press(app, ctrlS)
	teams := app.plan.Directory.Teams
	got := teams[len(teams)-1]
	if got.ID != "data-team" || got.Name != "Data team" {
		t.Fatalf("stored %+v", got)
	}
	if holders, ok := got.Holders["engineering-manager"]; !ok || len(holders) != 0 {
		t.Errorf("an explicit vacancy is an empty list, not an absent key: %v", got.Holders)
	}
	if !slices.Equal(got.Holders["it-contact"], []string{"alex-rivera", "sam-okafor"}) {
		t.Errorf("the listed holders keep their order: %v", got.Holders)
	}
	if _, ok := got.Holders["platform-team-lead"]; ok {
		t.Error("a role left as it was has no override")
	}
}

func TestPickingTheRolesOwnHoldersAgainRemovesTheOverride(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	app = press(app, tab, tab, down, enter)
	if !strings.Contains(plainView(app), "Alex Rivera") {
		t.Fatalf("the form shows the stored override:\n%s", plainView(app))
	}
	app = press(app, enter, down, enter, enter)
	if !strings.Contains(plainView(app), "Engineering manager") {
		t.Fatal("the override picker opens on the first role")
	}
	app = press(app, enter, typed("g"), space, enter, ctrlS)
	if got := app.plan.Directory.Teams[1]; len(got.Holders) != 0 {
		t.Errorf("the override is gone: %v", got.Holders)
	}
}

func TestEscapeInTheOverridePickerKeepsTheRowAsItWas(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	app = press(app, tab, tab, typed("a"), typed("Data team"), enter, down, enter, down, space, esc)
	if app.contacts.mode != contactsEditing || app.contacts.form.choosing {
		t.Fatal("esc closes only the picker")
	}
	if !strings.Contains(rowOf(t, app, "Engineering manager"), "same as the role") {
		t.Error("nothing changed")
	}
}

func TestRemovingShowsWhatItClearsBeforeAnythingChanges(t *testing.T) {
	t.Parallel()
	app, recorded := onContacts(t, acmePlan(t))
	before := app.plan
	asking := press(app, down, typed("x"))
	view := strings.Join(strings.Fields(plainView(asking)), " ")
	for _, want := range []string{
		"Remove the contact Sam Okafor?",
		"This also clears 1 answer and removes Sam Okafor from 1 role.",
		"Who can anyone ask for help with anything?",
		"Engineering manager role has nobody left",
		"Nothing is removed until you press y.",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("lacks %q:\n%s", want, plainView(asking))
		}
	}
	if !reflect.DeepEqual(asking.plan, before) || len(recorded.saved) != 0 {
		t.Fatal("the dry run changes and saves nothing")
	}
	removed := press(asking, typed("y"))
	if slices.Contains(contactIDs(removed.plan), "sam-okafor") {
		t.Fatal("y removes the contact")
	}
	if got := removed.plan.Answers.Sections["stuck-general"]["general_contacts"].List; !slices.Equal(got, []string{"alex-rivera"}) {
		t.Errorf("the answer lost the contact: %v", got)
	}
	if len(removed.plan.Directory.Roles[0].Holders) != 0 {
		t.Error("the role lost its holder")
	}
	if len(recorded.saved) != 1 || !reflect.DeepEqual(recorded.saved[0], removed.plan) {
		t.Errorf("the removal is saved at once, saved %d times", len(recorded.saved))
	}
	if removed.contacts.mode != contactsBrowsing || removed.contacts.cursor() != 1 {
		t.Error("the list returns")
	}
}

func TestAnyKeyButYKeepsTheEntry(t *testing.T) {
	t.Parallel()
	app, recorded := onContacts(t, acmePlan(t))
	before := app.plan
	for _, key := range []tea.KeyMsg{enter, esc, typed("n"), typed("x"), space} {
		kept := press(app, down, typed("x"), key)
		if !reflect.DeepEqual(kept.plan, before) || len(recorded.saved) != 0 || kept.contacts.mode != contactsBrowsing {
			t.Errorf("%v keeps the entry", key)
		}
		if !strings.Contains(plainView(kept), "nothing removed") {
			t.Errorf("%v: the status says nothing was removed", key)
		}
	}
}

func TestRemovingARoleCountsItsTeamOverridesAndTheAnswersItClears(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	asking := press(app, tab, typed("x"))
	view := strings.Join(strings.Fields(plainView(asking)), " ")
	for _, want := range []string{"This also clears 1 answer and removes 1 team override for this role.", "Which role owns onboarding here?", "Design team, as Engineering manager"} {
		if !strings.Contains(view, want) {
			t.Errorf("lacks %q:\n%s", want, plainView(asking))
		}
	}
	removed := press(asking, typed("y"))
	if _, ok := removed.plan.Answers.Sections["access"]["onboarding_owner_role"]; ok {
		t.Error("the answer is cleared")
	}
	if len(removed.questions) != len(app.questions) || answerState(removed.questions[removed.indexOf("access.onboarding_owner_role")]) != "unanswered" {
		t.Error("the Answer screen's questions are refreshed, the cleared one is unanswered")
	}
}

func TestRemovingATeamClearsTheTeamAnswer(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	asking := press(app, tab, tab, typed("x"))
	if !strings.Contains(plainView(asking), "This also clears 1 answer.") {
		t.Errorf("a team has no holders to clear:\n%s", plainView(asking))
	}
	if _, ok := press(asking, typed("y")).plan.Answers.Plan["team"]; ok {
		t.Error("the team answer is cleared")
	}
}

func TestRemovingAnEntryNothingUsesSaysNothingElseChanges(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, contactsWithFlags(t))
	asking := press(app, typed("G"), typed("x"))
	view := plainView(asking)
	if !strings.Contains(view, "Remove the contact Front desk?") || !strings.Contains(view, "Nothing else changes.") || strings.Contains(view, "Answers cleared") {
		t.Errorf("got:\n%s", view)
	}
}

func TestALongListOfClearedAnswersIsCutWithACount(t *testing.T) {
	t.Parallel()
	plan := acmePlan(t)
	for index := 0; index < 8; index++ {
		name := "extra" + strings.Repeat("x", index)
		plan.Variables = append(plan.Variables, model.Variable{Name: name, Description: "Extra question " + name, Type: model.VariableTypeContact})
		plan.Answers.Plan[name] = model.Value{Kind: model.ValueKindText, Text: "sam-okafor"}
	}
	app, _ := onContacts(t, plan)
	view := plainView(press(app, down, typed("x")))
	if !strings.Contains(view, "This also clears 9 answers") || !strings.Contains(view, "and 5 more") {
		t.Errorf("got:\n%s", view)
	}
}

func TestAnEmptyListSaysHowToAddTheFirstEntry(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, withoutDirectory(blankPlan(t)), model.AudienceOnboarder)
	app = press(resize(app, 120, 30).openOverview(), typed("c"))
	for _, want := range []string{"No contacts yet.", "Press a to add one."} {
		if !strings.Contains(plainView(app), want) {
			t.Errorf("lacks %q:\n%s", want, plainView(app))
		}
	}
	if press(app, typed("x"), enter).contacts.mode != contactsBrowsing || press(app, enter).contacts.mode != contactsBrowsing {
		t.Error("x and enter do nothing with no entry")
	}
	added := press(app, typed("a"), typed("Dana"), ctrlS)
	if got := contactIDs(added.plan); !slices.Equal(got, []string{"dana"}) {
		t.Errorf("the first contact: %v", got)
	}
}

func TestACreatedRoleNeedsContactsAndSaysSo(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, withoutDirectory(blankPlan(t)), model.AudienceOnboarder)
	app = press(resize(app, 120, 30).openOverview(), typed("c"), tab, typed("a"), typed("Lead"), enter)
	if !strings.Contains(strings.Join(strings.Fields(plainView(app)), " "), "No contacts yet. Add some on the Contacts list first") {
		t.Errorf("got:\n%s", plainView(app))
	}
	if saved := press(app, enter); len(saved.plan.Directory.Roles) != 1 {
		t.Error("a role with no holders can still be saved")
	}
	team := press(resize(startWithoutDirectory(t), 120, 30).openOverview(), typed("c"), tab, tab, typed("a"), typed("Team"), enter)
	if team.contacts.mode != contactsEditing || !team.contacts.form.onContactRow() {
		t.Error("enter on the name of a team moves to the team contact")
	}
	if saved := press(team, ctrlS); saved.contacts.mode != contactsBrowsing || len(saved.plan.Directory.Teams) != 1 {
		t.Error("ctrl+s saves a team with no roles to override and no contact")
	}
}

func startWithoutDirectory(t *testing.T) App {
	t.Helper()
	app, _ := startWith(t, withoutDirectory(blankPlan(t)), model.AudienceOnboarder)
	return app
}

func TestTheAnswerScreenSeesNewEntriesAfterReturningFromContacts(t *testing.T) {
	t.Parallel()
	app, recorded := startAnswering(t)
	app = resize(moveTo(t, app, "access.onboarding_owner_role"), 120, 30)
	onContacts := press(app, typed("c"))
	if onContacts.contacts.kind != kindRole {
		t.Fatalf("c from a role question opens the roles list, got %v", onContacts.contacts.kind)
	}
	onContacts = press(onContacts, typed("a"), typed("Security lead"), ctrlS)
	back := press(onContacts, esc)
	if back.screen != screenAnswer || back.questions[back.cursor].ID != "access.onboarding_owner_role" {
		t.Fatalf("esc returns to the question, got %v at %d", back.screen, back.cursor)
	}
	if !strings.Contains(plainView(back), "Security lead") {
		t.Errorf("the pick list shows the new role:\n%s", plainView(back))
	}
	if len(recorded.saved) != 1 {
		t.Errorf("only the directory edit was saved so far, saved %d times", len(recorded.saved))
	}
	chosen := press(back, down, down, down, enter)
	if got := chosen.plan.Answers.Sections["access"]["onboarding_owner_role"]; !sameValue(got, id("security-lead")) {
		t.Errorf("the new role can be picked at once: %+v", got)
	}
}

func TestTheEmptyPickListPointsAtTheContactsScreenAndTheKeyOpensIt(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, withoutDirectory(blankPlan(t)), model.AudienceOnboarder)
	app = resize(moveTo(t, app, "access.onboarding_owner_role"), 120, 30)
	if !strings.Contains(plainView(app), "Add them on the Contacts screen, press c.") {
		t.Fatalf("got:\n%s", plainView(app))
	}
	onContacts := press(app, typed("c"))
	if onContacts.screen != screenContacts || onContacts.contacts.kind != kindRole {
		t.Fatalf("c opens the list the question needs, got %v %v", onContacts.screen, onContacts.contacts.kind)
	}
	back := press(onContacts, typed("a"), typed("Owner"), ctrlS, esc)
	if !strings.Contains(plainView(back), "> Owner") {
		t.Errorf("the question now lists the new role:\n%s", plainView(back))
	}
}

func TestAnAnswerThatNamesARemovedEntryReadsAsNotAnsweredOnTheAnswerScreen(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	app = press(resize(app, 120, 30), typed("c"), tab, typed("x"), typed("y"), typed("w"))
	if app.screen != screenAnswer {
		t.Fatalf("w returns to the Answer screen, got %v", app.screen)
	}
	question := app.questions[app.indexOf("access.onboarding_owner_role")]
	if answerState(question) != "unanswered" || strings.Contains(plainView(moveTo(t, app, question.ID)), "Engineering manager (current)") {
		t.Error("the Answer screen shows the cleared answer as not answered")
	}
	if !strings.Contains(headerOf(app), "12 of 13") && !strings.Contains(headerOf(resize(app, 140, 30)), "12 of 13") {
		t.Errorf("the count follows the removal: %q", headerOf(resize(app, 140, 30)))
	}
}

func TestASaveThatFailsIsSaidAndTheEditStaysInMemory(t *testing.T) {
	t.Parallel()
	failing := func(model.Plan) error { return errors.New("disk full") }
	app := NewApp(NewBackend(), acmePlan(t), failing, model.AudienceOnboarder)
	app = press(resize(app, 120, 30), typed("c"), typed("a"), typed("Dana"), ctrlS)
	if !strings.Contains(plainView(app), "could not save: disk full") {
		t.Errorf("got:\n%s", plainView(app))
	}
}

func TestWhileAFormIsOpenLettersAreTextAndTheTabsShowNoKeys(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	form := press(app, typed("a"), typed("wors"))
	if form.screen != screenContacts || form.contacts.form.name() != "wors" {
		t.Error("letters that are screen keys are text in a form")
	}
	header := headerOf(form)
	if !strings.Contains(header, "[Contacts]") || strings.Contains(header, "Walk w") {
		t.Errorf("header %q", header)
	}
	hints := strings.Join(viewLines(form)[len(viewLines(form))-1:], "")
	for _, want := range []string{"enter next", "ctrl+s save", "esc cancel", "f1 help"} {
		if !strings.Contains(hints, want) {
			t.Errorf("the hint line lacks %q: %q", want, hints)
		}
	}
	if !strings.Contains(plainView(press(form, functionF)), "Keys") {
		t.Error("f1 opens help in a form")
	}
}

func TestEveryContactsModeHasItsHelp(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	cases := map[string]struct {
		app  App
		want string
	}{
		"browsing":            {press(app, typed("?")), "remove the chosen entry, after showing what it clears"},
		"contact":             {press(app, typed("a"), functionF), "enter on the last field, ctrl+s"},
		"holders":             {press(app, tab, typed("a"), enter, functionF), "the order you pick is the order kept"},
		"rows":                {press(app, tab, tab, typed("a"), enter, down, functionF), "change who holds it for this team"},
		"picker":              {press(app, tab, tab, typed("a"), enter, down, enter, functionF), "pick or unpick, or choose the top two rows"},
		"team contact":        {press(app, tab, tab, typed("a"), enter, functionF), "choose the team contact"},
		"team contact picker": {press(app, tab, tab, typed("a"), enter, enter, functionF), "choose the row"},
		"remove":              {press(app, down, typed("x"), typed("?")), "keep it and go back"},
	}
	for name, each := range cases {
		if got := strings.Join(strings.Fields(plainView(each.app)), " "); !strings.Contains(got, each.want) {
			t.Errorf("%s help lacks %q:\n%s", name, each.want, plainView(each.app))
		}
	}
	if !strings.Contains(plainView(press(moveTo(t, app, "plan.team_state"), typed("?"))), "c contacts") {
		t.Error("the Answer screen's help names the Contacts key")
	}
}

func TestTheJumpKeysLeaveContactsFromTheListButNotFromAForm(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	for key, want := range map[string]screen{"w": screenAnswer, "o": screenOverview, "r": screenReview, "s": screenSections} {
		if got := press(app, typed(key)).screen; got != want {
			t.Errorf("%s opened %v, want %v", key, got, want)
		}
	}
	if press(app, typed("a"), typed("o")).screen != screenContacts {
		t.Error("o is text in a form")
	}
	if press(app, typed("q")).screen == screenContacts {
		t.Error("q goes back from the list")
	}
}

func TestThePickerMovesTogglesAndKeepsOrder(t *testing.T) {
	t.Parallel()
	contacts := people()
	chooser := newPicker([]string{"inherit", "vacant"}, contacts, nil, specialInherit)
	if chooser.rows() != 5 || chooser.rowText(0) != "(o) inherit" || chooser.rowText(1) != "( ) vacant" {
		t.Fatalf("rows %d, %q, %q", chooser.rows(), chooser.rowText(0), chooser.rowText(1))
	}
	chooser = chooser.moved(1).toggled()
	if chooser.special != specialVacant || chooser.rowText(1) != "(o) vacant" {
		t.Errorf("space on a top row chooses it: %+v", chooser)
	}
	chooser = chooser.moved(3).toggled().moved(-1).toggled()
	if chooser.special != specialNone || !slices.Equal(chooser.picked, []string{"it-help", "sam-okafor"}) {
		t.Errorf("picking a contact leaves the top rows: %+v", chooser)
	}
	if chooser.orderText() != "Order: IT service desk, then Sam Okafor" || chooser.rowText(3) != "[2] Sam Okafor" {
		t.Errorf("order shown: %q %q", chooser.orderText(), chooser.rowText(3))
	}
	chooser = chooser.moved(1).toggled().moved(-1).toggled()
	if chooser.special != specialInherit || len(chooser.picked) != 0 {
		t.Errorf("unpicking the last contact goes back to the role's own holders: %+v", chooser)
	}
	empty := newPicker(nil, nil, nil, specialNone).toggled().moved(1)
	if empty.cursor != 0 || empty.rows() != 0 {
		t.Error("an empty picker has nothing to move or toggle")
	}
}

func TestAHolderThatIsNotAContactIsLeftOutOfThePicker(t *testing.T) {
	t.Parallel()
	chooser := newPicker(nil, people(), []string{"sam-okafor", "ghost", "sam-okafor"}, specialNone)
	if !slices.Equal(chooser.picked, []string{"sam-okafor"}) {
		t.Errorf("picked %v", chooser.picked)
	}
}

func TestADirectoryWithAProblemIsFlaggedByItsKind(t *testing.T) {
	t.Parallel()
	plan := acmePlan(t)
	plan.Directory.Roles[1].Holders = []string{"it-help", "nobody"}
	plan.Directory.Teams[1].Holders["ghost-role"] = []string{"alex-rivera"}
	plan.Directory.Contacts = append(plan.Directory.Contacts, model.Contact{ID: "sam-okafor", Name: "Sam Again", Email: "s@acme.example"})
	app, _ := onContacts(t, plan)
	if got := rowOf(t, press(app, tab), "IT contact"); !strings.Contains(got, "! holder is not a contact") {
		t.Errorf("role: %q", got)
	}
	if got := rowOf(t, press(app, tab, tab), "Design team"); !strings.Contains(got, "! role does not exist") {
		t.Errorf("team: %q", got)
	}
	if got := rowOf(t, app, "Sam Again"); !strings.Contains(got, "! id is shared") {
		t.Errorf("duplicate id: %q", got)
	}
	if !strings.Contains(flat(resize(press(app, tab, tab, down), 80, 30)), "! problem: This team overrides a role that does not exist.") {
		t.Error("an error reads as a problem in the detail")
	}
}

func TestEditingATeamDropsAnOverrideForARoleThatIsGone(t *testing.T) {
	t.Parallel()
	plan := acmePlan(t)
	plan.Directory.Teams[1].Holders["ghost-role"] = []string{"alex-rivera"}
	app, _ := onContacts(t, plan)
	app = press(app, tab, tab, down, enter, ctrlS)
	if _, ok := app.plan.Directory.Teams[1].Holders["ghost-role"]; ok {
		t.Error("saving the form cleans the override of a missing role")
	}
	if got := app.plan.Directory.Teams[1].Holders["engineering-manager"]; !slices.Equal(got, []string{"alex-rivera"}) {
		t.Errorf("and keeps the real one: %v", got)
	}
}

func teamContactOf(app App, id string) string {
	for _, team := range app.plan.Directory.Teams {
		if team.ID == id {
			return team.Contact
		}
	}
	return "missing"
}

func TestATeamFormStartsWithNoTeamContactAndPicksOneFromTheContacts(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	app = press(app, tab, tab, typed("a"), typed("Data team"), enter)
	if got := flat(app); !strings.Contains(got, "Who to contact for this team > No team contact") {
		t.Fatalf("the form offers the team contact row:\n%s", plainView(app))
	}
	app = press(app, enter)
	view := flat(app)
	for _, want := range []string{"Team contact, for Data team", "(o) No team contact", "( ) Alex Rivera", "( ) Platform team channel"} {
		if !strings.Contains(view, want) {
			t.Errorf("the picker lacks %q:\n%s", want, plainView(app))
		}
	}
	app = press(app, down, down, down, down, enter)
	if !strings.Contains(flat(app), "> Platform team channel") || app.contacts.form.choosing {
		t.Fatalf("enter chooses the row and closes the picker:\n%s", plainView(app))
	}
	app = press(app, ctrlS)
	if got := teamContactOf(app, "data-team"); got != "platform-team" {
		t.Errorf("the contact is saved with the team: %q", got)
	}
	if got := flat(resize(app, 200, 30)); !strings.Contains(got, "Contact: Platform team channel, #platform-team") {
		t.Errorf("the detail names the contact and its Slack handle:\n%s", plainView(app))
	}
}

func TestTheNoTeamContactRowClearsTheContact(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	app = press(app, tab, tab, enter, enter, enter)
	if !strings.Contains(flat(app), "(o) Platform team channel") {
		t.Fatalf("the picker opens on the stored contact:\n%s", plainView(app))
	}
	app = press(app, typed("g"), up, up, up, up, enter, ctrlS)
	if got := teamContactOf(app, "platform"); got != "" {
		t.Errorf("no team contact saves as empty: %q", got)
	}
	if !strings.Contains(flat(app), "Contact: none") {
		t.Errorf("the detail says there is no contact:\n%s", plainView(app))
	}
}

func TestEscapeInTheTeamContactPickerKeepsTheContact(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	app = press(app, tab, tab, enter, enter, enter, down, esc)
	if !app.contacts.form.onContactRow() || app.contacts.form.choosing {
		t.Fatal("esc closes only the picker")
	}
	if got := press(app, ctrlS); teamContactOf(got, "platform") != "platform-team" {
		t.Errorf("nothing changed: %q", teamContactOf(got, "platform"))
	}
}

func TestEditingAnOverrideOfATeamKeepsItsContact(t *testing.T) {
	t.Parallel()
	app, _ := onContacts(t, acmePlan(t))
	app = press(app, tab, tab, enter, enter, down, enter, down, space, enter, ctrlS)
	got := app.plan.Directory.Teams[0]
	if got.Contact != "platform-team" {
		t.Errorf("the contact is carried through the save: %+v", got)
	}
	if !slices.Equal(got.Holders["it-contact"], []string{"it-help"}) && len(got.Holders) == 0 {
		t.Errorf("the override was saved: %+v", got.Holders)
	}
	again := press(app, enter, ctrlS)
	if again.plan.Directory.Teams[0].Contact != "platform-team" {
		t.Error("saving a team with no change keeps the contact")
	}
}

func TestATeamContactThatIsNotAContactIsDroppedByTheFormAndFlaggedInTheList(t *testing.T) {
	t.Parallel()
	plan := acmePlan(t)
	plan.Directory.Teams[0].Contact = "ghost"
	app, _ := onContacts(t, plan)
	teams := press(app, tab, tab)
	if got := rowOf(t, teams, "Platform team"); !strings.Contains(got, "! team contact is not a contact") {
		t.Errorf("the row flags it: %q", got)
	}
	view := flat(resize(teams, 120, 30))
	for _, want := range []string{"! problem: The contact for this team is not", "Contact: someone who is not a contact"} {
		if !strings.Contains(view, want) {
			t.Errorf("the detail lacks %q:\n%s", want, plainView(teams))
		}
	}
	if !strings.Contains(flat(press(teams, enter, enter, enter)), "(o) No team contact") {
		t.Error("the form starts from no contact, so the person picks a real one")
	}
}

func TestRemovingATeamContactSaysItRemovesItAsTheContactOfTheTeam(t *testing.T) {
	t.Parallel()
	app, recorded := onContacts(t, acmePlan(t))
	asking := press(app, down, down, down, typed("x"))
	view := flat(asking)
	for _, want := range []string{
		"Remove the contact Platform team channel?",
		"This also removes it as the contact of 1 team.",
		"Holders removed",
		"Platform team, as the team contact",
	} {
		if !strings.Contains(view, want) {
			t.Errorf("lacks %q:\n%s", want, plainView(asking))
		}
	}
	if strings.Contains(view, "Left with nobody") || strings.Contains(view, "from 0") {
		t.Errorf("a cleared team contact leaves nothing vacant:\n%s", plainView(asking))
	}
	removed := press(asking, typed("y"))
	if got := teamContactOf(removed, "platform"); got != "" || len(recorded.saved) != 1 {
		t.Errorf("y clears the team contact and saves once: %q, saved %d", got, len(recorded.saved))
	}
}

func TestTheRemovalWordingJoinsHolderPlacesAndTeamContacts(t *testing.T) {
	t.Parallel()
	state := removalState{kind: kindContact, name: "Alex", cleared: model.Cleared{
		Answers: []string{"plan.x"},
		Holders: []string{"role:a", "team-contact:platform", "team-contact:design", "team:design/a"},
	}}
	want := "This also clears 1 answer and removes Alex from 1 role and 1 team override and removes it as the contact of 2 teams."
	if got := state.consequence(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestOnlyLowercaseCOpensTheContactsScreenAndCapitalPDoesNothing(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	app = resize(app, 120, 30)
	answering := moveTo(t, app, "plan.team_state")
	for name, from := range map[string]App{"answer": answering, "overview": app.openOverview(), "review": app, "sections": press(app, typed("s"))} {
		if got := press(from, typed("P")); got.screen != from.screen {
			t.Errorf("%s: capital P opened %v", name, got.screen)
		}
		if got := press(from, typed("c")); got.screen != screenContacts {
			t.Errorf("%s: c opened %v", name, got.screen)
		}
	}
	if got := press(answering, typed("p")); got.cursor != answering.cursor-1 {
		t.Error("p is back on the Answer screen")
	}
}

func TestCIsTextWhileTypingAnAnswerOrAFormField(t *testing.T) {
	t.Parallel()
	app, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	typing := press(moveTo(t, resize(app, 120, 30), "plan.organization"), typed("c"))
	if typing.screen != screenAnswer || !strings.Contains(plainView(typing), "c") {
		t.Fatalf("c is typed into the answer, got %v", typing.screen)
	}
	onList, _ := onContacts(t, acmePlan(t))
	form := press(onList, typed("a"), typed("c"))
	if form.contacts.form.name() != "c" || form.screen != screenContacts {
		t.Errorf("c is typed into a form field: %q", form.contacts.form.name())
	}
	if press(onList, typed("c")).screen != screenContacts {
		t.Error("c on the Contacts screen stays there")
	}
}
