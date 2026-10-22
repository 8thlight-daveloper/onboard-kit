package resolver

import "onboard-kit/internal/model"

func (f *filler) problems(section model.Section) []model.Problem {
	var problems []model.Problem
	reported := map[string]bool{}
	for _, variable := range section.Variables {
		if !f.used[variable.Name] || reported[variable.Name] {
			continue
		}
		reported[variable.Name] = true
		problems = append(problems, gapProblem(section.ID, variable, f.resolutions[variable.Name].gapCode))
	}
	return problems
}

func gapProblem(sectionID string, variable model.Variable, code string) model.Problem {
	audience := variable.Audience
	if audience == "" {
		audience = model.AudienceOnboarder
	}
	return model.Problem{
		Code:     code,
		Severity: model.SeverityWarning,
		Message:  "variable " + variable.Name + " has no value: " + code,
		Question: sectionID + "." + variable.Name,
		Section:  sectionID,
		Detail:   map[string]string{"audience": string(audience)},
	}
}
