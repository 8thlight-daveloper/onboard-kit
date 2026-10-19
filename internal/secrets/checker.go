package secrets

import "onboard-kit/internal/model"

type Checker struct{}

func New() Checker {
	return Checker{}
}

func (Checker) Check(value model.Value) []model.Problem {
	switch value.Kind {
	case model.ValueKindText:
		return textProblems(value.Text, nil)
	case model.ValueKindList:
		return listProblems(value.List)
	default:
		return nil
	}
}
