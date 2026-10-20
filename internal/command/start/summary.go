package start

import (
	"fmt"
	"io"

	"onboard-kit/internal/plan/refresher"
)

func summarize(stdout io.Writer, changes refresher.Report) {
	lines := []struct {
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
	printed := false
	for _, line := range lines {
		if line.count == 0 {
			continue
		}
		if !printed {
			fmt.Fprintln(stdout, "Refreshed from the organization's plan:")
			printed = true
		}
		fmt.Fprintf(stdout, "  %d %s\n", line.count, line.text)
	}
	if !printed {
		fmt.Fprintln(stdout, "Your plan is up to date with the organization's plan")
	}
}
