package authoringvalidator_test

import (
	contract "onboard-kit/contracts/plan/authoringvalidator"
	"onboard-kit/internal/plan/authoringvalidator"
)

func newValidator() contract.Validator {
	return authoringvalidator.New()
}
