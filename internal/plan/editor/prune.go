package editor

import (
	"maps"
	"slices"

	"onboard-kit/internal/model"
)

func (Editor) PruneHiddenAnswers(plan model.Plan) (model.Plan, model.Cleared, []model.Problem) {
	plan = copyEditableParts(plan)
	var hidden []string
	plan.Answers.Plan, hidden = removeUndeclared(plan.Answers.Plan, planScope, plan.Variables, hidden)
	for _, id := range slices.Sorted(maps.Keys(plan.Answers.Sections)) {
		values := plan.Answers.Sections[id]
		if len(values) == 0 {
			continue
		}
		values, hidden = removeUndeclared(values, id, sectionVariables(plan.Sections, id), hidden)
		if len(values) == 0 {
			delete(plan.Answers.Sections, id)
		}
	}
	return plan, model.Cleared{Answers: sortedUnique(hidden), Holders: []string{}}, []model.Problem{}
}

func removeUndeclared(values map[string]model.Value, scope string, variables []model.Variable, hidden []string) (map[string]model.Value, []string) {
	for name := range values {
		if _, declared := declaredIn(variables, name); declared {
			continue
		}
		delete(values, name)
		hidden = append(hidden, scope+"."+name)
	}
	return values, hidden
}

func sectionVariables(sections []model.Section, id string) []model.Variable {
	index := slices.IndexFunc(sections, func(section model.Section) bool { return section.ID == id })
	if index < 0 {
		return nil
	}
	return sections[index].Variables
}
