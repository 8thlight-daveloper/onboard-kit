package tui

import "onboard-kit/internal/model"

type choice struct {
	label string
	value model.Value
}

func choicesFor(question model.Question) []choice {
	var choices []choice
	switch question.Type {
	case model.VariableTypeBoolean:
		choices = append(choices,
			choice{"yes", model.Value{Kind: model.ValueKindFlag, Flag: true}},
			choice{"no", model.Value{Kind: model.ValueKindFlag, Flag: false}},
		)
	case model.VariableTypeChoice:
		for _, option := range question.Options {
			choices = append(choices, choice{option, model.Value{Kind: model.ValueKindText, Text: option}})
		}
	}
	if isDirectoryQuestion(question) {
		return directoryChoicesFor(question)
	}
	return append(choices,
		choice{"deferred, answer later", model.Value{Kind: model.ValueKindDeferred}},
		choice{"clear the answer", model.Value{}},
	)
}

func directoryChoicesFor(question model.Question) []choice {
	var choices []choice
	for _, entry := range question.Choices {
		label, _ := labelOf(question.Choices, entry.Value)
		choices = append(choices, choice{label, model.Value{Kind: model.ValueKindText, Text: entry.Value}})
	}
	if len(choices) == 0 || picksMany(question) {
		return choices
	}
	return append(choices,
		choice{"deferred, answer later", model.Value{Kind: model.ValueKindDeferred}},
		choice{"clear the answer", model.Value{}},
	)
}

func picksFromList(question model.Question) bool {
	return question.Type == model.VariableTypeBoolean || question.Type == model.VariableTypeChoice || isDirectoryQuestion(question)
}
