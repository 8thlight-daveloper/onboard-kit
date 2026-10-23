package validator_test

import (
	contract "onboard-kit/contracts/preparation/validator"
	implementation "onboard-kit/internal/preparation/validator"
)

func newValidator() contract.Validator {
	return implementation.New()
}
