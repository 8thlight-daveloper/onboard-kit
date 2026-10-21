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
	return buildQuestions(plan, ordered)
}

func buildQuestions(plan model.Plan, ordered []candidate) []model.Question {
	questions := make([]model.Question, len(ordered))
	for index, asked := range ordered {
		questions[index] = newQuestion(plan, asked)
	}
	for index, asked := range ordered {
		questions[index].Suggestions = suggest(plan.Answers, asked, questions, index)
	}
	return questions
}

func newQuestion(plan model.Plan, asked candidate) model.Question {
	variable := asked.variable
	answer, _ := storedAnswer(plan.Answers, asked.scope, variable.Name)
	question := model.Question{
		ID:              asked.id(),
		Scope:           asked.scope,
		Name:            variable.Name,
		Description:     variable.Description,
		Example:         variable.Example,
		Type:            typeOf(variable),
		Options:         optionsOf(variable),
		Choices:         choicesOf(plan.Directory, variable),
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

func optionsOf(variable model.Variable) []string {
	if isDirectoryType(variable.Type) {
		return nil
	}
	return variable.Options
}

func choicesOf(directory model.Directory, variable model.Variable) []model.Choice {
	if !isDirectoryType(variable.Type) {
		return nil
	}
	return directoryChoices(directory, variable.Type)
}
