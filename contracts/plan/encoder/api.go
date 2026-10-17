package encoder

import "onboard-kit/internal/model"

type Encoder interface {
	Encode(plan model.Plan) ([]byte, []model.Problem)
}
