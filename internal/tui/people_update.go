package tui

import (
	"slices"

	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

func (app App) updatePeople(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch app.people.mode {
	case peopleEditing:
		return app.updateForm(key)
	case peopleRemoving:
		return app.updateRemoval(key), nil
	}
	return app.updateBrowsing(key)
}

func (app App) updateBrowsing(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	count := len(entriesOf(app.plan.Directory, app.people.kind))
	switch key.String() {
	case "up", "k":
		app.people = app.people.moved(-1, count)
	case "down", "j":
		app.people = app.people.moved(1, count)
	case "g", "home":
		app.people = app.people.moved(-count, count)
	case "G", "end":
		app.people = app.people.moved(count, count)
	case "tab", "right":
		app.people = app.people.cycled(1)
	case "shift+tab", "left":
		app.people = app.people.cycled(-1)
	case "a":
		return app.openForm(""), nil
	case "enter", "e":
		return app.editSelected(), nil
	case "x", "delete":
		return app.askToRemove(), nil
	case "esc", "q":
		return app.closeSideScreen(), nil
	case "w":
		return app.returnToWalk(), nil
	case "o":
		return app.openOverview(), nil
	case "r":
		return app.openReview(), nil
	case "s":
		return app.openSections(), nil
	case "?":
		app.helpOpen = true
	}
	return app, nil
}

func (app App) updateRemoval(key tea.KeyMsg) App {
	switch key.String() {
	case "y":
		return app.confirmRemoval()
	case "?":
		app.helpOpen = true
		return app
	}
	app.people.mode = peopleBrowsing
	app.status = line{text: "nothing removed", tone: toneFaint}
	return app
}

func (app App) editSelected() App {
	entry, ok := app.selectedEntry()
	if !ok {
		return app
	}
	return app.openForm(entry.id)
}

func (app App) openForm(id string) App {
	app.people.form = newForm(app.plan.Directory, app.people.kind, id, app.formWidth())
	app.people.mode = peopleEditing
	return app
}

func (app App) formWidth() int {
	width, _ := app.frame().columns(answerShareNumerator, answerShareDenominator)
	return width
}

func (app App) updateForm(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := app.people.form
	switch key.String() {
	case "ctrl+s":
		return app.saveForm(), nil
	case "f1":
		app.helpOpen = true
		return app, nil
	case "esc":
		return app.cancelForm(), nil
	}
	switch {
	case form.choosing:
		return app.updatePicker(key), nil
	case form.onText():
		return app.updateFormText(key)
	case form.onHolders():
		return app.updateHolders(key), nil
	}
	return app.updateRoleRows(key), nil
}

func (app App) cancelForm() App {
	if app.people.form.choosing {
		app.people.form = app.people.form.closedPicker(false)
		return app
	}
	app.people.mode = peopleBrowsing
	app.status = line{text: "nothing saved", tone: toneFaint}
	return app
}

func (app App) updateFormText(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := app.people.form
	switch key.String() {
	case "enter":
		if form.focus == form.lastFocus() {
			return app.saveForm(), nil
		}
		app.people.form = form.focused(form.focus + 1)
		return app, nil
	case "down", "tab":
		app.people.form = form.focused(form.focus + 1)
		return app, nil
	case "up", "shift+tab":
		app.people.form = form.focused(form.focus - 1)
		return app, nil
	}
	fields := slices.Clone(form.fields)
	var command tea.Cmd
	fields[form.focus], command = fields[form.focus].Update(key)
	form.fields = fields
	form.problems = nil
	app.people.form = form
	return app, command
}

func (app App) updateHolders(key tea.KeyMsg) App {
	form := app.people.form
	switch key.String() {
	case "up", "k":
		if form.holders.cursor == 0 {
			form = form.focused(form.focus - 1)
			break
		}
		form.holders = form.holders.moved(-1)
	case "down", "j":
		form.holders = form.holders.moved(1)
	case " ":
		form.holders = form.holders.toggled()
	case "shift+tab", "tab":
		form = form.focused(0)
	case "enter":
		app.people.form = form
		return app.saveForm()
	}
	app.people.form = form
	return app
}

func (app App) updateRoleRows(key tea.KeyMsg) App {
	form := app.people.form
	switch key.String() {
	case "up", "k", "shift+tab":
		form = form.focused(form.focus - 1)
	case "down", "j", "tab":
		form = form.focused(form.focus + 1)
	case "enter", " ":
		form = form.openedPicker()
	}
	app.people.form = form
	return app
}

func (app App) updatePicker(key tea.KeyMsg) App {
	form := app.people.form
	switch key.String() {
	case "up", "k":
		form.chooser = form.chooser.moved(-1)
	case "down", "j":
		form.chooser = form.chooser.moved(1)
	case " ":
		form.chooser = form.chooser.toggled()
	case "enter":
		form = form.closedPicker(true)
	}
	app.people.form = form
	return app
}

func (app App) saveForm() App {
	form := app.people.form
	var next model.Plan
	var problems []model.Problem
	switch form.kind {
	case kindRole:
		next, problems = app.backend.SetRole(app.plan, form.role())
	case kindTeam:
		next, problems = app.backend.SetTeam(app.plan, form.team())
	default:
		next, problems = app.backend.SetContact(app.plan, form.contact())
	}
	if hasError(problems) {
		app.people.form.problems = plainProblems(problems, form.name())
		app.status = line{text: "not saved", tone: toneProblem}
		return app
	}
	return app.storedEntry(next, form.kind, form.id())
}

func (app App) storedEntry(next model.Plan, kind directoryKind, id string) App {
	app.plan = next
	keep := ""
	if question, ok := app.current(); ok {
		keep = question.ID
	}
	app = app.refreshed(keep)
	app.people.mode = peopleBrowsing
	app.people = app.people.pointedAt(next.Directory, kind, id)
	app = app.refreshedPeople()
	app.status = app.saveStatus(nil)
	return app
}
