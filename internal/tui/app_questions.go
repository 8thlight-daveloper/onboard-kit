package tui

import "onboard-kit/internal/model"

func (app App) current() (model.Question, bool) {
	if app.cursor < 0 || app.cursor >= len(app.questions) {
		return model.Question{}, false
	}
	return app.questions[app.cursor], true
}

func (app App) holdsAnswers() bool { return app.answeredCount() > 0 }

func (app App) answeredCount() int {
	count := 0
	for _, question := range app.questions {
		if stateOf(question.Answer) != "unanswered" {
			count++
		}
	}
	return count
}

func (app App) firstUnanswered() int {
	for index, question := range app.questions {
		if stateOf(question.Answer) == "unanswered" {
			return index
		}
	}
	return app.cursor
}

func (app App) indexOf(id string) int {
	for index, question := range app.questions {
		if question.ID == id {
			return index
		}
	}
	return -1
}

func (app App) canSwitchSections() bool { return app.audience != model.AudienceOnboardee }

func (app App) sectionTitle(scope string) string {
	if scope == "plan" {
		return "Plan"
	}
	for _, section := range app.plan.Sections {
		if section.ID == scope {
			return section.Title
		}
	}
	return scope
}

func isOn(section model.Section) bool {
	return section.Enabled == nil || *section.Enabled
}
