package encoder_test

import (
	contract "onboard-kit/contracts/preparation/encoder"
	"onboard-kit/internal/model"
)

type stub struct{}

func newEncoder() contract.Encoder { return stub{} }

func (stub) Encode(model.Preparation) ([]byte, []model.Problem) { return nil, nil }
