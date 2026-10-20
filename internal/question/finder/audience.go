package finder

import "onboard-kit/internal/model"

func forAudience(questions []model.Question, audience model.Audience) []model.Question {
	if audience == "" {
		return questions
	}
	chosen := []model.Question{}
	for _, question := range questions {
		if question.Audience == audience {
			chosen = append(chosen, question)
		}
	}
	return chosen
}

func audienceOf(variable model.Variable) model.Audience {
	if variable.Audience == "" {
		return model.DefaultAudience
	}
	return variable.Audience
}
