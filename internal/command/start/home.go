package start

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"onboard-kit/internal/model"
)

func stateHome(given string) (string, error) {
	if given != "" {
		return given, nil
	}
	if fromEnvironment := os.Getenv("OBK_HOME"); fromEnvironment != "" {
		return fromEnvironment, nil
	}
	if runtime.GOOS != "darwin" {
		if state := os.Getenv("XDG_STATE_HOME"); state != "" {
			return filepath.Join(state, "obk"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "state", "obk"), nil
	}
	config, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(config, "obk"), nil
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
