package refresher

import "onboard-kit/internal/model"

const planScope = "plan"

type question struct {
	scope    string
	name     string
	variable model.Variable
}

func (q question) id() string {
	return q.scope + "." + q.name
}

func (q question) isForOnboardee() bool {
	return q.variable.Audience == model.AudienceOnboardee
}

func declaredQuestions(plan model.Plan) map[string]question {
	questions := map[string]question{}
	add := func(scope string, variables []model.Variable) {
		for _, variable := range variables {
			asked := question{scope: scope, name: variable.Name, variable: variable}
			questions[asked.id()] = asked
		}
	}
	add(planScope, plan.Variables)
	for _, section := range plan.Sections {
		add(section.ID, section.Variables)
	}
	return questions
}
