package decoder_test

import (
	contract "onboard-kit/contracts/invitation/decoder"
	implementation "onboard-kit/internal/invitation/decoder"
)

func newDecoder() contract.Decoder { return implementation.New() }
