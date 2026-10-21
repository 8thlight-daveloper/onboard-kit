package finder

import (
	"onboard-kit/internal/condition/evaluator"
	"onboard-kit/internal/model"
)

func findActive(plan model.Plan, candidates []candidate) []candidate {
	conditions := evaluator.New()
	switches := collectSwitches(plan.Sections)
	decided := model.Answers{Plan: map[string]model.Value{}, Sections: map[string]map[string]model.Value{}}
	isActive := make([]bool, len(candidates))
	for _, position := range sequence(positionsNamed(candidates)) {
		asked := candidates[position]
		if !conditionHolds(conditions, asked, decided, switches) {
			continue
		}
		isActive[position] = true
		if value, found := decidedAnswer(plan, asked); found {
			record(decided, asked, value)
		}
	}
	active := []candidate{}
	for position, asked := range candidates {
		if isActive[position] {
			active = append(active, asked)
		}
	}
	return active
}

func conditionHolds(conditions evaluator.Evaluator, asked candidate, decided model.Answers, switches model.Switches) bool {
	if asked.variable.AskIf == nil {
		return true
	}
	return conditions.Holds(*asked.variable.AskIf, asked.scope, decided, switches)
}

func record(decided model.Answers, asked candidate, value model.Value) {
	if asked.scope == planScope {
		decided.Plan[asked.variable.Name] = value
		return
	}
	if decided.Sections[asked.scope] == nil {
		decided.Sections[asked.scope] = map[string]model.Value{}
	}
	decided.Sections[asked.scope][asked.variable.Name] = value
}

func collectSwitches(sections []model.Section) model.Switches {
	switches := make(model.Switches, len(sections))
	for _, section := range sections {
		switches[section.ID] = section.Enabled == nil || *section.Enabled
	}
	return switches
}

func orderActive(active []candidate) []candidate {
	ordered := make([]candidate, 0, len(active))
	for _, position := range sequence(positionsNamed(active)) {
		ordered = append(ordered, active[position])
	}
	return ordered
}

func decidedAnswer(plan model.Plan, asked candidate) (model.Value, bool) {
	value, found := storedAnswer(plan.Answers, asked.scope, asked.variable.Name)
	if !found || !isDirectoryType(asked.variable.Type) {
		return value, found
	}
	return foundAnswer(directoryChoices(plan.Directory, asked.variable.Type), value)
}
