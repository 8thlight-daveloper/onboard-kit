package assembler

import "strings"

func flatten(value string) string {
	return strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ", "\u2028", " ", "\u2029", " ").Replace(value)
}

func displayText(value string) string {
	var result strings.Builder
	for _, character := range flatten(value) {
		if strings.ContainsRune("\\`*_{}[]()<>#+-.!|", character) {
			result.WriteByte('\\')
		}
		result.WriteRune(character)
	}
	return result.String()
}

func displayCode(value string) string {
	value = flatten(value)
	longest, current := 0, 0
	for _, character := range value {
		if character == '`' {
			current++
			if current > longest {
				longest = current
			}
		} else {
			current = 0
		}
	}
	fence := strings.Repeat("`", longest+1)
	padding := ""
	if strings.HasPrefix(value, "`") || strings.HasSuffix(value, "`") || strings.HasPrefix(value, " ") || strings.HasSuffix(value, " ") {
		padding = " "
	}
	return fence + padding + value + padding + fence
}
