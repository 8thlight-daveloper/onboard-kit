package tui

import "strings"

const (
	railAnswered = "x"
	railDeferred = "~"
	railOpen     = "."
	railHere     = "◆"
	railGap      = "  "
)

func (app App) railPosition() int {
	switch app.screen {
	case screenWalk:
		return app.cursor
	case screenOverview:
		return app.overviewCursor
	}
	return -1
}

func (app App) railRow(width int, colors palette) string {
	if len(app.questions) == 0 || app.helpOpen {
		return ""
	}
	here := app.railPosition()
	var cells []string
	scope := ""
	for index, question := range app.questions {
		gap := ""
		if index > 0 && question.Scope != scope {
			gap = railGap
		}
		scope = question.Scope
		cells = append(cells, gap+app.railCell(index, here, colors))
	}
	row := strings.Join(cells, "")
	if cellWidth(row) > width {
		return ""
	}
	return row
}

func (app App) railCell(index int, here int, colors palette) string {
	if index == here {
		return colors.render(toneHere, true, railHere)
	}
	glyph := railOpen
	switch stateOf(app.questions[index].Answer) {
	case "settled":
		glyph = railAnswered
	case "deferred":
		glyph = railDeferred
	}
	return colors.render(toneFaint, true, glyph)
}
