package decoder_test

import (
	contract "onboard-kit/contracts/preparation/decoder"
	implementation "onboard-kit/internal/preparation/decoder"
)

func newDecoder() contract.Decoder {
	return implementation.New()
}
