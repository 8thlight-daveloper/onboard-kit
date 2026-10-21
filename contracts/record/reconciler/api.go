package reconciler

import "onboard-kit/internal/model"

type Reconciler interface {
	Reconcile(record model.Record, previous model.Plan, plan model.Plan, reference model.PlanReference) (model.Record, model.Reconciliation, []model.Problem)
}
