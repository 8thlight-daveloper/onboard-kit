package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

var (
	categories      = []string{"architecture", "behavior", "ux", "regenerative-process", "operations"}
	statuses        = []string{"proposed", "accepted", "superseded", "withdrawn"}
	approvals       = []string{"person", "agent", "unrecorded"}
	productRequired = []string{"behavior", "ux"}
	productAllowed  = []string{"architecture", "behavior", "ux"}
	digitRun        = regexp.MustCompile(`\d+`)
)

type result struct {
	problems []string
	notes    []string
}

func check(root string) (result, error) {
	records, problems := loadRecords(root)
	components, err := loadComponents(root)
	if err != nil {
		return result{}, err
	}
	contracts, err := loadContracts(root)
	if err != nil {
		return result{}, err
	}
	byNumber := map[string]record{}
	for _, one := range records {
		byNumber[one.number] = one
	}
	out := result{problems: problems}
	for _, one := range records {
		out.problems = append(out.problems, fieldProblems(one)...)
		out.problems = append(out.problems, affectsProblems(one, components)...)
		out.problems = append(out.problems, linkProblems(one, byNumber)...)
	}
	for _, one := range records {
		out.problems = append(out.problems, citationProblems(one, contracts)...)
		if note := uncitedNote(one, contracts); note != "" {
			out.notes = append(out.notes, note)
		}
	}
	sort.Strings(out.problems)
	return out, nil
}

func fieldProblems(one record) []string {
	var problems []string
	add := func(format string, args ...any) {
		problems = append(problems, one.file+": "+fmt.Sprintf(format, args...))
	}
	if !strings.HasPrefix(one.heading, "# "+one.number+". ") {
		add("the first heading must start with %q", "# "+one.number+". ")
	}
	if !contains(categories, one.front.Category) {
		add("category %q is not one of %s", one.front.Category, strings.Join(categories, ", "))
	}
	if !contains(statuses, one.front.Status) {
		add("status %q is not one of %s", one.front.Status, strings.Join(statuses, ", "))
	}
	if one.front.Status != "proposed" && !contains(approvals, one.front.Approval) {
		add("approval %q is not one of %s", one.front.Approval, strings.Join(approvals, ", "))
	}
	return problems
}

func affectsProblems(one record, components map[string]bool) []string {
	var problems []string
	category := one.front.Category
	if one.hasAffects && !contains(productAllowed, category) {
		problems = append(problems, fmt.Sprintf("%s: a %s record has no affects field", one.file, category))
	}
	if !one.hasAffects && contains(productRequired, category) {
		problems = append(problems, fmt.Sprintf("%s: a %s record needs an affects field, empty while the component is not built", one.file, category))
	}
	for _, id := range one.front.Affects {
		if !components[id] {
			problems = append(problems, fmt.Sprintf("%s: affects %q is not a component in architecture.yaml", one.file, id))
		}
	}
	return problems
}

func linkProblems(one record, byNumber map[string]record) []string {
	var problems []string
	links := map[string][]string{
		"builds-on": one.front.BuildsOn, "supersedes": one.front.Supersedes, "superseded-by": one.front.SupersededBy,
		"partly-supersedes": one.front.PartlySupersedes, "partly-superseded-by": one.front.PartlySupersededBy,
	}
	for field, numbers := range links {
		for _, number := range numbers {
			target, found := byNumber[number]
			switch {
			case number == one.number:
				problems = append(problems, fmt.Sprintf("%s: %s lists the record itself", one.file, field))
			case !found:
				problems = append(problems, fmt.Sprintf("%s: %s %s is not a record", one.file, field, number))
			default:
				problems = append(problems, reverseProblem(one, field, target)...)
			}
		}
	}
	return problems
}

func reverseProblem(one record, field string, target record) []string {
	reverse := map[string]struct {
		name   string
		values []string
	}{
		"supersedes":           {"superseded-by", target.front.SupersededBy},
		"superseded-by":        {"supersedes", target.front.Supersedes},
		"partly-supersedes":    {"partly-superseded-by", target.front.PartlySupersededBy},
		"partly-superseded-by": {"partly-supersedes", target.front.PartlySupersedes},
	}
	back, replacement := reverse[field]
	if !replacement || contains(back.values, one.number) {
		return nil
	}
	return []string{fmt.Sprintf("%s: %s %s, but %s does not list %s in %s", one.file, field, target.number, target.file, one.number, back.name)}
}

func citationProblems(one record, contracts map[string]map[string]bool) []string {
	if !one.hasAffects || !contains(productAllowed, one.front.Category) {
		return nil
	}
	var problems []string
	for _, component := range sortedKeys(contracts) {
		if contracts[component][one.number] && !contains(one.front.Affects, component) {
			problems = append(problems, fmt.Sprintf("%s: the contract of %s cites this record, so affects must list it", one.file, component))
		}
	}
	return problems
}

func uncitedNote(one record, contracts map[string]map[string]bool) string {
	var missing []string
	for _, component := range one.front.Affects {
		cited, hasContract := contracts[component]
		if hasContract && !cited[one.number] {
			missing = append(missing, component)
		}
	}
	if len(missing) == 0 {
		return ""
	}
	return fmt.Sprintf("%s: contracts of %s do not cite it", one.file, strings.Join(missing, ", "))
}

func loadComponents(root string) (map[string]bool, error) {
	text, err := os.ReadFile(filepath.Join(root, "architecture.yaml"))
	if err != nil {
		return nil, err
	}
	var file struct {
		Components map[string]yaml.Node `yaml:"components"`
	}
	if err := yaml.Unmarshal(text, &file); err != nil {
		return nil, fmt.Errorf("architecture.yaml: %w", err)
	}
	components := map[string]bool{}
	for id := range file.Components {
		components[id] = true
	}
	return components, nil
}

func loadContracts(root string) (map[string]map[string]bool, error) {
	contracts := map[string]map[string]bool{}
	base := filepath.Join(root, "contracts")
	err := filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || entry.Name() != "contract.md" {
			return err
		}
		text, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(base, filepath.Dir(path))
		contracts[filepath.ToSlash(relative)] = citedNumbers(string(text))
		return nil
	})
	return contracts, err
}

func citedNumbers(text string) map[string]bool {
	cited := map[string]bool{}
	for _, run := range digitRun.FindAllString(text, -1) {
		if len(run) == 4 && run[0] == '0' {
			cited[run] = true
		}
	}
	return cited
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
