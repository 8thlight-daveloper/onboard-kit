package encoder

import "onboard-kit/internal/model"

func (c *collector) outcomeSequence(outcomes []model.Outcome) node {
	return sequenceOf(outcomes, func(index int, outcome model.Outcome) node {
		return c.requireKeys(outcomeNode(outcome), indexed("outcomes", index))
	})
}

func outcomeNode(outcome model.Outcome) node {
	builder := &mappingBuilder{}
	builder.add("section", stringField(outcome.Section))
	builder.add("outcome", outcomeWordField(outcome.Kind))
	builder.add("on", stringField(string(outcome.On)))
	builder.add("sectionHash", stringField(outcome.SectionHash))
	builder.add("note", stringField(outcome.Note))
	builder.add("differently", stringField(outcome.Differently))
	return builder.build()
}

func outcomeWordField(kind model.OutcomeKind) node {
	if kind == model.OutcomeDeferred {
		return scalarNode(deferredMarker)
	}
	return stringField(string(kind))
}
