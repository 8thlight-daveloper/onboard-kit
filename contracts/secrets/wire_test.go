package secrets_test

import (
	contract "onboard-kit/contracts/secrets"
	implementation "onboard-kit/internal/secrets"
)

func newChecker() contract.Checker {
	return implementation.New()
}
