package encoder

import "onboard-kit/internal/model"

type Encoder interface {
	Encode(record model.Record) ([]byte, []model.Problem)
}
