package tui

import "strings"

type hint struct {
	key   string
	label string
}

func (app App) hints(tabKeysShown bool) []hint {
	if app.aboutOpen {
		return []hint{{"up/down", "scroll"}, {"esc", "back"}}
	}
	if app.helpOpen && app.about != nil {
		return []hint{{"a", "about"}, {"any key", "closes help"}}
	}
	if app.helpOpen {
		return []hint{{"any key", "closes help"}}
	}
	if app.stripFocused {
		return []hint{{"left/right", "choose a tab"}, {"enter", "open"}, {"down", "return"}, {"?", "help"}}
	}
	switch app.screen {
	case screenOverview:
		return app.overviewHints(tabKeysShown)
	case screenReview:
		return app.reviewHints(tabKeysShown)
	case screenSections:
		return app.sectionHints(tabKeysShown)
	case screenContacts:
		return app.contactsHints(tabKeysShown)
	case screenPreparation:
		return app.preparationHints(tabKeysShown)
	case screenInvitation:
		return app.invitationHints(tabKeysShown)
	}
	return app.answerHints(tabKeysShown)
}

func (app App) answerHints(tabKeysShown bool) []hint {
	switch {
	case app.entry.typing():
		return []hint{{"enter", "save"}, {"esc", "stop typing"}, {"down", "options"}, {"tab", "next page"}, {"ctrl+p", "back"}, {"f1", "help"}}
	case app.entry.onOptions():
		hints := []hint{{"Not typing.", ""}, {"up/down", "move"}, {"enter", "do it"}, {"tab", "next page"}, {"esc", app.leaveName()}}
		hints = append(hints, app.jumpHints(tabKeysShown)...)
		return append(hints, hint{"?", "help"})
	}
	hints := app.choosingHints()
	hints = append(hints, hint{"p", "back"}, hint{"d", "defer"})
	hints = append(hints, app.jumpHints(tabKeysShown)...)
	return append(hints, hint{"x", "clear"}, hint{"n", "skip"}, hint{"q", "quit"}, hint{"?", "help"})
}

func (app App) choosingHints() []hint {
	switch {
	case len(app.entry.choices) == 0:
		return nil
	case app.entry.pickingMany():
		return []hint{{"space", "pick"}, {"enter", "confirm"}}
	}
	return []hint{{"enter", "choose"}}
}

func (app App) leaveName() string {
	if app.returnTo == screenReview {
		return "review"
	}
	return "overview"
}

func (app App) sectionHints(tabKeysShown bool) []hint {
	hints := []hint{{"space", "switch"}, {"up/down", "move"}, {"esc", "back"}}
	hints = append(hints, app.jumpHints(tabKeysShown)...)
	return append(hints, hint{"?", "help"})
}

func (app App) contactsHints(tabKeysShown bool) []hint {
	switch app.contacts.mode {
	case contactsEditing:
		return app.formHints()
	case contactsRemoving:
		return []hint{{"y", "remove"}, {"esc", "keep"}, {"?", "help"}}
	}
	hints := []hint{{"a", "add"}, {"enter", "edit"}, {"x", "remove"}, {"left/right", "list"}}
	hints = append(hints, app.jumpHints(tabKeysShown)...)
	return append(hints, hint{"esc", "back"}, hint{"?", "help"})
}

func (app App) formHints() []hint {
	form := app.contacts.form
	switch {
	case form.choosing && form.chooser.single:
		return []hint{{"enter", "choose"}, {"esc", "back"}, {"f1", "help"}}
	case form.choosing:
		return []hint{{"space", "pick"}, {"enter", "confirm"}, {"esc", "back"}, {"f1", "help"}}
	case form.onRows():
		return []hint{{"enter", "change"}, {"ctrl+s", "save"}, {"esc", "cancel"}, {"f1", "help"}}
	case form.onHolders():
		return []hint{{"space", "pick"}, {"enter", "save"}, {"esc", "cancel"}, {"f1", "help"}}
	}
	return []hint{{"enter", "next"}, {"ctrl+s", "save"}, {"esc", "cancel"}, {"f1", "help"}}
}

func (app App) overviewHints(tabKeysShown bool) []hint {
	hints := []hint{{"enter", "open"}, {"esc", strings.ToLower(app.answerTabName())}}
	hints = append(hints, app.jumpHints(tabKeysShown)...)
	return append(hints, hint{"q", "quit"}, hint{"?", "help"})
}

func (app App) reviewHints(tabKeysShown bool) []hint {
	if app.review.pruning.asking {
		return []hint{{"y", "remove"}, {"esc", "keep"}, {"?", "help"}}
	}
	hints := []hint{{"enter", "open"}, {"left/right", "items/document"}, {"pgup/pgdn", "scroll"}}
	if app.review.leftOver > 0 {
		hints = append(hints, hint{"h", "left over"})
	}
	hints = append(hints, app.jumpHints(tabKeysShown)...)
	return append(hints, hint{"q", "quit"}, hint{"?", "help"})
}

func hintText(hints []hint) string {
	parts := make([]string, 0, len(hints))
	for _, entry := range hints {
		parts = append(parts, strings.TrimSpace(entry.key+" "+entry.label))
	}
	return strings.Join(parts, "  ")
}

func fitHints(hints []hint, width int) []hint {
	if len(hints) == 0 {
		return hints
	}
	last := hints[len(hints)-1]
	body := hints[:len(hints)-1]
	for len(body) > 0 && cellWidth(hintText(append(append([]hint{}, body...), last))) > width {
		body = body[:len(body)-1]
	}
	return append(append([]hint{}, body...), last)
}

func (app App) hintRow(width int, colors palette) string {
	layout := app.headerLayout(width, colors)
	return renderHints(app.hints(layout.keys), width, colors)
}

func renderHints(hints []hint, width int, colors palette) string {
	fitted := fitHints(hints, width)
	if cellWidth(hintText(fitted)) > width {
		return renderCell(faintLine(hintText(fitted)), width, true, colors)
	}
	parts := make([]string, 0, len(fitted))
	for _, entry := range fitted {
		if entry.label == "" {
			parts = append(parts, colors.render(toneStrong, true, entry.key))
			continue
		}
		parts = append(parts, colors.render(toneStrong, true, entry.key)+colors.render(toneFaint, true, " "+entry.label))
	}
	return strings.Join(parts, "  ")
}
