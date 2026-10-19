package instantiator

import "onboard-kit/internal/model"

type Instantiator interface {
	Instantiate(templates []model.Section, axisVariables []model.Variable, templateSet string, generatedOn model.Date) (model.Plan, []model.Problem)
}
