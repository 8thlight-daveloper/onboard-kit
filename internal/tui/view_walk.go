package tui

import (
	"fmt"
	"strings"

	"onboard-kit/internal/model"
)

func (a App) walkBody(l layout, p palette) []string {
	if len(a.questions) == 0 {
		return singlePane([]line{plainLine("No questions are active.")}, l, p)
	}
	right := a.walkPane(l.mainWidth)
	if !l.twoPanes {
		return singlePane(right, l, p)
	}
	outline, focus := a.outlineLines(l.listWidth, a.cursor, false)
	return sideBySide(scrollWindow(outline, focus, l.bodyHeight), clipLines(right, l.bodyHeight), l, p)
}

func (a App) walkPane(width int) []line {
	question := a.questions[a.cursor]
	content := a.questionLines(question, width)
	content = append(content, a.entryLines(width)...)
	content = append(content, a.problemLines(width)...)
	content = append(content, plainLine(""))
	return append(content, a.detailLines(question, width)...)
}

func (a App) questionLines(question model.Question, width int) []line {
	required := "optional"
	if question.Required {
		required = "* required"
	}
	content := []line{
		headingLine(a.groupTitle(question)),
		mutedLine(fmt.Sprintf("%d of %d  %s", a.cursor+1, len(a.questions), required)),
		plainLine(""),
	}
	content = append(content, lines(tonePlain, wrapText(question.Description, width))...)
	content = append(content, plainLine(""))
	return append(content, lines(tonePlain, hanging(stateMark(question.Answer)+" ", "Answer: "+answerText(question.Answer), width))...)
}

func answerText(value model.Value) string {
	switch stateOf(value) {
	case "unanswered":
		return "none yet"
	case "deferred":
		return "deferred, answer later"
	}
	return describeValue(value)
}

func (a App) entryLines(width int) []line {
	content := []line{plainLine("")}
	if a.entry.typing() {
		content = append(content, line{text: a.entry.input.View(), raw: true})
		return append(content, lines(toneMuted, a.entryHelp(width))...)
	}
	for index, option := range a.entry.choices {
		marker := "  "
		tone := tonePlain
		if index == a.entry.choice {
			marker = "> "
			tone = toneFocus
		}
		text := option.label
		if stateOf(a.entry.question.Answer) != "unanswered" && sameValue(option.value, a.entry.question.Answer) {
			text += " (current)"
		}
		content = append(content, line{text: marker + text, tone: tone})
	}
	return content
}

func (a App) entryHelp(width int) []string {
	question := a.entry.question
	var help []string
	switch {
	case question.Type == model.VariableTypeCommand:
		help = append(help, "Words separated by spaces. Quotes keep spaces together.")
	case question.Multiple:
		help = append(help, "Separate people with ;")
	}
	if len(question.Suggestions) > 0 {
		help = append(help, "Tab tries a suggestion: "+suggestionList(question))
	}
	var wrapped []string
	for _, text := range help {
		wrapped = append(wrapped, wrapText(text, width)...)
	}
	return wrapped
}

func (a App) problemLines(width int) []line {
	var content []line
	for _, problem := range a.entry.problems {
		content = append(content, lines(toneProblem, hanging("! problem: ", problem.Message, width))...)
	}
	return content
}

func suggestionList(question model.Question) string {
	texts := make([]string, 0, len(question.Suggestions))
	for _, suggestion := range question.Suggestions {
		texts = append(texts, describeValue(suggestion.Value))
	}
	return strings.Join(texts, ", ")
}
