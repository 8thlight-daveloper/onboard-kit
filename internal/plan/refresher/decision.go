package refresher

import "onboard-kit/internal/model"

type listing int

const (
	listingNone listing = iota
	listingTaken
	listingKept
	listingReview
)

type decision struct {
	question question
	answer   model.Value
	listing  listing
}

func decideAll(base, theirs, ours model.Plan, questions map[string]question) []decision {
	decisions := make([]decision, 0, len(questions))
	for _, asked := range questions {
		decisions = append(decisions, decide(asked, answerOf(base.Answers, asked), answerOf(theirs.Answers, asked), answerOf(ours.Answers, asked)))
	}
	return decisions
}

func decide(asked question, base, theirs, ours model.Value) decision {
	if asked.isForOnboardee() && !sameValue(ours, base) {
		return decideChangedNewcomerAnswer(asked, ours)
	}
	return decision{question: asked, answer: theirs, listing: takenIfDifferent(theirs, ours)}
}

func decideChangedNewcomerAnswer(asked question, ours model.Value) decision {
	switch {
	case isAbsent(ours):
		return decision{question: asked, answer: ours}
	case isValidAnswer(ours, asked.variable):
		return decision{question: asked, answer: ours, listing: listingKept}
	default:
		return decision{question: asked, listing: listingReview}
	}
}

func takenIfDifferent(theirs, ours model.Value) listing {
	if sameValue(theirs, ours) {
		return listingNone
	}
	return listingTaken
}
