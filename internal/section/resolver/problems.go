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
		problems = append(problems, gapProblem(section.ID, variable.Name, f.resolutions[variable.Name].gapCode))
	}
	return problems
}

func gapProblem(sectionID string, name string, code string) model.Problem {
	return model.Problem{
		Code:     code,
		Severity: model.SeverityWarning,
		Message:  "variable " + name + " has no value: " + code,
		Question: sectionID + "." + name,
		Section:  sectionID,
	}
}
