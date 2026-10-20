package planfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"onboard-kit/internal/command/report"
	"onboard-kit/internal/model"
	"onboard-kit/internal/plan/decoder"
	"onboard-kit/internal/plan/encoder"
	"onboard-kit/internal/plan/instantiator"
	"onboard-kit/internal/section/loader"
)

const planVariablesFile = "plan-variables.yaml"

func Open(path string, stderr io.Writer) (model.Plan, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return model.Plan{}, false
	}
	plan, problems := decoder.New().Decode(data)
	if len(problems) > 0 {
		report.Problems(stderr, path, problems)
		return model.Plan{}, false
	}
	return plan, true
}

func Encode(plan model.Plan, stderr io.Writer) ([]byte, bool) {
	data, problems := encoder.New().Encode(plan)
	if len(problems) > 0 {
		report.Problems(stderr, "plan", problems)
		return nil, false
	}
	return data, true
}

func Save(path string, plan model.Plan, stderr io.Writer) bool {
	data, ok := Encode(plan, stderr)
	if !ok {
		return false
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fmt.Fprintln(stderr, err)
		return false
	}
	return true
}

func FromTemplates(folder string, templateSet string, generatedOn string, stderr io.Writer) (model.Plan, bool) {
	sections, ok := loadTemplates(folder, stderr)
	if !ok {
		return model.Plan{}, false
	}
	variables, ok := loadPlanVariables(filepath.Join(folder, planVariablesFile), stderr)
	if !ok {
		return model.Plan{}, false
	}
	plan, problems := instantiator.New().Instantiate(sections, variables, templateSet, model.Date(dateOrToday(generatedOn)))
	if len(problems) > 0 {
		report.Problems(stderr, folder, problems)
		return model.Plan{}, false
	}
	return plan, true
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
		fmt.Fprintf(stderr, "no template files in %s\n", folder)
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
