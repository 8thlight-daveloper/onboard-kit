package tui

import (
	"fmt"
	"slices"
	"strings"

	"onboard-kit/internal/model"
)

const (
	answerShareNumerator   = 3
	answerShareDenominator = 5
	welcomeText            = "Welcome. A few questions about you, then your own onboarding document."
	deferredText           = "Deferred, answer later"
	requiredUnansweredText = "Required, not answered yet"
	unansweredText         = "Not answered yet"
	minimumPickRows        = 4
)

func (app App) inputWidth() int {
	answerWidth, _ := app.frame().columns(answerShareNumerator, answerShareDenominator)
	return answerWidth
}

func (app App) answerBody(screenFrame frame, colors palette) []string {
	if len(app.questions) == 0 {
		return stackPane(pane{lines: []line{bodyLine("No questions are active.")}, focused: true}, screenFrame.innerWidth(), screenFrame.bodyHeight, colors)
	}
	question := app.questions[app.cursor]
	height := screenFrame.bodyHeight
	answerWidth, notesWidth := screenFrame.columns(answerShareNumerator, answerShareDenominator)
	page, scrolled := app.pageLines(answerWidth, height, !screenFrame.wide(), colors)
	if screenFrame.wide() {
		notes := pane{lines: clipLines(app.rowNoteLines(question, notesWidth), height)}
		return joinPanes(pane{lines: page, focused: true}, notes, answerWidth, notesWidth, height, colors)
	}
	if !scrolled {
		page = append(append(page, blankLine()), app.rowNoteLines(question, answerWidth)...)
	}
	return stackPane(pane{lines: clipLines(page, height), focused: true}, answerWidth, height, colors)
}

func (app App) rowNoteLines(question model.Question, width int) []line {
	notes := app.noteLines(question, width)
	if first, last := app.pageBounds(app.cursor); last-first > 1 {
		notes[0] = faintLine(fmt.Sprintf("About question %d", app.cursor+1))
	}
	return notes
}

func (app App) pageHasOneRow() bool {
	first, last := app.pageBounds(app.cursor)
	return last-first == 1
}

func (app App) pageLines(width int, height int, narrow bool, colors palette) ([]line, bool) {
	header := app.pageHeader(width)
	area := max(height-len(header), 1)
	rows, spans, focused := app.pageRows(width, area-2, narrow && !app.pageHasOneRow(), colors)
	if len(rows) <= area {
		return append(header, rows...), false
	}
	shown := max(area-2, 1)
	start := scrollStart(rows, spans, focused, shown)
	if start == 0 {
		shown++
	}
	end := min(start+shown, len(rows))
	above, below := hiddenRows(slices.Delete(slices.Clone(spans), focused, focused+1), start, end)
	page := header
	if start > 0 {
		page = append(page, faintLine(moreText(above, "above")))
	}
	page = append(page, rows[start:end]...)
	return append(page, faintLine(moreText(below, "below"))), true
}

func hiddenRows(spans []rowSpan, start int, end int) (int, int) {
	above, below := 0, 0
	for _, span := range spans {
		switch {
		case span.start < start:
			above++
		case span.end > end:
			below++
		}
	}
	return above, below
}

func moreText(count int, where string) string {
	switch count {
	case 0:
		return ""
	case 1:
		return "1 more question " + where
	}
	return fmt.Sprintf("%d more questions %s", count, where)
}

func (app App) pageHeader(width int) []line {
	question := app.questions[app.cursor]
	var content []line
	if app.audience == model.AudienceOnboardee && app.answeredCount() == 0 {
		content = append(lines(toneFaint, wrapText(welcomeText, width)), blankLine())
	}
	content = append(content, strongLine(app.sectionTitle(question.Scope)))
	if first, last := app.pageBounds(app.cursor); last-first > 1 {
		content = append(content, lines(toneFaint, wrapText(fmt.Sprintf("%d of %d answered on this page", app.answeredOnPage(first, last), last-first), width))...)
	}
	position := fmt.Sprintf("Question %s, %s, %s", app.position(), requiredWords(question), stateWords(question))
	content = append(content, lines(toneFaint, wrapText(position, width))...)
	return append(content, blankLine())
}

