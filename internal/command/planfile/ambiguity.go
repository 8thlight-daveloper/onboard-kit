package planfile

import (
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/validator"
)

func Ambiguities(plan model.Plan) []model.Problem {
	var problems []model.Problem
	for _, problem := range validator.New().Validate(plan) {
		switch problem.Code {
		case "plan.duplicate-variable-name", "plan.duplicate-step-id", "plan.duplicate-section-id", "plan.duplicate-team-variable", "directory.duplicate-id":
			problems = append(problems, problem)
		}
	}
	return problems
}
