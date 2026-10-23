package encoder_test

import (
	contract "onboard-kit/contracts/invitation/encoder"
	implementation "onboard-kit/internal/invitation/encoder"
)

func newEncoder() contract.Encoder { return implementation.New() }
