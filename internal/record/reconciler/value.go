package reconciler

import (
	"maps"
	"slices"

	"onboard-kit/internal/model"
)

func isSettled(value model.Value) bool {
	switch value.Kind {
	case model.ValueKindText:
		return value.Text != ""
	case model.ValueKindFlag:
		return true
	case model.ValueKindList:
		return slices.ContainsFunc(value.List, isNonEmpty)
	default:
		return false
	}
}

func isNonEmpty(text string) bool {
	return text != ""
}

func copiedValue(value model.Value) model.Value {
	value.List = slices.Clone(value.List)
	return value
}

func copiedValues(values map[string]model.Value) map[string]model.Value {
	copied := maps.Clone(values)
	for name, value := range copied {
		copied[name] = copiedValue(value)
	}
	return copied
}
