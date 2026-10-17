package decoder

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

func (c *collector) variablesOf(node *yaml.Node, at location) []model.Variable {
	var variables []model.Variable
	c.items(node, at, codeEmptyItem, func(item *yaml.Node, place location) {
		variables = append(variables, c.variableOf(item, place))
	})
	return variables
}

func (c *collector) variableOf(node *yaml.Node, at location) model.Variable {
	var variable model.Variable
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return variable
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "name":
			variable.Name = c.stringOf(found.value, found.at)
		case "description":
			variable.Description = c.stringOf(found.value, found.at)
		case "type":
			variable.Type = model.VariableType(c.stringOf(found.value, found.at))
		case "options":
			variable.Options = c.stringListOf(found.value, found.at)
		case "pattern":
			variable.Pattern = c.stringOf(found.value, found.at)
		case "patternHint":
			variable.PatternHint = c.stringOf(found.value, found.at)
		case "required":
			variable.Required = c.flagOf(found.value, found.at)
		case "multiple":
			variable.Multiple = c.flagOf(found.value, found.at)
		case "default":
			variable.Default = c.optionalValueOf(found.value, found.at)
		case "phase":
			variable.Phase = model.Phase(c.stringOf(found.value, found.at))
		case "askIf":
			variable.AskIf = c.conditionOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return variable
}
