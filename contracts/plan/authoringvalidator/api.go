package authoringvalidator

import "onboard-kit/internal/model"

type Validator interface {
	Validate(plan model.Plan) []model.Problem
}
