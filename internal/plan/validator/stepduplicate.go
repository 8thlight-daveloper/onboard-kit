package validator

import "onboard-kit/internal/model"

func duplicateStepProblems(section model.Section) []model.Problem {
	counts := map[string]int{}
	var problems []model.Problem
	for _, step := range section.Steps {
		if step.ID == "" {
			continue
		}
		counts[step.ID]++
		if counts[step.ID] == 2 {
			problems = append(problems, model.Problem{
				Code:     "plan.duplicate-step-id",
				Severity: model.SeverityError,
				Message:  "Step id " + step.ID + " is repeated in section " + section.ID + ". Give each named step a unique id.",
				Section:  section.ID,
				Detail:   map[string]string{"step": step.ID},
			})
		}
	}
	return problems
}
