package resolver

import "onboard-kit/internal/model"

type Resolver struct{}

func New() Resolver {
	return Resolver{}
}

func (Resolver) Resolve(section model.Section, answers model.Answers) (model.Section, []model.Problem) {
	filler := newFiller(section, answers.Sections[section.ID])
	filled := fillSection(section, filler)
	return filled, filler.problems(section)
}
