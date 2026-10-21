package splitter

import (
	"slices"

	"onboard-kit/internal/model"
)

func sameValue(first, second model.Value) bool {
	if first.Kind != second.Kind {
		return false
	}
	switch first.Kind {
	case model.ValueKindText:
		return first.Text == second.Text
	case model.ValueKindFlag:
		return first.Flag == second.Flag
	case model.ValueKindList:
		return slices.Equal(first.List, second.List)
	}
	return true
}

func isAnswer(value model.Value) bool {
	return value.Kind != model.ValueKindNone
}
