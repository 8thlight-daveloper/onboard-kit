package main

import (
	"fmt"
	"sort"
	"strings"
)

type outcome struct {
	problems   []string
	undeclared []string
}

func evaluate(loaded rules, graph map[string]packageImports, contractExists func(string) bool) outcome {
	var out outcome
	add := func(format string, args ...any) { out.problems = append(out.problems, fmt.Sprintf(format, args...)) }
	for _, id := range sortedKeys(graph) {
		if _, ok := loaded.Components[id]; !ok {
			add("%s: the code has a package for it, but architecture.yaml has no component", id)
		}
	}
	for _, id := range sortedKeys(loaded.Components) {
		component := loaded.Components[id]
		kind, known := loaded.Kinds[component.Kind]
		if !known {
			add("%s: kind %q is not defined", id, component.Kind)
			continue
		}
		if !nameFits(loaded, id) {
			add("%s: the id is not noun/role and is not a one-word exception", id)
		}
		if component.Status != "" {
			continue
		}
		built, exists := graph[id]
		if !exists {
			add("%s: architecture.yaml lists it, and the code has no package for it", id)
			continue
		}
		if kind.Contract == "required" && !contractExists(id) {
			add("%s: a %s component needs a contracts folder", id, component.Kind)
		}
		if component.Imports == nil && component.Kind == "command" {
			out.undeclared = append(out.undeclared, id)
		}
		for _, imported := range built.components {
			if problem := importProblem(loaded, id, component, kind, imported); problem != "" {
				out.problems = append(out.problems, problem)
			}
		}
		for _, category := range kind.BannedCategories {
			if used := built.categories[category]; len(used) > 0 {
				add("%s: a %s component may not use the %s category, and it imports %s", id, component.Kind, category, strings.Join(used, ", "))
			}
		}
	}
	for _, id := range sortedKeys(loaded.Components) {
		for _, imported := range loaded.Components[id].Imports {
			if _, ok := loaded.Components[imported]; !ok {
				add("%s: declares an import of %s, which is not a component", id, imported)
			}
		}
	}
	out.problems = append(out.problems, cycleProblems(loaded, graph)...)
	sort.Strings(out.problems)
	return out
}

func importProblem(loaded rules, id string, component componentRule, kind kindRule, imported string) string {
	target, ok := loaded.Components[imported]
	if !ok {
		return ""
	}
	if !contains(kind.Imports, target.Kind) {
		return fmt.Sprintf("%s: a %s component may not import a %s component, and it imports %s", id, component.Kind, target.Kind, imported)
	}
	if component.Imports != nil && !contains(component.Imports, imported) {
		return fmt.Sprintf("%s: imports %s, which its declared imports do not list", id, imported)
	}
	return ""
}

func nameFits(loaded rules, id string) bool {
	if contains(loaded.Naming.DepthOneExceptions, id) {
		return true
	}
	parts := strings.Split(id, "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

func cycleProblems(loaded rules, graph map[string]packageImports) []string {
	state := map[string]int{}
	var problems []string
	var visit func(id string, path []string)
	visit = func(id string, path []string) {
		switch state[id] {
		case 1:
			start := 0
			for index, member := range path {
				if member == id {
					start = index
				}
			}
			problems = append(problems, "import cycle: "+strings.Join(append(append([]string{}, path[start:]...), id), " -> "))
			return
		case 2:
			return
		}
		state[id] = 1
		for _, imported := range graph[id].components {
			if _, ok := loaded.Components[imported]; ok {
				visit(imported, append(path, id))
			}
		}
		state[id] = 2
	}
	for _, id := range sortedKeys(graph) {
		visit(id, nil)
	}
	return problems
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
