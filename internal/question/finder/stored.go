package finder

import (
	"slices"

	"onboard-kit/internal/model"
)

func storedAnswer(answers model.Answers, scope string, name string) (model.Value, bool) {
	if scope == planScope {
		value, found := answers.Plan[name]
		return value, found
	}
	value, found := answers.Sections[scope][name]
	return value, found
}

func givenAnswers(answers model.Answers, name string) []model.Value {
	given := []model.Value{}
	if value, found := answers.Plan[name]; found {
		given = append(given, value)
	}
	sectionIDs := make([]string, 0, len(answers.Sections))
	for sectionID := range answers.Sections {
		sectionIDs = append(sectionIDs, sectionID)
	}
	slices.Sort(sectionIDs)
	for _, sectionID := range sectionIDs {
		if value, found := answers.Sections[sectionID][name]; found {
			given = append(given, value)
		}
	}
	return given
}
