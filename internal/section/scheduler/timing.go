package scheduler

import "onboard-kit/internal/model"

func effectiveTimings(sections []model.Section) map[string]model.Timing {
	definitions := make(map[string]model.Section, len(sections))
	for _, section := range sections {
		definitions[section.ID] = section
	}
	timings := make(map[string]model.Timing, len(sections))
	for _, section := range sections {
		timingOf(section.ID, definitions, timings)
	}
	return timings
}

func timingOf(identity string, definitions map[string]model.Section, timings map[string]model.Timing) model.Timing {
	if timing, held := timings[identity]; held {
		return timing
	}
	section := definitions[identity]
	timing := section.Timing
	if timing == "" {
		if section.Parent == "" {
			timing = model.DefaultTiming
		} else {
			timing = timingOf(section.Parent, definitions, timings)
		}
	}
	timings[identity] = timing
	return timing
}
