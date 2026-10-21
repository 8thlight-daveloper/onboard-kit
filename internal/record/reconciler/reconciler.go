package reconciler

import "onboard-kit/internal/model"

type Reconciler struct{}

func New() Reconciler {
	return Reconciler{}
}

func (Reconciler) Reconcile(record model.Record, previous model.Plan, plan model.Plan, reference model.PlanReference) (model.Record, model.Reconciliation, []model.Problem) {
	if problem, refused := refusal(record, plan); refused {
		return copiedRecord(record), model.Reconciliation{}, []model.Problem{problem}
	}
	questions := declaredQuestions(plan)
	answers, review := checkedAnswers(record.Answers, questions, plan.Directory)

	reconciled := copiedRecord(record)
	reconciled.Plan = model.PlanReference{ID: plan.ID, Hash: reference.Hash, URL: reference.URL}
	reconciled.Answers = answers

	return reconciled, reconciliationOf(record, previous, plan, questions, review), nil
}
