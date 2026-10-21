package finder

import (
	"regexp"
	"slices"

	"onboard-kit/internal/model"
)

func newOffer(question model.Question, pattern string) func(model.Value) bool {
	switch {
	case question.Type == model.VariableTypeBoolean:
		return func(value model.Value) bool { return value.Kind == model.ValueKindFlag }
	case isDirectoryType(question.Type):
		return offerChoices(question)
	case question.Type == model.VariableTypeCommand || question.Multiple:
		return func(value model.Value) bool { return value.Kind == model.ValueKindList && isSettled(value) }
	case question.Type == model.VariableTypeChoice:
		return func(value model.Value) bool { return isText(value) && slices.Contains(question.Options, value.Text) }
	case question.Type == model.VariableTypeString && pattern != "":
		return offerMatching(pattern)
	}
	return isText
}

func offerChoices(question model.Question) func(model.Value) bool {
	if question.Multiple {
		return func(value model.Value) bool {
			return value.Kind == model.ValueKindList && isSettled(value) && len(knownItems(question.Choices, value.List)) == len(value.List)
		}
	}
	return func(value model.Value) bool { return isText(value) && hasChoice(question.Choices, value.Text) }
}

func offerMatching(pattern string) func(model.Value) bool {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return func(model.Value) bool { return false }
	}
	return func(value model.Value) bool { return isText(value) && compiled.MatchString(value.Text) }
}

func isText(value model.Value) bool {
	return value.Kind == model.ValueKindText && value.Text != ""
}
