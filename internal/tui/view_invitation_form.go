package tui

import "fmt"

const (
	preaccessNotice = "Pre-access help is who the newcomer can reach before they have company access. Only contacts with a name and an email or Slack handle are offered."
	formStepNotice  = "Enter on the file, or ctrl+s, shows what will be written. Nothing is written until you press y there."
	guideGapRows    = 2
)

func (app App) invitationFormBody(screenFrame frame, colors palette) []string {
	content, focus := app.invitationLeftLines(colors)
	leftWidth, guideWidth, guideRows := app.invitationColumns(screenFrame)
	form := app.invitation.form
	scrolling := app.invitation.mode == invitationConfirming
	guide := append([]line{paneTitle(app.invitationGuideTitle(guideWidth, guideRows), scrolling, colors)}, windowFrom(app.invitationGuideLines(guideWidth), form.guideScroll, guideRows)...)
	if screenFrame.wide() {
		left := pane{lines: scrollWindow(content, focus, screenFrame.bodyHeight), focused: true}
		return joinPanes(left, pane{lines: guide, focused: true}, leftWidth, guideWidth, screenFrame.bodyHeight, colors)
	}
	combined := append(append(append([]line{}, content...), blankLine()), guide...)
	return stackPane(pane{lines: clipLines(combined, screenFrame.bodyHeight), focused: true}, leftWidth, screenFrame.bodyHeight, colors)
}

func (app App) invitationColumns(screenFrame frame) (int, int, int) {
	if screenFrame.wide() {
		left, right := screenFrame.columns(itemsShareNumerator, itemsShareDenominator)
		return left, right, screenFrame.documentRows()
	}
	width, _ := screenFrame.columns(answerShareNumerator, answerShareDenominator)
	content, _ := app.invitationLeftLines(newPalette())
	return width, width, max(screenFrame.bodyHeight-len(content)-guideGapRows, 1)
}

func (app App) invitationLeftLines(colors palette) ([]line, int) {
	width, _, _ := app.invitationPaneWidths(app.frame())
	if app.invitation.mode == invitationConfirming {
		return app.invitationConfirmLines(width), 0
	}
	return app.invitationFormLines(width, colors)
}

func (app App) invitationFormWidth() int {
	width, _, _ := app.invitationPaneWidths(app.frame())
	return width
}

func (app App) invitationPaneWidths(screenFrame frame) (int, int, bool) {
	if screenFrame.wide() {
		left, right := screenFrame.columns(itemsShareNumerator, itemsShareDenominator)
		return left, right, true
	}
	width, _ := screenFrame.columns(answerShareNumerator, answerShareDenominator)
	return width, width, false
}

func (app App) invitationFormLines(width int, colors palette) ([]line, int) {
	form := app.invitation.form
	content := []line{strongLine("Invitation for " + form.person.name()), faintLine("Account " + visibleText(form.person.Account))}
	content = append(content, lines(toneFaint, wrapText("Writes "+visibleText(form.output.Value()), width))...)
	content = append(content, blankLine())
	focus := 0
	for index := range invitationFieldCount {
		if index == form.focus {
			focus = len(content)
		}
		content = append(content, form.fieldLine(index, width, colors))
	}
	if form.choosing {
		content, focus = app.invitationPickerLines(content, width, colors)
	} else {
		content = append(content, blankLine())
		content = append(content, lines(toneFaint, wrapText(preaccessNotice, width))...)
		content = append(content, lines(toneFaint, wrapText(formStepNotice, width))...)
	}
	content = append(content, preparationProblemLines(form.problems, width)...)
	return append(content, reviewGroup("Guide problems", problemTexts(form.guideProblems), toneWarning, width)...), focus
}

func (form invitationForm) fieldLine(index int, width int, colors palette) line {
	mark := spaces(formMarkWidth)
	if index == form.focus {
		mark = colors.render(toneHere, true, "> ")
	}
	label := invitationFieldLabels[index]
	text := mark + colors.render(toneStrong, true, label+spaces(preparationLabelWidth-cellWidth(label)))
	valueWidth := width - formMarkWidth - preparationLabelWidth
	switch {
	case index == invitationOutputField && form.onOutput():
		return line{text: text + form.output.View(), raw: true}
	case index == invitationOutputField:
		return line{text: text + truncateCells(visibleText(form.output.Value()), valueWidth), raw: true}
	case index == invitationPreaccessField:
		return line{text: text + truncateCells(form.preaccessName(), valueWidth), raw: true}
	}
	return line{text: text + truncateCells(form.helpNames(), valueWidth), raw: true}
}

func (app App) invitationPickerLines(content []line, width int, colors palette) ([]line, int) {
	form := app.invitation.form
	content = append(content, blankLine(), strongLine(invitationFieldLabels[form.focus]+", for "+form.person.name()))
	focus := len(content)
	if form.chooser.rows() == 0 {
		content = append(content, lines(toneFaint, wrapText("No contact has a name and an email or Slack handle. Add one on the Contacts tab, press c, then come back.", width))...)
	}
	for row := 0; row < form.chooser.rows(); row++ {
		if row == form.chooser.cursor {
			focus = len(content)
		}
		content = append(content, hereLine(form.chooser.rowText(row), row == form.chooser.cursor, true, width, colors))
	}
	content = append(content, blankLine())
	if form.chooser.single {
		content = append(content, lines(toneFaint, wrapText("Only contacts with a name and an email or Slack handle are listed. Add one on the Contacts tab, press c.", width))...)
		return append(content, lines(toneFaint, wrapText("Enter chooses. Esc goes back without changing it.", width))...), focus
	}
	content = append(content, lines(toneFaint, wrapText("Space picks or unpicks. Enter confirms. Esc goes back without changing it. Pick order is kept.", width))...)
	if order := form.chooser.orderText(); order != "" {
		content = append(content, lines(toneFaint, hanging("", order, width))...)
	}
	return content, focus
}

func (app App) invitationConfirmLines(width int) []line {
	form := app.invitation.form
	content := []line{strongLine("Write the invitation for " + form.person.name() + "?"), faintLine("Account " + visibleText(form.person.Account)), blankLine()}
	content = append(content, bodyLine("File"))
	content = append(content, lines(toneBody, hanging("  ", visibleText(form.output.Value()), width))...)
	for _, row := range [][2]string{{"Pre-access help", form.preaccessName()}, {"General help", form.helpNames()}} {
		content = append(content, lines(toneBody, hanging(row[0]+spaces(preparationLabelWidth-cellWidth(row[0])), row[1], width))...)
	}
	content = append(content, blankLine())
	content = append(content, lines(toneBody, wrapText("The file holds the plan, this preparation, an empty record, and the first steps guide shown here. An existing file is never replaced.", width))...)
	content = append(content, blankLine())
	content = append(content, lines(toneBody, wrapText(sendsNothingNotice, width))...)
	return append(content, blankLine(), faintLine("Nothing is written until you press y."))
}

func (app App) invitationGuideLines(width int) []line {
	form := app.invitation.form
	if form.guide == "" {
		return lines(toneFaint, wrapText("The guide cannot be built from this help. The guide problems say why.", width))
	}
	var wrapped []string
	for _, source := range guideLines(form.guide) {
		wrapped = append(wrapped, wrapIndented(source, width)...)
	}
	return lines(toneBody, wrapped)
}

func (app App) invitationGuideTitle(width int, rows int) string {
	limit := max(len(app.invitationGuideLines(width))-rows, 0)
	if limit == 0 {
		return "First steps guide, all shown"
	}
	return fmt.Sprintf("First steps guide, %d%%", min(app.invitation.form.guideScroll, limit)*100/limit)
}
