package creator_test

import (
	contract "onboard-kit/contracts/catalog/creator"
	implementation "onboard-kit/internal/catalog/creator"
)

func newCommand() contract.Command { return implementation.New() }
