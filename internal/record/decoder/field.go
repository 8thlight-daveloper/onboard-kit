package decoder

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
