package tui

import (
	"slices"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (app App) updatePreparation(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch app.preparation.mode {
	case preparationAdding:
		return app.updateAdding(key)
	case preparationPerson:
		return app.updatePersonPage(key), nil
	case preparationEditing:
		return app.updatePreparationForm(key)
	case preparationRemoving:
		return app.updateArrangementRemoval(key), nil
	case preparationDropping:
		return app.updateDropping(key), nil
	case preparationReviewing:
		return app.updatePreparationReview(key), nil
	}
	return app.updatePeopleList(key), nil
}

func (app App) jumpedFromPreparation(key string) (App, bool) {
	switch key {
	case "w":
		return app.returnToAnswer(), true
	case "o":
		return app.openOverview(), true
	case "r":
		return app.openReview(), true
	case "s":
		return app.openSections(), true
	case "c":
		return app.openContacts(), true
	case "i":
		return app.openInvitation(), true
	case "?":
		app.helpOpen = true
		return app, true
	}
	return app, false
}

func (app App) updatePeopleList(key tea.KeyMsg) App {
	state := app.preparation
	last := len(state.people)
	switch key.String() {
	case "up", "k":
		app.preparation.cursor = max(state.cursor-1, 0)
	case "down", "j":
		app.preparation.cursor = min(state.cursor+1, last)
	case "g", "home":
		app.preparation.cursor = 0
	case "G", "end":
		app.preparation.cursor = last
	case "a":
		return app.addingPerson()
	case "enter", "e":
		if state.cursor >= len(state.people) {
			return app.addingPerson()
		}
		return app.openListedPerson(state.people[state.cursor].Account)
	case "esc", "q":
		return app.closeSideScreen()
	default:
		jumped, _ := app.jumpedFromPreparation(key.String())
		return jumped
	}
	return app
}

func (app App) openListedPerson(account string) App {
	opened, outcome := app.openedPerson(account)
	if outcome != PreparationDone {
		app.status = line{text: "cannot open " + account + ": " + outcome.sentence(), tone: toneProblem}
		return app
	}
	return opened
}

func (app App) addingPerson() App {
	input := newTextField("")
	input.Placeholder = "e.g. alex-doe"
	input.Width = max(app.formWidth()-formMarkWidth-preparationLabelWidth-1, inputMinimum)
	input.Focus()
	app.preparation.account = input
	app.preparation.accountProblem = ""
	app.preparation.mode = preparationAdding
	return app
}

func (app App) updateAdding(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		app.preparation.mode = preparationListing
		return app, nil
	case "f1":
		app.helpOpen = true
		return app, nil
	case "enter":
		return app.addedPerson(), nil
	}
	var command tea.Cmd
	app.preparation.account, command = app.preparation.account.Update(key)
	app.preparation.accountProblem = ""
	return app, command
}

func (app App) addedPerson() App {
	account := app.preparation.account.Value()
	if account == "" {
		app.preparation.accountProblem = "An account is needed. It names this person's folder."
		return app
	}
	opened, outcome := app.openedPerson(account)
	switch {
	case outcome == PreparationPathRefused:
		app.preparation.accountProblem = "Use letters, digits, dots, underscores, or hyphens, starting with a letter or a digit."
		return app
	case outcome != PreparationDone:
		app.preparation.accountProblem = outcome.sentence()
		return app
	case opened.preparation.person.loaded.Exists:
		opened.status = line{text: account + " already has a preparation. It is open.", tone: toneWarning}
	}
	return opened
}

func (app App) updatePersonPage(key tea.KeyMsg) App {
	person := app.preparation.person
	last := person.rowCount() - 1
	switch key.String() {
	case "up", "k":
		app.preparation.row = max(app.preparation.row-1, 0)
	case "down", "j":
		app.preparation.row = min(app.preparation.row+1, last)
	case "g", "home":
		app.preparation.row = 0
	case "G", "end":
		app.preparation.row = last
	case "enter", "e":
		return app.openedRow()
	case "a":
		return app.openedArrangementForm(-1)
	case "x", "delete":
		if _, ok := person.arrangementRow(app.preparation.row); ok {
			app.preparation.mode = preparationRemoving
		}
	case "esc", "q":
		if person.staged() {
			app.preparation.mode = preparationDropping
			return app
		}
		return app.closedPerson()
	default:
		jumped, _ := app.jumpedFromPreparation(key.String())
		return jumped
	}
	return app
}

func (app App) openedRow() App {
	person := app.preparation.person
	row := app.preparation.row
	switch row {
	case 0:
		app.preparation.form = recipientForm(person, app.formWidth())
		app.preparation.mode = preparationEditing
		return app
	case person.addRow():
		return app.openedArrangementForm(-1)
	case person.reviewRow():
		return app.openPreparationReview()
	}
	index, _ := person.arrangementRow(row)
	return app.openedArrangementForm(index)
}

