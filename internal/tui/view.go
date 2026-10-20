package tui

import (
	"fmt"
	"strings"

	"onboard-kit/internal/model"
)

func (a App) View() string {
	l := layoutFor(a.width, a.height)
	p := newPalette()
	if l.tooSmall() {
		return a.smallWindowMessage()
	}
	rows := []string{renderRow(a.header(), l.width, p)}
	rows = append(rows, a.body(l, p)...)
	rows = append(rows, renderRow(a.status, l.width, p))
	rows = append(rows, a.hintRow(l, p))
	return strings.Join(rows, "\n")
}

func (a App) smallWindowMessage() string {
	text := fmt.Sprintf("Please make the window larger, at least %d columns by %d rows.", minimumWidth, minimumHeight)
	rows := wrapText(text, max(a.width, 1))
	return strings.Join(rows[:min(len(rows), max(a.height, 1))], "\n")
}

func (a App) body(l layout, p palette) []string {
	if a.helpOpen {
		return singlePane(a.helpLines(l.width), l, p)
	}
	switch a.screen {
	case screenOverview:
		return a.overviewBody(l, p)
	case screenReview:
		return a.reviewBody(l, p)
	case screenSections:
		return singlePane(a.sectionLines(l), l, p)
	}
	return a.walkBody(l, p)
}

func (a App) header() line {
	title := "Onboarding plan"
	if a.audience == model.AudienceOnboardee {
		title = "Your onboarding"
	}
	return headingLine(title + "  " + a.screenName() + "  " + a.progress())
}

func (a App) screenName() string {
	switch a.screen {
	case screenOverview:
		return "Overview"
	case screenReview:
		return "Review"
	case screenSections:
		return "Sections"
	}
	return a.position()
}

func (a App) position() string {
	if len(a.questions) == 0 {
		return "No questions"
	}
	return fmt.Sprintf("%d of %d", a.cursor+1, len(a.questions))
}

func (a App) progress() string {
	return fmt.Sprintf("%d of %d answered", a.answeredCount(), len(a.questions))
}

func (a App) answeredCount() int {
	count := 0
	for _, question := range a.questions {
		if stateOf(question.Answer) != "unanswered" {
			count++
		}
	}
	return count
}

func (a App) sectionTitle(scope string) string {
	if scope == "plan" {
		return "Plan"
	}
	for _, section := range a.plan.Sections {
		if section.ID == scope {
			return section.Title
		}
	}
	return scope
}

func (a App) groupTitle(question model.Question) string {
	return a.sectionTitle(question.Scope)
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

func requiredMark(question model.Question) string {
	if question.Required {
		return "*"
	}
	return " "
}
