package decoder_test

import (
	contract "onboard-kit/contracts/preparation/decoder"
	"onboard-kit/internal/model"
)

type stubDecoder struct{}

func (stubDecoder) Decode([]byte) (model.Preparation, []model.Problem) {
	return model.Preparation{}, nil
}

func newDecoder() contract.Decoder {
	return stubDecoder{}
}
