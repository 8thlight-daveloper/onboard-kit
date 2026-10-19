package evaluator

import "onboard-kit/internal/model"

type Evaluator struct{}

func New() Evaluator {
	return Evaluator{}
}

func (Evaluator) Holds(condition model.Condition, scope string, answers model.Answers, switches model.Switches) bool {
	for _, clause := range condition.All {
		if !clauseHolds(clause, scope, answers, switches) {
			return false
		}
	}
	return true
}

func clauseHolds(clause model.Clause, scope string, answers model.Answers, switches model.Switches) bool {
	if clause.Section != "" {
		return sectionClauseHolds(clause, switches)
	}
	answer, found := findAnswer(clause.Question, scope, answers)
	return answerTestHolds(clause.Answer, answer, found && isSettled(answer))
}
