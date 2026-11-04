package walk

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"onboard-kit/internal/command/planfile"
	"onboard-kit/internal/model"
	"onboard-kit/internal/tui"
)

const (
	exitSuccess = 0
	exitProblem = 1
	exitUsage   = 2

	exitInterrupted = 130
	exitTerminated  = 143
)

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	options := flag.NewFlagSet("walk", flag.ContinueOnError)
	options.SetOutput(stderr)
	planPath := options.String("plan", "", "plan file to open, or to create when it does not exist")
	templates := options.String("templates", "", "folder of section templates to start a new plan from")
	templateSet := options.String("template-set", "local", "name recorded as the template set")
	generatedOn := options.String("generated-on", "", "date recorded in a new plan, today when left out")
	audience := options.String("audience", "onboarder", "whose questions to walk: onboarder or onboardee")
	root := options.String("root", "", "folder that holds the people's preparations, the plan's folder when left out")
	if options.Parse(args) != nil {
		return exitUsage
	}
	if *planPath == "" {
		fmt.Fprintln(stderr, "walk: --plan is required")
		return exitUsage
	}
	who, ok := parseAudience(*audience)
	if !ok {
		fmt.Fprintln(stderr, "walk: --audience is onboarder or onboardee")
		return exitUsage
	}
	if info, err := os.Stat(*root); *root != "" && (err != nil || !info.IsDir()) {
		fmt.Fprintf(stderr, "walk: --root %s is not a folder\n", *root)
		return exitUsage
	}
	plan, ok := startPlan(*planPath, *templates, *templateSet, *generatedOn, stderr)
	if !ok {
		return exitProblem
	}
	if who == model.AudienceOnboardee {
		return Walk(*planPath, plan, who, "", stdin, stdout, stderr)
	}
	return walk(*planPath, plan, who, "", newPreparationFiles(preparationRootFor(*planPath, *root), *planPath), stdin, stdout, stderr)
}

func preparationRootFor(planPath string, root string) string {
	if root == "" {
		return filepath.Dir(planPath)
	}
	return root
}

func Walk(planPath string, plan model.Plan, audience model.Audience, notice string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	return walk(planPath, plan, audience, notice, nil, stdin, stdout, stderr)
}

func walk(planPath string, plan model.Plan, audience model.Audience, notice string, preparation tui.PreparationPort, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	save := func(plan model.Plan) error {
		if !planfile.Save(planPath, plan, io.Discard) {
			return fmt.Errorf("the plan could not be written to %s", planPath)
		}
		return nil
	}
	err := tui.Run(tui.NewBackend(), plan, save, audience, notice, preparation, stdin, stdout)
	if errors.Is(err, tui.ErrInterrupted) {
		return exitInterrupted
	}
	if errors.Is(err, tui.ErrTerminated) {
		return exitTerminated
	}
	if err != nil {
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

func parseAudience(text string) (model.Audience, bool) {
	switch model.Audience(text) {
	case model.AudienceOnboarder, model.AudienceOnboardee:
		return model.Audience(text), true
	}
	return "", false
}
