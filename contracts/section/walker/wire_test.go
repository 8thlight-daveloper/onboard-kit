package walker_test

import (
	contract "onboard-kit/contracts/section/walker"
	implementation "onboard-kit/internal/section/walker"
)

func newWalker() contract.Walker {
	return implementation.New()
}
