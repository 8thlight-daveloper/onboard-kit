package decoder

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

func (c *collector) planReferenceOf(node *yaml.Node, at location) model.PlanReference {
	var reference model.PlanReference
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return reference
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "id":
			reference.ID = c.stringOf(found.value, found.at)
		case "hash":
			reference.Hash = c.stringOf(found.value, found.at)
		case "url":
			reference.URL = c.stringOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return reference
}
