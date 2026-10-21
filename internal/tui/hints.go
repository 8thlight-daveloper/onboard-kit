package tui

import "strings"

type hint struct {
	key   string
	label string
}

func (app App) hints(tabKeysShown bool) []hint {
	if app.helpOpen {
		return []hint{{"any key", "closes help"}}
	}
	switch app.screen {
	case screenOverview:
		return app.overviewHints(tabKeysShown)
	case screenReview:
		return app.reviewHints(tabKeysShown)
	case screenSections:
		return app.sectionHints(tabKeysShown)
	}
	return app.walkHints(tabKeysShown)
}

func (app App) walkHints(tabKeysShown bool) []hint {
	if app.entry.typing() {
		hints := []hint{{"enter", "save"}, {"shift+tab", "back"}}
		if len(app.entry.question.Suggestions) > 0 {
			hints = append(hints, hint{"tab", "suggest"})
		}
		return append(hints, hint{"ctrl+d", "defer"}, hint{"esc", app.leaveName()}, hint{"f1", "help"})
	}
	hints := app.choosingHints()
	hints = append(hints, hint{"p", "back"}, hint{"d", "defer"})
	hints = append(hints, app.jumpHints(tabKeysShown)...)
	return append(hints, hint{"x", "clear"}, hint{"n", "skip"}, hint{"q", "quit"}, hint{"?", "help"})
}

func (app App) choosingHints() []hint {
	switch {
	case len(app.entry.choices) == 0:
		return nil
	case app.entry.pickingMany():
		return []hint{{"space", "pick"}, {"enter", "confirm"}}
	}
	return []hint{{"enter", "choose"}}
}

func (app App) leaveName() string {
	if app.returnTo == screenReview {
		return "review"
	}
	return "overview"
}

func (app App) sectionHints(tabKeysShown bool) []hint {
	hints := []hint{{"space", "switch"}, {"up/down", "move"}, {"esc", "back"}}
	hints = append(hints, app.jumpHints(tabKeysShown)...)
	return append(hints, hint{"?", "help"})
}

func (app App) overviewHints(tabKeysShown bool) []hint {
	hints := []hint{{"enter", "open"}, {"esc", "walk"}}
	hints = append(hints, app.jumpHints(tabKeysShown)...)
	return append(hints, hint{"q", "quit"}, hint{"?", "help"})
}

func (app App) reviewHints(tabKeysShown bool) []hint {
	hints := []hint{{"enter", "open"}, {"tab", "items/document"}, {"pgup/pgdn", "scroll"}}
	hints = append(hints, app.jumpHints(tabKeysShown)...)
	return append(hints, hint{"q", "quit"}, hint{"?", "help"})
}

func hintText(hints []hint) string {
	parts := make([]string, 0, len(hints))
	for _, entry := range hints {
		parts = append(parts, entry.key+" "+entry.label)
	}
	return strings.Join(parts, "  ")
}

func fitHints(hints []hint, width int) []hint {
	if len(hints) == 0 {
		return hints
	}
	last := hints[len(hints)-1]
	body := hints[:len(hints)-1]
	for len(body) > 0 && cellWidth(hintText(append(append([]hint{}, body...), last))) > width {
		body = body[:len(body)-1]
	}
	return append(append([]hint{}, body...), last)
}

func (app App) hintRow(width int, colors palette) string {
	layout := app.headerLayout(width, colors)
	fitted := fitHints(app.hints(layout.keys), width)
	if cellWidth(hintText(fitted)) > width {
		return renderCell(faintLine(hintText(fitted)), width, true, colors)
	}
	parts := make([]string, 0, len(fitted))
	for _, entry := range fitted {
		parts = append(parts, colors.render(toneStrong, true, entry.key)+colors.render(toneFaint, true, " "+entry.label))
	}
	return strings.Join(parts, "  ")
}
