package validator

import (
	"regexp"

	"onboard-kit/internal/model"
)

func patternProblems(variable model.Variable, value model.Value) []model.Problem {
	if variable.Pattern == "" || value.Kind != model.ValueKindText {
		return nil
	}
	pattern, err := regexp.Compile(variable.Pattern)
	if err != nil || pattern.MatchString(value.Text) {
		return nil
	}
	return []model.Problem{{
		Code:     "answer.pattern-mismatch",
		Severity: model.SeverityError,
		Message:  patternMessage(variable),
	}}
}

func patternMessage(variable model.Variable) string {
	message := "This answer does not match the required pattern."
	if variable.PatternHint != "" {
		message += " " + variable.PatternHint
	}
	return message
}
