package start

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"onboard-kit/internal/command/planfile"
	"onboard-kit/internal/command/report"
	"onboard-kit/internal/command/walk"
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/refresher"
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
	refreshedOn := options.String("refreshed-on", "", "date recorded as the refresh date, today when left out")
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
	personal, ok := prepare(folder, *orgPlan, plan, *refreshedOn, stdout, stderr)
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

func prepare(folder string, orgPath string, orgPlan model.Plan, refreshedOn string, stdout io.Writer, stderr io.Writer) (string, bool) {
	if orgPlan.ID == "" {
		fmt.Fprintf(stderr, "start: %s has no id. Make the plan again with onboard-kit new, or add an id line\n", orgPath)
		return "", false
	}
	if err := os.MkdirAll(folder, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return "", false
	}
	name := organizationName(orgPlan)
	cached := filepath.Join(folder, name+".org.obk.yaml")
	personal := filepath.Join(folder, name+".obk.yaml")
	origin, ok := originOf(orgPath, orgPlan, refreshedOn, stderr)
	if !ok {
		return "", false
	}
	if _, err := os.Stat(personal); err != nil {
		return personal, firstPlan(personal, cached, orgPlan, origin, stderr)
	}
	return personal, refreshPlan(personal, cached, orgPlan, origin, stdout, stderr)
}

func originOf(orgPath string, orgPlan model.Plan, refreshedOn string, stderr io.Writer) (model.Origin, bool) {
	source, err := filepath.Abs(orgPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return model.Origin{}, false
	}
	data, ok := planfile.Encode(orgPlan, stderr)
	if !ok {
		return model.Origin{}, false
	}
	if refreshedOn == "" {
		refreshedOn = time.Now().UTC().Format("2006-01-02")
	}
	return model.Origin{
		Source:      source,
		ContentHash: fmt.Sprintf("%x", sha256.Sum256(data)),
		RefreshedOn: model.Date(refreshedOn),
	}, true
}

func firstPlan(personal string, cached string, orgPlan model.Plan, origin model.Origin, stderr io.Writer) bool {
	origin.PlanID = orgPlan.ID
	first := orgPlan
	first.ID = ""
	first.Origin = origin
	return planfile.Save(cached, orgPlan, stderr) && planfile.Save(personal, first, stderr)
}

func refreshPlan(personal string, cached string, orgPlan model.Plan, origin model.Origin, stdout io.Writer, stderr io.Writer) bool {
	ours, ok := planfile.Open(personal, stderr)
	if !ok {
		return false
	}
	var base model.Plan
	if _, err := os.Stat(cached); err == nil {
		if base, ok = planfile.Open(cached, stderr); !ok {
			return false
		}
	}
	refreshed, changes, problems := refresher.New().Refresh(base, orgPlan, ours, origin)
	if len(problems) > 0 {
		report.Problems(stderr, personal, problems)
		return false
	}
	if !planfile.Save(personal, refreshed, stderr) || !planfile.Save(cached, orgPlan, stderr) {
		return false
	}
	summarize(stdout, changes)
	return true
}
