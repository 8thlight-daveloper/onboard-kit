package encoder

import "onboard-kit/internal/model"

const (
	codeEmptyAnswer           = "encode.empty-answer"
	codeAmbiguousCheck        = "encode.ambiguous-check"
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
