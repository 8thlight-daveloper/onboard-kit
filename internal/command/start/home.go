package start

import (
	"strings"
	"unicode"

	"onboard-kit/internal/command/places"
	"onboard-kit/internal/model"
)

func stateHome(given string) (string, error) {
	if given != "" {
		return given, nil
	}
	home, err := places.StateHome()
	return home.Path, err
}

func organizationName(plan model.Plan) string {
	answer := plan.Answers.Plan["organization"]
	if answer.Kind != model.ValueKindText {
		return "org"
	}
	if name := slug(answer.Text); name != "" {
		return name
	}
	return "org"
}

func slug(text string) string {
	var name strings.Builder
	pending := false
	for _, character := range strings.ToLower(text) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			if pending && name.Len() > 0 {
				name.WriteByte('-')
			}
			pending = false
			name.WriteRune(character)
			continue
		}
		pending = true
	}
	return name.String()
}
