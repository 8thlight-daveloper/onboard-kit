package tui

import "fmt"

func (app App) helpLines(width int) []line {
	content := app.keyHelpLines(width)
	if app.about == nil {
		return content
	}
	return append(content, blankLine(), faintLine("Press a now to see the version and every place this tool reads or writes."))
}

func (app App) keyHelpLines(width int) []line {
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
	if app.screen == screenAnswer {
		content = append(content, blankLine())
		content = append(content, lines(toneFaint, wrapText(pageHelpNote, width))...)
	}
	content = append(content, blankLine(), faintLine(fmt.Sprintf("Marks: %s answered, %s deferred, %s not answered, %s you are here.", app.rail.answered, app.rail.deferred, app.rail.open, app.rail.here)))
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
		return app.withPreparationHelp([]hint{{"up, down, k, j", "move"}, {"space, enter", "switch the section on or off"}, pageHelp, {"esc, q", "back"}, {"w, o, r", "answer, overview, review"}, {"c", "contacts, roles, and teams"}})
	case screenContacts:
		return app.contactsHelp()
	case screenPreparation:
		return app.preparationHelp()
	case screenInvitation:
		return app.invitationHelp()
	}
	return app.answerHelp()
}

var pageHelp = hint{"tab, shift+tab", "next or previous screen"}

const pageHelpNote = "A page holds the plan's own questions, or one section's. * marks a required question. Text you type stays on its question until you leave the page, and only enter saves it."

func (app App) answerHelp() []hint {
	switch {
	case app.stripFocused:
		return app.withContactsHelp([]hint{
			{"left, right", "choose a tab"},
			{"enter", "open the chosen tab"},
			{"down, esc", "return to where you were"},
			pageHelp,
			{"o, r", "overview, review"},
		})
	case app.entry.typing():
		return []hint{
			{"enter", "save the text and move to the next row, or to the next page after the last row"},
			{"left, right", "move in the text"},
			{"esc, down", "stop typing, and choose from the list under the box. Your text is kept"},
			{"up", "the row above, or the tab strip from the first row"},
			pageHelp,
			{"ctrl+p", "back one row"},
			{"ctrl+n", "skip to the next row without changing the answer"},
			{"ctrl+d", "defer this row and move on"},
			{"empty box, enter", "clear the answer"},
			{"f1", "this help"},
		}
	case app.entry.onOptions():
		return app.withScreenHelp([]hint{
			{"up, down, k, j", "choose from the list. Up from Edit the answer goes back to typing, and down from the last option goes to the next row"},
			{"enter", "do what the chosen option says"},
			pageHelp,
			{"p, left", "back one row"},
			{"n, right", "skip to the next row without changing the answer"},
			{"d", "defer this row and move on"},
			{"x", "clear the answer and the box"},
			{"esc", "back to the overview or review"},
			{"q", "quit"},
		})
	}
	moving := []hint{{"up, down, k, j", "move the choice. Past the first or last choice, go to the row above or below. Up from the first row of the page goes to the tab strip"}}
	if app.entry.pickingMany() {
		moving = append(moving, hint{"space", "pick or unpick the choice, the order you pick is the order kept"})
	}
	return app.withScreenHelp(append(moving,
		app.enterHelp(),
		pageHelp,
		hint{"p, left", "back one row"},
		hint{"n, right", "skip to the next row without changing the answer"},
		hint{"d", "defer this row and move on"},
		hint{"x", "clear the answer"},
		hint{"esc", "back to the overview or review"},
		hint{"q", "quit"},
	))
}

func (app App) withScreenHelp(entries []hint) []hint {
	entries = append(entries, hint{"o", "overview of every question"}, hint{"r", "review: blockers, deferred, problems, document"})
	if app.canSwitchSections() {
		entries = append(entries, hint{"s", "switch sections on or off"})
	}
	return app.withContactsHelp(entries)
}

func (app App) enterHelp() hint {
	switch {
	case len(app.entry.choices) == 0:
		return hint{"enter", "move to the next row"}
	case app.entry.pickingMany():
		return hint{"enter", "accept the picks and move to the next row"}
	}
	return hint{"enter", "accept the answer and move to the next row"}
}

func (app App) overviewHelp() []hint {
	entries := []hint{
		{"up, down, k, j", "move"},
		pageHelp,
		{"pgup, pgdn", "move a page"},
		{"g, G", "first and last question"},
		{"enter", "open the question"},
		{"esc, w", "back to the question you left"},
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
		{"left, right", "focus the items or the document"},
		pageHelp,
		{"pgup, pgdn", "scroll the document"},
		{"enter", "open the question of the chosen item"},
		{"esc, w", "the first unanswered question"},
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
		{"left, right", "contacts, roles, or teams"},
		pageHelp,
		{"a", "add an entry to this list"},
		{"enter, e", "change the chosen entry"},
		{"x", "remove the chosen entry, after showing what it clears"},
		{"w, o, r, s", "answer, overview, review, sections"},
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
