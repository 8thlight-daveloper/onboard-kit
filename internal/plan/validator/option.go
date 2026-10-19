package validator

import (
	"slices"

	"onboard-kit/internal/model"
)

func optionProblems(variable model.Variable, value model.Value) []model.Problem {
	if value.Kind != model.ValueKindText || slices.Contains(variable.Options, value.Text) {
		return nil
	}
	return []model.Problem{{
		Code:     "answer.not-an-option",
		Severity: model.SeverityError,
		Message:  "This answer is not one of the options.",
	}}
}
