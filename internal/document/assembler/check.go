package assembler

import (
	"strings"

	"onboard-kit/internal/model"
)

func checkParts(checks []model.Check, text prose) []string {
	var parts []string
	for _, check := range checks {
		if len(check.Run) > 0 {
			parts = append(parts, "Check: `"+strings.Join(check.Run, wordSeparator)+"` exits with code 0.")
		}
		if text.present(check.Ask) {
			parts = append(parts, "Check: "+text.of(check.Ask))
		}
		if check.Env != "" {
			parts = append(parts, "Check: `"+check.Env+"` is set and not empty.")
		}
	}
	return parts
}
