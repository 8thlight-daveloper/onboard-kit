//go:build !contractstub

package reviewer_test

import (
	contract "onboard-kit/contracts/session/reviewer"
	implementation "onboard-kit/internal/session/reviewer"
)

func newReviewer() contract.Reviewer {
	return implementation.New()
}
