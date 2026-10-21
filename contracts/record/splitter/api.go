package splitter

import "onboard-kit/internal/model"

type Splitter interface {
	Split(plan model.Plan, joined model.Plan, previous model.Record) model.Record
}
