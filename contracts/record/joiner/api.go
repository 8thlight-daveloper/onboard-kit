package joiner

import "onboard-kit/internal/model"

type Joiner interface {
	Join(plan model.Plan, record model.Record) model.Plan
}
