package tui

import (
	"fmt"
	"strings"

	"onboard-kit/internal/model"
)

func (app App) View() string {
	screenFrame := app.frame()
	if screenFrame.tooSmall() {
		return app.smallWindowMessage()
	}
	colors := newPalette()
	inner := screenFrame.innerWidth()
	rows := []string{app.headerRow(inner, colors), app.railRow(inner, colors), ""}
	rows = append(rows, app.body(screenFrame, colors)...)
	rows = append(rows, renderCell(app.status, inner, true, colors), app.hintRow(inner, colors))
	margin := spaces(sideMargin)
	for index, row := range rows {
		rows[index] = strings.TrimRight(margin+row, " ")
	}
	return strings.Join(rows, "\n")
}

func (app App) smallWindowMessage() string {
	text := fmt.Sprintf("Please make the window larger, at least %d columns by %d rows.", minimumWidth, minimumHeight)
	rows := wrapText(text, max(app.width, 1))
	return strings.Join(rows[:min(len(rows), max(app.height, 1))], "\n")
}

func (app App) body(screenFrame frame, colors palette) []string {
	if app.helpOpen {
		return stackPane(pane{lines: app.helpLines(screenFrame.innerWidth()), focused: true}, screenFrame.innerWidth(), screenFrame.bodyHeight, colors)
	}
	switch app.screen {
	case screenOverview:
		return app.overviewBody(screenFrame, colors)
	case screenReview:
		return app.reviewBody(screenFrame, colors)
	case screenSections:
		return stackPane(pane{lines: app.sectionLines(screenFrame, colors), focused: true}, screenFrame.innerWidth(), screenFrame.bodyHeight, colors)
	}
	return app.walkBody(screenFrame, colors)
}

func (app App) headerRow(width int, colors palette) string {
	title := "Onboarding plan"
	if app.audience == model.AudienceOnboardee {
		title = "Your onboarding"
	}
	left := colors.render(toneStrong, true, title) + colors.render(toneFaint, true, "  "+app.screenName())
	right := app.progressText()
	gap := width - cellWidth(left) - cellWidth(right)
	if gap < 2 {
		return left
	}
	return left + spaces(gap) + colors.render(toneFaint, true, right)
}

func (app App) screenName() string {
	switch app.screen {
	case screenOverview:
		return "Overview"
	case screenReview:
		return "Review"
	case screenSections:
		return "Sections"
	}
	return "Walk"
}

func (app App) progressText() string {
	total := len(app.questions)
	left := total - app.answeredCount()
	if app.audience == model.AudienceOnboardee {
		switch {
		case total == 0:
			return "no questions"
		case left == 0:
			return "all answered"
		}
		return fmt.Sprintf("%d of %d to answer", left, total)
	}
	return fmt.Sprintf("%d of %d answered", app.answeredCount(), total)
}

func (app App) position() string {
	return fmt.Sprintf("%d of %d", app.cursor+1, len(app.questions))
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

func stateWords(value model.Value) string {
	switch stateOf(value) {
	case "settled":
		return "answered"
	case "deferred":
		return "deferred"
	}
	return "not answered yet"
}

func requiredMark(question model.Question) string {
	if question.Required {
		return "*"
	}
	return " "
}
