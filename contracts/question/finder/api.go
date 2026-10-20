package finder

import "onboard-kit/internal/model"

type Finder interface {
	Questions(plan model.Plan, audience model.Audience) []model.Question
	Blockers(plan model.Plan, audience model.Audience) []model.Blocker
}
