package validator_test

import (
	contract "onboard-kit/contracts/plan/validator"
	implementation "onboard-kit/internal/plan/validator"
)

func newValidator() contract.Validator {
	return implementation.New()
}
