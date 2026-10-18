package hasher

import "onboard-kit/internal/model"

type Hasher interface {
	Hash(section model.Section) string
}
