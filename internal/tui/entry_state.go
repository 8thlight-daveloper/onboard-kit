package tui

import (
	"github.com/charmbracelet/bubbles/textinput"

	"onboard-kit/internal/model"
)

const (
	inputPrompt   = "> "
	inputPadding  = 4
	inputMinimum  = 10
	examplePrefix = "e.g. "
)

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
	input.Prompt = inputPrompt
	input.SetValue(entryText(question))
	input.CursorEnd()
	input.Focus()
	entry.input = input
	return entry.withWidth(width)
}

func exampleText(question model.Question) string {
	if question.Example == "" {
		return ""
	}
	return examplePrefix + singleLine(question.Example)
}

func (entry entryState) typing() bool { return !picksFromList(entry.question) }

func (entry entryState) withWidth(width int) entryState {
	if !entry.typing() {
		return entry
	}
	entry.input.Width = max(width-inputPadding, inputMinimum)
	entry.input.Placeholder = truncateCells(exampleText(entry.question), entry.input.Width)
	return entry
}

func (entry entryState) styledInputView(colors palette) string {
	input := entry.input
	input.PromptStyle = colors.styles[toneHere]
	input.PlaceholderStyle = colors.styles[toneFaint]
	return input.View()
}

func (entry entryState) nextSuggestion() entryState {
	suggestions := entry.question.Suggestions
	if len(suggestions) == 0 {
		return entry
	}
	chosen := suggestions[entry.suggestion%len(suggestions)]
	entry.suggestion++
	entry.input.SetValue(describeValue(chosen.Value))
	entry.input.CursorEnd()
	return entry
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
