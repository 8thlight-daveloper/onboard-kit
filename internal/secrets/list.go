package secrets

import (
	"strconv"

	"onboard-kit/internal/model"
)

func listProblems(items []string) []model.Problem {
	var problems []model.Problem
	for position, item := range items {
		detail := map[string]string{"item": strconv.Itoa(position)}
		problems = append(problems, textProblems(item, detail)...)
	}
	return problems
}
