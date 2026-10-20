package finder

import "onboard-kit/internal/model"

type Finder struct{}

func New() Finder {
	return Finder{}
}

func (Finder) Questions(plan model.Plan, audience model.Audience) []model.Question {
	return forAudience(findQuestions(plan), audience)
}

func (Finder) Blockers(plan model.Plan, audience model.Audience) []model.Blocker {
	blockers := []model.Blocker{}
	for _, question := range forAudience(findQuestions(plan), audience) {
		if question.Required && !isAnswered(question.Answer) {
			blockers = append(blockers, model.Blocker{Question: question.ID})
		}
	}
	return blockers
}
