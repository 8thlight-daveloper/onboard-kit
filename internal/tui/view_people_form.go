package tui

import (
	"strings"
)

func (app App) formBody(screenFrame frame, colors palette) []string {
	width, _ := screenFrame.columns(answerShareNumerator, answerShareDenominator)
	content, focus := app.formLines(width, colors)
	return stackPane(pane{lines: scrollWindow(content, focus, screenFrame.bodyHeight), focused: true}, width, screenFrame.bodyHeight, colors)
}

func (form formState) title() string {
	if form.editing == "" {
		return "New " + form.kind.singular()
	}
	return "Edit " + form.kind.singular()
}

func (form formState) idText() string {
	if form.editing != "" {
		return "Id: " + form.editing + ", which never changes"
	}
	name := form.name()
	if slugOf(name) == "" {
		return "Id: made from the name"
	}
	id := form.id()
	if id != slugOf(name) {
		return "Id: " + id + ", made from the name, with a number because " + slugOf(name) + " is taken"
	}
	return "Id: " + id + ", made from the name"
}

func (app App) formLines(width int, colors palette) ([]line, int) {
	form := app.people.form
	content := []line{strongLine(form.title())}
	content = append(content, lines(toneFaint, wrapText(form.idText(), width))...)
	content = append(content, blankLine())
	focus := 0
	for index := range form.fields {
		if index == form.focus {
			focus = len(content)
		}
		content = append(content, form.fieldLine(index, colors))
	}
	switch {
	case form.kind == kindRole:
		content, focus = app.holderLines(content, focus, width, colors)
	case form.kind == kindTeam && form.choosing:
		content, focus = app.pickerLines(content, width, colors)
	case form.kind == kindTeam:
		content, focus = app.overrideLines(content, focus, width, colors)
	}
	return append(content, app.formProblemLines(width)...), focus
}

func (form formState) fieldLine(index int, colors palette) line {
	mark := spaces(formMarkWidth)
	if index == form.focus {
		mark = colors.render(toneHere, true, "> ")
	}
	label := form.labels[index] + spaces(formLabelWidth-cellWidth(form.labels[index]))
	return line{text: mark + colors.render(toneStrong, true, label) + form.fields[index].View(), raw: true}
}

func (app App) holderLines(content []line, focus int, width int, colors palette) ([]line, int) {
	form := app.people.form
	content = append(content, blankLine(), strongLine("Holders, in order"))
	if len(form.holders.contacts) == 0 {
		return append(content, lines(toneFaint, wrapText("No contacts yet. Add some on the Contacts list first, then pick them here.", width))...), focus
	}
	for row := 0; row < form.holders.rows(); row++ {
		selected := form.onHolders() && row == form.holders.cursor
		if selected {
			focus = len(content)
		}
		content = append(content, hereLine(form.holders.rowText(row), selected, true, width, colors))
	}
	content = append(content, blankLine())
	content = append(content, lines(toneFaint, wrapText("Space picks or unpicks. Enter saves. Pick order is kept.", width))...)
	if order := form.holders.orderText(); order != "" {
		content = append(content, lines(toneFaint, hanging("", order, width))...)
	}
	return content, focus
}

func (app App) overrideLines(content []line, focus int, width int, colors palette) ([]line, int) {
	form := app.people.form
	content = append(content, blankLine(), strongLine("Who holds each role for this team"))
	if len(form.rows) == 0 {
		return append(content, lines(toneFaint, wrapText("No roles yet. Add roles first, then a team can change who holds one.", width))...), focus
	}
	nameWidth := 0
	for _, row := range form.rows {
		nameWidth = max(nameWidth, cellWidth(singleLine(row.roleName)))
	}
	nameWidth = min(nameWidth, (width-rowMarkWidth)*nameColumnShare/nameColumnOutOf)
	for index, row := range form.rows {
		selected := form.onRows() && index == form.rowIndex()
		if selected {
			focus = len(content)
		}
		name := truncateCells(singleLine(row.roleName), nameWidth)
		text := name + spaces(nameWidth-cellWidth(name)) + summarySeparator + app.overrideText(row)
		content = append(content, hereLine(text, selected, true, width, colors))
	}
	content = append(content, blankLine())
	content = append(content, lines(toneFaint, wrapText("Enter changes who holds the chosen role for this team. Ctrl+s saves the team.", width))...)
	return content, focus
}

func (app App) overrideText(row overrideRow) string {
	switch row.special {
	case specialVacant:
		return "vacant for this team"
	case specialNone:
		return strings.Join(holderNames(app.plan.Directory, row.holders), ", ")
	}
	for _, role := range app.plan.Directory.Roles {
		if role.ID == row.roleID && len(role.Holders) > 0 {
			return "same as the role: " + strings.Join(holderNames(app.plan.Directory, role.Holders), ", ")
		}
	}
	return "same as the role: nobody"
}

func (app App) pickerLines(content []line, width int, colors palette) ([]line, int) {
	form := app.people.form
	team := form.name()
	if team == "" {
		team = "this team"
	}
	content = append(content, blankLine(), strongLine(form.rows[form.rowIndex()].roleName+", for "+team))
	focus := 0
	for row := 0; row < form.chooser.rows(); row++ {
		if row == form.chooser.cursor {
			focus = len(content)
		}
		content = append(content, hereLine(form.chooser.rowText(row), row == form.chooser.cursor, true, width, colors))
	}
	content = append(content, blankLine())
	content = append(content, lines(toneFaint, wrapText("Space picks. Enter confirms. Esc goes back without changing it. Pick order is kept.", width))...)
	if order := form.chooser.orderText(); order != "" {
		content = append(content, lines(toneFaint, hanging("", order, width))...)
	}
	return content, focus
}

func (app App) formProblemLines(width int) []line {
	var content []line
	for _, problem := range app.people.form.problems {
		content = append(content, lines(toneProblem, hanging("! problem: ", problem.Message, width))...)
	}
	if len(content) == 0 {
		return nil
	}
	return append([]line{blankLine()}, content...)
}

func (app App) removalLines(width int) []line {
	state := app.people.removal
	content := []line{strongLine("Remove the " + state.kind.singular() + " " + state.name + "?"), blankLine()}
	content = append(content, lines(toneBody, wrapText(state.consequence(), width))...)
	var answers []string
	for _, id := range state.cleared.Answers {
		answers = append(answers, app.questionText(id))
	}
	var places []string
	for _, place := range state.cleared.Holders {
		places = append(places, app.holderPlaceOf(place).text())
	}
	content = append(content, bulletLines("Answers cleared", answers, width)...)
	content = append(content, bulletLines("Holders removed", places, width)...)
	content = append(content, bulletLines("Left with nobody", app.vacantPlaces(state), width)...)
	return append(content, blankLine(), faintLine("Nothing is removed until you press y."))
}

func bulletLines(title string, texts []string, width int) []line {
	if len(texts) == 0 {
		return nil
	}
	content := []line{blankLine(), strongLine(title)}
	for _, text := range listedWithLimit(texts) {
		content = append(content, lines(toneBody, hanging("  - ", text, width))...)
	}
	return content
}
