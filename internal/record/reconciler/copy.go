package reconciler

import (
	"slices"

	"onboard-kit/internal/model"
)

func copiedRecord(record model.Record) model.Record {
	copied := record
	copied.Answers = copiedAnswers(record.Answers)
	copied.Outcomes = slices.Clone(record.Outcomes)
	copied.UnknownFields = copiedMap(record.UnknownFields)
	return copied
}

func copiedAnswers(answers model.Answers) model.Answers {
	copied := model.Answers{}
	if answers.Plan != nil {
		copied.Plan = copiedValues(answers.Plan)
	}
	if answers.Sections != nil {
		copied.Sections = make(map[string]map[string]model.Value, len(answers.Sections))
		for scope, values := range answers.Sections {
			copied.Sections[scope] = copiedValues(values)
		}
	}
	return copied
}

func copiedMap(fields map[string]any) map[string]any {
	if fields == nil {
		return nil
	}
	copied := make(map[string]any, len(fields))
	for name, field := range fields {
		copied[name] = copiedField(field)
	}
	return copied
}

func copiedField(field any) any {
	switch typed := field.(type) {
	case map[string]any:
		return copiedMap(typed)
	case []any:
		copied := make([]any, len(typed))
		for index, item := range typed {
			copied[index] = copiedField(item)
		}
		return copied
	default:
		return field
	}
}
