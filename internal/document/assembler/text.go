package assembler

import (
	"strings"
	"unicode"

	"onboard-kit/internal/model"
)

func trimmed(text string) string {
	return strings.TrimRightFunc(text, unicode.IsSpace)
}

func isBlank(text string) bool {
	return strings.TrimSpace(text) == ""
}

func onboardingBlanks(variables []model.Variable) []string {
	var blanks []string
	seen := map[string]bool{}
	for _, variable := range variables {
		if variable.Phase != model.PhaseOnboarding || seen[variable.Name] {
			continue
		}
		seen[variable.Name] = true
		blanks = append(blanks, "<"+variable.Name+">")
	}
	return blanks
}

func withBlanksQuoted(text string, blanks []string) string {
	for _, blank := range blanks {
		text = strings.ReplaceAll(text, blank, "`"+blank+"`")
	}
	return text
}

type prose struct {
	blanks []string
}

func (p prose) of(text string) string {
	return withBlanksQuoted(trimmed(text), p.blanks)
}

func (p prose) present(text string) bool {
	return !isBlank(text)
}
