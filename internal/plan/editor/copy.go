package editor

import (
	"slices"

	"onboard-kit/internal/model"
)

func copyEditableParts(plan model.Plan) model.Plan {
	plan.Answers = copyAnswers(plan.Answers)
	plan.Sections = copySections(plan.Sections)
	return plan
}

func copyAnswers(answers model.Answers) model.Answers {
	return model.Answers{
		Plan:     copyValues(answers.Plan),
		Sections: copySectionAnswers(answers.Sections),
	}
}

func copySectionAnswers(sections map[string]map[string]model.Value) map[string]map[string]model.Value {
	if sections == nil {
		return nil
	}
	copied := make(map[string]map[string]model.Value, len(sections))
	for id, values := range sections {
		copied[id] = copyValues(values)
	}
	return copied
}

func copyValues(values map[string]model.Value) map[string]model.Value {
	if values == nil {
		return nil
	}
	copied := make(map[string]model.Value, len(values))
	for name, value := range values {
		copied[name] = copyValue(value)
	}
	return copied
}

func copyValue(value model.Value) model.Value {
	value.List = slices.Clone(value.List)
	return value
}

func copySections(sections []model.Section) []model.Section {
	if sections == nil {
		return nil
	}
	copied := slices.Clone(sections)
	for index := range copied {
		copied[index].Enabled = copyFlag(copied[index].Enabled)
	}
	return copied
}

func copyFlag(flag *bool) *bool {
	if flag == nil {
		return nil
	}
	copied := *flag
	return &copied
}
