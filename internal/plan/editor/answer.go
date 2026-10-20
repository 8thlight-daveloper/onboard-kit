package editor

import (
	"strings"

	"onboard-kit/internal/model"
)

const planScope = "plan"

func applyAnswer(plan model.Plan, question string, value model.Value) model.Plan {
	scope, name, _ := strings.Cut(question, ".")
	plan = copyEditableParts(plan)
	if scope == planScope {
		plan.Answers.Plan = changedValues(plan.Answers.Plan, name, value)
		return plan
	}
	plan.Answers.Sections = changedSectionAnswers(plan.Answers.Sections, scope, name, value)
	return plan
}

func changedSectionAnswers(sections map[string]map[string]model.Value, scope, name string, value model.Value) map[string]map[string]model.Value {
	changed := changedValues(sections[scope], name, value)
	if len(changed) == 0 {
		delete(sections, scope)
		return sections
	}
	if sections == nil {
		sections = map[string]map[string]model.Value{}
	}
	sections[scope] = changed
	return sections
}

func changedValues(values map[string]model.Value, name string, value model.Value) map[string]model.Value {
	if isUnanswered(value) {
		delete(values, name)
		return values
	}
	if values == nil {
		values = map[string]model.Value{}
	}
	values[name] = copyValue(value)
	return values
}
