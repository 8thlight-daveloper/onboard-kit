package tui

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

func openSamLee(app App) App { return press(app, down, enter) }

func filledForm(app App) App {
	return press(openSamLee(app), enter, enter, down, enter, down, down, down, down, space, enter)
}

func TestTheInvitationTabNeedsAPortAndTheOnboarder(t *testing.T) {
	t.Parallel()
	port := newFakeInvitations(t)
	app := startInviting(t, port)
	if !strings.Contains(headerOf(resize(app, 140, 30)), "Preparation t  Invitation i") {
		t.Errorf("the strip names Invitation and its key: %q", headerOf(resize(app, 140, 30)))
	}
	plain, _ := startWith(t, acmePlan(t), model.AudienceOnboarder)
	plain = resize(plain.WithPreparation(withSamLee(t)), 140, 30)
	if strings.Contains(headerOf(plain), "Invitation") || press(plain, typed("i")).screen == screenInvitation {
		t.Error("without a port there is no Invitation tab")
	}
	onboardee, _ := startWith(t, blankPlan(t), model.AudienceOnboardee)
	onboardee = resize(onboardee.WithInvitation(port), 140, 30).openOverview()
	if strings.Contains(headerOf(onboardee), "Invitation") || press(onboardee, typed("i")).screen == screenInvitation {
		t.Errorf("the onboardee has no Invitation tab: %q", headerOf(onboardee))
	}
	if port.listed != 0 {
		t.Errorf("nothing was listed before the tab opened, got %d lists", port.listed)
	}
}

func TestIOpensTheInvitationFromEveryScreenAndEscReturns(t *testing.T) {
	t.Parallel()
	app := startInviting(t, newFakeInvitations(t))
	walking := moveTo(t, app, "plan.team_state")
	for name, from := range map[string]struct {
		app  App
		back screen
	}{
		"walk": {walking, screenWalk}, "overview": {app.openOverview(), screenOverview}, "review": {app, screenReview},
		"sections": {press(app, typed("s")), screenReview}, "contacts": {press(app, typed("c")), screenReview},
		"preparation": {press(app, typed("t")), screenReview},
	} {
		opened := press(from.app, typed("i"))
		if opened.screen != screenInvitation || !strings.Contains(headerOf(opened), "[Invitation]") {
			t.Errorf("%s: i opened %v with the header %q", name, opened.screen, headerOf(opened))
		}
		if got := press(opened, esc).screen; got != from.back {
			t.Errorf("%s: esc returned to %v, want %v", name, got, from.back)
		}
	}
	if got := press(app, typed("i"), typed("t")).screen; got != screenPreparation {
		t.Errorf("t opens the Preparation from the Invitation, got %v", got)
	}
}

func TestTheListShowsWhoIsReadyAndCountsWhatBlocksTheRest(t *testing.T) {
	t.Parallel()
	port := newFakeInvitations(t)
	view := plainView(onInvitation(t, port))
	for _, want := range []string{"People with a preparation", "> pat-kim  ! not ready, 3 to fix", "  sam-lee  Ready", sendsNothingNotice[:40]} {
		if !strings.Contains(view, want) {
			t.Errorf("the list lacks %q:\n%s", want, view)
		}
	}
	port.people[0].OwnerNamed = false
	if view := plainView(onInvitation(t, port)); !strings.Contains(view, "pat-kim  ! not ready, 4 to fix") {
		t.Errorf("a missing owner counts as one more:\n%s", view)
	}
	empty := onInvitation(t, &fakeInvitations{written: map[string]bool{}})
	if !strings.Contains(plainView(empty), "No one has a preparation here yet.") {
		t.Errorf("an empty list says where to add a person:\n%s", plainView(empty))
	}
}

func TestABlockedPersonShowsWhatBlocksAndNoForm(t *testing.T) {
	t.Parallel()
	port := newFakeInvitations(t)
	app := press(onInvitation(t, port), enter)
	if app.invitation.mode != invitationBlocked {
		t.Fatalf("a blocked person opens what blocks, got mode %v", app.invitation.mode)
	}
	view := plainView(app)
	for _, want := range []string{"Pat Kim is not ready for an invitation", "Preparation problems (1)", "The preparation belongs to another account.", "Blockers in the plan (2)", "required: "} {
		if !strings.Contains(view, want) {
			t.Errorf("the blocked page lacks %q:\n%s", want, view)
		}
	}
	if back := press(app, esc); back.invitation.mode != invitationListing || len(port.requests) != 0 {
		t.Errorf("esc goes back to the list and nothing is created: %v, %d requests", back.invitation.mode, len(port.requests))
	}
}

