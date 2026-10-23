package walker

import (
	"sort"

	"onboard-kit/internal/model"
)

func duplicateOutcomes(outcomes []model.Outcome) []model.Problem {
	counts := make(map[string]int, len(outcomes))
	for _, outcome := range outcomes {
		counts[outcome.Section]++
	}
	var problems []model.Problem
	for section, count := range counts {
		if count > 1 {
			problems = append(problems, model.Problem{
				Code:     "walk.duplicate-outcome",
				Severity: model.SeverityError,
				Section:  section,
				Message:  "Keep one recorded outcome for this section.",
				Detail:   map[string]string{"field": "outcomes"},
			})
		}
	}
	sort.Slice(problems, func(i, j int) bool { return problems[i].Section < problems[j].Section })
	return problems
}

func waitingOn(requires []string, outcomes map[string]model.Outcome) []string {
	seen := make(map[string]bool, len(requires))
	var waiting []string
	for _, section := range requires {
		if !seen[section] && outcomes[section].Kind != model.OutcomeVerified {
			waiting = append(waiting, section)
		}
		seen[section] = true
	}
	return waiting
}

func outcomeViews(outcomes []model.Outcome, hashes map[string]string) []OutcomeView {
	views := make([]OutcomeView, 0, len(outcomes))
	for _, outcome := range outcomes {
		current, present := hashes[outcome.Section]
		comparison := ComparisonUnknown
		if present && outcome.SectionHash != "" {
			comparison = ComparisonDifferent
			if outcome.SectionHash == current {
				comparison = ComparisonSame
			}
		}
		views = append(views, OutcomeView{Outcome: outcome, CurrentHash: current, Comparison: comparison})
	}
	sort.Slice(views, func(i, j int) bool { return views[i].Outcome.Section < views[j].Outcome.Section })
	return views
}
