package secrets

import "strings"

const (
	randomStringMinimumLength  = 32
	randomStringEdgeCharacters = "\"'`,;:.()[]{}<>"
)

func hasRandomString(text string) bool {
	for _, word := range strings.FieldsFunc(text, isSpace) {
		if looksRandom(strings.Trim(word, randomStringEdgeCharacters)) {
			return true
		}
	}
	return false
}

func looksRandom(word string) bool {
	if len(word) < randomStringMinimumLength {
		return false
	}
	var hasUpper, hasLower, hasDigit bool
	for index := 0; index < len(word); index++ {
		character := word[index]
		if !isRandomStringCharacter(character) {
			return false
		}
		hasUpper = hasUpper || isUpper(character)
		hasLower = hasLower || isLower(character)
		hasDigit = hasDigit || isDigit(character)
	}
	return hasUpper && hasLower && hasDigit
}
