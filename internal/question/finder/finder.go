package finder

import "onboard-kit/internal/model"

type Finder struct{}

func New() Finder {
	return Finder{}
}

func (Finder) Questions(plan model.Plan) []model.Question {
	return findQuestions(plan)
}

func (Finder) Blockers(plan model.Plan) []model.Blocker {
	blockers := []model.Blocker{}
	for _, question := range findQuestions(plan) {
		if question.Required && !isAnswered(question.Answer) {
			blockers = append(blockers, model.Blocker{Question: question.ID})
		}
	}
	return blockers
}
