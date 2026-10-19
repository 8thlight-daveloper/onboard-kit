package assembler

import "onboard-kit/internal/model"

func titleLine(answers model.Answers) string {
	organization, hasOrganization := settledText(answers.Plan["organization"])
	team, hasTeam := settledText(answers.Plan["team"])
	switch {
	case hasOrganization && hasTeam:
		return "# " + organization + " onboarding: " + team
	case hasOrganization:
		return "# " + organization + " onboarding"
	case hasTeam:
		return "# Onboarding: " + team
	default:
		return "# Onboarding"
	}
}

func settledText(value model.Value) (string, bool) {
	return value.Text, value.Kind == model.ValueKindText && value.Text != ""
}
