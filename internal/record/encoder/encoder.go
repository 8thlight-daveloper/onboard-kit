package encoder

import (
	"strings"

	"onboard-kit/internal/model"
)

type Encoder struct{}

func New() Encoder {
	return Encoder{}
}

func (Encoder) Encode(record model.Record) ([]byte, []model.Problem) {
	found := &collector{}
	document := found.recordNode(record)
	if len(found.problems) > 0 {
		return nil, found.problems
	}
	return []byte(strings.Join(render(document, 0), "\n") + "\n"), nil
}
