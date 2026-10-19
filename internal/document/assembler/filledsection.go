package assembler

import "onboard-kit/internal/model"

type filledSection struct {
	filled   model.Section
	declared []model.Variable
	problems []model.Problem
}

func (a Assembler) fillActiveSections(plan model.Plan) ([]filledSection, []model.Problem) {
	var sections []filledSection
	var problems []model.Problem
	for _, active := range a.selector.Select(plan) {
		filled, sectionProblems := a.resolver.Resolve(active, plan.Answers)
		sections = append(sections, filledSection{filled: filled, declared: active.Variables, problems: sectionProblems})
		problems = append(problems, sectionProblems...)
	}
	return sections, problems
}
