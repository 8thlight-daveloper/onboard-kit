package tui

func (app App) sectionLines(screenFrame frame, colors palette) []line {
	content := []line{bodyLine("Sections the plan can include. Space switches one on or off."), blankLine()}
	focus := 0
	for index, section := range app.plan.Sections {
		box := "[x] "
		if !isOn(section) {
			box = "[ ] "
		}
		if index == app.sectionCursor {
			focus = len(content)
		}
		content = append(content, hereLine(box+section.Title, index == app.sectionCursor, true, screenFrame.innerWidth(), colors))
	}
	return scrollWindow(content, focus, screenFrame.bodyHeight)
}
