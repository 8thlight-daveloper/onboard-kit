package tui

func (app App) helpLines(width int) []line {
	var content []line
	content = append(content, strongLine("Keys"))
	for _, entry := range app.helpEntries() {
		content = append(content, lines(toneBody, hanging("", entry.key+"  "+entry.label, width))...)
	}
	content = append(content, blankLine(), faintLine("Marks: x answered, ~ deferred, . not answered, ◆ you are here."))
	content = append(content, faintLine("Every accepted answer is saved at once. ctrl+c quits at any time."))
	return content
}

func (app App) helpEntries() []hint {
	switch app.screen {
	case screenOverview:
		return app.overviewHelp()
	case screenReview:
		return app.reviewHelp()
	case screenSections:
		return []hint{{"up, down, k, j", "move"}, {"space, enter", "switch the section on or off"}, {"esc, q", "back"}, {"w, o, r", "walk, overview, review"}}
	}
	return app.walkHelp()
}

func (app App) walkHelp() []hint {
	entries := []hint{
		app.enterHelp(),
		{"p, left, shift+tab", "back one question"},
		{"n, right", "skip forward without changing the answer"},
		{"d", "defer the question and move on"},
		{"x", "clear the answer"},
		{"o", "overview of every question"},
		{"r", "review: blockers, deferred, problems, document"},
		{"esc", "back to the overview or review"},
		{"q", "quit"},
	}
	if app.entry.typing() {
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
	moving := []hint{{"up, down, k, j", "move the choice"}}
	if app.entry.pickingMany() {
		moving = append(moving, hint{"space", "pick or unpick the choice, the order you pick is the order kept"})
	}
	entries = append(moving, entries...)
	if app.canSwitchSections() {
		entries = append(entries, hint{"s", "switch sections on or off"})
	}
	return entries
}

func (app App) enterHelp() hint {
	switch {
	case len(app.entry.choices) == 0:
		return hint{"enter", "move to the next question"}
	case app.entry.pickingMany():
		return hint{"enter", "accept the picks and move to the next question"}
	}
	return hint{"enter", "accept the answer and move to the next question"}
}

func (app App) overviewHelp() []hint {
	entries := []hint{
		{"up, down, k, j", "move"},
		{"pgup, pgdn", "move a page"},
		{"g, G", "first and last question"},
		{"enter", "open the question"},
		{"esc, w", "back to the walk"},
		{"r", "review"},
		{"q", "quit"},
	}
	if app.canSwitchSections() {
		entries = append(entries, hint{"s", "switch sections on or off"})
	}
	return entries
}

func (app App) reviewHelp() []hint {
	entries := []hint{
		{"up, down, k, j", "move between items, or scroll when the document has focus"},
		{"tab", "switch focus between the items and the document"},
		{"pgup, pgdn", "scroll the document"},
		{"enter", "open the question of the chosen item"},
		{"esc, w", "walk, from the first unanswered question"},
		{"o", "overview"},
		{"q", "quit"},
	}
	if app.canSwitchSections() {
		entries = append(entries, hint{"s", "switch sections on or off"})
	}
	return entries
}
