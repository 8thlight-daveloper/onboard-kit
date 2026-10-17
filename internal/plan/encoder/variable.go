package encoder

import "onboard-kit/internal/model"

func (c *collector) variableNode(variable model.Variable, path string) node {
	builder := &mappingBuilder{}
	builder.add("name", stringField(variable.Name))
	builder.add("description", stringField(variable.Description))
	builder.add("type", stringField(string(variable.Type)))
	builder.add("multiple", flagField(variable.Multiple))
	builder.add("options", listField(variable.Options))
	builder.add("pattern", stringField(variable.Pattern))
	builder.add("patternHint", stringField(variable.PatternHint))
	builder.add("required", flagField(variable.Required))
	builder.add("default", valueNode(variable.Default))
	builder.add("phase", stringField(string(variable.Phase)))
	builder.add("askIf", c.conditionNode(variable.AskIf, path+".askIf"))
	return c.requireKeys(builder.build(), path)
}

func (c *collector) variableSequence(variables []model.Variable, path string) node {
	return sequenceOf(variables, func(index int, variable model.Variable) node {
		return c.variableNode(variable, indexed(path, index))
	})
}
