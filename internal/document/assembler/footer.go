package assembler

import (
	"strings"

	"onboard-kit/internal/model"
)

func sectionFooter(section model.Section, text prose) string {
	var parts []string
	if text.present(section.Owner) {
		parts = append(parts, "Owner: "+text.of(section.Owner)+".")
	}
	if effectiveTiming(section.Timing) != model.TimingAnytime {
		parts = append(parts, "When: "+timingInWords(section.Timing)+".")
	}
	if escalation := escalationPart(section.Escalation); escalation != "" {
		parts = append(parts, escalation)
	}
	return strings.Join(parts, wordSeparator)
}
