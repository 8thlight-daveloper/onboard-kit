package resolver

import "onboard-kit/internal/model"

func isSettled(value model.Value) bool {
	switch value.Kind {
	case model.ValueKindText:
		return value.Text != ""
	case model.ValueKindFlag:
		return true
	case model.ValueKindList:
		return hasNonEmptyItem(value.List)
	default:
		return false
	}
}

func hasNonEmptyItem(items []string) bool {
	for _, item := range items {
		if item != "" {
			return true
		}
	}
	return false
}
