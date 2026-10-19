package hasher

import (
	"strings"
)

const hexDigits = "0123456789abcdef"

func quote(value string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for _, character := range value {
		writeCharacter(&builder, character)
	}
	builder.WriteByte('"')
	return builder.String()
}

func writeCharacter(builder *strings.Builder, character rune) {
	switch {
	case character == '"':
		builder.WriteString(`\"`)
	case character == '\\':
		builder.WriteString(`\\`)
	case character == '\n':
		builder.WriteString(`\n`)
	case character == '\r':
		builder.WriteString(`\r`)
	case character == '\t':
		builder.WriteString(`\t`)
	case character < 0x20:
		builder.WriteString(`\u00`)
		builder.WriteByte(hexDigits[character>>4])
		builder.WriteByte(hexDigits[character&0xf])
	default:
		builder.WriteRune(character)
	}
}
