//go:build contractstub

package reviewer_test

import (
	contract "onboard-kit/contracts/session/reviewer"
	"onboard-kit/internal/model"
)

type emptyReviewer struct{}

func newReviewer() contract.Reviewer {
	return emptyReviewer{}
}

func (emptyReviewer) Inspect(contract.Source) (contract.Session, []contract.Diagnostic) {
	return contract.Session{}, nil
}

func (emptyReviewer) Propose(contract.Source, contract.Intent) (contract.Proposal, []contract.Diagnostic) {
	return contract.Proposal{}, nil
}

func (emptyReviewer) Finalize(contract.Proposal, model.Timestamp) (contract.Publication, []contract.Diagnostic) {
	return contract.Publication{}, nil
}
