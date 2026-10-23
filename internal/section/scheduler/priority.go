package scheduler

import "onboard-kit/internal/model"

func precedes(left, right activity) bool {
	if left.section.Bootstrap != right.section.Bootstrap {
		return left.section.Bootstrap
	}
	leftTiming, rightTiming := timingRank(left.effectiveTime), timingRank(right.effectiveTime)
	if leftTiming != rightTiming {
		return leftTiming < rightTiming
	}
	if left.referenceRank != right.referenceRank {
		return left.referenceRank < right.referenceRank
	}
	return left.section.ID < right.section.ID
}

func timingRank(timing model.Timing) int {
	switch timing {
	case model.TimingBeforeDayOne:
		return 0
	case model.TimingWeekOne:
		return 1
	case model.TimingDays8To30:
		return 2
	case model.TimingDays31To90:
		return 3
	default:
		return 4
	}
}
