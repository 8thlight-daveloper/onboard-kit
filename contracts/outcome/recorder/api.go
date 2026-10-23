package recorder

import "onboard-kit/internal/model"

type Recorder interface {
	Record(plan model.Plan, record model.Record, proposal Proposal, on model.Timestamp) (model.Record, []model.Problem)
}

type Proposal struct {
	Section     string
	Kind        model.OutcomeKind
	Note        string
	Differently string
}
