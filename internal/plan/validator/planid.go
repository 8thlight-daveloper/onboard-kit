package validator

import (
	"strings"

	"onboard-kit/internal/model"
)

func duplicateSectionProblems(plan model.Plan) []model.Problem {
	counts := map[string]int{}
	problems := []model.Problem{}
	for _, section := range plan.Sections {
		counts[section.ID]++
		if counts[section.ID] == 2 {
			problems = append(problems, model.Problem{
				Code:     "plan.duplicate-section-id",
				Severity: model.SeverityError,
				Message:  "Two or more sections share an id.",
				Section:  section.ID,
			})
		}
	}
	return problems
}

func duplicateTeamVariableProblems(plan model.Plan) []model.Problem {
	names := []string{}
	for _, variable := range plan.Variables {
		if variable.Type == model.VariableTypeTeam {
			names = append(names, variable.Name)
		}
	}
	if len(names) < 2 {
		return nil
	}
	return []model.Problem{{
		Code:     "plan.duplicate-team-variable",
		Severity: model.SeverityError,
		Message:  "A plan holds more than one team variable.",
		Detail:   map[string]string{"variables": strings.Join(names, ",")},
	}}
}
