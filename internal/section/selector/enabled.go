package selector

import "onboard-kit/internal/model"

const enabledName = "enabled"

func isEnabled(section model.Section, answers model.Answers) bool {
	variable, declared := findEnabledVariable(section.Variables)
	if !declared {
		return true
	}
	answer := answers.Sections[section.ID][enabledName]
	if answer.Kind == model.ValueKindFlag {
		return answer.Flag
	}
	return variable.Default.Kind == model.ValueKindFlag && variable.Default.Flag
}

func findEnabledVariable(variables []model.Variable) (model.Variable, bool) {
	for _, variable := range variables {
		if variable.Name == enabledName {
			return variable, true
		}
	}
	return model.Variable{}, false
}
