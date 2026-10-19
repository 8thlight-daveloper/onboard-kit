package assembler

import (
	"strings"

	"onboard-kit/internal/model"
)

const continuationIndent = "  "

func stepItem(step model.Step, sectionTiming model.Timing, text prose) string {
	parts := []string{"- [ ] **" + text.of(step.Name) + ".**"}
	parts = append(parts, stepDetails(step, sectionTiming, text)...)
	return indentedContinuation(strings.Join(parts, wordSeparator))
}

func stepDetails(step model.Step, sectionTiming model.Timing, text prose) []string {
	var parts []string
	if text.present(step.Owner) {
		parts = append(parts, "Owner: "+text.of(step.Owner)+".")
	}
	if step.Timing != "" && effectiveTiming(step.Timing) != effectiveTiming(sectionTiming) {
		parts = append(parts, "When: "+timingInWords(step.Timing)+".")
	}
	if text.present(step.Notes) {
		parts = append(parts, text.of(step.Notes))
	}
	if text.present(step.Open) {
		parts = append(parts, "Open: "+text.of(step.Open)+".")
	}
	if step.Directory != "" {
		parts = append(parts, "Directory: `"+step.Directory+"`.")
	}
	if len(step.Run) > 0 {
		parts = append(parts, "Run: `"+strings.Join(step.Run, wordSeparator)+"`.")
	}
	if len(step.Env) > 0 {
		parts = append(parts, environmentPart(step.Env))
	}
	return append(parts, checkParts(step.Checks, text)...)
}

func indentedContinuation(item string) string {
	lines := strings.Split(item, lineBreak)
	for index := 1; index < len(lines); index++ {
		if lines[index] != "" {
			lines[index] = continuationIndent + lines[index]
		}
	}
	return strings.Join(lines, lineBreak)
}
