package scheduler

import "onboard-kit/internal/model"

type activity struct {
	section       model.Section
	effectiveTime model.Timing
	referenceRank int
}

type actionGraph struct {
	activities []activity
	successors []map[int]bool
	indegrees  []int
}

func graphOf(sections []model.Section, timings map[string]model.Timing) actionGraph {
	graph := actionGraph{
		activities: make([]activity, len(sections)),
		successors: make([]map[int]bool, len(sections)),
		indegrees:  make([]int, len(sections)),
	}
	indices := map[string]int{}
	for index, section := range sections {
		indices[section.ID] = index
		graph.activities[index] = activity{
			section: section, effectiveTime: timings[section.ID], referenceRank: index,
		}
	}
	for dependent, section := range sections {
		for _, target := range section.Requires {
			if predecessor, selected := indices[target]; selected {
				graph.add(predecessor, dependent)
			}
		}
		for _, target := range section.After {
			predecessor, selected := indices[target]
			if selected && sections[predecessor].Parent == section.Parent && timings[target] == timings[section.ID] {
				graph.add(predecessor, dependent)
			}
		}
	}
	return graph
}

func (graph actionGraph) add(predecessor, dependent int) {
	if graph.successors[predecessor] == nil {
		graph.successors[predecessor] = map[int]bool{}
	}
	if !graph.successors[predecessor][dependent] {
		graph.successors[predecessor][dependent] = true
		graph.indegrees[dependent]++
	}
}

func (graph actionGraph) orderedSections() []model.Section {
	ordered := make([]model.Section, 0, len(graph.activities))
	placed := make([]bool, len(graph.activities))
	for len(ordered) < len(graph.activities) {
		chosen := graph.firstFree(placed)
		ordered = append(ordered, graph.activities[chosen].section)
		placed[chosen] = true
		for dependent := range graph.successors[chosen] {
			graph.indegrees[dependent]--
		}
	}
	return ordered
}

func (graph actionGraph) firstFree(placed []bool) int {
	chosen := -1
	for index, candidate := range graph.activities {
		if placed[index] || graph.indegrees[index] != 0 {
			continue
		}
		if chosen < 0 || precedes(candidate, graph.activities[chosen]) {
			chosen = index
		}
	}
	return chosen
}
