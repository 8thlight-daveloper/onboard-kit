package loader

import "onboard-kit/internal/model"

type Loader interface {
	Load(data []byte) (model.Section, []model.Problem)
}
