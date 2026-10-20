package tui

func (a App) helpLines(width int) []line {
	var content []line
	content = append(content, headingLine("Keys"))
	for _, entry := range a.helpEntries() {
		content = append(content, lines(tonePlain, hanging("", entry.key+"  "+entry.label, width))...)
	}
	content = append(content, plainLine(""), mutedLine("Every accepted answer is saved at once. ctrl+c quits at any time."))
	return content
}

func (a App) helpEntries() []hint {
	switch a.screen {
	case screenOverview:
		return a.overviewHelp()
	case screenReview:
		return a.reviewHelp()
	case screenSections:
		return []hint{{"up, down, k, j", "move"}, {"space, enter", "switch the section on or off"}, {"esc, q", "back"}}
	}
	return a.walkHelp()
}

func (a App) walkHelp() []hint {
	entries := []hint{
		{"enter", "accept the answer and move to the next question"},
		{"p, left, shift+tab", "back one question"},
		{"n, right", "skip forward without changing the answer"},
		{"d", "defer the question and move on"},
		{"x", "clear the answer"},
		{"o", "overview of every question"},
		{"r", "review: blockers, deferred, problems, document"},
		{"esc", "back to the overview or review"},
		{"q", "quit"},
	}
	if a.entry.typing() {
		return []hint{
			{"enter", "accept the text and move to the next question"},
			{"tab", "try a suggestion"},
			{"shift+tab, ctrl+p", "back one question"},
			{"ctrl+n", "skip forward without changing the answer"},
			{"ctrl+d", "defer the question and move on"},
			{"empty box, enter", "clear the answer"},
			{"esc", "back to the overview or review"},
			{"f1", "this help"},
		}
	}
	entries = append([]hint{{"up, down, k, j", "move the choice"}}, entries...)
	if a.canSwitchSections() {
		entries = append(entries, hint{"s", "switch sections on or off"})
	}
	return entries
}

func (a App) overviewHelp() []hint {
	entries := []hint{
		{"up, down, k, j", "move"},
		{"pgup, pgdn", "move a page"},
		{"g, G", "first and last question"},
		{"enter", "open the question"},
		{"esc, w", "back to the walk"},
		{"r", "review"},
		{"q", "quit"},
	}
	if a.canSwitchSections() {
		entries = append(entries, hint{"s", "switch sections on or off"})
	}
	return entries
}

func (a App) reviewHelp() []hint {
	entries := []hint{
		{"up, down, k, j", "move between items, or scroll when the document has focus"},
		{"tab", "switch focus between the items and the document"},
		{"pgup, pgdn", "scroll the document"},
		{"enter", "open the question of the chosen item"},
		{"esc, w", "walk, from the first unanswered question"},
		{"o", "overview"},
		{"q", "quit"},
	}
	if a.canSwitchSections() {
		entries = append(entries, hint{"s", "switch sections on or off"})
	}
	return entries
}
