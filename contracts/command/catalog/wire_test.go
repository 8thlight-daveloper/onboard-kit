package catalog_test

import (
	contract "onboard-kit/contracts/command/catalog"
	implementation "onboard-kit/internal/command/catalog"
)

func newCommand() contract.Command { return implementation.New() }
