package loader

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

func (c *collector) stepsOf(node *yaml.Node, at location) []model.Step {
	var steps []model.Step
	c.items(node, at, codeEmptyItem, func(item *yaml.Node, place location) {
		steps = append(steps, c.stepOf(item, place))
	})
	return steps
}

func (c *collector) stepOf(node *yaml.Node, at location) model.Step {
	var step model.Step
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return step
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "id":
			step.ID = c.stringOf(found.value, found.at)
		case "name":
			step.Name = c.stringOf(found.value, found.at)
		case "notes":
			step.Notes = c.stringOf(found.value, found.at)
		case "owner":
			step.Owner = c.stringOf(found.value, found.at)
		case "timing":
			step.Timing = model.Timing(c.stringOf(found.value, found.at))
		case "open":
			step.Open = c.stringOf(found.value, found.at)
		case "directory":
			step.Directory = c.stringOf(found.value, found.at)
		case "escalation":
			step.Escalation = c.stringListOf(found.value, found.at)
		case "run":
			step.Run = c.stringListOf(found.value, found.at)
		case "env":
			step.Env = c.stringMapOf(found.value, found.at)
		case "checks":
			step.Checks = c.checksOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return step
}

func (c *collector) checksOf(node *yaml.Node, at location) []model.Check {
	var checks []model.Check
	c.items(node, at, codeEmptyItem, func(item *yaml.Node, place location) {
		checks = append(checks, c.checkOf(item, place))
	})
	return checks
}

func (c *collector) checkOf(node *yaml.Node, at location) model.Check {
	var check model.Check
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return check
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "run":
			check.Run = c.stringListOf(found.value, found.at)
		case "ask":
			check.Ask = c.stringOf(found.value, found.at)
		case "env":
			check.Env = c.stringOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return check
}
