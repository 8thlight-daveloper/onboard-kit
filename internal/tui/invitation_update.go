package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (app App) updateInvitation(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch app.invitation.mode {
	case invitationEditing:
		return app.updateInvitationForm(key)
	case invitationConfirming:
		return app.updateInvitationConfirmation(key), nil
	case invitationBlocked, invitationWritten:
		return app.updateInvitationNotice(key), nil
	}
	return app.updateInviteeList(key), nil
}

func (app App) jumpedFromInvitation(key string) App {
	switch key {
	case "w":
		return app.returnToAnswer()
	case "o":
		return app.openOverview()
	case "r":
		return app.openReview()
	case "s":
		return app.openSections()
	case "c":
		return app.openContacts()
	case "t":
		return app.openPreparation()
	case "?":
		app.helpOpen = true
	}
	return app
}

func (app App) updateInviteeList(key tea.KeyMsg) App {
	last := max(len(app.invitation.people)-1, 0)
	switch key.String() {
	case "up", "k":
		app.invitation.cursor = max(app.invitation.cursor-1, 0)
	case "down", "j":
		app.invitation.cursor = min(app.invitation.cursor+1, last)
	case "g", "home":
		app.invitation.cursor = 0
	case "G", "end":
		app.invitation.cursor = last
	case "enter", "e":
		return app.openedInvitee()
	case "esc", "q":
		return app.closeSideScreen()
	default:
		return app.jumpedFromInvitation(key.String())
	}
	return app
}

func (app App) updateInvitationNotice(key tea.KeyMsg) App {
	switch key.String() {
	case "esc", "q", "enter":
		app.invitation.mode = invitationListing
		return app.listedInvitations()
	}
	return app.jumpedFromInvitation(key.String())
}

func (app App) updateInvitationForm(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := app.invitation.form
	switch key.String() {
	case "ctrl+s":
		return app.confirmingInvitation(), nil
	case "f1":
		app.helpOpen = true
		return app, nil
	case "pgup", "pgdown":
		app.invitation.form = app.scrolledInvitationGuide(key.String())
		return app, nil
	case "esc":
		if form.choosing {
			app.invitation.form = form.closedPicker(false)
			return app, nil
		}
		app.invitation.mode = invitationListing
		app.status = line{text: "nothing written", tone: toneFaint}
		return app.listedInvitations(), nil
	}
	switch {
	case form.choosing:
		app.invitation.form = updatedInvitationPicker(form, key)
		return app, nil
	case form.onOutput():
		return app.updateInvitationOutput(key)
	}
	return app.updateInvitationPickRow(key), nil
}

func updatedInvitationPicker(form invitationForm, key tea.KeyMsg) invitationForm {
	switch key.String() {
	case "up", "k":
		form.chooser = form.chooser.moved(-1)
	case "down", "j":
		form.chooser = form.chooser.moved(1)
	case " ":
		form.chooser = form.chooser.toggled()
		if form.chooser.single {
			return form.closedPicker(true)
		}
	case "enter":
		if form.chooser.single {
			form.chooser = form.chooser.toggled()
		}
		return form.closedPicker(true)
	}
	return form
}

func (app App) updateInvitationPickRow(key tea.KeyMsg) App {
	form := app.invitation.form
	switch key.String() {
	case "up", "k", "shift+tab":
		form = form.focused(form.focus - 1)
	case "down", "j", "tab":
		form = form.focused(form.focus + 1)
	case "enter", " ":
		form = form.openedPicker()
	}
	app.invitation.form = form
	return app
}

func (app App) updateInvitationOutput(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := app.invitation.form
	switch key.String() {
	case "enter":
		return app.confirmingInvitation(), nil
	case "up", "shift+tab":
		app.invitation.form = form.focused(form.focus - 1)
		return app, nil
	case "down", "tab":
		return app, nil
	}
	var command tea.Cmd
	form.output, command = form.output.Update(key)
	form.problems = nil
	app.invitation.form = form
	return app, command
}

func (app App) confirmingInvitation() App {
	form := app.invitation.form
	if problems := form.missing(); len(problems) > 0 {
		app.invitation.form.problems = problems
		app.status = line{text: "nothing written", tone: toneProblem}
		return app
	}
	app.invitation.mode = invitationConfirming
	return app
}

func (app App) updateInvitationConfirmation(key tea.KeyMsg) App {
	switch key.String() {
	case "y":
		return app.createdInvitation()
	case "pgup", "pgdown", "up", "down", "k", "j":
		app.invitation.form = app.scrolledInvitationGuide(key.String())
	case "?":
		app.helpOpen = true
	case "esc", "q":
		app.invitation.mode = invitationEditing
		app.status = line{text: "nothing written", tone: toneFaint}
	}
	return app
}

func (app App) createdInvitation() App {
	form := app.invitation.form
	result := app.invitationPort.Create(form.request())
	switch {
	case result.Outcome.written():
		app.invitation.result = result
		app.invitation.mode = invitationWritten
		app.status = line{text: "wrote " + visibleText(result.Path), tone: toneFaint}
		if result.Outcome == InvitationWrittenLeftovers {
			app.status = line{text: result.Outcome.sentence(), tone: toneWarning}
		}
		return app
	case result.Outcome == InvitationOutputExists && result.FreeOutput != "":
		taken := form.output.Value()
		form.output.SetValue(result.FreeOutput)
		form.output.CursorEnd()
		form.problems = []string{"File: " + visibleText(taken) + " already exists, and it is never replaced. The next free name is filled in."}
		form = form.focused(invitationOutputField)
	default:
		form.problems = append([]string{result.Outcome.sentence()}, app.blockerTexts(result.Blockers)...)
	}
	app.invitation.form = form
	app.invitation.mode = invitationEditing
	app.status = line{text: "nothing written", tone: toneProblem}
	return app
}

func (app App) scrolledInvitationGuide(key string) invitationForm {
	form := app.invitation.form
	step := 1
	if key == "pgup" || key == "pgdown" {
		step = app.pageSize()
	}
	if key == "pgup" || key == "up" || key == "k" {
		step = -step
	}
	_, guideWidth, rows := app.invitationColumns(app.frame())
	limit := max(len(app.invitationGuideLines(guideWidth))-rows, 0)
	form.guideScroll = max(min(form.guideScroll+step, limit), 0)
	return form
}
