package assembler_test

import (
	contract "onboard-kit/contracts/firststeps/assembler"
	implementation "onboard-kit/internal/firststeps/assembler"
)

func newAssembler() contract.Assembler { return implementation.New() }
