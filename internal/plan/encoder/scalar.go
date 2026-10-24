package encoder

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const deferredMarker = "deferred"

func textNode(text string) node {
	if chomp, lines, ok := literalBlock(text); ok {
		return literalNode{chomp: chomp, lines: lines}
	}
	return scalarNode(keyText(text))
}

func stringField(text string) node {
	if text == "" {
		return nil
	}
	return textNode(text)
}

func booleanNode(flag bool) node {
	return scalarNode(strconv.FormatBool(flag))
}

func flagField(flag bool) node {
	if !flag {
		return nil
	}
	return booleanNode(true)
}

func listField(items []string) node {
	return sequenceOf(items, func(_ int, item string) node {
		return textNode(item)
	})
}

func keyText(text string) string {
	if isPlain(text) {
		return text
	}
	return quoted(text)
}

func isPlain(text string) bool {
	return text != "" &&
		startsPlainly(text[0]) &&
		!strings.HasSuffix(text, " ") &&
		!strings.HasSuffix(text, ":") &&
		!strings.ContainsFunc(text, unicode.IsControl) &&
		!strings.ContainsAny(text, "\u2028\u2029") &&
		!strings.Contains(text, ": ") &&
		!strings.Contains(text, " #") &&
		!isReservedWord(text)
}

func startsPlainly(first byte) bool {
	switch {
	case first >= 'a' && first <= 'z', first >= 'A' && first <= 'Z':
		return true
	}
	return first == '$' || first == '/' || first == '<'
}

func isReservedWord(text string) bool {
	switch strings.ToLower(text) {
	case "y", "n", "yes", "no", "true", "false", "on", "off", "null", "inf", "nan", deferredMarker:
		return true
	}
	return false
}

func quoted(text string) string {
	var out strings.Builder
	out.WriteByte('"')
	for _, character := range text {
		out.WriteString(escaped(character))
	}
	out.WriteByte('"')
	return out.String()
}

func escaped(character rune) string {
	switch character {
	case '\\':
		return `\\`
	case '"':
		return `\"`
	case '\n':
		return `\n`
	case '\t':
		return `\t`
	case '\r':
		return `\r`
	case '\u2028':
		return `\u2028`
	case '\u2029':
		return `\u2029`
	}
	if unicode.IsControl(character) {
		return fmt.Sprintf(`\x%02X`, character)
	}
	return string(character)
}

func literalBlock(text string) (string, []string, bool) {
	if !strings.Contains(text, "\n") || strings.HasSuffix(text, "\n\n") || hasControlOtherThanNewline(text) || strings.ContainsAny(text, "\u2028\u2029") {
		return "", nil, false
	}
	chomp := "-"
	body := text
	if strings.HasSuffix(text, "\n") {
		chomp = ""
		body = strings.TrimSuffix(text, "\n")
	}
	lines := strings.Split(body, "\n")
	if lines[0] == "" || strings.HasPrefix(lines[0], " ") {
		return "", nil, false
	}
	for _, line := range lines {
		if line != strings.TrimRightFunc(line, unicode.IsSpace) {
			return "", nil, false
		}
	}
	return chomp, lines, true
}

func hasControlOtherThanNewline(text string) bool {
	return strings.ContainsFunc(text, func(character rune) bool {
		return character != '\n' && unicode.IsControl(character)
	})
}
