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
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return model.Clause{}
	}
	found := c.entries(mapping, at)
	if hasEntry(found, "section") {
		return c.sectionClauseOf(found, at)
	}
	return c.questionClauseOf(found)
}

func hasEntry(found []entry, name string) bool {
	for _, candidate := range found {
		if candidate.name == name {
			return true
		}
	}
	return false
}

func (c *collector) questionClauseOf(found []entry) model.Clause {
	var clause model.Clause
	for _, item := range found {
		switch item.name {
		case "question":
			clause.Question = c.stringOf(item.value, item.at)
		case "answer":
			clause.Answer = c.answerTestOf(item.value, item.at)
		default:
			c.rejectUnknown(item)
		}
	}
	return clause
}

func (c *collector) sectionClauseOf(found []entry, at location) model.Clause {
	var clause model.Clause
	switched := false
	for _, item := range found {
		switch item.name {
		case "section":
			clause.Section = c.stringOf(item.value, item.at)
		case "enabled":
			clause.Enabled, switched = c.switchOf(item)
		default:
			c.rejectUnknown(item)
		}
	}
	if !switched {
		c.report(codeWrongType, at)
	}
	return clause
}

func (c *collector) switchOf(item entry) (bool, bool) {
	reported := len(c.findings)
	flag, present := c.optionalFlagOf(item.value, item.at)
	return flag, present || len(c.findings) > reported
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
		case "isSettled":
			test.IsSettled = c.optionalFlagPointerOf(found.value, found.at)
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
