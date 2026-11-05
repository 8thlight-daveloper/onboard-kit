package tui

import (
	"slices"

	"github.com/charmbracelet/bubbles/textinput"

	"onboard-kit/internal/model"
)

const (
	inputPrompt    = "> "
	blurredPrompt  = "  "
	inputPadding   = 4
	inputMinimum   = 10
	examplePrefix  = "e.g. "
	emptyFieldText = "Nothing typed yet"
	draftLabel     = "Not saved yet: "
	savedLabel     = "Saved: "
)

type entryFocus int

const (
	focusField entryFocus = iota
	focusOptions
)

type entryState struct {
	question model.Question
	choices  []choice
	choice   int
	picked   []string
	input    textinput.Model
	focus    entryFocus
	problems []model.Problem
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
	entry.choices = typedChoicesFor(question)
	input := textinput.New()
	input.Prompt = inputPrompt
	input.Focus()
	entry.input = input
	return entry.withText(entryText(question)).withWidth(width)
}

func exampleText(question model.Question) string {
	if question.Example == "" {
		return ""
	}
	return examplePrefix + singleLine(question.Example)
}

func (entry entryState) typed() bool { return !picksFromList(entry.question) }

func (entry entryState) typing() bool { return entry.typed() && entry.focus == focusField }

func (entry entryState) onOptions() bool { return entry.typed() && entry.focus == focusOptions }

func (entry entryState) withWidth(width int) entryState {
	if !entry.typed() {
		return entry
	}
	entry.input.Width = max(width-inputPadding, inputMinimum)
	entry.input.Placeholder = truncateCells(exampleText(entry.question), entry.input.Width)
	entry.input.SetCursor(entry.input.Position())
	return entry
}

func (entry entryState) withText(text string) entryState {
	entry.input.SetValue(text)
	entry.input.CursorEnd()
	return entry
}

func (entry entryState) resumed() entryState {
	entry.focus = focusField
	entry.choice = 0
	return entry
}

func (entry entryState) blurred() entryState {
	entry.focus = focusOptions
	entry.choice = 0
	return entry
}

func (entry entryState) hasDraft() bool {
	value := parseEntry(entry.question, entry.input.Value())
	return entry.typed() && value.Kind != model.ValueKindNone && !sameAnswer(value, entry.question.Answer)
}

func (entry entryState) keepsDraftFor(question model.Question) bool {
	return entry.hasDraft() && entry.question.ID == question.ID && sameAnswer(entry.question.Answer, question.Answer)
}

func (entry entryState) inputView(colors palette) string {
	input := entry.input
	input.PlaceholderStyle = colors.styles[toneFaint]
	input.PromptStyle = colors.styles[toneHere]
	return input.View()
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
