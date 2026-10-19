package finder

import (
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
		return slices.ContainsFunc(value.List, func(item string) bool { return item != "" })
	}
	return false
}

func isAnswered(value model.Value) bool {
	return isSettled(value) || value.Kind == model.ValueKindDeferred
}

func sameValue(left model.Value, right model.Value) bool {
	return left.Kind == right.Kind && left.Text == right.Text && left.Flag == right.Flag && slices.Equal(left.List, right.List)
}
