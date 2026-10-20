package walk

import (
	"flag"
	"fmt"
	"io"
	"os"

	"onboard-kit/internal/command/planfile"
	"onboard-kit/internal/model"
	"onboard-kit/internal/tui"
)

const (
	exitSuccess = 0
	exitProblem = 1
	exitUsage   = 2
)

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	options := flag.NewFlagSet("walk", flag.ContinueOnError)
	options.SetOutput(stderr)
	planPath := options.String("plan", "", "plan file to open, or to create when it does not exist")
	templates := options.String("templates", "", "folder of section templates to start a new plan from")
	templateSet := options.String("template-set", "local", "name recorded as the template set")
	generatedOn := options.String("generated-on", "", "date recorded in a new plan, today when left out")
	if options.Parse(args) != nil {
		return exitUsage
	}
	if *planPath == "" {
		fmt.Fprintln(stderr, "walk: --plan is required")
		return exitUsage
	}
	plan, ok := startPlan(*planPath, *templates, *templateSet, *generatedOn, stderr)
	if !ok {
		return exitProblem
	}
	save := func(plan model.Plan) error {
		if !planfile.Save(*planPath, plan, io.Discard) {
			return fmt.Errorf("the plan could not be written to %s", *planPath)
		}
		return nil
	}
	if err := tui.Run(tui.NewBackend(), plan, save, stdin, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return exitProblem
	}
	return exitSuccess
}

func startPlan(planPath string, templates string, templateSet string, generatedOn string, stderr io.Writer) (model.Plan, bool) {
	if _, err := os.Stat(planPath); err == nil {
		return planfile.Open(planPath, stderr)
	}
	if templates == "" {
		fmt.Fprintf(stderr, "walk: %s does not exist, give --templates to start a new plan\n", planPath)
		return model.Plan{}, false
	}
	plan, ok := planfile.FromTemplates(templates, templateSet, generatedOn, stderr)
	if !ok {
		return model.Plan{}, false
	}
	if !planfile.Save(planPath, plan, stderr) {
		return model.Plan{}, false
	}
	return plan, true
}
