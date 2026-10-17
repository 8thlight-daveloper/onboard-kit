package encoder

import (
	"strconv"

	"onboard-kit/internal/model"
)

func (c *collector) sectionSequence(sections []model.Section) node {
	return sequenceOf(sections, func(index int, section model.Section) node {
		return c.sectionNode(section, "sections["+strconv.Itoa(index)+"]")
	})
}

func (c *collector) sectionNode(section model.Section, path string) node {
	builder := &mappingBuilder{}
	builder.add("id", stringField(section.ID))
	builder.add("title", stringField(section.Title))
	builder.add("summary", stringField(section.Summary))
	builder.add("includeIf", conditionNode(section.IncludeIf))
	builder.add("timing", stringField(string(section.Timing)))
	builder.add("after", listField(section.After))
	builder.add("owner", stringField(section.Owner))
	builder.add("escalation", listField(section.Escalation))
	builder.add("verification", stringField(section.Verification))
	builder.add("body", stringField(section.Body))
	builder.add("variables", variableSequence(section.Variables))
	builder.add("steps", c.stepSequence(section.Steps, path+".steps"))
	builder.add("source", sourceNode(section.Source))
	c.addUnknownFields(builder, section.UnknownFields, path+".", isSectionKey)
	return builder.build()
}

func sourceNode(source *model.Source) node {
	if source == nil {
		return nil
	}
	builder := &mappingBuilder{}
	builder.add("template", stringField(source.Template))
	builder.add("hash", stringField(source.Hash))
	return builder.build()
}

func isSectionKey(key string) bool {
	switch key {
	case "id", "title", "summary", "includeIf", "timing", "after", "owner", "escalation", "verification", "body", "variables", "steps", "source":
		return true
	}
	return false
}
