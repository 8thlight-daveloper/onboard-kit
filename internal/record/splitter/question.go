package splitter

import "onboard-kit/internal/model"

const planScope = "plan"

type question struct {
	scope string
	name  string
}

func answersByQuestion(answers model.Answers) map[question]model.Value {
	flat := map[question]model.Value{}
	for name, value := range answers.Plan {
		flat[question{planScope, name}] = value
	}
	for scope, values := range answers.Sections {
		for name, value := range values {
			flat[question{scope, name}] = value
		}
	}
	return flat
}

func audiencesByQuestion(plan model.Plan) map[question]model.Audience {
	audiences := map[question]model.Audience{}
	for _, section := range plan.Sections {
		declareAll(audiences, section.ID, section.Variables)
	}
	declareAll(audiences, planScope, plan.Variables)
	return audiences
}

func declareAll(audiences map[question]model.Audience, scope string, variables []model.Variable) {
	declared := map[string]bool{}
	for _, variable := range variables {
		if declared[variable.Name] {
			continue
		}
		declared[variable.Name] = true
		audiences[question{scope, variable.Name}] = variable.Audience
	}
}
