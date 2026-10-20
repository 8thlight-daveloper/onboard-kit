package tui

import "strings"

type hint struct {
	key   string
	label string
}

func (app App) hints() []hint {
	if app.helpOpen {
		return []hint{{"any key", "closes help"}}
	}
	switch app.screen {
	case screenOverview:
		return app.overviewHints()
	case screenReview:
		return app.reviewHints()
	case screenSections:
		return []hint{{"space", "switch"}, {"up/down", "move"}, {"esc", "back"}, {"?", "help"}}
	}
	return app.walkHints()
}

func (app App) walkHints() []hint {
	if app.entry.typing() {
		hints := []hint{{"enter", "save"}, {"shift+tab", "back"}}
		if len(app.entry.question.Suggestions) > 0 {
			hints = append(hints, hint{"tab", "suggest"})
		}
		return append(hints, hint{"ctrl+d", "defer"}, hint{"esc", app.leaveName()}, hint{"f1", "help"})
	}
	hints := []hint{{"enter", "choose"}, {"p", "back"}, {"d", "defer"}, {"o", "overview"}, {"r", "review"}, {"x", "clear"}, {"n", "skip"}}
	if app.canSwitchSections() {
		hints = append(hints, hint{"s", "sections"})
	}
	return append(hints, hint{"q", "quit"}, hint{"?", "help"})
}

func (app App) leaveName() string {
	if app.returnTo == screenReview {
		return "review"
	}
	return "overview"
}

func (app App) overviewHints() []hint {
	hints := []hint{{"enter", "open"}, {"esc", "walk"}, {"r", "review"}}
	if app.canSwitchSections() {
		hints = append(hints, hint{"s", "sections"})
	}
	return append(hints, hint{"q", "quit"}, hint{"?", "help"})
}

func (app App) reviewHints() []hint {
	hints := []hint{{"enter", "open"}, {"tab", "items/document"}, {"pgup/pgdn", "scroll"}, {"w", "walk"}, {"o", "overview"}}
	if app.canSwitchSections() {
		hints = append(hints, hint{"s", "sections"})
	}
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
	fitted := fitHints(app.hints(), width)
	if cellWidth(hintText(fitted)) > width {
		return renderCell(faintLine(hintText(fitted)), width, true, colors)
	}
	parts := make([]string, 0, len(fitted))
	for _, entry := range fitted {
		parts = append(parts, colors.render(toneStrong, true, entry.key)+colors.render(toneFaint, true, " "+entry.label))
	}
	return strings.Join(parts, "  ")
}
