package assembler_test

import (
	contract "onboard-kit/contracts/document/assembler"
	implementation "onboard-kit/internal/document/assembler"
)

func newAssembler() contract.Assembler {
	return implementation.New()
}
