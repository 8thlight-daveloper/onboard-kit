package validator

import (
	"strconv"
	"strings"

	"onboard-kit/internal/model"
)

const minimumContactWays = 2

func personProblems(value model.Value) []model.Problem {
	switch value.Kind {
	case model.ValueKindText:
		return thinContactProblems(value.Text, nil)
	case model.ValueKindList:
		return peopleListProblems(value.List)
	default:
		return nil
	}
}

func peopleListProblems(items []string) []model.Problem {
	problems := []model.Problem{}
	present := 0
	for position, item := range items {
		if item == "" {
			continue
		}
		present++
		detail := map[string]string{"item": strconv.Itoa(position)}
		problems = append(problems, thinContactProblems(item, detail)...)
	}
	if present == 1 {
		problems = append(problems, model.Problem{
			Code:     "answer.single-person",
			Severity: model.SeverityWarning,
			Message:  "Only one person is listed, so there is a single point of failure.",
		})
	}
	return problems
}

func thinContactProblems(line string, detail map[string]string) []model.Problem {
	if countContactWays(line) >= minimumContactWays {
		return nil
	}
	return []model.Problem{{
		Code:     "answer.thin-contact",
		Severity: model.SeverityWarning,
		Message:  "This person has fewer than two ways to be reached.",
		Detail:   detail,
	}}
}

func countContactWays(line string) int {
	parts := strings.Split(line, ",")
	ways := 0
	for _, part := range parts[1:] {
		if strings.TrimSpace(part) != "" {
			ways++
		}
	}
	return ways
}
