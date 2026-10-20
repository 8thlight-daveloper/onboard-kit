package tui

func (a App) overviewBody(l layout, p palette) []string {
	if len(a.questions) == 0 {
		return singlePane([]line{plainLine("No questions are active.")}, l, p)
	}
	if !l.twoPanes {
		outline, focus := a.outlineLines(l.width, a.overviewCursor, true)
		return singlePane(scrollWindow(outline, focus, l.bodyHeight), l, p)
	}
	outline, focus := a.outlineLines(l.listWidth, a.overviewCursor, false)
	question := a.questions[a.overviewCursor]
	right := a.questionSummary(a.overviewCursor, l.mainWidth)
	right = append(right, a.detailLines(question, l.mainWidth)...)
	return sideBySide(scrollWindow(outline, focus, l.bodyHeight), clipLines(right, l.bodyHeight), l, p)
}

func (a App) questionSummary(index int, width int) []line {
	question := a.questions[index]
	saved := a.cursor
	a.cursor = index
	content := a.questionLines(question, width)
	a.cursor = saved
	return append(content, plainLine(""))
}
