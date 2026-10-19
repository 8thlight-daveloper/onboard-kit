package secrets

import "unicode"

func isSpace(character rune) bool {
	return unicode.IsSpace(character)
}
