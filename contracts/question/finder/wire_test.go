package finder_test

import (
	contract "onboard-kit/contracts/question/finder"
	implementation "onboard-kit/internal/question/finder"
)

func newFinder() contract.Finder {
	return implementation.New()
}
