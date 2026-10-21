package joiner

import "onboard-kit/internal/model"

type Joiner struct{}

func New() Joiner {
	return Joiner{}
}

func (Joiner) Join(plan model.Plan, record model.Record) model.Plan {
	joined := copyPlan(plan)
	joined.Answers = layerAnswers(plan.Answers, record.Answers)
	return joined
}
