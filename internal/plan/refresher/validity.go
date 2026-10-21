package refresher

import (
	"slices"

	"onboard-kit/internal/model"
)

func checkedAnswer(value model.Value, variable model.Variable, directory model.Directory) (model.Value, listing) {
	switch {
	case !isSettled(value):
		return value, listingKept
	case isDirectoryType(variable.Type):
		return checkedDirectoryAnswer(value, variable, directory)
	default:
		return keptIfListed(value, isValidAnswer(value, variable))
	}
}

func isValidAnswer(value model.Value, variable model.Variable) bool {
	if value.Kind != expectedKind(variable) {
		return false
	}
	if variable.Type != model.VariableTypeChoice {
		return true
	}
	if value.Kind == model.ValueKindList {
		return slices.IndexFunc(value.List, func(item string) bool {
			return item != "" && !slices.Contains(variable.Options, item)
		}) < 0
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
