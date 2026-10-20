package refresher_test

import (
	"onboard-kit/internal/model"

	contract "onboard-kit/contracts/plan/refresher"
	implementation "onboard-kit/internal/plan/refresher"
)

type wiredRefresher struct {
	inner interface {
		Refresh(base, theirs, ours model.Plan, origin model.Origin) (model.Plan, implementation.Report, []model.Problem)
	}
}

func (w wiredRefresher) Refresh(base, theirs, ours model.Plan, origin model.Origin) (model.Plan, contract.Report, []model.Problem) {
	plan, report, problems := w.inner.Refresh(base, theirs, ours, origin)
	return plan, contract.Report(report), problems
}

func newRefresher() contract.Refresher {
	return wiredRefresher{inner: implementation.New()}
}
