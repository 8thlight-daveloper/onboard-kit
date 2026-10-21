package tui

import (
	"slices"

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
	picked     []string
	input      textinput.Model
	suggestion int
	problems   []model.Problem
}

func newEntryState(question model.Question, width int) entryState {
	entry := entryState{question: question}
	if picksFromList(question) {
		entry.choices = choicesFor(question)
		entry.picked = pickedFrom(question)
		for index, option := range entry.choices {
			if len(entry.picked) > 0 && option.value.Text == entry.picked[0] {
				entry.choice = index
			}
			if !entry.pickingMany() && answerState(question) != "unanswered" && sameValue(option.value, question.Answer) {
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

func pickedFrom(question model.Question) []string {
	if !picksMany(question) {
		return nil
	}
	var ids []string
	switch question.Answer.Kind {
	case model.ValueKindText:
		ids = []string{question.Answer.Text}
	case model.ValueKindList:
		ids = question.Answer.List
	}
	var known []string
	for _, id := range ids {
		if _, ok := labelOf(question.Choices, id); ok && !slices.Contains(known, id) {
			known = append(known, id)
		}
	}
	return known
}

func (entry entryState) pickingMany() bool { return picksMany(entry.question) }

func (entry entryState) moved(delta int) entryState {
	entry.choice = max(min(entry.choice+delta, len(entry.choices)-1), 0)
	return entry
}

func (entry entryState) toggled() entryState {
	if !entry.pickingMany() || len(entry.choices) == 0 {
		return entry
	}
	id := entry.choices[entry.choice].value.Text
	if index := slices.Index(entry.picked, id); index >= 0 {
		entry.picked = slices.Delete(slices.Clone(entry.picked), index, index+1)
		return entry
	}
	entry.picked = append(slices.Clone(entry.picked), id)
	return entry
}

func (entry entryState) pickedValue() model.Value {
	if len(entry.picked) == 0 {
		return model.Value{}
	}
	return model.Value{Kind: model.ValueKindList, List: entry.picked}
}

func (entry entryState) pickOrder(index int) int {
	return slices.Index(entry.picked, entry.choices[index].value.Text) + 1
}