type rowSpan struct {
	start int
	end   int
	focus int
}

func (app App) pageRows(width int, height int, showExample bool, colors palette) ([]line, []rowSpan, int) {
	first, last := app.pageBounds(app.cursor)
	var content []line
	spans := make([]rowSpan, 0, last-first)
	for index := first; index < last; index++ {
		if index > first {
			content = append(content, blankLine())
		}
		row, focus := app.restingRowLines(app.questions[index], index+1, width), 0
		if index == app.cursor {
			row, focus = app.focusedRowLines(app.questions[index], index+1, width, height, showExample, colors)
		}
		spans = append(spans, rowSpan{start: len(content), end: len(content) + len(row), focus: len(content) + focus})
		content = append(content, row...)
	}
	return content, spans, app.cursor - first
}

func scrollStart(content []line, spans []rowSpan, focused int, height int) int {
	block := spans[focused]
	if block.end-block.start > height {
		return max(min(block.focus-height/2, len(content)-height), 0)
	}
	start := min(max((block.start+block.end)/2-height/2, block.end-height), block.start)
	start = max(min(start, len(content)-height), 0)
	for _, span := range spans {
		switch {
		case span.start >= start || start > span.end || span.end >= block.start:
		case block.end-span.start <= height:
			return span.start
		default:
			return span.end + 1
		}
	}
	return start
}

func rowPrefix(question model.Question, number int) string {
	return fmt.Sprintf("%s %d. ", requiredMark(question), number)
}

func descriptionLines(question model.Question, number int, weight tone, width int) []line {
	return lines(weight, hanging(rowPrefix(question, number), question.Description, width))
}

func rowDetail(question model.Question, number int, text string, width int) line {
	return faintLine(truncateCells(spaces(cellWidth(rowPrefix(question, number)))+text, width))
}

func (app App) focusedRowLines(question model.Question, number int, width int, height int, showExample bool, colors palette) ([]line, int) {
	content := descriptionLines(question, number, toneStrong, width)
	if example := exampleText(question); showExample && example != "" {
		content = append(content, rowDetail(question, number, example, width))
	}
	content = append(content, blankLine())
	problems := app.problemLines(width)
	entry, focus := app.entryLines(width, height-len(content)-len(problems), colors)
	focus += len(content)
	return append(append(content, entry...), problems...), focus
}

func (app App) restingRowLines(question model.Question, number int, width int) []line {
	content := descriptionLines(question, number, toneBody, width)
	if example := exampleText(question); example != "" {
		content = append(content, rowDetail(question, number, example, width))
	}
	return append(content, rowDetail(question, number, app.restingAnswer(question), width))
}

func (app App) restingAnswer(question model.Question) string {
	if text, drafted := app.drafts.typedDraft(question); drafted {
		return draftLabel + singleLine(text)
	}
	if picked, unsaved := app.drafts.unsavedPicks(question); unsaved {
		return draftLabel + strings.Join(knownLabels(question, picked), itemSeparator+" ")
	}
	switch answerState(question) {
	case "settled":
		return savedLabel + singleLine(describeAnswer(question, question.Answer))
	case "deferred":
		return deferredText
	}
	if question.Required {
		return requiredUnansweredText
	}
	return unansweredText
}

func requiredWords(question model.Question) string {
	if question.Required {
		return "required"
	}
	return "optional"
}

func (app App) entryLines(width int, height int, colors palette) ([]line, int) {
	if app.entry.typed() {
		return app.typedEntryLines(width, height, colors)
	}
	if len(app.entry.choices) == 0 && isDirectoryQuestion(app.entry.question) {
		return lines(toneBody, wrapText(emptyDirectoryText(app.entry.question, app.audience), width)), 0
	}
	rows := make([]line, 0, len(app.entry.choices))
	for index := range app.entry.choices {
		rows = append(rows, hereLine(app.choiceText(index), index == app.entry.choice && !app.stripFocused, true, width, colors))
	}
	tail := app.pickingHelp(width)
	if height-len(tail) < minimumPickRows {
		tail = nil
	}
	shown, position, focus := windowRowsAround(rows, app.entry.choice, height-len(tail))
	return append(append(shown, position...), tail...), focus
}

