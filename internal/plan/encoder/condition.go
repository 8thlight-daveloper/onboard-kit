package encoder

import "onboard-kit/internal/model"

func conditionNode(condition *model.Condition) node {
	if condition == nil {
		return nil
	}
	builder := &mappingBuilder{}
	builder.add("all", sequenceOf(condition.All, func(_ int, clause model.Clause) node {
		return clauseNode(clause)
	}))
	return builder.build()
}

func clauseNode(clause model.Clause) node {
	builder := &mappingBuilder{}
	builder.add("question", stringField(clause.Question))
	builder.add("answer", answerTestNode(clause.Answer))
	return builder.build()
}

func answerTestNode(test model.AnswerTest) node {
	builder := &mappingBuilder{}
	builder.add("is", valueNode(test.Is))
	builder.add("isNot", valueNode(test.IsNot))
	builder.add("isOneOf", valueSequence(test.IsOneOf))
	builder.add("isNotOneOf", valueSequence(test.IsNotOneOf))
	if test.IsAnswered != nil {
		builder.add("isAnswered", booleanNode(*test.IsAnswered))
	}
	return builder.build()
}

func valueSequence(values []model.Value) node {
	return sequenceOf(values, func(_ int, value model.Value) node {
		return valueNode(value)
	})
}
