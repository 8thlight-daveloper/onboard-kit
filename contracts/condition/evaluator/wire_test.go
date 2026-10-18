package evaluator_test

import (
	contract "onboard-kit/contracts/condition/evaluator"
	implementation "onboard-kit/internal/condition/evaluator"
)

func newEvaluator() contract.Evaluator {
	return implementation.New()
}
