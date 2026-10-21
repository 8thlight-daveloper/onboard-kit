package reconciler

import "onboard-kit/internal/model"

const planScope = "plan"

type questionKey struct {
	scope string
	name  string
}

func (key questionKey) id() string {
	return key.scope + "." + key.name
}

func declaredQuestions(plan model.Plan) map[questionKey]model.Variable {
	questions := map[questionKey]model.Variable{}
	add := func(scope string, variables []model.Variable) {
		for _, variable := range variables {
			questions[questionKey{scope: scope, name: variable.Name}] = variable
		}
	}
	add(planScope, plan.Variables)
	for _, section := range plan.Sections {
		add(section.ID, section.Variables)
	}
	return questions
}

func isForOnboardee(variable model.Variable) bool {
	return variable.Audience == model.AudienceOnboardee
}
