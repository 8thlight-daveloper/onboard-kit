package validator

import (
	"slices"

	"onboard-kit/internal/model"
)

func answerProblems(directory model.Directory, severity model.Severity, variable model.Variable, value model.Value) []model.Problem {
	if isUnanswered(value) {
		return nil
	}
	if problem, wrong := wrongKind(variable, value); wrong {
		return []model.Problem{problem}
	}
	switch typeOf(variable) {
	case model.VariableTypeChoice:
		return optionProblems(variable, value)
	case model.VariableTypeString:
		return patternProblems(variable, value)
	case model.VariableTypePerson:
		return personProblems(value)
	case model.VariableTypeContact, model.VariableTypeRole, model.VariableTypeTeam:
		return directoryIDProblems(directory, variable, value, severity)
	default:
		return nil
	}
}

func typeOf(variable model.Variable) model.VariableType {
	if variable.Type == "" {
		return model.DefaultVariableType
	}
	return variable.Type
}

func isUnanswered(value model.Value) bool {
	switch value.Kind {
	case model.ValueKindText:
		return value.Text == ""
	case model.ValueKindList:
		return !slices.ContainsFunc(value.List, func(item string) bool { return item != "" })
	default:
		return value.Kind == model.ValueKindNone || value.Kind == model.ValueKindDeferred
	}
}
