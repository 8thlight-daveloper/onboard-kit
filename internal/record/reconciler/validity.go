package reconciler

import (
	"slices"

	"onboard-kit/internal/model"
)

func isValidAnswer(value model.Value, variable model.Variable) bool {
	if value.Kind != expectedKind(variable) {
		return false
	}
	if variable.Type != model.VariableTypeChoice {
		return true
	}
	if value.Kind == model.ValueKindList {
		return !slices.ContainsFunc(value.List, func(item string) bool {
			return item != "" && !slices.Contains(variable.Options, item)
		})
	}
	return slices.Contains(variable.Options, value.Text)
}

func expectedKind(variable model.Variable) model.ValueKind {
	switch {
	case variable.Multiple || variable.Type == model.VariableTypeCommand:
		return model.ValueKindList
	case variable.Type == model.VariableTypeBoolean:
		return model.ValueKindFlag
	default:
		return model.ValueKindText
	}
}
