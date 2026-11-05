package walk

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"onboard-kit/internal/command/places"
	"onboard-kit/internal/command/planfile"
	"onboard-kit/internal/command/registry"
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

func Run(args []string, version string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
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
	return edit(editing{planPath: *planPath, plan: plan, audience: who, root: *root, version: version}, stdin, stdout, stderr)
}

func Open(planPath string, version string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	plan, ok := planfile.Open(planPath, stderr)
	if !ok {
		return exitProblem
	}
	return edit(editing{planPath: planPath, plan: plan, audience: model.AudienceOnboarder, version: version}, stdin, stdout, stderr)
}

type editing struct {
	planPath string
	plan     model.Plan
	audience model.Audience
	root     string
	version  string
}

func edit(opened editing, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	home, homeErr := places.StateHome()
	if homeErr == nil {
		remember(home.Path, opened.planPath)
	}
	preparationRoot := preparationRootFor(opened.planPath, opened.root)
	context := places.Context{Home: home, PlanPath: absolute(opened.planPath), RecordsRoot: absolute(preparationRoot)}
	if homeErr != nil {
		context.HomeProblem = "The state home cannot be found: " + homeErr.Error()
	}
	about := places.About(opened.version, opened.plan.Provenance.TemplateSet+", as this plan records it", context, places.RegistryProblems(home))
	if opened.audience == model.AudienceOnboardee {
		return walk(opened.planPath, opened.plan, opened.audience, "", nil, nil, &about, stdin, stdout, stderr)
	}
	return walk(opened.planPath, opened.plan, opened.audience, "", newPreparationFiles(preparationRoot, opened.planPath), newInvitationFiles(preparationRoot, opened.planPath), &about, stdin, stdout, stderr)
}

func remember(home string, planPath string) {
	if path, err := filepath.Abs(planPath); err == nil {
		registry.Record(home, path, time.Now())
	}
}

func absolute(path string) string {
	if full, err := filepath.Abs(path); err == nil {
		return full
	}
	return path
}

func preparationRootFor(planPath string, root string) string {
	if root == "" {
		return filepath.Dir(planPath)
	}
	return root
}

func Walk(planPath string, plan model.Plan, audience model.Audience, notice string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	return walk(planPath, plan, audience, notice, nil, nil, nil, stdin, stdout, stderr)
}

func walk(planPath string, plan model.Plan, audience model.Audience, notice string, preparation tui.PreparationPort, invitation tui.InvitationPort, about *tui.About, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	save := func(plan model.Plan) error {
		if !planfile.Save(planPath, plan, io.Discard) {
			return fmt.Errorf("the plan could not be written to %s", planPath)
		}
		return nil
	}
	err := tui.Run(tui.NewBackend(), plan, save, audience, notice, preparation, invitation, about, stdin, stdout)
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
