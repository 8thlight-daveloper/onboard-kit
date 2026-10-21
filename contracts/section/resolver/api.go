package resolver

import "onboard-kit/internal/model"

type Resolver interface {
	Resolve(section model.Section, answers model.Answers, directory model.Directory, team string) (model.Section, []model.Problem)
}
