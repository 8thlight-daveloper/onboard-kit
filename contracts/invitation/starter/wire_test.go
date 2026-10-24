package starter_test

import (
	contract "onboard-kit/contracts/invitation/starter"
	implementation "onboard-kit/internal/invitation/starter"
)

func newStarter(facts contract.Runtime) contract.Command { return implementation.New(facts) }
