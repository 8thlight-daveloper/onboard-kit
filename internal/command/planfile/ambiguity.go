package planfile

import (
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/validator"
)

func AssemblyProblems(plan model.Plan) []model.Problem {
	var problems []model.Problem
	for _, problem := range validator.New().Validate(plan) {
		if isAmbiguity(problem.Code) {
			problems = append(problems, problem)
			continue
		}
		switch problem.Code {
		case "plan.parent-unknown", "plan.parent-self", "plan.parent-cycle",
			"plan.requires-unknown", "plan.requires-self", "plan.requires-cycle",
			"plan.bootstrap-requires-nonbootstrap", "plan.after-unknown", "plan.after-self",
			"plan.after-cycle", "plan.bootstrap-after-nonbootstrap", "plan.ordering-conflict":
			problems = append(problems, problem)
		}
	}
	return problems
}

func OnlyAmbiguities(problems []model.Problem) bool {
	for _, problem := range problems {
		if !isAmbiguity(problem.Code) {
			return false
		}
	}
	return true
}

func isAmbiguity(code string) bool {
	switch code {
	case "plan.duplicate-variable-name", "plan.duplicate-step-id", "plan.duplicate-section-id", "plan.duplicate-team-variable", "directory.duplicate-id":
		return true
	}
	return false
}
