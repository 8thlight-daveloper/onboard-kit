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
	problems := planWideProblems(plan)
	active := activeQuestions(plan)
	for _, stored := range storedAnswers(plan) {
		problems = append(problems, validateStored(plan, active, stored)...)
	}
	return problems
}

func (Validator) CheckValue(plan model.Plan, question string, value model.Value) []model.Problem {
	target := parseQuestion(question)
	variable, declarations := declaredVariable(plan, target)
	if declarations == 0 {
		return []model.Problem{unknownQuestion(plan, target, model.SeverityError)}
	}
	if declarations > 1 {
		return []model.Problem{ambiguousQuestion(plan, target)}
	}
	if !activeQuestions(plan)[target.id] {
		return []model.Problem{inactiveQuestion(plan, target)}
	}
	return checkActive(plan, target, variable, value, model.SeverityError)
}

func validateStored(plan model.Plan, active map[string]bool, stored storedAnswer) []model.Problem {
	variable, declarations := declaredVariable(plan, stored.question)
	switch {
	case declarations == 0:
		unknown := unknownQuestion(plan, stored.question, model.SeverityWarning)
		return append([]model.Problem{unknown}, secretProblems(plan, stored.question, stored.value)...)
	case declarations > 1:
		return secretProblems(plan, stored.question, stored.value)
	case active[stored.question.id]:
		return checkActive(plan, stored.question, variable, stored.value, model.SeverityWarning)
	default:
		return secretProblems(plan, stored.question, stored.value)
	}
}

func checkActive(plan model.Plan, target question, variable model.Variable, value model.Value, unknownIDSeverity model.Severity) []model.Problem {
	problems := annotate(plan, target, answerProblems(plan.Directory, unknownIDSeverity, variable, value))
	return append(problems, secretProblems(plan, target, value)...)
}

func secretProblems(plan model.Plan, target question, value model.Value) []model.Problem {
	return annotate(plan, target, secrets.New().Check(value))
}

func activeQuestions(plan model.Plan) map[string]bool {
	active := map[string]bool{}
	for _, asked := range finder.New().Questions(plan, "") {
		active[asked.ID] = true
	}
	return active
}
