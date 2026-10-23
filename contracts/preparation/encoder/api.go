package encoder

import "onboard-kit/internal/model"

type Encoder interface {
	Encode(model.Preparation) ([]byte, []model.Problem)
}
