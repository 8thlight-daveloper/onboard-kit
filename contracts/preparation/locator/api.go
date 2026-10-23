package locator

import "onboard-kit/internal/model"

type Paths struct {
	Record      string
	Preparation string
}

type Locator interface {
	Locate(resultsPath, account string) (Paths, []model.Problem)
}
