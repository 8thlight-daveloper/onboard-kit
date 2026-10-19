package secrets

import "onboard-kit/internal/model"

type Checker interface {
	Check(value model.Value) []model.Problem
}
