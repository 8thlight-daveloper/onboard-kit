package decoder_test

import (
	contract "onboard-kit/contracts/record/decoder"
	implementation "onboard-kit/internal/record/decoder"
)

func newDecoder() contract.Decoder {
	return implementation.New()
}
