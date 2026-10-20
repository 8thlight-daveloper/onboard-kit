package refresher

import "onboard-kit/internal/model"

type Refresher interface {
	Refresh(base, theirs, ours model.Plan, origin model.Origin) (model.Plan, Report, []model.Problem)
}

type Report struct {
	NewQuestions     []string
	RemovedQuestions []string
	AnswersTaken     []string
	AnswersKept      []string
	AnswersToReview  []string
	NewSections      []string
	RemovedSections  []string
}
