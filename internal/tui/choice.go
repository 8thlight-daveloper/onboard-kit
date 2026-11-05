package tui

import "onboard-kit/internal/model"

const (
	deferLabel      = "deferred, answer later"
	clearLabel      = "clear the answer"
	skipLabel       = "skip, change nothing"
	editLabel       = "Edit the answer"
	suggestionLabel = "Use the suggestion: "
)

type choiceAction int

const (
	choiceAnswer choiceAction = iota
	choiceEdit
	choiceSuggestion
	choiceSkip
	choiceClear
)

type choice struct {
	label  string
	value  model.Value
	action choiceAction
}

func answerChoice(label string, value model.Value) choice {
	return choice{label: label, value: value}
}

func deferChoice() choice {
	return answerChoice(deferLabel, model.Value{Kind: model.ValueKindDeferred})
}

func choicesFor(question model.Question) []choice {
	var choices []choice
	switch question.Type {
	case model.VariableTypeBoolean:
		choices = append(choices,
			answerChoice("yes", model.Value{Kind: model.ValueKindFlag, Flag: true}),
			answerChoice("no", model.Value{Kind: model.ValueKindFlag, Flag: false}),
		)
	case model.VariableTypeChoice:
		for _, option := range question.Options {
			choices = append(choices, answerChoice(option, model.Value{Kind: model.ValueKindText, Text: option}))
		}
	}
	if isDirectoryQuestion(question) {
		return directoryChoicesFor(question)
	}
	return append(choices, deferChoice(), answerChoice(clearLabel, model.Value{}))
}

func directoryChoicesFor(question model.Question) []choice {
	var choices []choice
	for _, entry := range question.Choices {
		label, _ := labelOf(question.Choices, entry.Value)
		choices = append(choices, answerChoice(label, model.Value{Kind: model.ValueKindText, Text: entry.Value}))
	}
	if len(choices) == 0 || picksMany(question) {
		return choices
	}
	return append(choices, deferChoice(), answerChoice(clearLabel, model.Value{}))
}

func typedChoicesFor(question model.Question) []choice {
	choices := []choice{{label: editLabel, action: choiceEdit}}
	for _, suggestion := range question.Suggestions {
		if sameValue(suggestion.Value, question.Answer) {
			continue
		}
		choices = append(choices, choice{label: suggestionLabel + singleLine(textOf(question, suggestion.Value)), value: suggestion.Value, action: choiceSuggestion})
	}
	return append(choices, deferChoice(), choice{label: skipLabel, action: choiceSkip}, choice{label: clearLabel, action: choiceClear})
}

func picksFromList(question model.Question) bool {
	return question.Type == model.VariableTypeBoolean || question.Type == model.VariableTypeChoice || isDirectoryQuestion(question)
}
