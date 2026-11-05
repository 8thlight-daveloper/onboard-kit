package tui

import tea "github.com/charmbracelet/bubbletea"

const (
	draftKeptText    = "Your text is kept, not saved yet."
	draftDroppedText = "Your unsaved text on the last question was dropped."
)

func (app App) tabIndex(target screen) int {
	for index, entry := range app.tabs() {
		if entry.screen == target {
			return index
		}
	}
	return 0
}

func (app App) focusedStrip() App {
	app.stripFocused = true
	app.stripChoice = app.tabIndex(app.screen)
	return app
}

func (app App) updateStrip(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	last := len(app.tabs()) - 1
	switch key.String() {
	case "left", "h":
		app.stripChoice = max(app.stripChoice-1, 0)
	case "right", "l":
		app.stripChoice = min(app.stripChoice+1, last)
	case "enter":
		return app.openScreen(app.tabs()[app.stripChoice].screen), nil
	case "down", "j", "esc":
		app.stripFocused = false
	case "tab", "shift+tab":
		return app.pageMoved(pageStep(key.String())), nil
	case "q":
		if app.screen == screenAnswer {
			return app.quit()
		}
	case "?":
		app.helpOpen = true
	default:
		if target, ok := app.tabWithKey(key.String()); ok {
			return app.openScreen(target), nil
		}
	}
	return app, nil
}

func (app App) tabWithKey(key string) (screen, bool) {
	for _, entry := range app.tabs() {
		if entry.key == key {
			return entry.screen, true
		}
	}
	return app.screen, false
}

func pageStep(key string) int {
	switch key {
	case "tab":
		return 1
	case "shift+tab":
		return -1
	}
	return 0
}

func (app App) tabMovesPages() bool {
	switch app.screen {
	case screenReview:
		return !app.review.pruning.asking
	case screenContacts:
		return app.contacts.mode == contactsBrowsing
	case screenPreparation:
		mode := app.preparation.mode
		return mode == preparationListing || mode == preparationAdding || mode == preparationPerson || mode == preparationReviewing
	case screenInvitation:
		mode := app.invitation.mode
		return mode == invitationListing || mode == invitationBlocked || mode == invitationWritten
	}
	return true
}

func (app App) pageMoved(delta int) App {
	tabs := app.tabs()
	next := (app.tabIndex(app.screen) + delta + len(tabs)) % len(tabs)
	return app.openScreen(tabs[next].screen)
}

func (app App) openScreen(target screen) App {
	app.stripFocused = false
	if target == app.screen {
		return app
	}
	leaving := app.screen == screenAnswer && app.entry.hasDraft()
	opened := app.openedScreen(target)
	if leaving && opened.screen != screenAnswer {
		opened.status = line{text: draftKeptText, tone: toneWarning}
	}
	return opened
}

func (app App) openedScreen(target screen) App {
	switch target {
	case screenAnswer:
		if app.screen == screenReview {
			return app.answerFromReview()
		}
		return app.returnToAnswer()
	case screenOverview:
		return app.openOverview()
	case screenReview:
		return app.openReview()
	case screenSections:
		return app.openSections()
	case screenContacts:
		return app.openContacts()
	case screenPreparation:
		return app.openPreparation()
	case screenInvitation:
		return app.openInvitation()
	}
	return app
}
