package refresher

import (
	"slices"

	"onboard-kit/internal/model"
)

func sameValue(left, right model.Value) bool {
	if left.Kind != right.Kind {
		return false
	}
	switch left.Kind {
	case model.ValueKindText:
		return left.Text == right.Text
	case model.ValueKindFlag:
		return left.Flag == right.Flag
	case model.ValueKindList:
		return slices.Equal(left.List, right.List)
	default:
		return true
	}
}

func isAbsent(value model.Value) bool {
	return value.Kind == model.ValueKindNone
}

func isSettled(value model.Value) bool {
	switch value.Kind {
	case model.ValueKindText:
		return value.Text != ""
	case model.ValueKindFlag:
		return true
	case model.ValueKindList:
		return slices.ContainsFunc(value.List, nonEmpty)
	default:
		return false
	}
}

func nonEmpty(text string) bool {
	return text != ""
}
