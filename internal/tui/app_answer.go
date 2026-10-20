package tui

import "onboard-kit/internal/model"

func (app App) submit(value model.Value) (App, bool) {
	question, ok := app.current()
	if !ok {
		return app, false
	}
	next, problems := app.backend.Submit(app.plan, question.ID, value)
	if hasError(problems) {
		app.entry.problems = problems
		app.status = line{text: "not saved", tone: toneProblem}
		return app, false
	}
	app.plan = next
	app = app.refreshed(question.ID)
	app.status = app.saveStatus(problems)
	return app, true
}

func (app App) saveStatus(problems []model.Problem) line {
	if err := app.save(app.plan); err != nil {
		return line{text: "could not save: " + err.Error(), tone: toneProblem}
	}
	if len(problems) > 0 {
		return line{text: "saved. warning: " + firstMessage(problems), tone: toneWarning}
	}
	return line{text: "saved", tone: toneFaint}
}

func (app App) refreshed(keep string) App {
	app.questions = app.backend.Questions(app.plan, app.audience)
	app.cursor = min(app.cursor, max(len(app.questions)-1, 0))
	if index := app.indexOf(keep); index >= 0 {
		app.cursor = index
	}
	app.overviewCursor = min(app.overviewCursor, max(len(app.questions)-1, 0))
	return app
}

func hasError(problems []model.Problem) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func firstMessage(problems []model.Problem) string {
	if len(problems) == 0 {
		return ""
	}
	return problems[0].Message
}
