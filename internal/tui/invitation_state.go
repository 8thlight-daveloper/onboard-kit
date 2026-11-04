package tui

import (
	"strings"
	"unicode"

	"onboard-kit/internal/model"
)

type invitationMode int

const (
	invitationListing invitationMode = iota
	invitationBlocked
	invitationEditing
	invitationConfirming
	invitationWritten
)

type invitationState struct {
	mode        invitationMode
	people      []InvitationPerson
	listRefusal InvitationOutcome
	cursor      int
	form        invitationForm
	result      InvitationResult
}

func (app App) WithInvitation(port InvitationPort) App {
	app.invitationPort = port
	return app
}

func (app App) canInvite() bool {
	return app.invitationPort != nil && app.audience != model.AudienceOnboardee
}

func (person InvitationPerson) blocking() int {
	if person.Refusal != InvitationDone {
		return 1
	}
	count := len(person.PreparationErrors) + len(person.Blockers)
	if !person.OwnerNamed {
		count++
	}
	return count
}

func (person InvitationPerson) ready() bool { return person.blocking() == 0 }

func (person InvitationPerson) name() string {
	if name := strings.TrimFunc(person.Preparation.Recipient.Name, unicode.IsSpace); name != "" {
		return visibleText(singleLine(name))
	}
	return visibleText(person.Account)
}

func (app App) openInvitation() App {
	if !app.canInvite() {
		return app
	}
	app = app.leavingForSideScreen()
	app.screen = screenInvitation
	app.invitation.mode = invitationListing
	return app.listedInvitations()
}

func (app App) listedInvitations() App {
	people, outcome := app.invitationPort.People()
	app.invitation.people = people
	app.invitation.listRefusal = outcome
	app.invitation.cursor = max(min(app.invitation.cursor, len(people)-1), 0)
	return app
}

func (app App) chosenInvitee() (InvitationPerson, bool) {
	state := app.invitation
	if state.cursor < 0 || state.cursor >= len(state.people) {
		return InvitationPerson{}, false
	}
	return state.people[state.cursor], true
}

func (app App) openedInvitee() App {
	person, ok := app.chosenInvitee()
	switch {
	case !ok:
		return app
	case !person.ready():
		app.invitation.mode = invitationBlocked
		return app
	}
	app.invitation.form = newInvitationForm(person, app.invitationFormWidth())
	app.invitation.mode = invitationEditing
	return app
}

func (app App) invitationTyping() bool {
	return app.invitation.mode == invitationEditing && !app.invitation.form.choosing && app.invitation.form.onOutput()
}
