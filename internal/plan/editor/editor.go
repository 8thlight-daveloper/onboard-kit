package editor

import (
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/validator"
)

type Editor struct{}

func New() Editor {
	return Editor{}
}

func (Editor) Submit(plan model.Plan, question string, value model.Value) (model.Plan, []model.Problem) {
	problems := validator.New().CheckValue(plan, question, value)
	if hasError(problems) {
		return copyEditableParts(plan), problems
	}
	return applyAnswer(plan, question, value), problems
}

func hasError(problems []model.Problem) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}
