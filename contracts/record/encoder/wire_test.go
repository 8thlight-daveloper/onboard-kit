package encoder_test

import (
	contract "onboard-kit/contracts/record/encoder"
	implementation "onboard-kit/internal/record/encoder"
)

func newEncoder() contract.Encoder {
	return implementation.New()
}
