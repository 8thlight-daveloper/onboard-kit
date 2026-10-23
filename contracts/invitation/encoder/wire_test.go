package encoder_test

import (
	contract "onboard-kit/contracts/invitation/encoder"
	"onboard-kit/internal/model"
)

type stubEncoder struct{}

func (stubEncoder) Encode(model.InvitationFiles) ([]byte, []model.Problem) { return nil, nil }

func newEncoder() contract.Encoder { return stubEncoder{} }
