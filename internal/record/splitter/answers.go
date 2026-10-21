package splitter

import "onboard-kit/internal/model"

func keptAnswers(plan model.Plan, joined model.Plan, previous model.Record) model.Answers {
	rules := answerRules{
		audiences:   audiencesByQuestion(joined),
		planAnswers: answeredOnly(answersByQuestion(plan.Answers)),
		heldAnswers: answeredOnly(answersByQuestion(previous.Answers)),
	}
	var kept model.Answers
	for asked, value := range answeredOnly(answersByQuestion(joined.Answers)) {
		if rules.keeps(asked, value) {
			store(&kept, asked, copyOfValue(value))
		}
	}
	return kept
}

func store(answers *model.Answers, asked question, value model.Value) {
	if asked.scope == planScope {
		if answers.Plan == nil {
			answers.Plan = map[string]model.Value{}
		}
		answers.Plan[asked.name] = value
		return
	}
	if answers.Sections == nil {
		answers.Sections = map[string]map[string]model.Value{}
	}
	if answers.Sections[asked.scope] == nil {
		answers.Sections[asked.scope] = map[string]model.Value{}
	}
	answers.Sections[asked.scope][asked.name] = value
}
