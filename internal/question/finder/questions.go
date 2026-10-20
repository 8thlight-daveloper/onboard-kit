package finder

import (
	"onboard-kit/internal/model"
	"onboard-kit/internal/section/selector"
)

func findQuestions(plan model.Plan) []model.Question {
	sections := selector.New().Select(plan)
	candidates := listCandidates(plan, sections)
	active := findActive(plan, candidates)
	ordered := orderActive(active)
	return buildQuestions(plan.Answers, ordered)
}

func buildQuestions(answers model.Answers, ordered []candidate) []model.Question {
	questions := make([]model.Question, len(ordered))
	for index, asked := range ordered {
		questions[index] = newQuestion(answers, asked)
	}
	for index, asked := range ordered {
		questions[index].Suggestions = suggest(answers, asked, questions, index)
	}
	return questions
}

func newQuestion(answers model.Answers, asked candidate) model.Question {
	variable := asked.variable
	answer, _ := storedAnswer(answers, asked.scope, variable.Name)
	question := model.Question{
		ID:              asked.id(),
		Scope:           asked.scope,
		Name:            variable.Name,
		Description:     variable.Description,
		Example:         variable.Example,
		Type:            typeOf(variable),
		Options:         variable.Options,
		Required:        variable.Required,
		Multiple:        variable.Multiple,
		Audience:        audienceOf(variable),
		Answer:          answer,
		DeclaredDefault: variable.Default,
	}
	question.Group = groupKey(question)
	return question
}

func typeOf(variable model.Variable) model.VariableType {
	if variable.Type == "" {
		return model.DefaultVariableType
	}
	return variable.Type
}
