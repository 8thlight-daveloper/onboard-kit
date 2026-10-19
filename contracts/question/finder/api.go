package finder

import "onboard-kit/internal/model"

type Finder interface {
	Questions(plan model.Plan) []model.Question
	Blockers(plan model.Plan) []model.Blocker
}
