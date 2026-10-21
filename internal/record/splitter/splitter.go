package splitter

import "onboard-kit/internal/model"

type Splitter struct{}

func New() Splitter {
	return Splitter{}
}

func (Splitter) Split(plan model.Plan, joined model.Plan, previous model.Record) model.Record {
	result := copyOfRecord(previous)
	result.Answers = keptAnswers(plan, joined, previous)
	return result
}
