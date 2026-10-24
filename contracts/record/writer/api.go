package writer

import "onboard-kit/internal/model"

type Writer interface {
	Write(plan model.Plan, record model.Record, intent Intent) (model.Record, []model.Problem)
}

type Operation string

const (
	OperationSubmit     Operation = "submit"
	OperationFollowPlan Operation = "follow-plan"
)

type Intent struct {
	Operation Operation
	Question  string
	Value     model.Value
}
