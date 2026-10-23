package encoder_test

import (
	contract "onboard-kit/contracts/preparation/encoder"
	implementation "onboard-kit/internal/preparation/encoder"
)

func newEncoder() contract.Encoder { return implementation.New() }
