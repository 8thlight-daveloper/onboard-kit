package joiner_test

import (
	contract "onboard-kit/contracts/record/joiner"
	implementation "onboard-kit/internal/record/joiner"
)

func newJoiner() contract.Joiner {
	return implementation.New()
}
