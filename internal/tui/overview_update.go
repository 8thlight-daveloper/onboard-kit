package tui

import tea "github.com/charmbracelet/bubbletea"

func (a App) updateOverview(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	last := max(len(a.questions)-1, 0)
	switch key.String() {
	case "up", "k":
		a.overviewCursor = max(a.overviewCursor-1, 0)
	case "down", "j":
		a.overviewCursor = min(a.overviewCursor+1, last)
	case "pgup":
		a.overviewCursor = max(a.overviewCursor-a.pageSize(), 0)
	case "pgdown":
		a.overviewCursor = min(a.overviewCursor+a.pageSize(), last)
	case "g", "home":
		a.overviewCursor = 0
	case "G", "end":
		a.overviewCursor = last
	case "enter":
		a.returnTo = screenOverview
		return a.jumpTo(a.overviewCursor), nil
	case "esc", "w":
		return a.returnToWalk(), nil
	case "r":
		return a.openReview(), nil
	case "s":
		return a.openSections(), nil
	case "q":
		return a.quit()
	case "?":
		a.helpOpen = true
	}
	return a, nil
}

func (a App) pageSize() int { return max(layoutFor(a.width, a.height).bodyHeight-1, 1) }
