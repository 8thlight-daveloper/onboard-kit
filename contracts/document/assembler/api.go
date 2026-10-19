package assembler

import "onboard-kit/internal/model"

type Assembler interface {
	Assemble(plan model.Plan) (string, []model.Problem)
}
