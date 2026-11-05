package tui

import (
	"os"
	"strings"
)

const railGap = "  "

type railGlyphs struct {
	answered string
	deferred string
	open     string
	here     string
}

func unicodeRailGlyphs() railGlyphs {
	return railGlyphs{answered: "✓", deferred: "~", open: "·", here: "◆"}
}

func plainRailGlyphs() railGlyphs {
	return railGlyphs{answered: "x", deferred: "~", open: ".", here: "*"}
}

func railGlyphsForTerminal() railGlyphs {
	for _, name := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		value := strings.ToLower(os.Getenv(name))
		if value == "" {
			continue
		}
		if strings.Contains(value, "utf-8") || strings.Contains(value, "utf8") {
			return unicodeRailGlyphs()
		}
		return plainRailGlyphs()
	}
	return unicodeRailGlyphs()
}

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
	if len(app.questions) == 0 || app.helpOpen || app.aboutOpen {
		return ""
	}
	here := app.railPosition()
	var cells []string
	scope := ""
	for index, question := range app.questions {
		gap := ""
		if index > 0 && app.sectionRoot(question.Scope) != app.sectionRoot(scope) {
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
		return colors.render(toneHere, true, app.rail.here)
	}
	glyph := app.rail.open
	switch answerState(app.questions[index]) {
	case "settled":
		glyph = app.rail.answered
	case "deferred":
		glyph = app.rail.deferred
	}
	return colors.render(toneFaint, true, glyph)
}
