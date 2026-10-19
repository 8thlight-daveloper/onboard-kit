package encoder

import "onboard-kit/internal/model"

func (c *collector) conditionNode(condition *model.Condition, path string) node {
	if condition == nil {
		return nil
	}
	builder := &mappingBuilder{}
	builder.add("all", sequenceOf(condition.All, func(index int, clause model.Clause) node {
		return c.clauseNode(clause, indexed(path+".all", index))
	}))
	return builder.build()
}

func (c *collector) clauseNode(clause model.Clause, path string) node {
	builder := &mappingBuilder{}
	builder.add("question", stringField(clause.Question))
	builder.add("answer", c.answerTestNode(clause.Answer, path+".answer"))
	return c.requireKeys(builder.build(), path)
}

func (c *collector) answerTestNode(test model.AnswerTest, path string) node {
	builder := &mappingBuilder{}
	builder.add("is", valueNode(test.Is))
	builder.add("isNot", valueNode(test.IsNot))
	builder.add("isOneOf", c.valueSequence(test.IsOneOf, path+".isOneOf"))
	builder.add("isNotOneOf", c.valueSequence(test.IsNotOneOf, path+".isNotOneOf"))
	if test.IsSettled != nil {
		builder.add("isSettled", booleanNode(*test.IsSettled))
	}
	return builder.build()
}

func (c *collector) valueSequence(values []model.Value, path string) node {
	return sequenceOf(values, func(index int, value model.Value) node {
		return c.answerNode(value, indexed(path, index))
	})
}