func (app App) openedArrangementForm(index int) App {
	app.preparation.form = arrangementForm(app.preparation.person, index, app.formWidth())
	app.preparation.mode = preparationEditing
	return app
}

func (app App) updatePreparationForm(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := app.preparation.form
	switch key.String() {
	case "ctrl+s":
		return app.stagedForm(), nil
	case "f1":
		app.helpOpen = true
		return app, nil
	case "esc":
		if form.choosing {
			app.preparation.form = form.closedPicker(false)
			return app, nil
		}
		app.preparation.mode = preparationPerson
		app.status = line{text: "nothing staged", tone: toneFaint}
		return app, nil
	}
	switch {
	case form.choosing:
		app.preparation.form = updatedPreparationPicker(form, key)
		return app, nil
	case form.onText():
		return app.updatePreparationText(key)
	}
	return app.updatePickRow(key), nil
}

func updatedPreparationPicker(form preparationForm, key tea.KeyMsg) preparationForm {
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

func (app App) updatePreparationText(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := app.preparation.form
	switch key.String() {
	case "enter":
		if form.onLast() {
			return app.stagedForm(), nil
		}
		app.preparation.form = form.focused(form.focus + 1)
		return app, nil
	case "down", "tab":
		app.preparation.form = form.focused(form.focus + 1)
		return app, nil
	case "up", "shift+tab":
		app.preparation.form = form.focused(form.focus - 1)
		return app, nil
	}
	fields := slices.Clone(form.fields)
	var command tea.Cmd
	var input textinput.Model
	input, command = fields[form.focus].input.Update(key)
	fields[form.focus].input = input
	form.fields = fields
	form.problems = nil
	app.preparation.form = form
	return app, command
}

func (app App) updatePickRow(key tea.KeyMsg) App {
	form := app.preparation.form
	switch key.String() {
	case "up", "k", "shift+tab":
		form = form.focused(form.focus - 1)
	case "down", "j", "tab":
		form = form.focused(form.focus + 1)
	case "enter", " ":
		form = form.openedPicker()
	}
	app.preparation.form = form
	return app
}

func (app App) stagedForm() App {
	form := app.preparation.form
	if problem, refused := form.idProblem(); refused {
		app.preparation.form.problems = []string{problem}
		app.status = line{text: "not staged", tone: toneProblem}
		return app
	}
	person, problems := app.preparation.person.withChange(form.change())
	if problems != nil {
		app.preparation.form.problems = form.labelled(problems)
		app.status = line{text: "not staged", tone: toneProblem}
		return app
	}
	app.preparation.person = person
	app.preparation.mode = preparationPerson
	if form.arrangement && form.editing == "" {
		app.preparation.row = len(person.arrangements())
	}
	app.status = line{text: "staged, not saved", tone: toneFaint}
	return app
}

func (app App) updateArrangementRemoval(key tea.KeyMsg) App {
	switch key.String() {
	case "y":
		return app.stagedRemoval()
	case "?":
		app.helpOpen = true
		return app
	}
	app.preparation.mode = preparationPerson
	app.status = line{text: "nothing removed", tone: toneFaint}
	return app
}

func (app App) stagedRemoval() App {
	index, _ := app.preparation.person.arrangementRow(app.preparation.row)
	held := app.preparation.person.arrangements()[index]
	person, problems := app.preparation.person.withChange(preparationChange{kind: changeRemoval, arrangement: held})
	app.preparation.mode = preparationPerson
	if problems != nil {
		app.status = line{text: "not removed: " + firstMessage(problems), tone: toneProblem}
		return app
	}
	app.preparation.person = person
	app.preparation.row = min(app.preparation.row, person.rowCount()-1)
	app.status = line{text: "removal staged, not saved", tone: toneFaint}
	return app
}

func (app App) updateDropping(key tea.KeyMsg) App {
	switch key.String() {
	case "y":
		app = app.closedPerson()
		app.status = line{text: "changes dropped, nothing saved", tone: toneFaint}
		return app
	case "?":
		app.helpOpen = true
		return app
	}
	app.preparation.mode = preparationPerson
	app.status = line{text: "changes kept, not saved", tone: toneFaint}
	return app
}

func (app App) updatePreparationReview(key tea.KeyMsg) App {
	review := app.preparation.review
	switch key.String() {
	case "y":
		return app.savedPreparation()
	case "tab":
		app.preparation.review.focusGuide = !review.focusGuide
	case "up", "k":
		app.preparation.review = app.scrolledPreparationReview(-1)
	case "down", "j":
		app.preparation.review = app.scrolledPreparationReview(1)
	case "pgup", "b":
		app.preparation.review = app.scrolledPreparationReview(-app.pageSize())
	case "pgdown", "f", " ":
		app.preparation.review = app.scrolledPreparationReview(app.pageSize())
	case "esc", "q":
		app.preparation.mode = preparationPerson
		app.status = line{text: "nothing saved", tone: toneFaint}
	default:
		jumped, _ := app.jumpedFromPreparation(key.String())
		return jumped
	}
	return app
}
