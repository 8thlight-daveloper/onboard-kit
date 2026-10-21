package assembler

import "onboard-kit/internal/model"

func inVariableOrder(section model.Section, reported []model.Problem) []model.Problem {
	byQuestion := map[string]model.Problem{}
	for _, problem := range reported {
		if _, seen := byQuestion[problem.Question]; !seen {
			byQuestion[problem.Question] = problem
		}
	}
	var ordered []model.Problem
	for _, variable := range section.Variables {
		question := section.ID + "." + variable.Name
		if problem, found := byQuestion[question]; found {
			ordered = append(ordered, problem)
			delete(byQuestion, question)
		}
	}
	return ordered
}
