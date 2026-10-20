package finder

import "onboard-kit/internal/model"

const planScope = "plan"

type candidate struct {
	scope    string
	variable model.Variable
}

func (asked candidate) id() string {
	return asked.scope + "." + asked.variable.Name
}

func listCandidates(plan model.Plan, sections []model.Section) []candidate {
	candidates := planningCandidates(planScope, plan.Variables)
	for _, section := range sections {
		candidates = append(candidates, planningCandidates(section.ID, section.Variables)...)
	}
	return candidates
}

func planningCandidates(scope string, variables []model.Variable) []candidate {
	candidates := []candidate{}
	for _, variable := range variables {
		if variable.Audience != model.AudienceOnboardee {
			candidates = append(candidates, candidate{scope: scope, variable: variable})
		}
	}
	return candidates
}
