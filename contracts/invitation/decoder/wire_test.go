package decoder_test

import (
	contract "onboard-kit/contracts/invitation/decoder"
	"onboard-kit/internal/model"
)

type stubDecoder struct{}

func (stubDecoder) Decode([]byte) (model.InvitationFiles, []model.Problem) {
	return model.InvitationFiles{}, nil
}

func newDecoder() contract.Decoder { return stubDecoder{} }
