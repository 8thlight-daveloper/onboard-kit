package tui

func (app App) helpLines(width int) []line {
	var content []line
	content = append(content, strongLine("Keys"))
	for _, entry := range app.helpEntries() {
		content = append(content, lines(toneBody, hanging("", entry.key+"  "+entry.label, width))...)
	}
	if app.screen == screenPreparation {
		content = append(content, blankLine())
		return append(content, lines(toneFaint, wrapText("Nothing in a preparation is written until you review it and press y. ctrl+c quits at any time and drops staged changes.", width))...)
	}
	if app.screen == screenInvitation {
		content = append(content, blankLine())
		return append(content, lines(toneFaint, wrapText(invitationHelpNote, width))...)
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
		return app.withPreparationHelp([]hint{{"up, down, k, j", "move"}, {"space, enter", "switch the section on or off"}, {"esc, q", "back"}, {"w, o, r", "walk, overview, review"}, {"c", "contacts, roles, and teams"}})
	case screenContacts:
		return app.contactsHelp()
	case screenPreparation:
		return app.preparationHelp()
	case screenInvitation:
		return app.invitationHelp()
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
	return app.withContactsHelp(entries)
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
	return app.withContactsHelp(entries)
}

func (app App) reviewHelp() []hint {
	if app.review.pruning.asking {
		return []hint{{"y", "remove the left over answers listed above"}, {"any other key", "keep them and go back"}}
	}
	entries := []hint{
		{"up, down, k, j", "move between items, or scroll when the document has focus"},
		{"tab", "switch focus between the items and the document"},
		{"pgup, pgdn", "scroll the document"},
		{"enter", "open the question of the chosen item"},
		{"esc, w", "walk, from the first unanswered question"},
		{"o", "overview"},
		{"h", "show the answers left over from questions the plan no longer asks, and remove them after you confirm"},
		{"q", "quit"},
	}
	if app.canSwitchSections() {
		entries = append(entries, hint{"s", "switch sections on or off"})
	}
	return app.withContactsHelp(entries)
}

func (app App) withContactsHelp(entries []hint) []hint {
	if !app.canEditDirectory() {
		return entries
	}
	return app.withPreparationHelp(append(entries, hint{"c", "contacts: add, change, or remove contacts, roles, and teams"}))
}

func (app App) withPreparationHelp(entries []hint) []hint {
	if app.canPrepare() {
		entries = append(entries, hint{"t", "preparation: what is arranged for each newcomer"})
	}
	return app.withInvitationHelp(entries)
}

func (app App) withInvitationHelp(entries []hint) []hint {
	if !app.canInvite() {
		return entries
	}
	return append(entries, hint{"i", "invitation: write the file a newcomer opens"})
}

func (app App) contactsHelp() []hint {
	switch app.contacts.mode {
	case contactsEditing:
		return app.formHelp()
	case contactsRemoving:
		return []hint{{"y", "remove it, and clear what the list above says"}, {"any other key", "keep it and go back"}}
	}
	return app.withPreparationHelp([]hint{
		{"up, down, k, j", "move"},
		{"tab, shift+tab, left, right", "contacts, roles, or teams"},
		{"a", "add an entry to this list"},
		{"enter, e", "change the chosen entry"},
		{"x", "remove the chosen entry, after showing what it clears"},
		{"w, o, r, s", "walk, overview, review, sections"},
		{"esc, q", "back"},
	})
}

func (app App) formHelp() []hint {
	form := app.contacts.form
	switch {
	case form.choosing && form.chooser.single:
		return []hint{{"up, down, k, j", "move"}, {"enter, space", "choose the row"}, {"esc", "go back without changing it"}}
	case form.choosing:
		return []hint{{"up, down, k, j", "move"}, {"space", "pick or unpick, or choose the top two rows"}, {"enter", "confirm"}, {"esc", "go back without changing it"}}
	case form.onContactRow():
		return []hint{{"up, down, k, j", "choose a row"}, {"enter, space", "choose the team contact"}, {"ctrl+s", "save the team"}, {"esc", "cancel"}}
	case form.onRows():
		return []hint{{"up, down, k, j", "choose a role"}, {"enter, space", "change who holds it for this team"}, {"ctrl+s", "save the team"}, {"esc", "cancel"}}
	case form.onHolders():
		return []hint{{"up, down, k, j", "move"}, {"space", "pick or unpick, the order you pick is the order kept"}, {"enter, ctrl+s", "save"}, {"tab", "back to the name"}, {"esc", "cancel"}}
	}
	return []hint{{"enter, down, tab", "next field"}, {"shift+tab, up", "previous field"}, {"enter on the last field, ctrl+s", "save"}, {"esc", "cancel"}}
}
