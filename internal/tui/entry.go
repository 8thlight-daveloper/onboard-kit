package tui

import (
	"strings"
	"unicode"

	"onboard-kit/internal/model"
)

const itemSeparator = ";"

func parseEntry(question model.Question, text string) model.Value {
	if strings.TrimSpace(text) == "" {
		return model.Value{}
	}
	switch {
	case question.Type == model.VariableTypeCommand:
		return model.Value{Kind: model.ValueKindList, List: splitWords(text)}
	case question.Multiple:
		return model.Value{Kind: model.ValueKindList, List: splitItems(text)}
	}
	return model.Value{Kind: model.ValueKindText, Text: text}
}

func entryText(question model.Question) string {
	return textOf(question, question.Answer)
}

func textOf(question model.Question, value model.Value) string {
	if value.Kind == model.ValueKindText {
		return value.Text
	}
	if value.Kind != model.ValueKindList {
		return ""
	}
	if question.Type == model.VariableTypeCommand {
		return joinWords(value.List)
	}
	return strings.Join(value.List, itemSeparator+" ")
}

func splitItems(text string) []string {
	var items []string
	for _, part := range strings.Split(text, itemSeparator) {
		if item := strings.TrimSpace(part); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func splitWords(text string) []string {
	var words []string
	var current strings.Builder
	var quote rune
	started := false
	for _, character := range text {
		switch {
		case quote != 0 && character == quote:
			quote = 0
		case quote != 0:
			current.WriteRune(character)
		case character == '"' || character == '\'':
			quote = character
			started = true
		case unicode.IsSpace(character):
			if started || current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
				started = false
			}
		default:
			current.WriteRune(character)
		}
	}
	if started || current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

func joinWords(words []string) string {
	quoted := make([]string, 0, len(words))
	for _, word := range words {
		if word == "" || strings.ContainsAny(word, " \t\"'") {
			quoted = append(quoted, "\""+strings.ReplaceAll(word, "\"", "'")+"\"")
			continue
		}
		quoted = append(quoted, word)
	}
	return strings.Join(quoted, " ")
}

func describeValue(value model.Value) string {
	switch value.Kind {
	case model.ValueKindText:
		return value.Text
	case model.ValueKindFlag:
		if value.Flag {
			return "yes"
		}
		return "no"
	case model.ValueKindList:
		return strings.Join(value.List, itemSeparator+" ")
	case model.ValueKindDeferred:
		return "deferred"
	}
	return ""
}

func stateOf(value model.Value) string {
	switch value.Kind {
	case model.ValueKindDeferred:
		return "deferred"
	case model.ValueKindFlag:
		return "settled"
	case model.ValueKindText:
		if value.Text != "" {
			return "settled"
		}
	case model.ValueKindList:
		for _, item := range value.List {
			if item != "" {
				return "settled"
			}
		}
	}
	return "unanswered"
}
