package finder

import "onboard-kit/internal/model"

func isFound(choices []model.Choice, value model.Value) bool {
	switch value.Kind {
	case model.ValueKindDeferred:
		return true
	case model.ValueKindText:
		return hasChoice(choices, value.Text)
	case model.ValueKindList:
		return len(knownItems(choices, value.List)) > 0
	}
	return false
}

func knownItems(choices []model.Choice, items []string) []string {
	known := []string{}
	for _, item := range items {
		if hasChoice(choices, item) {
			known = append(known, item)
		}
	}
	return known
}

func foundAnswer(choices []model.Choice, value model.Value) (model.Value, bool) {
	if !isFound(choices, value) {
		return model.Value{}, false
	}
	if value.Kind == model.ValueKindList {
		return model.Value{Kind: model.ValueKindList, List: knownItems(choices, value.List)}, true
	}
	return value, true
}
