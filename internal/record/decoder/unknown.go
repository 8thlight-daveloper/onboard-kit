package decoder

import (
	"strconv"

	"go.yaml.in/yaml/v3"
)

func (c *collector) keepUnknown(fields map[string]any, found entry) map[string]any {
	if fields == nil {
		fields = map[string]any{}
	}
	fields[found.name] = c.unknownOf(found.value, found.at)
	return fields
}

func (c *collector) unknownOf(node *yaml.Node, at location) any {
	if !c.prepare(node, at, false) {
		return nil
	}
	switch node.Kind {
	case yaml.SequenceNode:
		return c.unknownListOf(node, at)
	case yaml.MappingNode:
		return c.unknownMapOf(node, at)
	}
	return unknownScalarOf(node)
}

func unknownScalarOf(node *yaml.Node) any {
	if !isPlain(node) {
		return node.Value
	}
	switch classify(node) {
	case scalarNull:
		return nil
	case scalarBoolean:
		return node.Value == "true"
	case scalarInteger:
		number, _ := strconv.ParseInt(node.Value, 10, 64)
		return number
	case scalarDecimal:
		number, _ := strconv.ParseFloat(node.Value, 64)
		return number
	}
	return node.Value
}

func (c *collector) unknownListOf(sequence *yaml.Node, at location) []any {
	list := make([]any, len(sequence.Content))
	for index, item := range sequence.Content {
		list[index] = c.unknownOf(item, at.item(index, item))
	}
	return list
}

func (c *collector) unknownMapOf(mapping *yaml.Node, at location) map[string]any {
	values := map[string]any{}
	for _, found := range c.entries(mapping, at) {
		values[found.name] = c.unknownOf(found.value, found.at)
	}
	return values
}
