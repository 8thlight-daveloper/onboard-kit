package resolver_test

import (
	contract "onboard-kit/contracts/section/resolver"
	implementation "onboard-kit/internal/section/resolver"
)

func newResolver() contract.Resolver {
	return implementation.New()
}
