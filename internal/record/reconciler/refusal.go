package reconciler

import "onboard-kit/internal/model"

const (
	codeMissingPlanID = "reconcile.missing-plan-id"
	codeDifferentPlan = "reconcile.different-plan"
)

func refusal(record model.Record, plan model.Plan) (model.Problem, bool) {
	if plan.ID == "" {
		return model.Problem{
			Code:     codeMissingPlanID,
			Severity: model.SeverityError,
			Message:  "The new plan has no id, so a record cannot be reconciled with it.",
		}, true
	}
	if record.Plan.ID != "" && record.Plan.ID != plan.ID {
		return model.Problem{
			Code:     codeDifferentPlan,
			Severity: model.SeverityError,
			Message:  "This record belongs to another plan.",
			Detail:   map[string]string{"expected": record.Plan.ID, "found": plan.ID},
		}, true
	}
	return model.Problem{}, false
}
