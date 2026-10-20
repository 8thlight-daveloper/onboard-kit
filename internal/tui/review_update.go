package tui

import tea "github.com/charmbracelet/bubbletea"

func (a App) updateReview(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	scrolling := a.review.focusDocument
	switch key.String() {
	case "up", "k":
		return a.reviewStep(-1, scrolling), nil
	case "down", "j":
		return a.reviewStep(1, scrolling), nil
	case "pgup", "b":
		return a.scrollDocument(-a.pageSize()), nil
	case "pgdown", "f", " ":
		return a.scrollDocument(a.pageSize()), nil
	case "g", "home":
		return a.reviewEdge(false, scrolling), nil
	case "G", "end":
		return a.reviewEdge(true, scrolling), nil
	case "tab":
		a.review.focusDocument = !a.review.focusDocument
	case "enter":
		return a.openReviewItem(), nil
	case "esc", "w":
		return a.walkFromReview(), nil
	case "o":
		return a.openOverview(), nil
	case "s":
		return a.openSections(), nil
	case "q":
		return a.quit()
	case "?":
		a.helpOpen = true
	}
	return a, nil
}

func (a App) reviewStep(delta int, scrolling bool) App {
	if scrolling {
		return a.scrollDocument(delta)
	}
	a.review = a.review.moved(delta)
	return a
}

func (a App) reviewEdge(end bool, scrolling bool) App {
	switch {
	case scrolling && end:
		a.review = a.review.scrolled(a.documentOffsetLimit(), a.documentOffsetLimit())
	case scrolling:
		a.review.scroll = 0
	case end:
		a.review.cursor = max(len(a.review.items)-1, 0)
	default:
		a.review.cursor = 0
	}
	return a
}

func (a App) scrollDocument(delta int) App {
	a.review = a.review.scrolled(a.documentOffsetLimit(), delta)
	return a
}

func (a App) openReviewItem() App {
	item, ok := a.review.selected()
	if !ok {
		return a
	}
	index := a.indexOf(item.questionID)
	if index < 0 {
		a.status = line{text: "no question to open for this item", tone: toneWarning}
		return a
	}
	a.returnTo = screenReview
	return a.openQuestion(index)
}

func (a App) walkFromReview() App {
	if len(a.questions) == 0 {
		return a
	}
	a.returnTo = screenReview
	return a.openQuestion(a.firstUnanswered())
}
