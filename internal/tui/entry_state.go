package tui

import (
	"github.com/charmbracelet/bubbles/textinput"

	"onboard-kit/internal/model"
)

const inputMargin = 6

type entryState struct {
	question   model.Question
	choices    []choice
	choice     int
	input      textinput.Model
	suggestion int
	problems   []model.Problem
}

func newEntryState(question model.Question, width int) entryState {
	entry := entryState{question: question}
	if picksFromList(question) {
		entry.choices = choicesFor(question)
		for index, option := range entry.choices {
			if stateOf(question.Answer) != "unanswered" && sameValue(option.value, question.Answer) {
				entry.choice = index
			}
		}
		return entry
	}
	input := textinput.New()
	input.SetValue(entryText(question))
	input.CursorEnd()
	input.Focus()
	input.Width = max(width-inputMargin, 10)
	entry.input = input
	return entry
}

func (e entryState) typing() bool { return !picksFromList(e.question) }

func (e entryState) withWidth(width int) entryState {
	e.input.Width = max(width-inputMargin, 10)
	return e
}

func (e entryState) nextSuggestion() entryState {
	suggestions := e.question.Suggestions
	if len(suggestions) == 0 {
		return e
	}
	chosen := suggestions[e.suggestion%len(suggestions)]
	e.suggestion++
	e.input.SetValue(describeValue(chosen.Value))
	e.input.CursorEnd()
	return e
}

func sameValue(left model.Value, right model.Value) bool {
	return left.Kind == right.Kind && describeValue(left) == describeValue(right)
}

func sameAnswer(left model.Value, right model.Value) bool {
	if stateOf(left) == "unanswered" && stateOf(right) == "unanswered" {
		return true
	}
	return sameValue(left, right)
}
