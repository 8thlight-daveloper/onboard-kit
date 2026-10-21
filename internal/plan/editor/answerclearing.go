package editor

import (
	"maps"
	"slices"

	"onboard-kit/internal/model"
)

func clearAnswers(plan model.Plan, kind model.VariableType, id string) (model.Plan, []string) {
	cleared := []string{}
	plan.Answers.Plan, cleared = clearScope(plan.Answers.Plan, planScope, plan.Variables, kind, id, cleared)
	for _, section := range plan.Sections {
		var values map[string]model.Value
		values, cleared = clearScope(plan.Answers.Sections[section.ID], section.ID, section.Variables, kind, id, cleared)
		if len(values) == 0 {
			delete(plan.Answers.Sections, section.ID)
		}
	}
	slices.Sort(cleared)
	return plan, slices.Compact(cleared)
}

func clearScope(values map[string]model.Value, scope string, variables []model.Variable, kind model.VariableType, id string, cleared []string) (map[string]model.Value, []string) {
	for _, name := range slices.Sorted(maps.Keys(values)) {
		variable, declared := declaredIn(variables, name)
		if !declared || variable.Type != kind {
			continue
		}
		remaining, changed := withoutID(values[name], id)
		if !changed {
			continue
		}
		cleared = append(cleared, scope+"."+name)
		if isUnanswered(remaining) {
			delete(values, name)
			continue
		}
		values[name] = remaining
	}
	return values, cleared
}

func declaredIn(variables []model.Variable, name string) (model.Variable, bool) {
	index := slices.IndexFunc(variables, func(variable model.Variable) bool { return variable.Name == name })
	if index < 0 {
		return model.Variable{}, false
	}
	return variables[index], true
}

func withoutID(value model.Value, id string) (model.Value, bool) {
	switch value.Kind {
	case model.ValueKindText:
		if value.Text == id {
			return model.Value{}, true
		}
		return value, false
	case model.ValueKindList:
		kept, removed := withoutItem(value.List, id)
		value.List = kept
		return value, removed
	default:
		return value, false
	}
}
