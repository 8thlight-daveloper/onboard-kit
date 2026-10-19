package finder

import "onboard-kit/internal/model"

func suggest(answers model.Answers, asked candidate, questions []model.Question, position int) []model.Suggestion {
	question := questions[position]
	offers := newOffer(question, asked.variable.Pattern)
	suggestions := []model.Suggestion{}
	defaults := []model.Value{}
	for _, member := range questions {
		if member.Group == question.Group {
			defaults = append(defaults, member.DeclaredDefault)
		}
	}
	suggestions = appendDistinct(suggestions, defaults, model.SuggestionSourceDefault, offers)
	return appendDistinct(suggestions, givenAnswers(answers, question.Name), model.SuggestionSourceGiven, offers)
}

func appendDistinct(suggestions []model.Suggestion, values []model.Value, source model.SuggestionSource, offers func(model.Value) bool) []model.Suggestion {
	seen := []model.Value{}
	for _, value := range values {
		if !offers(value) || containsValue(seen, value) {
			continue
		}
		seen = append(seen, value)
		suggestions = append(suggestions, model.Suggestion{Value: value, Source: source})
	}
	return suggestions
}

func containsValue(values []model.Value, target model.Value) bool {
	for _, value := range values {
		if sameValue(value, target) {
			return true
		}
	}
	return false
}
