package resolver

func nameEnd(text string, start int) int {
	if start >= len(text) || !isLowercaseLetter(text[start]) {
		return start
	}
	end := start + 1
	for end < len(text) && isNameCharacter(text[end]) {
		end++
	}
	return end
}

func isLowercaseLetter(character byte) bool {
	return character >= 'a' && character <= 'z'
}

func isNameCharacter(character byte) bool {
	return isLowercaseLetter(character) || (character >= '0' && character <= '9') || character == '_'
}

func exactName(text string) (string, bool) {
	if len(text) < 2 || text[0] != '$' {
		return "", false
	}
	end := nameEnd(text, 1)
	if end == 1 || end != len(text) {
		return "", false
	}
	return text[1:], true
}
