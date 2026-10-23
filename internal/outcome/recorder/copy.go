package recorder

import "onboard-kit/internal/model"

func copyRecord(record model.Record) model.Record {
	result := record
	result.Outcomes = copySlice(record.Outcomes)
	result.Answers.Plan = copyValues(record.Answers.Plan)
	if record.Answers.Sections != nil {
		result.Answers.Sections = make(map[string]map[string]model.Value, len(record.Answers.Sections))
		for section, values := range record.Answers.Sections {
			result.Answers.Sections[section] = copyValues(values)
		}
	}
	result.UnknownFields = copyUnknownMap(record.UnknownFields)
	return result
}

func copySlice[T any](values []T) []T {
	if values == nil {
		return nil
	}
	result := make([]T, len(values))
	copy(result, values)
	return result
}

func copyValues(values map[string]model.Value) map[string]model.Value {
	if values == nil {
		return nil
	}
	result := make(map[string]model.Value, len(values))
	for key, value := range values {
		value.List = copySlice(value.List)
		result[key] = value
	}
	return result
}

func copyUnknownMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	result := make(map[string]any, len(values))
	for key, value := range values {
		result[key] = copyUnknown(value)
	}
	return result
}

func copyUnknown(value any) any {
	switch value := value.(type) {
	case map[string]any:
		return copyUnknownMap(value)
	case []any:
		if value == nil {
			return []any(nil)
		}
		result := make([]any, len(value))
		for index, item := range value {
			result[index] = copyUnknown(item)
		}
		return result
	default:
		return value
	}
}
