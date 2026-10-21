package decoder

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

func (c *collector) planFrom(root *yaml.Node) model.Plan {
	plan := model.Plan{SchemaVersion: supportedVersion}
	c.open(root, rootLocation())
	for _, found := range c.entries(root, rootLocation()) {
		switch found.name {
		case versionKey:
			c.open(found.value, found.at)
		case "id":
			plan.ID = c.stringOf(found.value, found.at)
		case "origin":
			plan.Origin = c.originOf(found.value, found.at)
		case "directory":
			plan.Directory = c.directoryOf(found.value, found.at)
		case "provenance":
			plan.Provenance = c.provenanceOf(found.value, found.at)
		case "variables":
			plan.Variables = c.variablesOf(found.value, found.at)
		case "answers":
			plan.Answers = c.answersOf(found.value, found.at)
		case "sections":
			plan.Sections = c.sectionsOf(found.value, found.at)
		default:
			plan.UnknownFields = c.keepUnknown(plan.UnknownFields, found)
		}
	}
	return plan
}

func (c *collector) provenanceOf(node *yaml.Node, at location) model.Provenance {
	var provenance model.Provenance
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return provenance
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "templateSet":
			provenance.TemplateSet = c.stringOf(found.value, found.at)
		case "generatedOn":
			provenance.GeneratedOn = model.Date(c.stringOf(found.value, found.at))
		default:
			c.rejectUnknown(found)
		}
	}
	return provenance
}
