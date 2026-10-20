package decoder

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

func (c *collector) sectionsOf(node *yaml.Node, at location) []model.Section {
	var sections []model.Section
	c.items(node, at, codeEmptyItem, func(item *yaml.Node, place location) {
		sections = append(sections, c.sectionOf(item, place))
	})
	return sections
}

func (c *collector) sectionOf(node *yaml.Node, at location) model.Section {
	var section model.Section
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return section
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "id":
			section.ID = c.stringOf(found.value, found.at)
		case "title":
			section.Title = c.stringOf(found.value, found.at)
		case "summary":
			section.Summary = c.stringOf(found.value, found.at)
		case "parent":
			section.Parent = c.stringOf(found.value, found.at)
		case "owner":
			section.Owner = c.stringOf(found.value, found.at)
		case "timing":
			section.Timing = model.Timing(c.stringOf(found.value, found.at))
		case "enabled":
			section.Enabled = c.optionalFlagPointerOf(found.value, found.at)
		case "verification":
			section.Verification = c.stringOf(found.value, found.at)
		case "body":
			section.Body = c.stringOf(found.value, found.at)
		case "includeIf":
			section.IncludeIf = c.conditionOf(found.value, found.at)
		case "after":
			section.After = c.stringListOf(found.value, found.at)
		case "escalation":
			section.Escalation = c.stringListOf(found.value, found.at)
		case "variables":
			section.Variables = c.variablesOf(found.value, found.at)
		case "steps":
			section.Steps = c.stepsOf(found.value, found.at)
		case "source":
			section.Source = c.sourceOf(found.value, found.at)
		default:
			section.UnknownFields = c.keepUnknown(section.UnknownFields, found)
		}
	}
	return section
}

func (c *collector) sourceOf(node *yaml.Node, at location) *model.Source {
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return nil
	}
	source := &model.Source{}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "template":
			source.Template = c.stringOf(found.value, found.at)
		case "hash":
			source.Hash = c.stringOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return source
}
