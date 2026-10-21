package decoder

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

func (c *collector) outcomesOf(node *yaml.Node, at location) []model.Outcome {
	var outcomes []model.Outcome
	c.items(node, at, codeEmptyItem, func(item *yaml.Node, place location) {
		outcomes = append(outcomes, c.outcomeOf(item, place))
	})
	return outcomes
}

func (c *collector) outcomeOf(node *yaml.Node, at location) model.Outcome {
	var outcome model.Outcome
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return outcome
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "section":
			outcome.Section = c.stringOf(found.value, found.at)
		case "outcome":
			outcome.Kind = model.OutcomeKind(c.stringOf(found.value, found.at))
		case "on":
			outcome.On = model.Timestamp(c.stringOf(found.value, found.at))
		case "sectionHash":
			outcome.SectionHash = c.stringOf(found.value, found.at)
		case "note":
			outcome.Note = c.stringOf(found.value, found.at)
		case "differently":
			outcome.Differently = c.stringOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return outcome
}
