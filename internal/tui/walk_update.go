package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

func (app App) updateWalk(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if app.entry.typing() {
		return app.updateTyping(key)
	}
	return app.updateChoosing(key)
}

func (app App) updateChoosing(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		app.entry = app.entry.moved(-1)
	case "down", "j":
		app.entry = app.entry.moved(1)
	case " ":
		app.entry = app.entry.toggled()
	case "enter":
		return app.confirmChoice(), nil
	case "d":
		return app.accept(model.Value{Kind: model.ValueKindDeferred}), nil
	case "x":
		return app.clear(), nil
	case "p", "left", "shift+tab", "ctrl+p":
		return app.moveTo(app.cursor - 1), nil
	case "n", "right", "ctrl+n":
		return app.moveTo(app.cursor + 1), nil
	case "o":
		return app.openOverview(), nil
	case "r":
		return app.openReview(), nil
	case "s":
		return app.openSections(), nil
	case "c":
		return app.openContacts(), nil
	case "t":
		return app.openPreparation(), nil
	case "i":
		return app.openInvitation(), nil
	case "esc":
		return app.leaveWalk(), nil
	case "q":
		return app.quit()
	case "?":
		app.helpOpen = true
	}
	return app, nil
}

func (app App) confirmChoice() App {
	switch {
	case app.entry.pickingMany():
		if len(app.entry.picked) == 0 && answerState(app.entry.question) != "settled" {
			return app.moveTo(app.cursor + 1)
		}
		return app.accept(app.entry.pickedValue())
	case len(app.entry.choices) == 0:
		return app.moveTo(app.cursor + 1)
	}
	return app.accept(app.entry.choices[app.entry.choice].value)
}

func (app App) updateTyping(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		return app.acceptText(), nil
	case "tab":
		app.entry = app.entry.nextSuggestion()
		return app, nil
	case "ctrl+d":
		return app.accept(model.Value{Kind: model.ValueKindDeferred}), nil
	case "shift+tab", "ctrl+p":
		return app.moveTo(app.cursor - 1), nil
	case "ctrl+n":
		return app.moveTo(app.cursor + 1), nil
	case "esc":
		return app.leaveWalk(), nil
	case "f1":
		app.helpOpen = true
		return app, nil
	}
	var command tea.Cmd
	app.entry.input, command = app.entry.input.Update(key)
	return app, command
}

func (app App) leaveWalk() App {
	if app.returnTo == screenReview {
		return app.openReview()
	}
	return app.openOverview()
}

func (app App) accept(value model.Value) App {
	question, ok := app.current()
	if !ok {
		return app
	}
	if sameAnswer(question.Answer, value) {
		return app.moveTo(app.cursor + 1)
	}
	next, saved := app.submit(value)
	if !saved {
		return next
	}
	status := next.status
	moved := next.moveTo(next.cursor + 1)
	moved.status = status
	return moved
}

func (app App) acceptText() App {
	question, ok := app.current()
	if !ok {
		return app
	}
	value := parseEntry(question, app.entry.input.Value())
	if value.Kind == model.ValueKindNone && answerState(question) != "settled" {
		return app.moveTo(app.cursor + 1)
	}
	return app.accept(value)
}

func (app App) clear() App {
	question, ok := app.current()
	if !ok || answerState(question) == "unanswered" {
		return app
	}
	next, saved := app.submit(model.Value{})
	if !saved {
		return next
	}
	status := next.status
	reopened := next.openQuestion(next.cursor)
	reopened.status = status
	return reopened
}
