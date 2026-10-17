package decoder_test

import (
	contract "onboard-kit/contracts/plan/decoder"
	implementation "onboard-kit/internal/plan/decoder"
)

func newDecoder() contract.Decoder {
	return implementation.New()
}
