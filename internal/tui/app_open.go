package tui

func (app App) openQuestion(index int) App {
	app.cursor = max(min(index, len(app.questions)-1), 0)
	if question, ok := app.current(); ok {
		app.entry = newEntryState(question, app.inputWidth())
	}
	app.screen = screenWalk
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
	app.cameFrom = app.screen
	app.screen = screenSections
	return app
}

func (app App) closeSections() App {
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

func (app App) returnToWalk() App {
	if len(app.questions) == 0 {
		return app
	}
	return app.openQuestion(app.cursor)
}

func (app App) moveTo(index int) App {
	if index >= len(app.questions) {
		return app.openReview()
	}
	return app.openQuestion(max(index, 0))
}

func (app App) jumpTo(index int) App {
	if index < 0 || index >= len(app.questions) {
		return app
	}
	return app.openQuestion(index)
}
