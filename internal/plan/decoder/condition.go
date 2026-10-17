package decoder

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

func (c *collector) conditionOf(node *yaml.Node, at location) *model.Condition {
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return nil
	}
	condition := &model.Condition{}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "all":
			condition.All = c.clausesOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return condition
}

func (c *collector) clausesOf(node *yaml.Node, at location) []model.Clause {
	var clauses []model.Clause
	c.items(node, at, codeEmptyItem, func(item *yaml.Node, place location) {
		clauses = append(clauses, c.clauseOf(item, place))
	})
	return clauses
}

func (c *collector) clauseOf(node *yaml.Node, at location) model.Clause {
	var clause model.Clause
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return clause
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "question":
			clause.Question = c.stringOf(found.value, found.at)
		case "answer":
			clause.Answer = c.answerTestOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return clause
}

func (c *collector) answerTestOf(node *yaml.Node, at location) model.AnswerTest {
	var test model.AnswerTest
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return test
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "is":
			test.Is = c.optionalValueOf(found.value, found.at)
		case "isNot":
			test.IsNot = c.optionalValueOf(found.value, found.at)
		case "isOneOf":
			test.IsOneOf = c.answerListOf(found.value, found.at)
		case "isNotOneOf":
			test.IsNotOneOf = c.answerListOf(found.value, found.at)
		case "isAnswered":
			test.IsAnswered = c.optionalFlagPointerOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return test
}

func (c *collector) optionalFlagPointerOf(node *yaml.Node, at location) *bool {
	flag, present := c.optionalFlagOf(node, at)
	if !present {
		return nil
	}
	return &flag
}
