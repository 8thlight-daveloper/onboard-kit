package scheduler_test

import (
	contract "onboard-kit/contracts/section/scheduler"
	"onboard-kit/internal/model"
)

type draftScheduler struct{}

func (draftScheduler) Schedule(model.Plan) ([]model.Section, []model.Problem) {
	return nil, nil
}

func newScheduler() contract.Scheduler {
	return draftScheduler{}
}
