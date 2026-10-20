package selector

import (
	"slices"
	"strings"

	"onboard-kit/internal/model"
)

func orderSiblings(siblings []model.Section, inherited model.Timing) []model.Section {
	ordered := make([]model.Section, 0, len(siblings))
	for _, group := range groupByTiming(siblings, inherited) {
		ordered = append(ordered, orderWithinTiming(group)...)
	}
	return ordered
}

func groupByTiming(sections []model.Section, inherited model.Timing) [][]model.Section {
	byTiming := map[model.Timing][]model.Section{}
	for _, section := range sections {
		timing := effectiveTiming(section, inherited)
		byTiming[timing] = append(byTiming[timing], section)
	}
	timings := make([]model.Timing, 0, len(byTiming))
	for timing := range byTiming {
		timings = append(timings, timing)
	}
	slices.SortFunc(timings, compareTimings)
	groups := make([][]model.Section, 0, len(timings))
	for _, timing := range timings {
		groups = append(groups, byTiming[timing])
	}
	return groups
}

func compareTimings(left model.Timing, right model.Timing) int {
	if rankOrder := timingRank(left) - timingRank(right); rankOrder != 0 {
		return rankOrder
	}
	return strings.Compare(string(left), string(right))
}

func orderWithinTiming(sections []model.Section) []model.Section {
	remaining := slices.Clone(sections)
	slices.SortStableFunc(remaining, func(left model.Section, right model.Section) int {
		return strings.Compare(left.ID, right.ID)
	})
	ordered := make([]model.Section, 0, len(remaining))
	for len(remaining) > 0 {
		next := indexOfNext(remaining)
		ordered = append(ordered, remaining[next])
		remaining = slices.Delete(remaining, next, next+1)
	}
	return ordered
}

func indexOfNext(remaining []model.Section) int {
	pending := pendingIDs(remaining)
	for index, section := range remaining {
		if isFree(section, pending) {
			return index
		}
	}
	return 0
}

func pendingIDs(remaining []model.Section) map[string]bool {
	pending := make(map[string]bool, len(remaining))
	for _, section := range remaining {
		pending[section.ID] = true
	}
	return pending
}

func isFree(section model.Section, pending map[string]bool) bool {
	for _, id := range section.After {
		if id != section.ID && pending[id] {
			return false
		}
	}
	return true
}
