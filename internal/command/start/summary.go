package start

import (
	"fmt"
	"strings"

	"onboard-kit/internal/plan/refresher"
)

func summarize(changes refresher.Report) string {
	counts := []struct {
		count int
		text  string
	}{
		{len(changes.NewQuestions), "new questions"},
		{len(changes.RemovedQuestions), "questions removed"},
		{len(changes.AnswersTaken), "answers updated from the organization"},
		{len(changes.AnswersToReview), "answers to review"},
		{len(changes.NewSections), "new sections"},
		{len(changes.RemovedSections), "sections removed"},
	}
	var parts []string
	for _, entry := range counts {
		if entry.count > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", entry.count, entry.text))
		}
	}
	if len(parts) == 0 {
		return "Your plan is up to date with the organization's plan"
	}
	return "Refreshed from the organization's plan: " + strings.Join(parts, ", ")
}
