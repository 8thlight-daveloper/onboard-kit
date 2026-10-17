package encoder

import "onboard-kit/internal/model"

func variableNode(variable model.Variable) node {
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
	builder.add("askIf", conditionNode(variable.AskIf))
	return builder.build()
}

func variableSequence(variables []model.Variable) node {
	return sequenceOf(variables, func(_ int, variable model.Variable) node {
		return variableNode(variable)
	})
}
