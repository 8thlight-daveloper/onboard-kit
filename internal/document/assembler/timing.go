package assembler

import (
	"strings"

	"onboard-kit/internal/model"
)

func effectiveTiming(timing model.Timing) model.Timing {
	if timing == "" {
		return model.DefaultTiming
	}
	return timing
}

func timingInWords(timing model.Timing) string {
	return strings.ReplaceAll(string(effectiveTiming(timing)), "-", " ")
}

func capitalized(text string) string {
	return strings.ToUpper(text[:1]) + text[1:]
}
