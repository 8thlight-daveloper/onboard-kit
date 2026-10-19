package hasher

import "strings"

type fields []string

func (existing fields) text() string {
	return "{" + strings.Join(existing, ",") + "}"
}

func (existing fields) withRaw(key string, raw string, present bool) fields {
	if !present {
		return existing
	}
	return append(existing, quote(key)+":"+raw)
}

func (existing fields) withString(key string, value string) fields {
	return existing.withRaw(key, quote(value), value != "")
}

func (existing fields) withFlag(key string, value bool) fields {
	return existing.withRaw(key, "true", value)
}

func (existing fields) withStrings(key string, values []string) fields {
	return existing.withRaw(key, stringListText(values), len(values) > 0)
}

func listText(items []string) string {
	return "[" + strings.Join(items, ",") + "]"
}

func stringListText(values []string) string {
	items := make([]string, len(values))
	for index, value := range values {
		items[index] = quote(value)
	}
	return listText(items)
}
