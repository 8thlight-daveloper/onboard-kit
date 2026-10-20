package refresher

import "onboard-kit/internal/model"

const (
	codeMissingPlanID = "refresh.missing-plan-id"
	codeDifferentPlan = "refresh.different-plan"
)

func refusal(theirs, ours model.Plan) (model.Problem, bool) {
	if theirs.ID == "" {
		return model.Problem{
			Code:     codeMissingPlanID,
			Severity: model.SeverityError,
			Message:  "The new org plan has no id, so it cannot be refreshed from.",
		}, true
	}
	if ours.Origin.PlanID != "" && ours.Origin.PlanID != theirs.ID {
		return model.Problem{
			Code:     codeDifferentPlan,
			Severity: model.SeverityError,
			Message:  "This personal plan comes from another organization's plan.",
			Detail:   map[string]string{"expected": ours.Origin.PlanID, "found": theirs.ID},
		}, true
	}
	return model.Problem{}, false
}
