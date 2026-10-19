package instantiator

import "onboard-kit/internal/model"

const schemaVersion = 1

type Instantiator struct{}

func New() Instantiator {
	return Instantiator{}
}

func (Instantiator) Instantiate(templates []model.Section, axisVariables []model.Variable, templateSet string, generatedOn model.Date) (model.Plan, []model.Problem) {
	problems := append(duplicateIDProblems(templates), namelessVariableProblems(axisVariables)...)
	if len(problems) > 0 {
		return model.Plan{}, problems
	}
	return model.Plan{
		SchemaVersion: schemaVersion,
		Provenance:    model.Provenance{TemplateSet: templateSet, GeneratedOn: generatedOn},
		Variables:     copyVariables(axisVariables),
		Sections:      instantiateSections(templates),
	}, nil
}
