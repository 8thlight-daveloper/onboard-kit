package report

import (
	"fmt"
	"io"

	"onboard-kit/internal/model"
)

func Problems(writer io.Writer, source string, problems []model.Problem) {
	lines := make([]string, 0, len(problems))
	for _, problem := range problems {
		lines = append(lines, line(source, problem))
	}
	for _, text := range lines {
		fmt.Fprintln(writer, text)
	}
}

func line(source string, problem model.Problem) string {
	place := problem.Detail["path"]
	if place == "" {
		place = problem.Question
	}
	if place == "" {
		return fmt.Sprintf("%s: %s: %s", source, problem.Code, problem.Message)
	}
	return fmt.Sprintf("%s: %s: %s: %s", source, place, problem.Code, problem.Message)
}

func Gaps(writer io.Writer, source string, gaps []model.Problem) {
	for _, gap := range gaps {
		fmt.Fprintf(writer, "%s: %s: %s: %s\n", source, gap.Severity, gap.Question, gap.Message)
	}
}
