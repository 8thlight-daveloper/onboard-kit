package loader

import "go.yaml.in/yaml/v3"

func (c *collector) stringOf(node *yaml.Node, at location) string {
	if !c.prepare(node, at, true) || isNull(node) {
		return ""
	}
	if node.Kind != yaml.ScalarNode || (isPlain(node) && classify(node) == scalarBoolean) {
		c.report(codeWrongType, at)
		return ""
	}
	return node.Value
}

func (c *collector) flagOf(node *yaml.Node, at location) bool {
	flag, _ := c.optionalFlagOf(node, at)
	return flag
}

func (c *collector) optionalFlagOf(node *yaml.Node, at location) (bool, bool) {
	if !c.prepare(node, at, false) || isNull(node) {
		return false, false
	}
	if !isPlain(node) || classify(node) != scalarBoolean {
		c.report(codeWrongType, at)
		return false, false
	}
	return node.Value == "true", true
}

func (c *collector) stringListOf(node *yaml.Node, at location) []string {
	var list []string
	c.items(node, at, codeEmptyItem, func(item *yaml.Node, place location) {
		list = append(list, c.stringOf(item, place))
	})
	return list
}

func (c *collector) stringMapOf(node *yaml.Node, at location) map[string]string {
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return nil
	}
	values := map[string]string{}
	for _, found := range c.entries(mapping, at) {
		values[found.name] = c.stringOf(found.value, found.at)
	}
	return values
}
