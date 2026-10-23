package reconciler

import (
	"maps"
	"slices"

	"onboard-kit/internal/model"
	"onboard-kit/internal/section/hasher"
)

func reconciliationOf(record model.Record, previous model.Plan, plan model.Plan, questions map[questionKey]model.Variable, review map[string]bool) model.Reconciliation {
	reconciliation := model.Reconciliation{
		AnswersToReview: sortedKeys(review),
		ChangedOutcomes: changedOutcomes(record.Outcomes, plan),
	}
	if hasNothingToCompare(previous) {
		return reconciliation
	}
	previousQuestions := declaredQuestions(previous)
	reconciliation.NewQuestions = sortedIDs(questions, previousQuestions)
	reconciliation.RemovedQuestions = sortedIDs(previousQuestions, questions)
	reconciliation.NewSections = sortedSectionIDs(plan, previous)
	reconciliation.RemovedSections = sortedSectionIDs(previous, plan)
	return reconciliation
}

func hasNothingToCompare(previous model.Plan) bool {
	return len(previous.Variables) == 0 && len(previous.Sections) == 0
}

func sortedKeys(set map[string]bool) []string {
	if len(set) == 0 {
		return nil
	}
	return slices.Sorted(maps.Keys(set))
}

func sortedIDs(including map[questionKey]model.Variable, excluding map[questionKey]model.Variable) []string {
	ids := map[string]bool{}
	for key := range including {
		if _, present := excluding[key]; !present {
			ids[key.id()] = true
		}
	}
	return sortedKeys(ids)
}

func sortedSectionIDs(including model.Plan, excluding model.Plan) []string {
	present := map[string]bool{}
	for _, section := range excluding.Sections {
		present[section.ID] = true
	}
	ids := map[string]bool{}
	for _, section := range including.Sections {
		if !present[section.ID] {
			ids[section.ID] = true
		}
	}
	return sortedKeys(ids)
}

func changedOutcomes(outcomes []model.Outcome, plan model.Plan) []string {
	hashes := map[string]string{}
	definitionHasher := hasher.New()
	for _, section := range plan.Sections {
		hashes[section.ID] = definitionHasher.Hash(section)
	}
	changed := map[string]bool{}
	for _, outcome := range outcomes {
		current, present := hashes[outcome.Section]
		if present && outcome.SectionHash != "" && outcome.SectionHash != current {
			changed[outcome.Section] = true
		}
	}
	return sortedKeys(changed)
}
