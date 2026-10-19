package selector_test

import (
	contract "onboard-kit/contracts/section/selector"
	implementation "onboard-kit/internal/section/selector"
)

func newSelector() contract.Selector {
	return implementation.New()
}
