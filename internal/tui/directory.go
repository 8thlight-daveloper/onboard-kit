package tui

import (
	"slices"
	"strings"

	"onboard-kit/internal/model"
)

func isDirectoryQuestion(question model.Question) bool {
	switch question.Type {
	case model.VariableTypeContact, model.VariableTypeRole, model.VariableTypeTeam:
		return true
	}
	return false
}

func picksMany(question model.Question) bool {
	return isDirectoryQuestion(question) && question.Multiple
}

func directoryNoun(question model.Question) string {
	switch question.Type {
	case model.VariableTypeRole:
		return "roles"
	case model.VariableTypeTeam:
		return "teams"
	}
	return "contacts"
}

func labelOf(choices []model.Choice, id string) (string, bool) {
	index := slices.IndexFunc(choices, func(each model.Choice) bool { return each.Value == id })
	if index < 0 {
		return "", false
	}
	found := choices[index]
	sameName := 0
	for _, each := range choices {
		if each.Label == found.Label {
			sameName++
		}
	}
	if sameName > 1 {
		return found.Label + " (" + found.Value + ")", true
	}
	return found.Label, true
}

func knownLabels(question model.Question, ids []string) []string {
	var labels []string
	for _, id := range ids {
		if label, ok := labelOf(question.Choices, id); ok {
			labels = append(labels, label)
		}
	}
	return labels
}

func describeAnswer(question model.Question, value model.Value) string {
	if !isDirectoryQuestion(question) {
		return describeValue(value)
	}
	switch value.Kind {
	case model.ValueKindText:
		return strings.Join(knownLabels(question, []string{value.Text}), "")
	case model.ValueKindList:
		return strings.Join(knownLabels(question, value.List), personSeparator+" ")
	}
	return describeValue(value)
}

func answerState(question model.Question) string {
	state := stateOf(question.Answer)
	if state == "settled" && isDirectoryQuestion(question) && describeAnswer(question, question.Answer) == "" {
		return "unanswered"
	}
	return state
}

func emptyDirectoryText(question model.Question, audience model.Audience) string {
	if audience == model.AudienceOnboardee {
		return "No " + directoryNoun(question) + " to pick from yet. Ask your onboarder to add this."
	}
	return "No " + directoryNoun(question) + " to pick from yet. Add them on the People screen, press P."
}
