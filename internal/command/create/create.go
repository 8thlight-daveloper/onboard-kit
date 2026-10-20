package create

import (
	"flag"
	"fmt"
	"io"
	"os"

	"onboard-kit/internal/command/planfile"
)

const (
	exitSuccess = 0
	exitProblem = 1
	exitUsage   = 2
)

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	options := flag.NewFlagSet("new", flag.ContinueOnError)
	options.SetOutput(stderr)
	templates := options.String("templates", "", "folder of section template files, and an optional plan-variables.yaml")
	output := options.String("out", "", "file to write, standard output when left out")
	templateSet := options.String("template-set", "local", "name recorded as the template set")
	generatedOn := options.String("generated-on", "", "date recorded in the plan, today when left out")
	if options.Parse(args) != nil {
		return exitUsage
	}
	if *templates == "" {
		fmt.Fprintln(stderr, "new: --templates is required")
		return exitUsage
	}
	plan, ok := planfile.FromTemplates(*templates, *templateSet, *generatedOn, stderr)
	if !ok {
		return exitProblem
	}
	data, ok := planfile.Encode(plan, stderr)
	if !ok {
		return exitProblem
	}
	return write(*output, data, stdout, stderr)
}

func write(path string, data []byte, stdout io.Writer, stderr io.Writer) int {
	if path == "" {
		if _, err := stdout.Write(data); err != nil {
			fmt.Fprintln(stderr, err)
			return exitProblem
		}
		return exitSuccess
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fmt.Fprintln(stderr, err)
		return exitProblem
	}
	return exitSuccess
}
