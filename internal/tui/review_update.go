package tui

import tea "github.com/charmbracelet/bubbletea"

func (app App) updateReview(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if app.review.pruning.asking {
		return app.updatePruning(key), nil
	}
	scrolling := app.review.focusDocument
	switch key.String() {
	case "up", "k":
		return app.reviewStep(-1, scrolling), nil
	case "down", "j":
		return app.reviewStep(1, scrolling), nil
	case "pgup", "b":
		return app.scrollDocument(-app.pageSize()), nil
	case "pgdown", "f", " ":
		return app.scrollDocument(app.pageSize()), nil
	case "g", "home":
		return app.reviewEdge(false, scrolling), nil
	case "G", "end":
		return app.reviewEdge(true, scrolling), nil
	case "tab":
		app.review.focusDocument = !app.review.focusDocument
	case "enter":
		return app.openReviewItem(), nil
	case "esc", "w":
		return app.walkFromReview(), nil
	case "o":
		return app.openOverview(), nil
	case "s":
		return app.openSections(), nil
	case "c":
		return app.openContacts(), nil
	case "h":
		return app.askToPrune(), nil
	case "q":
		return app.quit()
	case "?":
		app.helpOpen = true
	}
	return app, nil
}

func (app App) reviewStep(delta int, scrolling bool) App {
	if scrolling {
		return app.scrollDocument(delta)
	}
	app.review = app.review.moved(delta)
	return app
}

func (app App) reviewEdge(end bool, scrolling bool) App {
	switch {
	case scrolling && end:
		app.review = app.review.scrolled(app.documentOffsetLimit(), app.documentOffsetLimit())
	case scrolling:
		app.review.scroll = 0
	case end:
		app.review.cursor = max(len(app.review.items)-1, 0)
	default:
		app.review.cursor = 0
	}
	return app
}

func (app App) scrollDocument(delta int) App {
	app.review = app.review.scrolled(app.documentOffsetLimit(), delta)
	return app
}

func (app App) openReviewItem() App {
	item, ok := app.review.selected()
	if !ok {
		return app
	}
	index := app.indexOf(item.questionID)
	if index < 0 {
		app.status = line{text: "no question to open for this item", tone: toneWarning}
		return app
	}
	app.returnTo = screenReview
	return app.openQuestion(index)
}

func (app App) walkFromReview() App {
	if len(app.questions) == 0 {
		return app
	}
	app.returnTo = screenReview
	return app.openQuestion(app.firstUnanswered())
}
