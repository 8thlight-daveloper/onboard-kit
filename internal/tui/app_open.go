package tui

import "onboard-kit/internal/model"

func (app App) openQuestion(index int) App {
	app.cursor = max(min(index, len(app.questions)-1), 0)
	if question, ok := app.current(); ok {
		app = app.enteredQuestion(question)
	}
	app.screen = screenAnswer
	return app
}

func (app App) enteredQuestion(question model.Question) App {
	entry := newEntryState(question, app.inputWidth())
	if app.entry.question.ID != "" && app.entry.question.Scope == question.Scope {
		app.drafts = app.drafts.kept(app.entry)
		app.entry = app.drafts.restored(entry)
		return app
	}
	if dropped := app.draftRows(); len(dropped) > 0 {
		app.status = draftDroppedLine(dropped)
	}
	app.drafts = nil
	app.entry = entry
	return app
}

func (app App) openOverview() App {
	app.overviewCursor = app.cursor
	app.screen = screenOverview
	app.returnTo = screenOverview
	return app
}

func (app App) openReview() App {
	document, _ := app.backend.Assemble(app.plan)
	app.review = newReviewState(app.backend, app.plan, app.audience, app.questions, document)
	app.screen = screenReview
	app.returnTo = screenReview
	return app
}

func (app App) openSections() App {
	if !app.canSwitchSections() {
		return app
	}
	app = app.leavingForSideScreen()
	app.screen = screenSections
	return app
}

func (app App) openContacts() App {
	if !app.canEditDirectory() {
		return app
	}
	if app.contacts.mode == contactsEditing {
		app = app.leavingForSideScreen()
		app.screen = screenContacts
		return app.refocusedForm()
	}
	if question, ok := app.current(); ok && app.screen == screenAnswer {
		if kind, found := kindOfQuestion(question); found {
			app.contacts.kind = kind
		}
	}
	app = app.leavingForSideScreen()
	app.contacts.mode = contactsBrowsing
	app.screen = screenContacts
	return app.refreshedContacts()
}

func (app App) leavingForSideScreen() App {
	if app.screen != screenSections && app.screen != screenContacts && app.screen != screenPreparation && app.screen != screenInvitation {
		app.cameFrom = app.screen
	}
	return app
}

func (app App) closeSideScreen() App {
	switch app.cameFrom {
	case screenOverview:
		app.screen = screenOverview
		app.overviewCursor = min(app.overviewCursor, max(len(app.questions)-1, 0))
		return app
	case screenReview:
		return app.openReview()
	}
	return app.openQuestion(app.cursor)
}

func (app App) returnToAnswer() App {
	if len(app.questions) == 0 {
		return app
	}
	return app.openQuestion(app.cursor)
}

func (app App) moveTo(index int) App {
	if index >= len(app.questions) {
		drafted := app.draftRows()
		reviewed := app.openReview()
		if len(drafted) > 0 {
			reviewed.status = draftKeptLine(drafted)
		}
		return reviewed
	}
	return app.openQuestion(max(index, 0))
}

func (app App) jumpTo(index int) App {
	if index < 0 || index >= len(app.questions) {
		return app
	}
	return app.openQuestion(index)
}
