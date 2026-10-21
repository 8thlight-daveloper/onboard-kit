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
		decisions = append(decisions, decide(asked, theirs.Directory, answerOf(base.Answers, asked), answerOf(theirs.Answers, asked), answerOf(ours.Answers, asked)))
	}
	return decisions
}

func decide(asked question, directory model.Directory, base, theirs, ours model.Value) decision {
	if asked.isForOnboardee() && !sameValue(ours, base) {
		return decideChangedNewcomerAnswer(asked, directory, ours)
	}
	return decision{question: asked, answer: theirs, listing: takenIfDifferent(theirs, ours)}
}

func decideChangedNewcomerAnswer(asked question, directory model.Directory, ours model.Value) decision {
	if isAbsent(ours) {
		return decision{question: asked, answer: ours}
	}
	checked, outcome := checkedAnswer(ours, asked.variable, directory)
	return decision{question: asked, answer: checked, listing: outcome}
}

func takenIfDifferent(theirs, ours model.Value) listing {
	if sameValue(theirs, ours) {
		return listingNone
	}
	return listingTaken
}
