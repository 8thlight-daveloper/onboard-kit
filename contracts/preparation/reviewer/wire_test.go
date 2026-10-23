package reviewer_test

import (
	contract "onboard-kit/contracts/preparation/reviewer"
	implementation "onboard-kit/internal/preparation/reviewer"
)

func newReviewer() contract.Reviewer { return implementation.New() }
