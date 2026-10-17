package decoder

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

func (c *collector) answersOf(node *yaml.Node, at location) model.Answers {
	var answers model.Answers
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return answers
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "plan":
			answers.Plan = c.answerMapOf(found.value, found.at)
		case "sections":
			answers.Sections = c.sectionAnswersOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return answers
}

func (c *collector) answerMapOf(node *yaml.Node, at location) map[string]model.Value {
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return nil
	}
	values := map[string]model.Value{}
	for _, found := range c.entries(mapping, at) {
		values[found.name] = c.answerOf(found.value, found.at)
	}
	return values
}

func (c *collector) sectionAnswersOf(node *yaml.Node, at location) map[string]map[string]model.Value {
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return nil
	}
	scopes := map[string]map[string]model.Value{}
	for _, found := range c.entries(mapping, at) {
		scope := c.answerMapOf(found.value, found.at)
		if scope == nil {
			scope = map[string]model.Value{}
		}
		scopes[found.name] = scope
	}
	return scopes
}
