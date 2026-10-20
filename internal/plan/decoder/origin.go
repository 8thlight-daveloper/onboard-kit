package decoder

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

func (c *collector) originOf(node *yaml.Node, at location) model.Origin {
	var origin model.Origin
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return origin
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "planId":
			origin.PlanID = c.stringOf(found.value, found.at)
		case "source":
			origin.Source = c.stringOf(found.value, found.at)
		case "contentHash":
			origin.ContentHash = c.stringOf(found.value, found.at)
		case "refreshedOn":
			origin.RefreshedOn = model.Date(c.stringOf(found.value, found.at))
		default:
			c.rejectUnknown(found)
		}
	}
	return origin
}
