package tui

import (
	"fmt"

	"onboard-kit/internal/model"
)

const (
	rowPrefixWidth      = 8
	answerColumnMinimum = 36
	rowTextShare        = 3
	rowTextOutOf        = 5
	sectionIndent       = 2
)

func (app App) overviewBody(screenFrame frame, colors palette) []string {
	if len(app.questions) == 0 {
		return stackPane(pane{lines: []line{bodyLine("No questions are active.")}, focused: true}, screenFrame.innerWidth(), screenFrame.bodyHeight, colors)
	}
	listWidth, previewWidth := screenFrame.columns(answerShareNumerator, answerShareDenominator)
	if !screenFrame.wide() {
		listWidth = screenFrame.innerWidth()
	}
	rows, focus := app.outlineLines(listWidth, colors)
	list := pane{lines: scrollWindow(rows, focus, screenFrame.bodyHeight), focused: true}
	if !screenFrame.wide() {
		return stackPane(list, listWidth, screenFrame.bodyHeight, colors)
	}
	preview := pane{lines: clipLines(app.previewLines(app.questions[app.overviewCursor], previewWidth), screenFrame.bodyHeight)}
	return joinPanes(list, preview, listWidth, previewWidth, screenFrame.bodyHeight, colors)
}

func (app App) outlineLines(width int, colors palette) ([]line, int) {
	var content []line
	focus := 0
	heading := "\x00"
	for index, question := range app.questions {
		indent := spaces(sectionIndent * app.sectionDepth(question.Scope))
		if question.Scope != heading {
			if heading != "\x00" {
				content = append(content, blankLine())
			}
			heading = question.Scope
			content = append(content, strongLine(indent+app.sectionTitle(question.Scope)))
		}
		if index == app.overviewCursor {
			focus = len(content)
		}
		content = append(content, outlineRow(question, indent, index == app.overviewCursor, width, colors))
	}
	return content, focus
}

func outlineRow(question model.Question, indent string, selected bool, width int, colors palette) line {
	available := width - rowPrefixWidth - len(indent)
	text := truncateCells(singleLine(question.Description), available)
	if available >= answerColumnMinimum && answerState(question) != "unanswered" {
		descriptionWidth := available * rowTextShare / rowTextOutOf
		answerWidth := available - descriptionWidth - 2
		description := truncateCells(singleLine(question.Description), descriptionWidth)
		answer := truncateCells(singleLine(answerText(question)), answerWidth)
		text = description + spaces(descriptionWidth-cellWidth(description)) + "  " + answer
	}
	body := indent + stateMark(question) + " " + requiredMark(question) + " " + text
	return hereLine(body, selected, true, width, colors)
}

func answerText(question model.Question) string {
	switch answerState(question) {
	case "unanswered":
		return "none yet"
	case "deferred":
		return "deferred, answer later"
	}
	return describeAnswer(question, question.Answer)
}

func (app App) previewLines(question model.Question, width int) []line {
	index := app.indexOf(question.ID)
	content := []line{
		faintLine(app.sectionTitle(question.Scope)),
		faintLine(fmt.Sprintf("%d of %d, %s, %s", index+1, len(app.questions), requiredWords(question), stateWords(question))),
		blankLine(),
	}
	content = append(content, lines(toneFaint, wrapText(question.Description, width))...)
	content = append(content, blankLine())
	content = append(content, lines(toneFaint, hanging("", "Answer: "+answerText(question), width))...)
	content = append(content, blankLine())
	return append(content, app.noteLines(question, width)...)
}
