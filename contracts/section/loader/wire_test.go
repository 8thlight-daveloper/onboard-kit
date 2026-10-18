package loader_test

import (
	contract "onboard-kit/contracts/section/loader"
	implementation "onboard-kit/internal/section/loader"
)

func newLoader() contract.Loader {
	return implementation.New()
}
