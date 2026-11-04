package tui

func (app App) preparationHints(tabKeysShown bool) []hint {
	state := app.preparation
	var hints []hint
	switch state.mode {
	case preparationAdding:
		return []hint{{"enter", "open"}, {"esc", "cancel"}, {"f1", "help"}}
	case preparationEditing:
		return app.preparationFormHints()
	case preparationRemoving:
		return []hint{{"y", "remove"}, {"esc", "keep"}, {"?", "help"}}
	case preparationDropping:
		return []hint{{"y", "drop"}, {"esc", "keep"}, {"?", "help"}}
	case preparationReviewing:
		hints = []hint{{"y", "save"}, {"tab", "changes/guide"}, {"pgup/pgdn", "scroll"}, {"esc", "back"}}
	case preparationPerson:
		hints = []hint{{"enter", "open"}, {"a", "add arrangement"}}
		if _, ok := state.person.arrangementRow(state.row); ok {
			hints = append(hints, hint{"x", "remove"})
		}
		hints = append(hints, hint{"esc", "people"})
	default:
		hints = []hint{{"enter", "open"}, {"a", "add person"}, {"esc", "back"}}
	}
	hints = append(hints, app.jumpHints(tabKeysShown)...)
	return append(hints, hint{"?", "help"})
}

func (app App) preparationFormHints() []hint {
	form := app.preparation.form
	switch {
	case form.choosing && form.chooser.single:
		return []hint{{"enter", "choose"}, {"esc", "back"}, {"f1", "help"}}
	case form.choosing:
		return []hint{{"space", "pick"}, {"enter", "confirm"}, {"esc", "back"}, {"f1", "help"}}
	case form.onText() && form.onLast():
		return []hint{{"enter", "stage"}, {"ctrl+s", "stage"}, {"esc", "cancel"}, {"f1", "help"}}
	case form.onText():
		return []hint{{"enter", "next"}, {"ctrl+s", "stage"}, {"esc", "cancel"}, {"f1", "help"}}
	}
	return []hint{{"enter", "pick"}, {"ctrl+s", "stage"}, {"esc", "cancel"}, {"f1", "help"}}
}

func (app App) preparationHelp() []hint {
	state := app.preparation
	switch state.mode {
	case preparationAdding:
		return []hint{{"letters", "type the account, which names the person's folder"}, {"enter", "open the person, with nothing written yet"}, {"esc", "cancel"}}
	case preparationEditing:
		return app.preparationFormHelp()
	case preparationRemoving:
		return []hint{{"y", "stage the removal, with nothing written yet"}, {"any other key", "keep the arrangement and go back"}}
	case preparationDropping:
		return []hint{{"y", "drop the staged changes, leaving the file as it is"}, {"any other key", "keep them and go back"}}
	case preparationReviewing:
		return app.withJumpHelp([]hint{
			{"y", "save the changes listed here to the file"},
			{"tab", "switch focus between the changes and the guide"},
			{"up, down, pgup, pgdn", "scroll the pane with focus"},
			{"esc, q", "back to the person, with the changes still staged"},
		})
	case preparationPerson:
		return app.withJumpHelp([]hint{
			{"up, down, k, j", "move"},
			{"enter, e", "change the recipient or the chosen arrangement, or review and save"},
			{"a", "add an arrangement"},
			{"x", "remove the chosen arrangement, after you confirm"},
			{"esc, q", "back to the people, after you confirm when changes are staged"},
		})
	}
	return app.withJumpHelp([]hint{
		{"up, down, k, j", "move"},
		{"enter, e", "open the chosen person"},
		{"a", "add a person"},
		{"esc, q", "back"},
	})
}

func (app App) withJumpHelp(entries []hint) []hint {
	return app.withInvitationHelp(append(entries, hint{"w, o, r, s, c", "walk, overview, review, sections, contacts"}))
}

func (app App) preparationFormHelp() []hint {
	form := app.preparation.form
	switch {
	case form.choosing && form.chooser.single:
		return []hint{{"up, down, k, j", "move"}, {"enter, space", "choose the row"}, {"esc", "go back without changing it"}}
	case form.choosing:
		return []hint{{"up, down, k, j", "move"}, {"space", "pick or unpick, the order you pick is the order kept"}, {"enter", "confirm"}, {"esc", "go back without changing it"}}
	}
	return []hint{
		{"letters", "type into a text field"},
		{"down, tab", "next field"},
		{"enter", "next field, or open the pick list of the chosen row"},
		{"shift+tab, up", "previous field"},
		{"enter on the last field, ctrl+s", "stage the form, with nothing written yet"},
		{"esc", "cancel"},
	}
}
