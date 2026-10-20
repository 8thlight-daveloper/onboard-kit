package refresher

import (
	"maps"
	"slices"

	"onboard-kit/internal/model"
)

type Report struct {
	NewQuestions     []string
	RemovedQuestions []string
	AnswersTaken     []string
	AnswersKept      []string
	AnswersToReview  []string
	NewSections      []string
	RemovedSections  []string
}

func reportOf(theirs, ours model.Plan, theirQuestions, ourQuestions map[string]question, decisions []decision) Report {
	return Report{
		NewQuestions:     sortedKeysMissingFrom(theirQuestions, ourQuestions),
		RemovedQuestions: sortedKeysMissingFrom(ourQuestions, theirQuestions),
		AnswersTaken:     questionsListed(decisions, listingTaken),
		AnswersKept:      questionsListed(decisions, listingKept),
		AnswersToReview:  questionsListed(decisions, listingReview),
		NewSections:      sectionsMissingFrom(theirs, ours),
		RemovedSections:  sectionsMissingFrom(ours, theirs),
	}
}

func sortedKeysMissingFrom(wanted, present map[string]question) []string {
	var missing []string
	for id := range wanted {
		if _, found := present[id]; !found {
			missing = append(missing, id)
		}
	}
	slices.Sort(missing)
	return missing
}

func questionsListed(decisions []decision, wanted listing) []string {
	var ids []string
	for _, made := range decisions {
		if made.listing == wanted {
			ids = append(ids, made.question.id())
		}
	}
	slices.Sort(ids)
	return ids
}

func sectionsMissingFrom(wanted, present model.Plan) []string {
	presentIDs := map[string]bool{}
	for _, section := range present.Sections {
		presentIDs[section.ID] = true
	}
	missing := map[string]bool{}
	for _, section := range wanted.Sections {
		if !presentIDs[section.ID] {
			missing[section.ID] = true
		}
	}
	return slices.Sorted(maps.Keys(missing))
}
