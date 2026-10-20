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
	candidates := variableCandidates(planScope, plan.Variables)
	for _, section := range sections {
		candidates = append(candidates, variableCandidates(section.ID, section.Variables)...)
	}
	return candidates
}

func variableCandidates(scope string, variables []model.Variable) []candidate {
	candidates := []candidate{}
	for _, variable := range variables {
		candidates = append(candidates, candidate{scope: scope, variable: variable})
	}
	return candidates
}
