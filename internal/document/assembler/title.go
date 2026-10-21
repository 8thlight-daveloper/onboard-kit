package assembler

import "onboard-kit/internal/model"

func titleLine(plan model.Plan, teamIdentifier string) string {
	organization, hasOrganization := settledText(plan.Answers.Plan["organization"])
	team, hasTeam := teamTitle(plan, teamIdentifier)
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
