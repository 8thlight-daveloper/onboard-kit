package invite_test

import (
	contract "onboard-kit/contracts/command/invite"
	implementation "onboard-kit/internal/command/invite"
)

func newCommand() contract.Command { return implementation.New() }
