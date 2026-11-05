package tui

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"onboard-kit/internal/model"
)

type rowDraft struct {
	text   string
	picked []string
	answer model.Value
}

type rowDrafts map[string]rowDraft

func (drafts rowDrafts) kept(entry entryState) rowDrafts {
	next := rowDrafts{}
	maps.Copy(next, drafts)
	delete(next, entry.question.ID)
	if entry.hasDraft() || entry.hasUnsavedPicks() {
		next[entry.question.ID] = rowDraft{text: entry.input.Value(), picked: slices.Clone(entry.picked), answer: entry.question.Answer}
	}
	return next
}

func (drafts rowDrafts) restored(entry entryState) entryState {
	draft, found := drafts[entry.question.ID]
	if !found || !sameAnswer(draft.answer, entry.question.Answer) {
		return entry
	}
	if entry.typed() {
		return entry.withText(draft.text)
	}
	entry.picked = slices.Clone(draft.picked)
	return entry
}

func (drafts rowDrafts) typedDraft(question model.Question) (string, bool) {
	draft, found := drafts[question.ID]
	if !found || picksFromList(question) || !sameAnswer(draft.answer, question.Answer) {
		return "", false
	}
	value := parseEntry(question, draft.text)
	return draft.text, value.Kind != model.ValueKindNone && !sameAnswer(value, question.Answer)
}

func (drafts rowDrafts) unsavedPicks(question model.Question) ([]string, bool) {
	draft, found := drafts[question.ID]
	if !found || !picksMany(question) || !sameAnswer(draft.answer, question.Answer) {
		return nil, false
	}
	return draft.picked, !slices.Equal(draft.picked, pickedFrom(question))
}

func (app App) pageBounds(index int) (int, int) {
	if index < 0 || index >= len(app.questions) {
		return 0, 0
	}
	scope := app.questions[index].Scope
	first, last := index, index+1
	for first > 0 && app.questions[first-1].Scope == scope {
		first--
	}
	for last < len(app.questions) && app.questions[last].Scope == scope {
		last++
	}
	return first, last
}

func (app App) onFirstRow() bool {
	first, _ := app.pageBounds(app.cursor)
	return app.cursor == first
}

func (app App) onLastRow() bool {
	_, last := app.pageBounds(app.cursor)
	return app.cursor == last-1
}

func (app App) rowAbove() App {
	if app.onFirstRow() {
		return app.focusedStrip()
	}
	moved := app.openQuestion(app.cursor - 1)
	moved.entry = moved.entry.landedAtBottom()
	return moved
}

func (app App) rowBelow() App {
	if app.onLastRow() {
		return app
	}
	moved := app.openQuestion(app.cursor + 1)
	moved.entry = moved.entry.landedAtTop()
	return moved
}

func (app App) draftRows() []int {
	var numbers []int
	scope := app.entry.question.Scope
	for index, question := range app.questions {
		if question.Scope != scope {
			continue
		}
		if question.ID == app.entry.question.ID {
			if app.entry.hasDraft() {
				numbers = append(numbers, index+1)
			}
			continue
		}
		if _, drafted := app.drafts.typedDraft(question); drafted {
			numbers = append(numbers, index+1)
		}
	}
	return numbers
}

func questionNumbers(numbers []int) string {
	words := make([]string, 0, len(numbers))
	for _, number := range numbers {
		words = append(words, strconv.Itoa(number))
	}
	if len(words) == 1 {
		return "question " + words[0]
	}
	return "questions " + strings.Join(words[:len(words)-1], ", ") + " and " + words[len(words)-1]
}

func draftKeptLine(numbers []int) line {
	return line{text: fmt.Sprintf("Your text on %s is kept, not saved yet.", questionNumbers(numbers)), tone: toneWarning}
}

func draftDroppedLine(numbers []int) line {
	return line{text: fmt.Sprintf("Your unsaved text on %s was dropped.", questionNumbers(numbers)), tone: toneWarning}
}

func (app App) answeredOnPage(first int, last int) int {
	count := 0
	for _, question := range app.questions[first:last] {
		if answerState(question) != "unanswered" {
			count++
		}
	}
	return count
}
