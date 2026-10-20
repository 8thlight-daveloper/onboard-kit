package editor

import (
	"slices"

	"onboard-kit/internal/model"
)

func isUnanswered(value model.Value) bool {
	switch value.Kind {
	case model.ValueKindNone:
		return true
	case model.ValueKindText:
		return value.Text == ""
	case model.ValueKindList:
		return !slices.ContainsFunc(value.List, func(item string) bool { return item != "" })
	default:
		return false
	}
}
