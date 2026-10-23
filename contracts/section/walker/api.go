package walker

import "onboard-kit/internal/model"

type Walker interface {
	Walk(plan model.Plan, record model.Record) (View, []model.Problem)
}

type Comparison string

const (
	ComparisonUnknown   Comparison = "unknown"
	ComparisonSame      Comparison = "same"
	ComparisonDifferent Comparison = "different"
)

type View struct {
	Sections  []SectionView
	Outcomes  []OutcomeView
	Questions []model.Question
	Blockers  []model.Blocker
}

type SectionView struct {
	Section     model.Section
	CurrentHash string
	WaitingOn   []string
	Escalation  []string
}

type OutcomeView struct {
	Outcome     model.Outcome
	CurrentHash string
	Comparison  Comparison
}
