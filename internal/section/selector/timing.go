package selector

import "onboard-kit/internal/model"

var timingRanks = map[model.Timing]int{
	model.TimingBeforeDayOne: 0,
	model.TimingWeekOne:      1,
	model.TimingDays8To30:    2,
	model.TimingDays31To90:   3,
	model.TimingAnytime:      4,
}

func effectiveTiming(section model.Section, inherited model.Timing) model.Timing {
	if section.Timing == "" {
		return inherited
	}
	return section.Timing
}

func timingRank(timing model.Timing) int {
	rank, known := timingRanks[timing]
	if !known {
		return len(timingRanks)
	}
	return rank
}
