package create

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"onboard-kit/internal/command/report"
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/decoder"
	"onboard-kit/internal/plan/encoder"
	"onboard-kit/internal/plan/instantiator"
	"onboard-kit/internal/section/loader"
)

const planVariablesFile = "plan-variables.yaml"

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
	sections, loaded := loadTemplates(*templates, stderr)
	if !loaded {
		return exitProblem
	}
	variables, loaded := loadPlanVariables(filepath.Join(*templates, planVariablesFile), stderr)
	if !loaded {
		return exitProblem
	}
	plan, problems := instantiator.New().Instantiate(sections, variables, *templateSet, model.Date(dateOrToday(*generatedOn)))
	if len(problems) > 0 {
		report.Problems(stderr, *templates, problems)
		return exitProblem
	}
	data, problems := encoder.New().Encode(plan)
	if len(problems) > 0 {
		report.Problems(stderr, "plan", problems)
		return exitProblem
	}
	return write(*output, data, stdout, stderr)
}

func dateOrToday(given string) string {
	if given != "" {
		return given
	}
	return time.Now().UTC().Format("2006-01-02")
}

func loadTemplates(folder string, stderr io.Writer) ([]model.Section, bool) {
	paths, err := filepath.Glob(filepath.Join(folder, "*.yaml"))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return nil, false
	}
	sort.Strings(paths)
	var sections []model.Section
	ok := true
	for _, path := range paths {
		if filepath.Base(path) == planVariablesFile {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(stderr, err)
			ok = false
			continue
		}
		section, problems := loader.New().Load(data)
		if len(problems) > 0 {
			report.Problems(stderr, path, problems)
			ok = false
			continue
		}
		sections = append(sections, section)
	}
	if ok && len(sections) == 0 {
		fmt.Fprintf(stderr, "new: no template files in %s\n", folder)
		return nil, false
	}
	return sections, ok
}

func loadPlanVariables(path string, stderr io.Writer) ([]model.Variable, bool) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, true
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return nil, false
	}
	plan, problems := decoder.New().Decode(data)
	if len(problems) > 0 {
		report.Problems(stderr, path, problems)
		return nil, false
	}
	return plan.Variables, true
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
		fmt.Fprintln(stderr, strings.TrimSpace(err.Error()))
		return exitProblem
	}
	return exitSuccess
}
