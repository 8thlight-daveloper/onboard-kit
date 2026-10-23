package scheduler

import "onboard-kit/internal/model"

type Scheduler interface {
	Schedule(plan model.Plan) ([]model.Section, []model.Problem)
}
