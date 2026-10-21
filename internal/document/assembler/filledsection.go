package assembler

import "onboard-kit/internal/model"

type filledSection struct {
	filled   model.Section
	declared []model.Variable
	depth    int
	problems []model.Problem
}

func (a Assembler) fillActiveSections(plan model.Plan, context planContext) ([]filledSection, []model.Problem) {
	var sections []filledSection
	var problems []model.Problem
	parents := parentsByID(plan.Sections)
	for _, active := range a.selector.Select(plan) {
		filled, sectionProblems := a.fillSection(active, context)
		sections = append(sections, filledSection{filled: filled, declared: active.Variables, depth: depthOf(active, parents), problems: sectionProblems})
		problems = append(problems, sectionProblems...)
	}
	return sections, problems
}

func (a Assembler) fillSection(active model.Section, context planContext) (model.Section, []model.Problem) {
	filled, reported := a.resolver.Resolve(active, context.answers, context.directory, context.team)
	filled, escalationProblems := a.fillEscalations(filled, active, context)
	reported = append(reported, escalationProblems...)
	reported = append(reported, a.vacantRoleProblems(active, context)...)
	return filled, inVariableOrder(active, reported)
}
