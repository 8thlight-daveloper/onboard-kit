package tui

import tea "github.com/charmbracelet/bubbletea"

func (app App) updateOverview(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	last := max(len(app.questions)-1, 0)
	switch key.String() {
	case "up", "k":
		app.overviewCursor = max(app.overviewCursor-1, 0)
	case "down", "j":
		app.overviewCursor = min(app.overviewCursor+1, last)
	case "pgup":
		app.overviewCursor = max(app.overviewCursor-app.pageSize(), 0)
	case "pgdown":
		app.overviewCursor = min(app.overviewCursor+app.pageSize(), last)
	case "g", "home":
		app.overviewCursor = 0
	case "G", "end":
		app.overviewCursor = last
	case "enter":
		app.returnTo = screenOverview
		return app.jumpTo(app.overviewCursor), nil
	case "esc", "w":
		return app.returnToWalk(), nil
	case "r":
		return app.openReview(), nil
	case "s":
		return app.openSections(), nil
	case "c":
		return app.openContacts(), nil
	case "t":
		return app.openPreparation(), nil
	case "q":
		return app.quit()
	case "?":
		app.helpOpen = true
	}
	return app, nil
}

func (app App) pageSize() int { return max(app.frame().bodyHeight-1, 1) }
