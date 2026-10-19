package evaluator

import "onboard-kit/internal/model"

func isEqual(answer model.Value, expected model.Value) bool {
	if answer.Kind != expected.Kind {
		return false
	}
	switch answer.Kind {
	case model.ValueKindText:
		return answer.Text == expected.Text
	case model.ValueKindFlag:
		return answer.Flag == expected.Flag
	default:
		return false
	}
}

func isList(value model.Value) bool {
	return value.Kind == model.ValueKindList
}

func isEqualToAny(answer model.Value, expected []model.Value) bool {
	for _, value := range expected {
		if isEqual(answer, value) {
			return true
		}
	}
	return false
}

func containsList(values []model.Value) bool {
	for _, value := range values {
		if isList(value) {
			return true
		}
	}
	return false
}
