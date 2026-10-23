package guide_test

import (
	contract "onboard-kit/contracts/preparation/guide"
	implementation "onboard-kit/internal/preparation/guide"
)

func newCommand() contract.Command { return implementation.New() }
