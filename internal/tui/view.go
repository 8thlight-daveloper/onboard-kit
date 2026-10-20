package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"onboard-kit/internal/model"
)

func titleStyle() lipgloss.Style { return lipgloss.NewStyle().Bold(true) }
func cursorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
}
func dimStyle() lipgloss.Style  { return lipgloss.NewStyle().Faint(true) }
func warnStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(lipgloss.Color("3")) }
func problemStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
}

func (a App) View() string {
	switch a.screen {
	case screenEdit:
		return a.viewEdit()
	case screenSections:
		return a.viewSections()
	case screenReview:
		return titleStyle().Render("Review") + "\n" + a.review.View() + "\n" + dimStyle().Render("up and down scroll, esc goes back")
	}
	return a.viewQuestions()
}

func (a App) viewQuestions() string {
	var view strings.Builder
	view.WriteString(titleStyle().Render("Onboarding plan") + "  " + dimStyle().Render(a.progress()) + "\n\n")
	if len(a.questions) == 0 {
		view.WriteString("No questions are active.\n")
	}
	scope := "\x00"
	first, last := a.window(len(a.questions), a.cursor, max(a.height-8, 5))
	for index := first; index < last; index++ {
		question := a.questions[index]
		if question.Scope != scope {
			scope = question.Scope
			view.WriteString(dimStyle().Render(scopeTitle(scope)) + "\n")
		}
		view.WriteString(a.questionLine(index, question) + "\n")
	}
	view.WriteString("\n" + a.messageLine())
	view.WriteString(dimStyle().Render("enter answer  d defer  x clear  r review  q quit"+a.sectionsHint()) + "\n")
	return view.String()
}

func (a App) questionLine(index int, question model.Question) string {
	marker := "  "
	if index == a.cursor {
		marker = "> "
	}
	required := " "
	if question.Required {
		required = "*"
	}
	line := fmt.Sprintf("%s%s %s %-34s %s", marker, stateMark(question.Answer), required, truncate(question.Description, 34), describeValue(question.Answer))
	if index == a.cursor {
		return cursorStyle().Render(line)
	}
	return line
}

func (a App) viewEdit() string {
	question := a.edit.question
	var view strings.Builder
	view.WriteString(titleStyle().Render(question.Description) + "\n")
	view.WriteString(dimStyle().Render(question.ID+", "+string(effectiveType(question))) + "\n\n")
	if picksFromList(question) {
		for index, option := range a.edit.choices {
			marker := "  "
			if index == a.edit.choice {
				marker = "> "
			}
			view.WriteString(marker + option.label + "\n")
		}
		view.WriteString("\n" + dimStyle().Render("up and down, enter chooses, esc cancels") + "\n")
	} else {
		view.WriteString(a.edit.input.View() + "\n\n")
		view.WriteString(a.entryHelp(question))
	}
	for _, problem := range a.edit.problems {
		view.WriteString(problemStyle().Render(problem.Message) + "\n")
	}
	return view.String()
}

func (a App) entryHelp(question model.Question) string {
	var help strings.Builder
	switch {
	case question.Type == model.VariableTypeCommand:
		help.WriteString(dimStyle().Render("words separated by spaces, quotes keep spaces together") + "\n")
	case question.Multiple:
		help.WriteString(dimStyle().Render("separate people with ;") + "\n")
	}
	if len(question.Suggestions) > 0 {
		help.WriteString(dimStyle().Render("tab tries a suggestion: "+suggestionList(question)) + "\n")
	}
	help.WriteString(dimStyle().Render("enter saves, ctrl+d defers, esc cancels") + "\n")
	return help.String()
}

func (a App) viewSections() string {
	var view strings.Builder
	view.WriteString(titleStyle().Render("Sections") + "\n\n")
	for index, section := range a.plan.Sections {
		marker := "  "
		if index == a.sectionCursor {
			marker = "> "
		}
		box := "[x]"
		if !isOn(section) {
			box = "[ ]"
		}
		line := marker + box + " " + section.Title
		if index == a.sectionCursor {
			line = cursorStyle().Render(line)
		}
		view.WriteString(line + "\n")
	}
	view.WriteString("\n" + a.messageLine())
	view.WriteString(dimStyle().Render("space turns a section on or off, esc goes back") + "\n")
	return view.String()
}

func (a App) messageLine() string {
	if a.message == "" {
		return ""
	}
	if strings.HasPrefix(a.message, "saved") {
		return dimStyle().Render(a.message) + "\n"
	}
	return warnStyle().Render(a.message) + "\n"
}

func (a App) sectionsHint() string {
	if a.audience == model.AudienceOnboardee {
		return ""
	}
	return "  s sections"
}

func (a App) progress() string {
	settled := 0
	for _, question := range a.questions {
		if stateOf(question.Answer) != "unanswered" {
			settled++
		}
	}
	return fmt.Sprintf("%d of %d answered", settled, len(a.questions))
}

func (a App) window(total int, cursor int, size int) (int, int) {
	if total <= size {
		return 0, total
	}
	first := max(min(cursor-size/2, total-size), 0)
	return first, first + size
}

func stateMark(value model.Value) string {
	switch stateOf(value) {
	case "settled":
		return "[x]"
	case "deferred":
		return "[~]"
	}
	return "[ ]"
}

func scopeTitle(scope string) string {
	if scope == "plan" {
		return "Plan"
	}
	return scope
}

func effectiveType(question model.Question) model.VariableType {
	if question.Type == "" {
		return model.DefaultVariableType
	}
	return question.Type
}

func suggestionList(question model.Question) string {
	var texts []string
	for _, suggestion := range question.Suggestions {
		texts = append(texts, describeValue(suggestion.Value))
	}
	return strings.Join(texts, ", ")
}

func truncate(text string, width int) string {
	runes := []rune(text)
	if len(runes) <= width {
		return text
	}
	return string(runes[:width-1]) + "…"
}
