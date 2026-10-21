package tui

import "fmt"

const itemsShareNumerator, itemsShareDenominator = 2, 5

func (app App) reviewColumns(screenFrame frame) (int, int) {
	if !screenFrame.wide() {
		return screenFrame.innerWidth(), screenFrame.innerWidth()
	}
	return screenFrame.columns(itemsShareNumerator, itemsShareDenominator)
}

func (app App) reviewBody(screenFrame frame, colors palette) []string {
	if app.review.pruning.asking {
		width, _ := screenFrame.columns(answerShareNumerator, answerShareDenominator)
		return stackPane(pane{lines: clipLines(app.pruningLines(width), screenFrame.bodyHeight), focused: true}, width, screenFrame.bodyHeight, colors)
	}
	itemsWidth, documentWidth := app.reviewColumns(screenFrame)
	itemsFocused := !app.review.focusDocument
	rows, focus := app.reviewItemLines(itemsWidth, itemsFocused, colors)
	items := pane{lines: append([]line{paneTitle("Items", itemsFocused, colors)}, scrollWindow(rows, focus, screenFrame.documentRows())...), focused: itemsFocused}
	document := pane{lines: append([]line{paneTitle(app.documentTitle(), !itemsFocused, colors)}, app.documentView(screenFrame)...), focused: !itemsFocused}
	switch {
	case screenFrame.wide():
		return joinPanes(items, document, itemsWidth, documentWidth, screenFrame.bodyHeight, colors)
	case app.review.focusDocument:
		return stackPane(document, documentWidth, screenFrame.bodyHeight, colors)
	}
	return stackPane(items, itemsWidth, screenFrame.bodyHeight, colors)
}

func paneTitle(name string, focused bool, colors palette) line {
	if !focused {
		return line{text: "  " + name, tone: toneFaint}
	}
	return line{text: colors.render(toneHere, true, "> ") + colors.render(toneStrong, true, name), raw: true}
}

func (app App) documentTitle() string {
	limit := app.documentOffsetLimit()
	if limit == 0 {
		return "Document, all shown"
	}
	return fmt.Sprintf("Document, %d%%", app.review.scroll*100/limit)
}

func (app App) documentView(screenFrame frame) []line {
	wrapped := app.documentLines()
	first := min(app.review.scroll, len(wrapped))
	last := min(first+screenFrame.documentRows(), len(wrapped))
	return lines(toneBody, wrapped[first:last])
}

func (app App) reviewItemLines(width int, focused bool, colors palette) ([]line, int) {
	state := app.review
	content := []line{bodyLine(fmt.Sprintf("Answered %d of %d.", state.answered, state.total))}
	if state.leftOver > 0 {
		content = append(content, lines(toneFaint, wrapText(leftOverNotice(state.leftOver), width))...)
	}
	if len(state.items) == 0 {
		return append(content, bodyLine("Nothing is waiting.")), 0
	}
	focus := 0
	groups := []struct {
		kind  reviewKind
		title string
	}{{reviewBlocker, "Blockers"}, {reviewDeferred, "Deferred"}, {reviewProblem, "Problems"}}
	for _, group := range groups {
		if state.counts(group.kind) == 0 {
			continue
		}
		content = append(content, blankLine(), strongLine(fmt.Sprintf("%s (%d)", group.title, state.counts(group.kind))))
		for index, item := range state.items {
			if item.kind != group.kind {
				continue
			}
			if index == state.cursor {
				focus = len(content)
			}
			content = append(content, itemLines(item, index == state.cursor, focused, width, colors)...)
		}
	}
	return content, focus
}

func itemLines(item reviewItem, selected bool, focused bool, width int, colors palette) []line {
	text := item.label + ": " + item.text
	if selected {
		rows := wrapText(text, width-2)
		content := []line{hereLine(rows[0], true, focused, width, colors)}
		for _, row := range rows[1:] {
			content = append(content, line{text: "  " + row, tone: toneStrong})
		}
		return content
	}
	weight := toneBody
	switch item.severity {
	case "error":
		weight = toneProblem
	case "warning":
		weight = toneWarning
	}
	return lines(weight, hanging("  ", text, width))
}
