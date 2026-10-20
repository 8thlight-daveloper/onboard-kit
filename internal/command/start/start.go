package start

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"onboard-kit/internal/command/planfile"
	"onboard-kit/internal/command/walk"
	"onboard-kit/internal/model"
)

const (
	exitSuccess = 0
	exitProblem = 1
	exitUsage   = 2
)

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	options := flag.NewFlagSet("start", flag.ContinueOnError)
	options.SetOutput(stderr)
	orgPlan := options.String("org-plan", "", "the organization's plan file")
	home := options.String("home", "", "folder for your plan, a per-user folder when left out")
	noWalk := options.Bool("no-walk", false, "write your plan and stop, without the walkthrough")
	if options.Parse(args) != nil {
		return exitUsage
	}
	if *orgPlan == "" {
		fmt.Fprintln(stderr, "start: --org-plan is required")
		return exitUsage
	}
	plan, ok := planfile.Open(*orgPlan, stderr)
	if !ok {
		return exitProblem
	}
	folder, err := stateHome(*home)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitProblem
	}
	personal, ok := prepare(folder, plan, stderr)
	if !ok {
		return exitProblem
	}
	if *noWalk {
		fmt.Fprintln(stdout, personal)
		return exitSuccess
	}
	current, ok := planfile.Open(personal, stderr)
	if !ok {
		return exitProblem
	}
	return walk.Walk(personal, current, model.AudienceOnboardee, stdin, stdout, stderr)
}

func prepare(folder string, orgPlan model.Plan, stderr io.Writer) (string, bool) {
	if err := os.MkdirAll(folder, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return "", false
	}
	name := organizationName(orgPlan)
	cached := filepath.Join(folder, name+".org.obk.yaml")
	personal := filepath.Join(folder, name+".obk.yaml")
	if !planfile.Save(cached, orgPlan, stderr) {
		return "", false
	}
	if _, err := os.Stat(personal); err == nil {
		return personal, true
	}
	if !planfile.Save(personal, orgPlan, stderr) {
		return "", false
	}
	return personal, true
}
