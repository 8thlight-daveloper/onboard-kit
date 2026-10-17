package decoder

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

func (c *collector) answerOf(node *yaml.Node, at location) model.Value {
	if isNull(node) {
		c.report(codeEmptyAnswer, at)
		return model.Value{}
	}
	return c.optionalValueOf(node, at)
}

func (c *collector) optionalValueOf(node *yaml.Node, at location) model.Value {
	if !c.prepare(node, at, true) || isNull(node) {
		return model.Value{}
	}
	switch node.Kind {
	case yaml.SequenceNode:
		return c.listValueOf(node, at)
	case yaml.ScalarNode:
		return scalarValueOf(node)
	}
	c.report(codeWrongType, at)
	return model.Value{}
}

func scalarValueOf(node *yaml.Node) model.Value {
	if isPlain(node) {
		switch {
		case node.Value == unknownYetText:
			return model.Value{Kind: model.ValueKindUnknown}
		case classify(node) == scalarBoolean:
			return model.Value{Kind: model.ValueKindFlag, Flag: node.Value == "true"}
		}
	}
	return model.Value{Kind: model.ValueKindText, Text: node.Value}
}

func (c *collector) listValueOf(node *yaml.Node, at location) model.Value {
	list := []string{}
	c.eachItem(node, at, codeEmptyItem, func(item *yaml.Node, place location) {
		list = append(list, c.stringOf(item, place))
	})
	return model.Value{Kind: model.ValueKindList, List: list}
}

func (c *collector) answerListOf(node *yaml.Node, at location) []model.Value {
	var values []model.Value
	c.items(node, at, codeEmptyAnswer, func(item *yaml.Node, place location) {
		values = append(values, c.optionalValueOf(item, place))
	})
	return values
}
