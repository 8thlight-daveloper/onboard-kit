package resolver

import "onboard-kit/internal/model"

type Resolver interface {
	Resolve(section model.Section, answers model.Answers) (model.Section, []model.Problem)
}
