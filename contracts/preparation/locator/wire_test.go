package locator_test

import (
	contract "onboard-kit/contracts/preparation/locator"
	implementation "onboard-kit/internal/preparation/locator"
)

func newLocator() contract.Locator { return implementation.New() }