func TestTheFormProposesTheDefaultFileAndShowsTheGuide(t *testing.T) {
	t.Parallel()
	app := resize(openSamLee(onInvitation(t, newFakeInvitations(t))), 140, 40)
	if app.invitation.mode != invitationEditing {
		t.Fatalf("a ready person opens the form, got mode %v", app.invitation.mode)
	}
	view := plainView(app)
	for _, want := range []string{"Invitation for Sam Lee", "Pre-access help   not chosen", "General help      nobody", "File              plans/obk-invitation-sam-lee.zip", "First steps guide", "Organization: ACME Corp"} {
		if !strings.Contains(view, want) {
			t.Errorf("the form lacks %q:\n%s", want, view)
		}
	}
}

func TestOnlyContactsWithARouteCanBeThePreaccessHelp(t *testing.T) {
	t.Parallel()
	app := press(openSamLee(onInvitation(t, newFakeInvitations(t))), enter)
	chooser := app.invitation.form.chooser
	var rows []string
	for row := 0; row < chooser.rows(); row++ {
		rows = append(rows, chooser.rowText(row))
	}
	if want := []string{"( ) Alex Rivera", "( ) IT service desk"}; !reflect.DeepEqual(rows, want) {
		t.Errorf("the pick list offers the contacts the port names: %v", rows)
	}
}

func TestPickingHelpChangesTheGuide(t *testing.T) {
	t.Parallel()
	app := openSamLee(onInvitation(t, newFakeInvitations(t)))
	before := app.invitation.form.guide
	picked := press(app, enter, down, enter)
	if picked.invitation.form.preaccess != "it-help" {
		t.Fatalf("enter chooses the pre-access help, got %q", picked.invitation.form.preaccess)
	}
	if picked.invitation.form.guide == before || !strings.Contains(picked.invitation.form.guide, "IT service desk") {
		t.Errorf("the guide names the chosen help:\n%s", picked.invitation.form.guide)
	}
}

func TestThePreaccessHelpIsRequiredBeforeTheLastStep(t *testing.T) {
	t.Parallel()
	port := newFakeInvitations(t)
	app := press(openSamLee(onInvitation(t, port)), ctrlS)
	if app.invitation.mode != invitationEditing || !strings.Contains(plainView(app), "! problem: Pre-access help: Choose who the") {
		t.Errorf("the form stays open with the problem:\n%s", plainView(app))
	}
	if len(port.requests) != 0 {
		t.Errorf("nothing is created, got %d requests", len(port.requests))
	}
}

func TestNothingIsWrittenBeforeY(t *testing.T) {
	t.Parallel()
	port := newFakeInvitations(t)
	app := filledForm(onInvitation(t, port))
	confirming := resize(press(app, down, enter), 140, 40)
	if confirming.invitation.mode != invitationConfirming {
		t.Fatalf("enter on the file goes on to the last step, got mode %v:\n%s", confirming.invitation.mode, plainView(confirming))
	}
	view := plainView(confirming)
	for _, want := range []string{"Write the invitation for Sam Lee?", "File              plans/obk-invitation-sam-lee.zip", "Pre-access help   Alex Rivera", "General help      Engineering manager (role)", "Nothing is written until you press y."} {
		if !strings.Contains(view, want) {
			t.Errorf("the last step lacks %q:\n%s", want, view)
		}
	}
	scrolled := press(confirming, tea.KeyMsg{Type: tea.KeyPgDown}, typed("x"), typed("n"))
	if len(port.requests) != 0 || scrolled.invitation.mode != invitationConfirming {
		t.Errorf("only y writes, got %d requests in mode %v", len(port.requests), scrolled.invitation.mode)
	}
	if back := press(confirming, esc); back.invitation.mode != invitationEditing || len(port.requests) != 0 {
		t.Errorf("esc goes back to the form with nothing written")
	}
	written := press(confirming, typed("y"))
	want := InvitationRequest{Account: "sam-lee", Output: "plans/obk-invitation-sam-lee.zip", Preaccess: "alex-rivera", Help: []InvitationHelp{{"contact", "alex-rivera"}, {"role", "engineering-manager"}}}
	if len(port.requests) != 1 || !reflect.DeepEqual(port.requests[0], want) {
		t.Errorf("y creates once with the chosen help, pre-access first: %+v", port.requests)
	}
	if written.invitation.mode != invitationWritten {
		t.Errorf("after writing, the result shows, got mode %v", written.invitation.mode)
	}
}

