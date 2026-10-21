package encoder

import (
	"strconv"

	"onboard-kit/internal/model"
)

const recordKind = "record"

func (c *collector) recordNode(record model.Record) node {
	builder := &mappingBuilder{}
	builder.add("schemaVersion", scalarNode(strconv.Itoa(record.SchemaVersion)))
	builder.add("kind", scalarNode(recordKind))
	builder.add("account", stringField(record.Account))
	builder.add("startDate", stringField(string(record.StartDate)))
	builder.add("plan", planReferenceNode(record.Plan))
	builder.add("answers", c.answersNode(record.Answers))
	builder.add("outcomes", c.outcomeSequence(record.Outcomes))
	c.addUnknownFields(builder, record.UnknownFields, "", isRecordKey)
	return builder.build()
}

func planReferenceNode(reference model.PlanReference) node {
	builder := &mappingBuilder{}
	builder.add("id", stringField(reference.ID))
	builder.add("hash", stringField(reference.Hash))
	builder.add("url", stringField(reference.URL))
	return builder.build()
}

func isRecordKey(key string) bool {
	switch key {
	case "schemaVersion", "kind", "account", "startDate", "plan", "answers", "outcomes":
		return true
	}
	return false
}
