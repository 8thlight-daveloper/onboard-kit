package validator

import (
	"slices"
	"strings"

	"onboard-kit/internal/model"
)

type storedAnswer struct {
	question question
	value    model.Value
}

func storedAnswers(plan model.Plan) []storedAnswer {
	stored := []storedAnswer{}
	for name, value := range plan.Answers.Plan {
		stored = append(stored, storedAnswer{question: question{id: planScope + "." + name, scope: planScope, name: name}, value: value})
	}
	for scope, answers := range plan.Answers.Sections {
		for name, value := range answers {
			stored = append(stored, storedAnswer{question: question{id: scope + "." + name, scope: scope, name: name}, value: value})
		}
	}
	slices.SortStableFunc(stored, func(left, right storedAnswer) int {
		return strings.Compare(left.question.id, right.question.id)
	})
	return stored
}
