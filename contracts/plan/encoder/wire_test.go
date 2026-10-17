package encoder_test

import (
	contract "onboard-kit/contracts/plan/encoder"
	implementation "onboard-kit/internal/plan/encoder"
)

func newEncoder() contract.Encoder {
	return implementation.New()
}
