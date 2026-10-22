package assemble

import (
	"flag"
	"fmt"
	"io"
	"os"

	"onboard-kit/internal/command/planfile"
	"onboard-kit/internal/command/report"
	"onboard-kit/internal/document/assembler"
)

const (
	exitSuccess = 0
	exitProblem = 1
	exitUsage   = 2
)

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	options := flag.NewFlagSet("assemble", flag.ContinueOnError)
	options.SetOutput(stderr)
	planPath := options.String("plan", "", "plan file to assemble")
	output := options.String("out", "", "file to write, standard output when left out")
	if options.Parse(args) != nil {
		return exitUsage
	}
	if *planPath == "" {
		fmt.Fprintln(stderr, "assemble: --plan is required")
		return exitUsage
	}
	plan, ok := planfile.Open(*planPath, stderr)
	if !ok {
		return exitProblem
	}
	if problems := planfile.Ambiguities(plan); len(problems) > 0 {
		report.Problems(stderr, *planPath, problems)
		return exitProblem
	}
	document, gaps := assembler.New().Assemble(plan)
	report.Gaps(stderr, *planPath, gaps)
	return write(*output, document, stdout, stderr)
}

func write(path string, document string, stdout io.Writer, stderr io.Writer) int {
	if path == "" {
		if _, err := io.WriteString(stdout, document); err != nil {
			fmt.Fprintln(stderr, err)
			return exitProblem
		}
		return exitSuccess
	}
	if err := os.WriteFile(path, []byte(document), 0o644); err != nil {
		fmt.Fprintln(stderr, err)
		return exitProblem
	}
	return exitSuccess
}
