package validator

import "onboard-kit/internal/model"

type Validator interface {
	Validate(plan model.Plan) []model.Problem
	CheckValue(plan model.Plan, question string, value model.Value) []model.Problem
}
