package scheduler

import (
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/structurevalidator"
	"onboard-kit/internal/section/selector"
)

type Scheduler struct{}

func New() Scheduler {
	return Scheduler{}
}

func (Scheduler) Schedule(plan model.Plan) ([]model.Section, []model.Problem) {
	if problems := structurevalidator.New().Validate(plan); len(problems) != 0 {
		return nil, problems
	}
	selected := selector.New().Select(plan)
	timings := effectiveTimings(plan.Sections)
	graph := graphOf(selected, timings)
	return graph.orderedSections(), nil
}
