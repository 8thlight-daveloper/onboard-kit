package reviewer

import "onboard-kit/internal/model"

type HelpReference struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type Context struct {
	Account     string          `json:"account"`
	GeneralHelp []HelpReference `json:"generalHelp"`
}

type Review struct {
	PreparationProblems []model.Problem
	OnboarderBlockers   []model.Blocker
	FirstSteps          string
	FirstStepsProblems  []model.Problem
}

type Reviewer interface {
	Review(Context, model.Plan, model.Preparation) Review
}
