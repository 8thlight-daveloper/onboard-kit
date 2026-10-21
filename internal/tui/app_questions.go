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
		if answerState(question) != "unanswered" {
			count++
		}
	}
	return count
}

func (app App) firstUnanswered() int {
	for index, question := range app.questions {
		if answerState(question) == "unanswered" {
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

func (app App) sectionAncestors(scope string) []string {
	parents := map[string]string{}
	for _, section := range app.plan.Sections {
		parents[section.ID] = section.Parent
	}
	var ancestors []string
	seen := map[string]bool{scope: true}
	for parent := parents[scope]; parent != "" && !seen[parent]; parent = parents[parent] {
		seen[parent] = true
		ancestors = append(ancestors, parent)
	}
	return ancestors
}

func (app App) sectionDepth(scope string) int { return len(app.sectionAncestors(scope)) }

func (app App) sectionRoot(scope string) string {
	ancestors := app.sectionAncestors(scope)
	if len(ancestors) == 0 {
		return scope
	}
	return ancestors[len(ancestors)-1]
}