func (app App) typedEntryLines(width int, height int, colors palette) ([]line, int) {
	content := []line{app.fieldLine(width, colors)}
	content = append(content, lines(toneFaint, app.entryHelp(width))...)
	content = append(content, blankLine())
	first := 1
	if app.entry.onOptions() {
		first = 0
	}
	focused := app.entry.onOptions() && !app.stripFocused
	rows := make([]line, 0, len(app.entry.choices))
	for index := first; index < len(app.entry.choices); index++ {
		rows = append(rows, optionLine(app.entry.choices[index].label, focused, index == app.entry.choice, width, colors))
	}
	shown, position, focus := windowRowsAround(rows, max(app.entry.choice-first, 0), height-len(content))
	if !focused {
		focus = -len(content)
	}
	return append(append(content, shown...), position...), len(content) + focus
}

func (app App) fieldLine(width int, colors palette) line {
	if app.typingNow() {
		return line{text: app.entry.inputView(colors), raw: true}
	}
	text := singleLine(app.entry.input.Value())
	switch {
	case strings.TrimSpace(text) == "":
		return faintLine(blurredPrompt + emptyFieldText)
	case app.entry.hasDraft():
		return faintLine(truncateCells(blurredPrompt+draftLabel+text, width))
	}
	return faintLine(truncateCells(blurredPrompt+savedLabel+text, width))
}

func optionLine(label string, focused bool, chosen bool, width int, colors palette) line {
	switch {
	case !focused:
		return line{text: "  " + label, tone: toneFaint}
	case !chosen:
		return line{text: "  " + label}
	}
	return barLine(label, width, colors)
}

func (app App) choiceText(index int) string {
	option := app.entry.choices[index]
	if app.entry.pickingMany() {
		mark := "[ ] "
		if order := app.entry.pickOrder(index); order > 0 {
			mark = fmt.Sprintf("[%d] ", order)
		}
		return mark + option.label
	}
	if answerState(app.entry.question) != "unanswered" && sameValue(option.value, app.entry.question.Answer) {
		return option.label + " (current)"
	}
	return option.label
}

func (app App) pickingHelp(width int) []line {
	if !app.entry.pickingMany() || len(app.entry.choices) == 0 {
		return nil
	}
	content := []line{blankLine()}
	content = append(content, lines(toneFaint, wrapText("Space picks or unpicks. Enter confirms. Pick order is kept.", width))...)
	if len(app.entry.picked) > 1 {
		order := strings.Join(knownLabels(app.entry.question, app.entry.picked), ", then ")
		content = append(content, lines(toneFaint, hanging("", "Order: "+order, width))...)
	}
	return content
}

func windowRows(rows []line, focus int, height int) ([]line, []line) {
	shown, position, _ := windowRowsAround(rows, focus, height)
	return shown, position
}

func windowRowsAround(rows []line, focus int, height int) ([]line, []line, int) {
	if len(rows) <= height {
		return rows, nil, focus
	}
	if height < 2 {
		return rows[focus : focus+1], nil, 0
	}
	shown := scrollWindow(rows, focus, height-1)
	first := min(max(focus-(height-1)/2, 0), len(rows)-len(shown))
	return shown, []line{faintLine(fmt.Sprintf("Showing %d to %d of %d", first+1, first+len(shown), len(rows)))}, focus - first
}

func (app App) entryHelp(width int) []string {
	question := app.entry.question
	var help []string
	switch {
	case question.Type == model.VariableTypeCommand:
		help = append(help, "Words separated by spaces. Quotes keep spaces together.")
	case question.Multiple:
		help = append(help, "Separate people with ;")
	}
	var wrapped []string
	for _, text := range help {
		wrapped = append(wrapped, wrapText(text, width)...)
	}
	return wrapped
}

func (app App) problemLines(width int) []line {
	var content []line
	for _, problem := range app.entry.problems {
		content = append(content, lines(toneProblem, hanging("! problem: ", problem.Message, width))...)
	}
	if len(content) == 0 {
		return nil
	}
	return append([]line{blankLine()}, content...)
}
