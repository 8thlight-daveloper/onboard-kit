package structurevalidator_test

import (
	contract "onboard-kit/contracts/plan/structurevalidator"
	implementation "onboard-kit/internal/plan/structurevalidator"
)

func newValidator() contract.Validator {
	return implementation.New()
}
