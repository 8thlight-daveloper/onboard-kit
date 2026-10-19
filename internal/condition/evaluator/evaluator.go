package evaluator

import "onboard-kit/internal/model"

type Evaluator struct{}

func New() Evaluator {
	return Evaluator{}
}

func (Evaluator) Holds(condition model.Condition, scope string, answers model.Answers) bool {
	for _, clause := range condition.All {
		if !clauseHolds(clause, scope, answers) {
			return false
		}
	}
	return true
}

func clauseHolds(clause model.Clause, scope string, answers model.Answers) bool {
	answer, found := findAnswer(clause.Question, scope, answers)
	return answerTestHolds(clause.Answer, answer, found && isSettled(answer))
}
