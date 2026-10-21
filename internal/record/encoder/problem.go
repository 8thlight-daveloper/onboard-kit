package encoder

import (
	"strconv"

	"onboard-kit/internal/model"
)

const (
	codeEmptyAnswer           = "encode.empty-answer"
	codeEmptyItem             = "encode.empty-item"
	codeUnknownFieldCollision = "encode.unknown-field-collision"
	codeUnsupportedValue      = "encode.unsupported-value"
)

type collector struct {
	problems []model.Problem
}

func (c *collector) report(code string, path string, message string) {
	c.problems = append(c.problems, model.Problem{
		Code:     code,
		Severity: model.SeverityError,
		Message:  message,
		Detail:   map[string]string{"path": path},
	})
}

func (c *collector) requireKeys(item node, path string) node {
	if item == nil {
		c.report(codeEmptyItem, path, "an item would write no keys")
	}
	return item
}

func indexed(path string, index int) string {
	return path + "[" + strconv.Itoa(index) + "]"
}
