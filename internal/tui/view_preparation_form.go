package tui

func (app App) preparationFormBody(screenFrame frame, colors palette) []string {
	width, _ := screenFrame.columns(answerShareNumerator, answerShareDenominator)
	content, focus := app.preparationFormLines(width, colors)
	return stackPane(pane{lines: scrollWindow(content, focus, screenFrame.bodyHeight), focused: true}, width, screenFrame.bodyHeight, colors)
}

func (form preparationForm) title() string {
	switch {
	case !form.arrangement:
		return "Recipient"
	case form.editing == "":
		return "New arrangement"
	}
	return "Arrangement " + visibleText(form.editing)
}

func (app App) preparationFormLines(width int, colors palette) ([]line, int) {
	form := app.preparation.form
	content := []line{strongLine(form.title())}
	if form.arrangement && form.editing != "" {
		content = append(content, lines(toneFaint, wrapText("Id: "+visibleText(form.editing)+", which never changes", width))...)
	}
	content = append(content, blankLine())
	focus := 0
	for index := range form.fields {
		if index == form.focus {
			focus = len(content)
		}
		content = append(content, form.preparationFieldLine(index, colors))
	}
	if form.choosing {
		content, focus = app.preparationPickerLines(content, width, colors)
	} else {
		content = append(content, blankLine())
		content = append(content, lines(toneFaint, wrapText("Enter moves to the next field and stages the form on the last. Ctrl+s stages it. Nothing is written until you review and save.", width))...)
	}
	return append(content, preparationProblemLines(form.problems, width)...), focus
}

func (form preparationForm) preparationFieldLine(index int, colors palette) line {
	field := form.fields[index]
	mark := spaces(formMarkWidth)
	if index == form.focus {
		mark = colors.render(toneHere, true, "> ")
	}
	label := colors.render(toneStrong, true, field.label+spaces(preparationLabelWidth-cellWidth(field.label)))
	if field.kind.typed() {
		return line{text: mark + label + field.input.View(), raw: true}
	}
	value := displayValues(field.kind, field.values(), form.plan)
	return line{text: mark + label + value, raw: true}
}

func (app App) preparationPickerLines(content []line, width int, colors palette) ([]line, int) {
	form := app.preparation.form
	content = append(content, blankLine(), strongLine(form.current().label+", for "+visibleText(app.preparation.person.loaded.Account)))
	focus := len(content)
	if form.chooser.rows() == 0 {
		content = append(content, lines(toneFaint, wrapText("Nothing to pick from yet. Add it to the plan first, then come back.", width))...)
	}
	for row := 0; row < form.chooser.rows(); row++ {
		if row == form.chooser.cursor {
			focus = len(content)
		}
		content = append(content, hereLine(form.chooser.rowText(row), row == form.chooser.cursor, true, width, colors))
	}
	content = append(content, blankLine())
	if form.chooser.single {
		return append(content, lines(toneFaint, wrapText("Enter chooses. Esc goes back without changing it.", width))...), focus
	}
	content = append(content, lines(toneFaint, wrapText("Space picks or unpicks. Enter confirms. Esc goes back without changing it. Pick order is kept.", width))...)
	if order := form.chooser.orderText(); order != "" {
		content = append(content, lines(toneFaint, hanging("", order, width))...)
	}
	return content, focus
}

func preparationProblemLines(problems []string, width int) []line {
	if len(problems) == 0 {
		return nil
	}
	content := []line{blankLine()}
	for _, problem := range problems {
		content = append(content, lines(toneProblem, hanging("! problem: ", problem, width))...)
	}
	return content
}
