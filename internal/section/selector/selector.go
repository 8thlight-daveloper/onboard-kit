package selector

import "onboard-kit/internal/model"

type Selector struct{}

func New() Selector {
	return Selector{}
}

func (Selector) Select(plan model.Plan) []model.Section {
	active := activeSections(plan)
	return arrangeTree(active)
}
