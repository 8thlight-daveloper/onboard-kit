package reconciler

import "onboard-kit/internal/model"

type listing int

const (
	listingKept listing = iota
	listingReview
)

func checkedAnswer(value model.Value, variable model.Variable, directory model.Directory) (model.Value, listing) {
	switch {
	case !isSettled(value):
		return value, listingKept
	case isDirectoryType(variable.Type):
		return checkedDirectoryAnswer(value, variable, directory)
	default:
		return keptIfValid(value, isValidAnswer(value, variable))
	}
}

func keptIfValid(value model.Value, valid bool) (model.Value, listing) {
	if valid {
		return value, listingKept
	}
	return model.Value{}, listingReview
}

func checkedAnswers(answers model.Answers, questions map[questionKey]model.Variable, directory model.Directory) (model.Answers, map[string]bool) {
	review := map[string]bool{}
	check := func(scope string, values map[string]model.Value) map[string]model.Value {
		checked := map[string]model.Value{}
		for name, value := range values {
			key := questionKey{scope: scope, name: name}
			variable, declared := questions[key]
			if !declared || !isForOnboardee(variable) {
				checked[name] = copiedValue(value)
				continue
			}
			result, outcome := checkedAnswer(value, variable, directory)
			if outcome == listingReview {
				review[key.id()] = true
			}
			if result.Kind != model.ValueKindNone {
				checked[name] = copiedValue(result)
			}
		}
		return checked
	}

	result := model.Answers{}
	if answers.Plan != nil {
		result.Plan = check(planScope, answers.Plan)
	}
	for scope, values := range answers.Sections {
		checked := check(scope, values)
		if len(checked) == 0 {
			continue
		}
		if result.Sections == nil {
			result.Sections = map[string]map[string]model.Value{}
		}
		result.Sections[scope] = checked
	}
	return result, review
}
