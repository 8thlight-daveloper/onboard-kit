package valuejson

import "onboard-kit/internal/model"

func EncodeProblems(problems []model.Problem) []any {
	encoded := make([]any, 0, len(problems))
	for _, problem := range problems {
		encoded = append(encoded, encodeProblem(problem))
	}
	return encoded
}

func encodeProblem(problem model.Problem) map[string]any {
	encoded := map[string]any{
		"code":     problem.Code,
		"severity": string(problem.Severity),
		"message":  problem.Message,
	}
	if problem.Question != "" {
		encoded["question"] = problem.Question
	}
	if problem.Section != "" {
		encoded["section"] = problem.Section
	}
	if len(problem.Detail) > 0 {
		encoded["detail"] = problem.Detail
	}
	return encoded
}

func EncodeQuestions(questions []model.Question) []any {
	encoded := make([]any, 0, len(questions))
	for _, question := range questions {
		encoded = append(encoded, encodeQuestion(question))
	}
	return encoded
}

func encodeQuestion(question model.Question) map[string]any {
	suggestions := make([]any, 0, len(question.Suggestions))
	for _, suggestion := range question.Suggestions {
		suggestions = append(suggestions, map[string]any{"value": Encode(suggestion.Value), "source": string(suggestion.Source)})
	}
	options := question.Options
	if options == nil {
		options = []string{}
	}
	choices := make([]any, 0, len(question.Choices))
	for _, choice := range question.Choices {
		choices = append(choices, map[string]any{"value": choice.Value, "label": choice.Label})
	}
	return map[string]any{
		"question":        question.ID,
		"scope":           question.Scope,
		"name":            question.Name,
		"description":     question.Description,
		"type":            string(question.Type),
		"options":         options,
		"choices":         choices,
		"required":        question.Required,
		"multiple":        question.Multiple,
		"audience":        string(question.Audience),
		"answer":          Encode(question.Answer),
		"declaredDefault": Encode(question.DeclaredDefault),
		"group":           question.Group,
		"suggestions":     suggestions,
	}
}
