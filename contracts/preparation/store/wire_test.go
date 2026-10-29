package store_test

import (
	contract "onboard-kit/contracts/preparation/store"
	implementation "onboard-kit/internal/preparation/store"
)

func newStore() contract.Store { return implementation.New() }
