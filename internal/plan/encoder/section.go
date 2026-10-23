package encoder

import "onboard-kit/internal/model"

func (c *collector) sectionSequence(sections []model.Section) node {
	return sequenceOf(sections, func(index int, section model.Section) node {
		return c.sectionNode(section, indexed("sections", index))
	})
}

func (c *collector) sectionNode(section model.Section, path string) node {
	builder := &mappingBuilder{}
	builder.add("id", stringField(section.ID))
	builder.add("title", stringField(section.Title))
	builder.add("summary", stringField(section.Summary))
	builder.add("includeIf", c.conditionNode(section.IncludeIf, path+".includeIf"))
	builder.add("parent", stringField(section.Parent))
	builder.add("bootstrap", flagField(section.Bootstrap))
	builder.add("timing", stringField(string(section.Timing)))
	if section.Enabled != nil {
		builder.add("enabled", booleanNode(*section.Enabled))
	}
	builder.add("after", listField(section.After))
	builder.add("requires", listField(section.Requires))
	builder.add("owner", stringField(section.Owner))
	builder.add("escalation", listField(section.Escalation))
	builder.add("verification", stringField(section.Verification))
	builder.add("body", stringField(section.Body))
	builder.add("variables", c.variableSequence(section.Variables, path+".variables"))
	builder.add("steps", c.stepSequence(section.Steps, path+".steps"))
	builder.add("source", sourceNode(section.Source))
	c.addUnknownFields(builder, section.UnknownFields, path+".", isSectionKey)
	return c.requireKeys(builder.build(), path)
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
	case "id", "title", "summary", "includeIf", "parent", "bootstrap", "timing", "enabled", "after", "requires", "owner", "escalation", "verification", "body", "variables", "steps", "source":
		return true
	}
	return false
}
