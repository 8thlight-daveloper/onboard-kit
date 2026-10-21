package splitter

import (
	"slices"

	"onboard-kit/internal/model"
)

func copyOfRecord(record model.Record) model.Record {
	copied := record
	copied.Answers = model.Answers{}
	copied.Outcomes = slices.Clone(record.Outcomes)
	copied.UnknownFields = copyOfFields(record.UnknownFields)
	return copied
}

func copyOfValue(value model.Value) model.Value {
	value.List = slices.Clone(value.List)
	return value
}

func copyOfFields(fields map[string]any) map[string]any {
	if fields == nil {
		return nil
	}
	copied := make(map[string]any, len(fields))
	for name, field := range fields {
		copied[name] = copyOfUnknown(field)
	}
	return copied
}

func copyOfUnknown(field any) any {
	switch typed := field.(type) {
	case map[string]any:
		return copyOfFields(typed)
	case []any:
		return copyOfItems(typed)
	}
	return field
}

func copyOfItems(items []any) []any {
	if items == nil {
		return nil
	}
	copied := make([]any, len(items))
	for index, item := range items {
		copied[index] = copyOfUnknown(item)
	}
	return copied
}
