package prepare_test

import (
	contract "onboard-kit/contracts/command/prepare"
	implementation "onboard-kit/internal/command/prepare"
)

func newCommand() contract.Command { return implementation.New() }
