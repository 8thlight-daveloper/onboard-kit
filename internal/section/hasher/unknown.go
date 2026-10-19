package hasher

import (
	"sort"
	"strconv"
	"strings"
)

func unknownText(value any) string {
	switch typed := value.(type) {
	case string:
		return quote(typed)
	case bool:
		return strconv.FormatBool(typed)
	case int:
		return strconv.FormatInt(int64(typed), 10)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		return decimalText(typed)
	case []any:
		return anyListText(typed)
	case []string:
		return stringListText(typed)
	case map[string]any:
		return mapText(typed)
	default:
		return "null"
	}
}

func decimalText(number float64) string {
	text := strconv.FormatFloat(number, 'f', -1, 64)
	if !strings.Contains(text, ".") {
		text += ".0"
	}
	return text
}

func anyListText(values []any) string {
	items := make([]string, len(values))
	for index, value := range values {
		items[index] = unknownText(value)
	}
	return listText(items)
}

func mapText(entries map[string]any) string {
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	members := make(fields, 0, len(keys))
	for _, key := range keys {
		members = members.withRaw(key, unknownText(entries[key]), true)
	}
	return members.text()
}
