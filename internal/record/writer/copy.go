package writer

import "onboard-kit/internal/model"

func copyRecord(record model.Record) model.Record {
	record.Answers.Plan = copyMap(record.Answers.Plan, copyValue)
	record.Answers.Sections = copyMap(record.Answers.Sections, copyValues)
	record.Outcomes = copySlice(record.Outcomes, identity[model.Outcome])
	record.UnknownFields = copyMap(record.UnknownFields, copyUnknown)
	return record
}

func copyValues(values map[string]model.Value) map[string]model.Value {
	return copyMap(values, copyValue)
}

func copyValue(value model.Value) model.Value {
	value.List = copySlice(value.List, identity[string])
	return value
}

func copyProblems(problems []model.Problem) []model.Problem {
	return copySlice(problems, func(problem model.Problem) model.Problem {
		problem.Detail = copyMap(problem.Detail, identity[string])
		return problem
	})
}

func copyUnknown(value any) any {
	switch item := value.(type) {
	case map[string]any:
		return copyMap(item, copyUnknown)
	case []any:
		return copySlice(item, copyUnknown)
	default:
		return value
	}
}

func copyMap[T any](values map[string]T, copyItem func(T) T) map[string]T {
	if values == nil {
		return nil
	}
	result := make(map[string]T, len(values))
	for key, value := range values {
		result[key] = copyItem(value)
	}
	return result
}

func copySlice[T any](values []T, copyItem func(T) T) []T {
	if values == nil {
		return nil
	}
	result := make([]T, len(values))
	for index, value := range values {
		result[index] = copyItem(value)
	}
	return result
}

func identity[T any](value T) T {
	return value
}
