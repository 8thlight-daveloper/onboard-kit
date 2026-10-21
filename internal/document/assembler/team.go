package assembler

import "onboard-kit/internal/model"

func teamIdentifier(plan model.Plan) string {
	for _, variable := range plan.Variables {
		if variable.Type != model.VariableTypeTeam {
			continue
		}
		identifier, settled := settledText(plan.Answers.Plan[variable.Name])
		if settled && findTeam(plan.Directory, identifier) != nil {
			return identifier
		}
		return ""
	}
	return ""
}

func findTeam(directory model.Directory, identifier string) *model.Team {
	for index := range directory.Teams {
		if directory.Teams[index].ID == identifier {
			return &directory.Teams[index]
		}
	}
	return nil
}

func teamTitle(plan model.Plan, identifier string) (string, bool) {
	if team := findTeam(plan.Directory, identifier); team != nil && identifier != "" {
		return team.Name, true
	}
	return settledText(plan.Answers.Plan["team"])
}
