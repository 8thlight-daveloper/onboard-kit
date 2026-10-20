package tui

import (
	"fmt"
	"strings"

	"onboard-kit/internal/model"
)

const (
	answerShareNumerator   = 3
	answerShareDenominator = 5
	welcomeText            = "Welcome. A few questions about you, then your own onboarding document."
)

func (app App) inputWidth() int {
	answerWidth, _ := app.frame().columns(answerShareNumerator, answerShareDenominator)
	return answerWidth
}

func (app App) walkBody(screenFrame frame, colors palette) []string {
	if len(app.questions) == 0 {
		return stackPane(pane{lines: []line{bodyLine("No questions are active.")}, focused: true}, screenFrame.innerWidth(), screenFrame.bodyHeight, colors)
	}
	question := app.questions[app.cursor]
	answerWidth, notesWidth := screenFrame.columns(answerShareNumerator, answerShareDenominator)
	answer := app.answerLines(question, answerWidth, colors)
	if screenFrame.wide() {
		notes := pane{lines: clipLines(app.noteLines(question, notesWidth), screenFrame.bodyHeight)}
		answerPane := pane{lines: clipLines(answer, screenFrame.bodyHeight), focused: true}
		return joinPanes(answerPane, notes, answerWidth, notesWidth, screenFrame.bodyHeight, colors)
	}
	combined := append(append(answer, blankLine()), app.noteLines(question, answerWidth)...)
	return stackPane(pane{lines: clipLines(combined, screenFrame.bodyHeight), focused: true}, answerWidth, screenFrame.bodyHeight, colors)
}

func (app App) answerLines(question model.Question, width int, colors palette) []line {
	var content []line
	if app.audience == model.AudienceOnboardee && app.answeredCount() == 0 {
		content = append(lines(toneFaint, wrapText(welcomeText, width)), blankLine())
	}
	content = append(content,
		strongLine(app.sectionTitle(question.Scope)),
		faintLine(fmt.Sprintf("%s, %s, %s", app.position(), requiredWords(question), stateWords(question.Answer))),
		blankLine(),
	)
	content = append(content, lines(toneBody, wrapText(question.Description, width))...)
	content = append(content, blankLine())
	content = append(content, app.entryLines(width, colors)...)
	return append(content, app.problemLines(width)...)
}

func requiredWords(question model.Question) string {
	if question.Required {
		return "required"
	}
	return "optional"
}

func (app App) entryLines(width int, colors palette) []line {
	if app.entry.typing() {
		content := []line{{text: app.entry.styledInputView(colors), raw: true}}
		return append(content, lines(toneFaint, app.entryHelp(width))...)
	}
	content := make([]line, 0, len(app.entry.choices))
	for index, option := range app.entry.choices {
		text := option.label
		if stateOf(app.entry.question.Answer) != "unanswered" && sameValue(option.value, app.entry.question.Answer) {
			text += " (current)"
		}
		content = append(content, hereLine(text, index == app.entry.choice, true, width, colors))
	}
	return content
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
	if len(question.Suggestions) > 0 {
		help = append(help, "Tab tries a suggestion: "+suggestionList(question))
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

func suggestionList(question model.Question) string {
	texts := make([]string, 0, len(question.Suggestions))
	for _, suggestion := range question.Suggestions {
		texts = append(texts, describeValue(suggestion.Value))
	}
	return strings.Join(texts, ", ")
}
