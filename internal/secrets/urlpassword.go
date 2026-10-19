package secrets

import "regexp"

var (
	urlWithPassword   = regexp.MustCompile(`[A-Za-z][A-Za-z0-9+.-]*://[^\s/?#@:]*:([^\s/?#@]+)@`)
	variableReference = regexp.MustCompile(`^\$[a-z][a-z0-9_]*$`)
)

func hasURLPassword(text string) bool {
	for _, match := range urlWithPassword.FindAllStringSubmatch(text, -1) {
		if !variableReference.MatchString(match[1]) {
			return true
		}
	}
	return false
}
