package validator_test

import (
	contract "onboard-kit/contracts/preparation/validator"
	"onboard-kit/internal/model"
)

type unimplemented struct{}

func (unimplemented) Validate(contract.Context, model.Preparation) []model.Problem {
	return nil
}

func newValidator() contract.Validator {
	return unimplemented{}
}
