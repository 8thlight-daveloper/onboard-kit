package reviewer_test

import (
	contract "onboard-kit/contracts/preparation/reviewer"
	"onboard-kit/internal/model"
)

type stubReviewer struct{}

func (stubReviewer) Review(contract.Context, model.Plan, model.Preparation) contract.Review {
	return contract.Review{}
}

func newReviewer() contract.Reviewer { return stubReviewer{} }
