package assembler_test

import (
	contract "onboard-kit/contracts/firststeps/assembler"
	"onboard-kit/internal/model"
)

type emptyAssembler struct{}

func (emptyAssembler) Assemble(contract.Context, model.Plan, model.Preparation) (string, []model.Problem) {
	return "", nil
}

func newAssembler() contract.Assembler { return emptyAssembler{} }
