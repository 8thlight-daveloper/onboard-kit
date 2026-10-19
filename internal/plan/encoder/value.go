package encoder

import "onboard-kit/internal/model"

func valueNode(value model.Value) node {
	switch value.Kind {
	case model.ValueKindText:
		return textNode(value.Text)
	case model.ValueKindFlag:
		return booleanNode(value.Flag)
	case model.ValueKindList:
		return valueList(value.List)
	case model.ValueKindDeferred:
		return scalarNode(deferredMarker)
	}
	return nil
}

func valueList(items []string) node {
	if len(items) == 0 {
		return scalarNode("[]")
	}
	return listField(items)
}

func (c *collector) answersNode(answers model.Answers) node {
	builder := &mappingBuilder{}
	builder.add("plan", c.answerMap(answers.Plan, "answers.plan"))
	builder.add("sections", c.sectionAnswersNode(answers.Sections))
	return builder.build()
}

func (c *collector) sectionAnswersNode(sections map[string]map[string]model.Value) node {
	builder := &mappingBuilder{}
	for _, id := range sortedKeys(sections) {
		builder.add(id, c.answerMap(sections[id], "answers.sections."+id))
	}
	return builder.build()
}

func (c *collector) answerMap(answers map[string]model.Value, path string) node {
	builder := &mappingBuilder{}
	for _, name := range sortedKeys(answers) {
		builder.add(name, c.answerNode(answers[name], path+"."+name))
	}
	return builder.build()
}

func (c *collector) answerNode(value model.Value, path string) node {
	answer := valueNode(value)
	if answer == nil {
		c.report(codeEmptyAnswer, path, "an answer has no value")
	}
	return answer
}
