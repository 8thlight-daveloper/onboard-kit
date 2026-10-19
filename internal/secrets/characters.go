package secrets

func isUpper(character byte) bool {
	return character >= 'A' && character <= 'Z'
}

func isLower(character byte) bool {
	return character >= 'a' && character <= 'z'
}

func isDigit(character byte) bool {
	return character >= '0' && character <= '9'
}

func isAlphanumeric(character byte) bool {
	return isUpper(character) || isLower(character) || isDigit(character)
}

func isUpperOrDigit(character byte) bool {
	return isUpper(character) || isDigit(character)
}

func isAlphanumericOrHyphen(character byte) bool {
	return isAlphanumeric(character) || character == '-'
}

func isWordCharacter(character byte) bool {
	return isAlphanumeric(character) || character == '_'
}

func isURLSafe(character byte) bool {
	return isAlphanumeric(character) || character == '_' || character == '-'
}

func isRandomStringCharacter(character byte) bool {
	return isAlphanumeric(character) || character == '_' || character == '+' || character == '=' || character == '-'
}
