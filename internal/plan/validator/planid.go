package validator

import (
	"strings"

	"onboard-kit/internal/model"
)

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
