package tui

import (
	"onboard-kit/internal/model"
)

const (
	rowPrefixWidth      = 8
	answerColumnMinimum = 30
)

func (a App) outlineLines(width int, selected int, showAnswer bool) ([]line, int) {
	var content []line
	focus := 0
	heading := "\x00"
	for index, question := range a.questions {
		if title := a.groupTitle(question); title != heading {
			heading = title
			content = append(content, headingLine(title))
		}
		if index == selected {
			focus = len(content)
		}
		content = append(content, outlineRow(question, index == selected, width, showAnswer))
	}
	return content, focus
}

func outlineRow(question model.Question, selected bool, width int, showAnswer bool) line {
	marker := "  "
	tone := tonePlain
	if selected {
		marker = "> "
		tone = toneFocus
	}
	prefix := marker + stateMark(question.Answer) + " " + requiredMark(question) + " "
	available := width - rowPrefixWidth
	text := truncateCells(singleLine(question.Description), available)
	if showAnswer && available >= answerColumnMinimum && stateOf(question.Answer) != "unanswered" {
		descriptionWidth := available * 3 / 5
		answerWidth := available - descriptionWidth - 2
		description := truncateCells(singleLine(question.Description), descriptionWidth)
		text = description + spaces(descriptionWidth-cellWidth(description)) + "  " + truncateCells(singleLine(answerText(question.Answer)), answerWidth)
	}
	return line{text: prefix + text, tone: tone}
}
