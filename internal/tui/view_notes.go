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

func (app App) noteLines(question model.Question, width int) []line {
	content := []line{faintLine("About this question"), blankLine()}
	for _, fact := range app.noteFacts(question) {
		content = append(content, lines(toneFaint, hanging("", fact, width))...)
	}
	return content
}

func (app App) noteFacts(question model.Question) []string {
	var facts []string
	if question.Example != "" {
		facts = append(facts, "Example: "+singleLine(question.Example))
	}
	typeText := string(effectiveType(question))
	if len(question.Options) > 0 {
		typeText += ": " + strings.Join(question.Options, ", ")
	}
	if question.Multiple {
		typeText += ", several allowed"
	}
	facts = append(facts,
		"Type: "+typeText,
		"Answered by: "+string(question.Audience),
		"Name: "+question.ID,
	)
	for _, suggestion := range question.Suggestions {
		facts = append(facts, "Suggestion ("+string(suggestion.Source)+"): "+describeValue(suggestion.Value))
	}
	return facts
}
