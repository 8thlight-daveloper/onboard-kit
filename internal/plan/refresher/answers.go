package refresher

import (
	"maps"
	"slices"

	"onboard-kit/internal/model"
)

func answerOf(answers model.Answers, asked question) model.Value {
	if asked.scope == planScope {
		return answers.Plan[asked.name]
	}
	return answers.Sections[asked.scope][asked.name]
}

func answersAfter(ours model.Answers, decisions []decision) model.Answers {
	answers := copyAnswers(ours)
	for _, made := range decisions {
		answers = withAnswer(answers, made.question, made.answer)
	}
	return answers
}

func withAnswer(answers model.Answers, asked question, value model.Value) model.Answers {
	if asked.scope == planScope {
		answers.Plan = changedValues(answers.Plan, asked.name, value)
		return answers
	}
	changed := changedValues(answers.Sections[asked.scope], asked.name, value)
	if len(changed) == 0 {
		delete(answers.Sections, asked.scope)
		return answers
	}
	if answers.Sections == nil {
		answers.Sections = map[string]map[string]model.Value{}
	}
	answers.Sections[asked.scope] = changed
	return answers
}

func changedValues(values map[string]model.Value, name string, value model.Value) map[string]model.Value {
	if isAbsent(value) {
		delete(values, name)
		return values
	}
	if values == nil {
		values = map[string]model.Value{}
	}
	values[name] = copyValue(value)
	return values
}

func copyAnswers(answers model.Answers) model.Answers {
	copied := model.Answers{Plan: copyValues(answers.Plan)}
	if answers.Sections != nil {
		copied.Sections = make(map[string]map[string]model.Value, len(answers.Sections))
		for id, values := range answers.Sections {
			copied.Sections[id] = copyValues(values)
		}
	}
	return copied
}

func copyValues(values map[string]model.Value) map[string]model.Value {
	copied := maps.Clone(values)
	for name, value := range copied {
		copied[name] = copyValue(value)
	}
	return copied
}

func copyValue(value model.Value) model.Value {
	value.List = slices.Clone(value.List)
	return value
}
