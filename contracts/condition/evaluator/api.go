package evaluator

import "onboard-kit/internal/model"

type Evaluator interface {
	Holds(condition model.Condition, scope string, answers model.Answers, switches model.Switches) bool
}
