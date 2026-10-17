package encoder

import (
	"strconv"

	"onboard-kit/internal/model"
)

func (c *collector) planNode(plan model.Plan) node {
	builder := &mappingBuilder{}
	builder.add("schemaVersion", scalarNode(strconv.Itoa(plan.SchemaVersion)))
	builder.add("provenance", provenanceNode(plan.Provenance))
	builder.add("variables", c.variableSequence(plan.Variables, "variables"))
	builder.add("answers", c.answersNode(plan.Answers))
	builder.add("sections", c.sectionSequence(plan.Sections))
	c.addUnknownFields(builder, plan.UnknownFields, "", isPlanKey)
	return builder.build()
}

func provenanceNode(provenance model.Provenance) node {
	builder := &mappingBuilder{}
	builder.add("templateSet", stringField(provenance.TemplateSet))
	builder.add("generatedOn", stringField(string(provenance.GeneratedOn)))
	return builder.build()
}

func isPlanKey(key string) bool {
	switch key {
	case "schemaVersion", "provenance", "variables", "answers", "sections":
		return true
	}
	return false
}
