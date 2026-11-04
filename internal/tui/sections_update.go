package tui

import tea "github.com/charmbracelet/bubbletea"

func (app App) updateSections(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc", "q":
		return app.closeSideScreen(), nil
	case "w":
		return app.returnToWalk(), nil
	case "o":
		return app.openOverview(), nil
	case "r":
		return app.openReview(), nil
	case "c":
		return app.openContacts(), nil
	case "t":
		return app.openPreparation(), nil
	case "i":
		return app.openInvitation(), nil
	case "up", "k":
		app.sectionCursor = max(app.sectionCursor-1, 0)
	case "down", "j":
		app.sectionCursor = min(app.sectionCursor+1, max(len(app.plan.Sections)-1, 0))
	case " ", "enter":
		return app.toggleSection(), nil
	case "?":
		app.helpOpen = true
	}
	return app, nil
}

func (app App) toggleSection() App {
	if app.sectionCursor >= len(app.plan.Sections) {
		return app
	}
	section := app.plan.Sections[app.sectionCursor]
	next, problems := app.backend.SetSectionEnabled(app.plan, section.ID, !isOn(section))
	if hasError(problems) {
		app.status = line{text: firstMessage(problems), tone: toneProblem}
		return app
	}
	keep := ""
	if question, ok := app.current(); ok {
		keep = question.ID
	}
	app.plan = next
	app = app.refreshed(keep)
	app.status = app.saveStatus(problems)
	return app
}
