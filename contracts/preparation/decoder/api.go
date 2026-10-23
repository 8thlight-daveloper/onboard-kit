package decoder

import "onboard-kit/internal/model"

type Decoder interface {
	Decode(data []byte) (model.Preparation, []model.Problem)
}
