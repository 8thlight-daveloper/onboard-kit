package splitter

import "onboard-kit/internal/model"

type answerRules struct {
	audiences   map[question]model.Audience
	planAnswers map[question]model.Value
	heldAnswers map[question]model.Value
}

func (c answerRules) keeps(asked question, value model.Value) bool {
	_, held := c.heldAnswers[asked]
	audience, declared := c.audiences[asked]
	if !declared {
		return held
	}
	if audience != model.AudienceOnboardee {
		return false
	}
	return held || c.differsFromPlan(asked, value)
}

func (c answerRules) differsFromPlan(asked question, value model.Value) bool {
	seeded, seededAnswer := c.planAnswers[asked]
	return !seededAnswer || !sameValue(seeded, value)
}

func answeredOnly(flat map[question]model.Value) map[question]model.Value {
	answered := map[question]model.Value{}
	for asked, value := range flat {
		if isAnswer(value) {
			answered[asked] = value
		}
	}
	return answered
}
