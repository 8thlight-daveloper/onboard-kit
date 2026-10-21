package splitter_test

import (
	contract "onboard-kit/contracts/record/splitter"
	implementation "onboard-kit/internal/record/splitter"
)

func newSplitter() contract.Splitter {
	return implementation.New()
}
