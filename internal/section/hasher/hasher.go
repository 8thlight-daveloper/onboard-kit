package hasher

import (
	"crypto/sha256"
	"encoding/hex"

	"onboard-kit/internal/model"
)

type Hasher struct{}

func New() Hasher {
	return Hasher{}
}

func (Hasher) Hash(section model.Section) string {
	sum := sha256.Sum256([]byte(sectionText(section)))
	return hex.EncodeToString(sum[:])
}
