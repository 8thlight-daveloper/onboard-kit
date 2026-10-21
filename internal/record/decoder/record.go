package decoder

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

func (c *collector) recordFrom(root *yaml.Node) model.Record {
	record := model.Record{SchemaVersion: supportedVersion}
	c.open(root, rootLocation())
	for _, found := range c.entries(root, rootLocation()) {
		switch found.name {
		case versionKey, kindKey:
			c.open(found.value, found.at)
		case "account":
			record.Account = c.stringOf(found.value, found.at)
		case "startDate":
			record.StartDate = model.Date(c.stringOf(found.value, found.at))
		case "plan":
			record.Plan = c.planReferenceOf(found.value, found.at)
		case "answers":
			record.Answers = c.answersOf(found.value, found.at)
		case "outcomes":
			record.Outcomes = c.outcomesOf(found.value, found.at)
		default:
			record.UnknownFields = c.keepUnknown(record.UnknownFields, found)
		}
	}
	return record
}
