package validator

import (
	"onboard-kit/internal/model"
	"onboard-kit/internal/question/finder"
	"onboard-kit/internal/secrets"
)

type Validator struct{}

func New() Validator {
	return Validator{}
}

func (Validator) Validate(plan model.Plan) []model.Problem {
	active := activeQuestions(plan)
	problems := []model.Problem{}
	for _, stored := range storedAnswers(plan) {
		problems = append(problems, validateStored(plan, active, stored)...)
	}
	return problems
}

func (Validator) CheckValue(plan model.Plan, question string, value model.Value) []model.Problem {
	target := parseQuestion(question)
	variable, declared := declaredVariable(plan, target)
	if !declared {
		return []model.Problem{unknownQuestion(plan, target, model.SeverityError)}
	}
	if !activeQuestions(plan)[target.id] {
		return []model.Problem{inactiveQuestion(plan, target)}
	}
	return checkActive(plan, target, variable, value)
}

func validateStored(plan model.Plan, active map[string]bool, stored storedAnswer) []model.Problem {
	variable, declared := declaredVariable(plan, stored.question)
	switch {
	case !declared:
		unknown := unknownQuestion(plan, stored.question, model.SeverityWarning)
		return append([]model.Problem{unknown}, secretProblems(plan, stored.question, stored.value)...)
	case active[stored.question.id]:
		return checkActive(plan, stored.question, variable, stored.value)
	default:
		return secretProblems(plan, stored.question, stored.value)
	}
}

func checkActive(plan model.Plan, target question, variable model.Variable, value model.Value) []model.Problem {
	problems := annotate(plan, target, answerProblems(variable, value))
	return append(problems, secretProblems(plan, target, value)...)
}

func secretProblems(plan model.Plan, target question, value model.Value) []model.Problem {
	return annotate(plan, target, secrets.New().Check(value))
}

func activeQuestions(plan model.Plan) map[string]bool {
	active := map[string]bool{}
	for _, asked := range finder.New().Questions(plan) {
		active[asked.ID] = true
	}
	return active
}
