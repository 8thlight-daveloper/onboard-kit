package authoringvalidator

import "onboard-kit/internal/model"

func (a *analysis) finishGraph() {
	for source, targets := range a.askEdges {
		for target := range targets {
			if a.declarations[source].unique && a.declarations[target].unique {
				addEdge(a.edges, source, target)
			}
		}
	}
	if !a.duplicateSections {
		for source, d := range a.declarations {
			if !d.unique || d.section < 0 {
				continue
			}
			sections := []int{d.section}
			ancestors, known := a.ancestors(d.section)
			if known {
				sections = append(sections, ancestors...)
			}
			for _, section := range sections {
				for target := range a.includeEdges[section] {
					if a.declarations[target].unique {
						addEdge(a.edges, source, target)
					}
				}
			}
		}
		for index, d := range a.declarations {
			if d.unique && reaches(a.edges, index, index, map[int]bool{}) {
				a.emit("condition-cycle", d.path, d)
			}
		}
	}
	queue := make([]int, 0, len(a.used))
	for index := range a.used {
		queue = append(queue, index)
	}
	for len(queue) > 0 {
		source := queue[0]
		queue = queue[1:]
		for target := range a.askEdges[source] {
			if !a.used[target] {
				a.used[target] = true
				queue = append(queue, target)
			}
		}
	}
	for index, d := range a.declarations {
		if d.unique && legalName(d.variable.Name) && d.variable.Audience == model.AudienceOnboardee && !a.used[index] {
			a.emit("unused-onboardee", d.path, d)
		}
	}
}

func reaches(graph map[int]map[int]bool, start, target int, seen map[int]bool) bool {
	if seen[start] {
		return false
	}
	seen[start] = true
	for next := range graph[start] {
		if next == target || reaches(graph, next, target, seen) {
			return true
		}
	}
	return false
}
