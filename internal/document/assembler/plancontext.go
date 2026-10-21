package assembler

import "onboard-kit/internal/model"

type planContext struct {
	answers   model.Answers
	directory model.Directory
	team      string
}
