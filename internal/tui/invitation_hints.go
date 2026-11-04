package tui

func (app App) invitationHints(tabKeysShown bool) []hint {
	var hints []hint
	switch app.invitation.mode {
	case invitationEditing:
		return app.invitationFormHints()
	case invitationConfirming:
		return []hint{{"y", "write"}, {"pgup/pgdn", "guide"}, {"esc", "back"}, {"?", "help"}}
	case invitationBlocked:
		hints = []hint{{"esc", "people"}}
	case invitationWritten:
		hints = []hint{{"enter", "people"}}
	default:
		hints = []hint{{"enter", "open"}, {"esc", "back"}}
	}
	hints = append(hints, app.jumpHints(tabKeysShown)...)
	return append(hints, hint{"?", "help"})
}

func (app App) invitationFormHints() []hint {
	form := app.invitation.form
	switch {
	case form.choosing && form.chooser.single:
		return []hint{{"enter", "choose"}, {"esc", "back"}, {"f1", "help"}}
	case form.choosing:
		return []hint{{"space", "pick"}, {"enter", "confirm"}, {"esc", "back"}, {"f1", "help"}}
	case form.onOutput():
		return []hint{{"enter", "next"}, {"pgup/pgdn", "guide"}, {"esc", "cancel"}, {"f1", "help"}}
	}
	return []hint{{"enter", "pick"}, {"ctrl+s", "next"}, {"pgup/pgdn", "guide"}, {"esc", "cancel"}, {"f1", "help"}}
}

func (app App) invitationHelp() []hint {
	form := app.invitation.form
	switch app.invitation.mode {
	case invitationEditing:
		switch {
		case form.choosing && form.chooser.single:
			return []hint{{"up, down, k, j", "move"}, {"enter, space", "choose the row"}, {"esc", "go back without changing it"}}
		case form.choosing:
			return []hint{{"up, down, k, j", "move"}, {"space", "pick or unpick, the order you pick is the order kept"}, {"enter", "confirm"}, {"esc", "go back without changing it"}}
		}
		return []hint{
			{"up, down, tab, shift+tab", "move between the fields"},
			{"enter", "open the pick list of the chosen row, or go on from the file"},
			{"letters", "type the file name"},
			{"ctrl+s", "go on to the last step, with nothing written yet"},
			{"pgup, pgdn", "scroll the first steps guide"},
			{"esc", "back to the people, with nothing written"},
		}
	case invitationConfirming:
		return []hint{
			{"y", "write the invitation file"},
			{"up, down, pgup, pgdn", "scroll the first steps guide"},
			{"esc, q", "back to the form, with nothing written"},
		}
	case invitationBlocked:
		return app.withInvitationJumpHelp([]hint{{"esc, q, enter", "back to the people"}})
	case invitationWritten:
		return app.withInvitationJumpHelp([]hint{{"enter, esc, q", "back to the people"}})
	}
	return app.withInvitationJumpHelp([]hint{
		{"up, down, k, j", "move"},
		{"enter, e", "open the chosen person: the form when ready, or what blocks the invitation"},
		{"esc, q", "back"},
	})
}

func (app App) withInvitationJumpHelp(entries []hint) []hint {
	entries = append(entries, hint{"w, o, r, s, c", "walk, overview, review, sections, contacts"})
	if app.canPrepare() {
		entries = append(entries, hint{"t", "preparation: what is arranged for each newcomer"})
	}
	return entries
}
