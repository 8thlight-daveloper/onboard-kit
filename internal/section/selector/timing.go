package selector

import "onboard-kit/internal/model"

var timingRanks = map[model.Timing]int{
	model.TimingBeforeDayOne: 0,
	model.TimingWeekOne:      1,
	model.TimingDays8To30:    2,
	model.TimingDays31To90:   3,
	model.TimingAnytime:      4,
}

func timingOf(section model.Section) model.Timing {
	if section.Timing == "" {
		return model.DefaultTiming
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
