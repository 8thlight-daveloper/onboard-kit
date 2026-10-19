package hasher

import (
	"sort"

	"onboard-kit/internal/model"
)

func valueText(value model.Value) string {
	switch value.Kind {
	case model.ValueKindText:
		return quote(value.Text)
	case model.ValueKindFlag:
		if value.Flag {
			return "true"
		}
		return "false"
	case model.ValueKindList:
		return stringListText(value.List)
	case model.ValueKindDeferred:
		return `{"deferred":true}`
	default:
		return "null"
	}
}

func hasValue(value model.Value) bool {
	return value.Kind != model.ValueKindNone
}

func valueSetText(values []model.Value) string {
	items := make([]string, len(values))
	for index, value := range values {
		items[index] = valueText(value)
	}
	return setText(items)
}

func setText(items []string) string {
	sorted := append([]string(nil), items...)
	sort.Strings(sorted)
	unique := sorted[:0]
	for index, item := range sorted {
		if index == 0 || item != sorted[index-1] {
			unique = append(unique, item)
		}
	}
	return listText(unique)
}

func quotedSetText(values []string) string {
	items := make([]string, len(values))
	for index, value := range values {
		items[index] = quote(value)
	}
	return setText(items)
}
