package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"onboard-kit/internal/model"
)

func (a App) updateWalk(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.entry.typing() {
		return a.updateTyping(key)
	}
	return a.updateChoosing(key)
}

func (a App) updateChoosing(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		a.entry.choice = max(a.entry.choice-1, 0)
	case "down", "j":
		a.entry.choice = min(a.entry.choice+1, len(a.entry.choices)-1)
	case "enter":
		return a.accept(a.entry.choices[a.entry.choice].value), nil
	case "d":
		return a.accept(model.Value{Kind: model.ValueKindDeferred}), nil
	case "x":
		return a.clear(), nil
	case "p", "left", "shift+tab", "ctrl+p":
		return a.moveTo(a.cursor - 1), nil
	case "n", "right", "ctrl+n":
		return a.moveTo(a.cursor + 1), nil
	case "o":
		return a.openOverview(), nil
	case "r":
		return a.openReview(), nil
	case "s":
		return a.openSections(), nil
	case "esc":
		return a.leaveWalk(), nil
	case "q":
		return a.quit()
	case "?":
		a.helpOpen = true
	}
	return a, nil
}

func (a App) updateTyping(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		return a.acceptText(), nil
	case "tab":
		a.entry = a.entry.nextSuggestion()
		return a, nil
	case "ctrl+d":
		return a.accept(model.Value{Kind: model.ValueKindDeferred}), nil
	case "shift+tab", "ctrl+p":
		return a.moveTo(a.cursor - 1), nil
	case "ctrl+n":
		return a.moveTo(a.cursor + 1), nil
	case "esc":
		return a.leaveWalk(), nil
	case "f1":
		a.helpOpen = true
		return a, nil
	}
	var command tea.Cmd
	a.entry.input, command = a.entry.input.Update(key)
	return a, command
}

func (a App) leaveWalk() App {
	if a.returnTo == screenReview {
		return a.openReview()
	}
	return a.openOverview()
}

func (a App) accept(value model.Value) App {
	question, ok := a.current()
	if !ok {
		return a
	}
	if sameAnswer(question.Answer, value) {
		return a.moveTo(a.cursor + 1)
	}
	next, saved := a.submit(value)
	if !saved {
		return next
	}
	status := next.status
	moved := next.moveTo(next.cursor + 1)
	moved.status = status
	return moved
}

func (a App) acceptText() App {
	question, ok := a.current()
	if !ok {
		return a
	}
	value := parseEntry(question, a.entry.input.Value())
	if value.Kind == model.ValueKindNone && stateOf(question.Answer) != "settled" {
		return a.moveTo(a.cursor + 1)
	}
	return a.accept(value)
}

func (a App) clear() App {
	question, ok := a.current()
	if !ok || stateOf(question.Answer) == "unanswered" {
		return a
	}
	next, saved := a.submit(model.Value{})
	if !saved {
		return next
	}
	status := next.status
	reopened := next.openQuestion(next.cursor)
	reopened.status = status
	return reopened
}
