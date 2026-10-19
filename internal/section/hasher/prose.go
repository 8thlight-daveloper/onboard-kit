package hasher

import "strings"

func prose(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	lines := strings.Split(value, "\n")
	for index, line := range lines {
		lines[index] = strings.TrimRight(line, " \t")
	}
	return strings.Trim(strings.Join(lines, "\n"), " \t\n")
}

func (existing fields) withProse(key string, value string) fields {
	return existing.withString(key, prose(value))
}
