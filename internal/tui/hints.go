package tui

import "strings"

type hint struct {
	key   string
	label string
}

const helpHint = "? help"

func (a App) hints() []hint {
	if a.helpOpen {
		return []hint{{"any key", "closes help"}}
	}
	switch a.screen {
	case screenOverview:
		return a.overviewHints()
	case screenReview:
		return a.reviewHints()
	case screenSections:
		return []hint{{"space", "switch"}, {"up/down", "move"}, {"esc", "back"}, {"?", "help"}}
	}
	return a.walkHints()
}

func (a App) walkHints() []hint {
	if a.entry.typing() {
		hints := []hint{{"enter", "save"}, {"shift+tab", "back"}}
		if len(a.entry.question.Suggestions) > 0 {
			hints = append(hints, hint{"tab", "suggest"})
		}
		return append(hints, hint{"ctrl+d", "defer"}, hint{"esc", a.leaveName()}, hint{"f1", "help"})
	}
	hints := []hint{{"enter", "choose"}, {"p", "back"}, {"d", "defer"}, {"o", "overview"}, {"r", "review"}, {"x", "clear"}, {"n", "skip"}}
	if a.canSwitchSections() {
		hints = append(hints, hint{"s", "sections"})
	}
	return append(hints, hint{"q", "quit"}, hint{"?", "help"})
}

func (a App) leaveName() string {
	if a.returnTo == screenReview {
		return "review"
	}
	return "overview"
}

func (a App) overviewHints() []hint {
	hints := []hint{{"enter", "open"}, {"esc", "walk"}, {"r", "review"}}
	if a.canSwitchSections() {
		hints = append(hints, hint{"s", "sections"})
	}
	return append(hints, hint{"q", "quit"}, hint{"?", "help"})
}

func (a App) reviewHints() []hint {
	hints := []hint{{"enter", "open"}, {"tab", "items/document"}, {"pgup/pgdn", "scroll"}, {"w", "walk"}, {"o", "overview"}}
	if a.canSwitchSections() {
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

func (a App) hintRow(l layout, p palette) string {
	fitted := fitHints(a.hints(), l.width)
	if cellWidth(hintText(fitted)) > l.width {
		return renderRow(mutedLine(hintText(fitted)), l.width, p)
	}
	parts := make([]string, 0, len(fitted))
	for _, entry := range fitted {
		parts = append(parts, p.key.Render(entry.key)+p.style(toneMuted).Render(" "+entry.label))
	}
	return strings.Join(parts, "  ")
}
