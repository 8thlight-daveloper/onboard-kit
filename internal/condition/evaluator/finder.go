package evaluator

import (
	"strings"

	"onboard-kit/internal/model"
)

const planScope = "plan"

func findAnswer(question string, scope string, answers model.Answers) (model.Value, bool) {
	questionScope, name := splitQuestion(question, scope)
	if questionScope == planScope {
		value, found := answers.Plan[name]
		return value, found
	}
	value, found := answers.Sections[questionScope][name]
	return value, found
}

func splitQuestion(question string, scope string) (string, string) {
	questionScope, name, qualified := strings.Cut(question, ".")
	if !qualified {
		return scope, question
	}
	return questionScope, name
}
