package assembler

import "onboard-kit/internal/model"

type filledSection struct {
	filled   model.Section
	declared []model.Variable
	depth    int
	problems []model.Problem
}

func (a Assembler) fillActiveSections(plan model.Plan) ([]filledSection, []model.Problem) {
	var sections []filledSection
	var problems []model.Problem
	parents := parentsByID(plan.Sections)
	for _, active := range a.selector.Select(plan) {
		filled, sectionProblems := a.resolver.Resolve(active, plan.Answers, plan.Directory, "")
		sections = append(sections, filledSection{filled: filled, declared: active.Variables, depth: depthOf(active, parents), problems: sectionProblems})
		problems = append(problems, sectionProblems...)
	}
	return sections, problems
}
