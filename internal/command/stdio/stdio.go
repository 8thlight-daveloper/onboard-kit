package stdio

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"onboard-kit/internal/command/planfile"
	"onboard-kit/internal/model"
)

const (
	exitSuccess = 0
	exitProblem = 1
	exitUsage   = 2
)

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	options := flag.NewFlagSet("stdio", flag.ContinueOnError)
	options.SetOutput(stderr)
	planPath := options.String("plan", "", "plan file to open")
	templates := options.String("templates", "", "folder of section templates to start a new plan from")
	templateSet := options.String("template-set", "local", "name recorded as the template set")
	generatedOn := options.String("generated-on", "", "date recorded in a new plan, today when left out")
	if options.Parse(args) != nil {
		return exitUsage
	}
	plan, ok := startPlan(*planPath, *templates, *templateSet, *generatedOn, stderr)
	if !ok {
		return exitProblem
	}
	return serve(newSession(plan), stdin, stdout, stderr)
}

func startPlan(planPath string, templates string, templateSet string, generatedOn string, stderr io.Writer) (model.Plan, bool) {
	switch {
	case planPath != "" && templates != "":
		fmt.Fprintln(stderr, "stdio: give --plan or --templates, not both")
		return model.Plan{}, false
	case planPath != "":
		return planfile.Open(planPath, stderr)
	case templates != "":
		return planfile.FromTemplates(templates, templateSet, generatedOn, stderr)
	}
	fmt.Fprintln(stderr, "stdio: --plan or --templates is required")
	return model.Plan{}, false
}

func serve(current *session, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	scanner := bufio.NewScanner(stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(false)
	for scanner.Scan() {
		if err := encoder.Encode(current.handle(scanner.Bytes())); err != nil {
			fmt.Fprintln(stderr, err)
			return exitProblem
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(stderr, err)
		return exitProblem
	}
	return exitSuccess
}