func TestTheResultNamesThePathAndTheLineForTheNewcomer(t *testing.T) {
	t.Parallel()
	app := press(filledForm(onInvitation(t, newFakeInvitations(t))), ctrlS, typed("y"))
	view := plainView(app)
	for _, want := range []string{"Invitation written", "plans/obk-invitation-sam-lee.zip", fmt.Sprintf(handoverLine, "Sam Lee", "plans/obk-invitation-sam-lee.zip"), "This tool sends nothing."} {
		if !strings.Contains(view, want) {
			t.Errorf("the result lacks %q:\n%s", want, view)
		}
	}
	if handoverLine != "Give this file to %s. They run: obk %s" {
		t.Errorf("the line for the newcomer keeps the agreed wording: %q", handoverLine)
	}
	if back := press(app, enter); back.invitation.mode != invitationListing {
		t.Errorf("enter goes back to the people, got mode %v", back.invitation.mode)
	}
}

func TestASecondInvitationTakesTheNextFreeName(t *testing.T) {
	t.Parallel()
	port := newFakeInvitations(t)
	app := press(filledForm(onInvitation(t, port)), ctrlS, typed("y"), enter)
	again := openSamLee(app)
	if got := again.invitation.form.output.Value(); got != "plans/obk-invitation-sam-lee-2.zip" {
		t.Errorf("the default skips the file that exists: %q", got)
	}
}

func TestAFileThatAppearedIsNeverReplacedAndTheNextNameIsOffered(t *testing.T) {
	t.Parallel()
	port := newFakeInvitations(t)
	app := filledForm(onInvitation(t, port))
	port.written["plans/obk-invitation-sam-lee.zip"] = true
	refused := press(app, ctrlS, typed("y"))
	if refused.invitation.mode != invitationEditing || !refused.invitation.form.onOutput() {
		t.Fatalf("the form opens again on the file, got mode %v", refused.invitation.mode)
	}
	if got := refused.invitation.form.output.Value(); got != "plans/obk-invitation-sam-lee-2.zip" {
		t.Errorf("the next free name is filled in: %q", got)
	}
	if problems := refused.invitation.form.problems; len(problems) != 1 || !strings.Contains(problems[0], "already exists, and it is never replaced") {
		t.Errorf("the form says why: %q", problems)
	}
	written := press(refused, enter, typed("y"))
	if written.invitation.result.Path != "plans/obk-invitation-sam-lee-2.zip" || len(port.requests) != 2 {
		t.Errorf("the second try writes the free name: %+v", written.invitation.result)
	}
}

func TestARefusedCreationSaysWhyAndStaysOnTheForm(t *testing.T) {
	t.Parallel()
	port := newFakeInvitations(t)
	port.outcomes = []InvitationOutcome{InvitationBlocked}
	app := press(filledForm(onInvitation(t, port)), ctrlS, typed("y"))
	if app.invitation.mode != invitationEditing || len(app.invitation.form.problems) == 0 || app.invitation.form.problems[0] != InvitationBlocked.sentence() {
		t.Errorf("the refusal shows on the form: %q", app.invitation.form.problems)
	}
	if !strings.Contains(plainView(app), "! problem: The plan has required questions") {
		t.Errorf("the refusal is on screen:\n%s", plainView(app))
	}
}

func TestLettersInTheFileNameAreText(t *testing.T) {
	t.Parallel()
	app := press(openSamLee(onInvitation(t, newFakeInvitations(t))), down, down, typed("qitwo"))
	if app.screen != screenInvitation || !strings.HasSuffix(app.invitation.form.output.Value(), "qitwo") {
		t.Errorf("letters go into the file name: %v %q", app.screen, app.invitation.form.output.Value())
	}
	if strings.Contains(headerOf(app), "Invitation i") {
		t.Errorf("while typing the tabs show no keys: %q", headerOf(app))
	}
}

func TestTheCommandPathIsQuotedOnlyWhenTheShellNeedsIt(t *testing.T) {
	t.Parallel()
	for path, want := range map[string]string{
		"plans/obk-invitation-sam.zip": "plans/obk-invitation-sam.zip",
		"My plans/obk-invitation.zip":  "'My plans/obk-invitation.zip'",
		"it's/obk-invitation.zip":      `'it'\''s/obk-invitation.zip'`,
	} {
		if got := commandPath(path); got != want {
			t.Errorf("%q: got %q, want %q", path, got, want)
		}
	}
}
