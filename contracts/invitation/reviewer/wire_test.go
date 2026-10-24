package reviewer_test

import (
	contract "onboard-kit/contracts/invitation/reviewer"
	implementation "onboard-kit/internal/invitation/reviewer"
)

func newReviewer() contract.Reviewer { return implementation.New() }
