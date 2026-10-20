package tui

import tea "github.com/charmbracelet/bubbletea"

func (a App) updateSections(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc", "q":
		return a.closeSections(), nil
	case "up", "k":
		a.sectionCursor = max(a.sectionCursor-1, 0)
	case "down", "j":
		a.sectionCursor = min(a.sectionCursor+1, max(len(a.plan.Sections)-1, 0))
	case " ", "enter":
		return a.toggleSection(), nil
	case "?":
		a.helpOpen = true
	}
	return a, nil
}

func (a App) toggleSection() App {
	if a.sectionCursor >= len(a.plan.Sections) {
		return a
	}
	section := a.plan.Sections[a.sectionCursor]
	next, problems := a.backend.SetSectionEnabled(a.plan, section.ID, !isOn(section))
	if hasError(problems) {
		a.status = line{text: firstMessage(problems), tone: toneProblem}
		return a
	}
	keep := ""
	if question, ok := a.current(); ok {
		keep = question.ID
	}
	a.plan = next
	a.refresh(keep)
	a.status = a.saveStatus(problems)
	return a
}
