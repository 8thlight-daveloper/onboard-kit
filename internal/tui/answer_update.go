package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

const (
	stoppedTypingText = "Stopped typing. "
	answerKeptText    = "The answer is kept as it was."
	suggestionText    = "The suggestion is in the box. Enter saves it."
)

func (app App) updateAnswer(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case app.entry.typing():
		return app.updateTyping(key)
	case app.entry.onOptions():
		return app.updateOptions(key)
	}
	return app.updateChoosing(key)
}

func (app App) updateChoosing(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		if app.entry.choice == 0 {
			return app.focusedStrip(), nil
		}
		app.entry = app.entry.moved(-1)
	case "down", "j":
		app.entry = app.entry.moved(1)
	case " ":
		app.entry = app.entry.toggled()
	case "enter":
		return app.confirmChoice(), nil
	}
	return app.updateAnswerLetter(key)
}

func (app App) updateOptions(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		if app.entry.choice == 0 {
			app.entry = app.entry.resumed()
			return app, nil
		}
		app.entry = app.entry.moved(-1)
	case "down", "j":
		app.entry = app.entry.moved(1)
	case "enter":
		return app.chosenOption(), nil
	}
	return app.updateAnswerLetter(key)
}

func (app App) updateAnswerLetter(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "d":
		return app.accept(model.Value{Kind: model.ValueKindDeferred}), nil
	case "x":
		return app.clear(), nil
	case "p", "left", "ctrl+p":
		return app.moveTo(app.cursor - 1), nil
	case "n", "right", "ctrl+n":
		return app.moveTo(app.cursor + 1), nil
	case "esc":
		return app.leaveAnswer(), nil
	case "q":
		return app.quit()
	case "?":
		app.helpOpen = true
		return app, nil
	}
	if target, ok := app.tabWithKey(key.String()); ok {
		return app.openScreen(target), nil
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

func (app App) chosenOption() App {
	option := app.entry.choices[app.entry.choice]
	switch option.action {
	case choiceEdit:
		app.entry = app.entry.resumed()
	case choiceSuggestion:
		app.entry = app.entry.withText(textOf(app.entry.question, option.value)).resumed()
		app.status = line{text: suggestionText, tone: toneStrong}
	case choiceSkip:
		return app.moveTo(app.cursor + 1)
	case choiceClear:
		return app.clear()
	default:
		return app.accept(option.value)
	}
	return app
}

func (app App) updateTyping(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		return app.acceptText(), nil
	case "esc", "down":
		return app.stoppedTyping(), nil
	case "up":
		return app.focusedStrip(), nil
	case "ctrl+d":
		return app.accept(model.Value{Kind: model.ValueKindDeferred}), nil
	case "ctrl+p":
		return app.moveTo(app.cursor - 1), nil
	case "ctrl+n":
		return app.moveTo(app.cursor + 1), nil
	case "f1":
		app.helpOpen = true
		return app, nil
	}
	var command tea.Cmd
	app.entry.input, command = app.entry.input.Update(key)
	return app, command
}

func (app App) stoppedTyping() App {
	app.entry = app.entry.blurred()
	kept := answerKeptText
	if app.entry.hasDraft() {
		kept = draftKeptText
	}
	app.status = line{text: stoppedTypingText + kept, tone: toneStrong}
	return app
}

func (app App) leaveAnswer() App {
	leaving := app.entry.hasDraft()
	if app.returnTo == screenReview {
		app = app.openReview()
	} else {
		app = app.openOverview()
	}
	if leaving {
		app.status = line{text: draftKeptText, tone: toneWarning}
	}
	return app
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
	moved.status = status.followedBy(moved.status)
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
	if !ok {
		return app
	}
	if app.entry.typed() {
		app.entry = app.entry.withText("").resumed()
	}
	if answerState(question) == "unanswered" {
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
