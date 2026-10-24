package writer

import "onboard-kit/internal/model"

func answerAt(answers model.Answers, scope, name string) model.Value {
	if scope == "plan" {
		return answers.Plan[name]
	}
	return answers.Sections[scope][name]
}

func putAnswer(answers *model.Answers, scope, name string, value model.Value) {
	if scope == "plan" {
		if answers.Plan == nil {
			answers.Plan = make(map[string]model.Value)
		}
		answers.Plan[name] = copyValue(value)
		return
	}
	if answers.Sections == nil {
		answers.Sections = make(map[string]map[string]model.Value)
	}
	if answers.Sections[scope] == nil {
		answers.Sections[scope] = make(map[string]model.Value)
	}
	answers.Sections[scope][name] = copyValue(value)
}

func removeAnswer(answers *model.Answers, scope, name string) {
	if scope == "plan" {
		delete(answers.Plan, name)
		return
	}
	values := answers.Sections[scope]
	if _, present := values[name]; !present {
		return
	}
	delete(values, name)
	if len(values) == 0 {
		delete(answers.Sections, scope)
	}
}
