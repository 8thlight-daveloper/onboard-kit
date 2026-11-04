package tui

import "fmt"

func (app App) preparationReviewColumns(screenFrame frame) (int, int) {
	if !screenFrame.wide() {
		return screenFrame.innerWidth(), screenFrame.innerWidth()
	}
	return screenFrame.columns(itemsShareNumerator, itemsShareDenominator)
}

func (app App) preparationReviewBody(screenFrame frame, colors palette) []string {
	changesWidth, guideWidth := app.preparationReviewColumns(screenFrame)
	review := app.preparation.review
	rows := screenFrame.documentRows()
	changes := windowFrom(app.reviewChangeLines(changesWidth), review.scroll, rows)
	guide := windowFrom(app.reviewGuideLines(guideWidth), review.guideScroll, rows)
	left := pane{lines: append([]line{paneTitle("Changes", !review.focusGuide, colors)}, changes...), focused: !review.focusGuide}
	right := pane{lines: append([]line{paneTitle(app.guideTitle(guideWidth), review.focusGuide, colors)}, guide...), focused: review.focusGuide}
	switch {
	case screenFrame.wide():
		return joinPanes(left, right, changesWidth, guideWidth, screenFrame.bodyHeight, colors)
	case review.focusGuide:
		return stackPane(right, guideWidth, screenFrame.bodyHeight, colors)
	}
	return stackPane(left, changesWidth, screenFrame.bodyHeight, colors)
}

func windowFrom(content []line, offset int, height int) []line {
	first := min(max(offset, 0), max(len(content)-height, 0))
	return content[first:min(first+height, len(content))]
}

func (app App) guideTitle(width int) string {
	limit := max(len(app.reviewGuideLines(width))-app.belowHeading(app.frame()).documentRows(), 0)
	if limit == 0 {
		return "First steps guide, all shown"
	}
	return fmt.Sprintf("First steps guide, %d%%", min(app.preparation.review.guideScroll, limit)*100/limit)
}

func (app App) reviewChangeLines(width int) []line {
	review := app.preparation.review
	person := app.preparation.person
	content := []line{strongLine("Save the preparation for " + visibleText(person.loaded.Account) + "?")}
	if person.loaded.FoldersMissing {
		content = append(content, lines(toneFaint, wrapText(foldersNotice, width))...)
	}
	content = append(content, blankLine())
	switch {
	case !person.loaded.Exists:
		content = append(content, lines(toneBody, wrapText("A new preparation. Saving creates the file.", width))...)
		content = append(content, blankLine())
	case len(review.changes) == 0:
		content = append(content, lines(toneBody, wrapText("No changes. Saving writes nothing.", width))...)
	}
	if len(review.changes) > 0 {
		content = append(content, strongLine(fmt.Sprintf("Changes (%d)", len(review.changes))))
	}
	for _, change := range review.changes {
		content = append(content, lines(toneBody, hanging("  ", change.text, width))...)
		for _, detail := range change.details {
			content = append(content, lines(toneBody, hanging("    ", detail, width))...)
		}
	}
	content = append(content, reviewGroup("Preparation problems", problemTexts(review.problems), toneProblem, width)...)
	if len(review.problems) > 0 {
		content = append(content, blankLine(), strongLine("Blockers in the plan"))
		content = append(content, lines(toneFaint, hanging("  ", "Not checked until the preparation problems are fixed.", width))...)
	} else {
		content = append(content, reviewGroup("Blockers in the plan", app.blockerTexts(review.blockers), toneBody, width)...)
	}
	content = append(content, reviewGroup("Guide problems", problemTexts(review.guideProblems), toneWarning, width)...)
	content = append(content, blankLine())
	if review.refusal != "" {
		return append(content, lines(toneProblem, hanging("! problem: ", review.refusal, width))...)
	}
	return append(content, faintLine("Nothing is written until you press y."))
}

func reviewGroup(title string, texts []string, weight tone, width int) []line {
	if len(texts) == 0 {
		return nil
	}
	content := []line{blankLine(), strongLine(fmt.Sprintf("%s (%d)", title, len(texts)))}
	for _, text := range texts {
		content = append(content, lines(weight, hanging("  ", text, width))...)
	}
	return content
}

func (app App) reviewGuideLines(width int) []line {
	review := app.preparation.review
	switch {
	case len(review.problems) > 0:
		return lines(toneFaint, wrapText("The guide is not built until the preparation problems are fixed.", width))
	case review.guide == "":
		return lines(toneFaint, wrapText("The guide cannot be built. The guide problems say why.", width))
	}
	var wrapped []string
	for _, source := range guideLines(review.guide) {
		wrapped = append(wrapped, wrapIndented(source, width)...)
	}
	return lines(toneBody, wrapped)
}

func (app App) scrolledPreparationReview(delta int) preparationReview {
	review := app.preparation.review
	changesWidth, guideWidth := app.preparationReviewColumns(app.frame())
	rows := app.belowHeading(app.frame()).documentRows()
	if review.focusGuide {
		limit := max(len(app.reviewGuideLines(guideWidth))-rows, 0)
		review.guideScroll = max(min(review.guideScroll+delta, limit), 0)
		return review
	}
	limit := max(len(app.reviewChangeLines(changesWidth))-rows, 0)
	review.scroll = max(min(review.scroll+delta, limit), 0)
	return review
}
