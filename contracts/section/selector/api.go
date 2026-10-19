package selector

import "onboard-kit/internal/model"

type Selector interface {
	Select(plan model.Plan) []model.Section
}
