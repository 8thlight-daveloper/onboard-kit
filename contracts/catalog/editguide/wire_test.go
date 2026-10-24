package editguide_test

import (
	contract "onboard-kit/contracts/catalog/editguide"
	"onboard-kit/internal/catalog/editguide"
)

func newCommand() contract.Command { return editguide.New() }
