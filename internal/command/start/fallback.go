package start

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"onboard-kit/internal/command/planfile"
	"onboard-kit/internal/command/walk"
	"onboard-kit/internal/model"
)

func startFromCopy(orgPath string, folder string, noWalk bool, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	source, err := filepath.Abs(orgPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitProblem
	}
	personal, current, found := personalPlanFrom(source, folder)
	if !found {
		fmt.Fprintf(stderr, "start: %s cannot be read, and there is no earlier copy of it in %s\n", orgPath, folder)
		return exitProblem
	}
	notice := fmt.Sprintf("The organization's plan cannot be reached. Using your copy as of %s.", current.Origin.RefreshedOn)
	if noWalk {
		fmt.Fprintln(stdout, notice)
		fmt.Fprintln(stdout, personal)
		return exitSuccess
	}
	return walk.Walk(personal, current, model.AudienceOnboardee, notice, stdin, stdout, stderr)
}

func personalPlanFrom(source string, folder string) (string, model.Plan, bool) {
	paths, err := filepath.Glob(filepath.Join(folder, "*.obk.yaml"))
	if err != nil {
		return "", model.Plan{}, false
	}
	for _, path := range paths {
		if strings.HasSuffix(path, ".org.obk.yaml") {
			continue
		}
		plan, ok := planfile.Open(path, io.Discard)
		if ok && plan.Origin.Source == source {
			return path, plan, true
		}
	}
	return "", model.Plan{}, false
}
