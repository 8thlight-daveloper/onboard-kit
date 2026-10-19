package instantiator_test

import (
	contract "onboard-kit/contracts/plan/instantiator"
	implementation "onboard-kit/internal/plan/instantiator"
)

func newInstantiator() contract.Instantiator {
	return implementation.New()
}
