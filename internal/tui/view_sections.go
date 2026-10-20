package tui

func (a App) sectionLines(l layout) []line {
	content := []line{plainLine("Sections the plan can include. Space switches one on or off."), plainLine("")}
	focus := 0
	for index, section := range a.plan.Sections {
		marker := "  "
		tone := tonePlain
		if index == a.sectionCursor {
			marker = "> "
			tone = toneFocus
			focus = len(content)
		}
		box := "[x] "
		if !isOn(section) {
			box = "[ ] "
		}
		content = append(content, line{text: marker + box + section.Title, tone: tone})
	}
	return scrollWindow(content, focus, l.bodyHeight)
}
