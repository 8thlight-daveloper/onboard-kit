package tui

import (
	"strings"

	"onboard-kit/internal/model"
)

func effectiveType(question model.Question) model.VariableType {
	if question.Type == "" {
		return model.DefaultVariableType
	}
	return question.Type
}

func (a App) detailLines(question model.Question, width int) []line {
	content := []line{headingLine("Details")}
	typeText := string(effectiveType(question))
	if len(question.Options) > 0 {
		typeText += ": " + strings.Join(question.Options, ", ")
	}
	if question.Multiple {
		typeText += ", several allowed"
	}
	facts := []string{
		"Type: " + typeText,
		"Section: " + a.sectionTitle(question.Scope),
		"Answered by: " + string(question.Audience),
		"Name: " + question.ID,
	}
	for _, suggestion := range question.Suggestions {
		facts = append(facts, "Suggestion ("+string(suggestion.Source)+"): "+describeValue(suggestion.Value))
	}
	for _, fact := range facts {
		content = append(content, lines(toneMuted, hanging("", fact, width))...)
	}
	return content
}
