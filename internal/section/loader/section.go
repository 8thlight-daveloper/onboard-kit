package loader

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

func (c *collector) sectionFrom(root *yaml.Node) model.Section {
	var section model.Section
	c.open(root, rootLocation())
	for _, found := range c.entries(root, rootLocation()) {
		switch found.name {
		case versionKey:
			c.open(found.value, found.at)
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
		default:
			c.rejectUnknown(found)
		}
	}
	return section
}
