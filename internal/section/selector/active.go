package selector

import (
	"onboard-kit/internal/condition/evaluator"
	"onboard-kit/internal/model"
)

func activeSections(plan model.Plan) []model.Section {
	conditions := evaluator.New()
	switches := collectSwitches(plan.Sections)
	byID := indexByID(plan.Sections)
	passes := map[string]bool{}
	for _, section := range plan.Sections {
		passes[section.ID] = isSwitchedOn(section) && includeIfPasses(conditions, section, plan.Answers, switches)
	}
	remaining := map[string]bool{}
	for _, section := range plan.Sections {
		if isActive(section, byID, passes) {
			remaining[section.ID] = true
		}
	}
	for changed := true; changed; {
		changed = false
		for _, section := range plan.Sections {
			if remaining[section.ID] && !dependenciesActive(section, remaining) {
				delete(remaining, section.ID)
				changed = true
			}
		}
	}
	active := make([]model.Section, 0, len(remaining))
	for _, section := range plan.Sections {
		if remaining[section.ID] {
			active = append(active, section)
		}
	}
	return active
}

func dependenciesActive(section model.Section, remaining map[string]bool) bool {
	if section.Parent != "" && !remaining[section.Parent] {
		return false
	}
	for _, target := range section.Requires {
		if !remaining[target] {
			return false
		}
	}
	return true
}

func indexByID(sections []model.Section) map[string]model.Section {
	byID := make(map[string]model.Section, len(sections))
	for _, section := range sections {
		if _, seen := byID[section.ID]; !seen {
			byID[section.ID] = section
		}
	}
	return byID
}

func isActive(section model.Section, byID map[string]model.Section, passes map[string]bool) bool {
	visited := map[string]bool{}
	current := section
	for {
		if !passes[current.ID] {
			return false
		}
		if current.Parent == "" {
			return true
		}
		visited[current.ID] = true
		parent, found := byID[current.Parent]
		if !found || visited[parent.ID] {
			return false
		}
		current = parent
	}
}

func includeIfPasses(conditions evaluator.Evaluator, section model.Section, answers model.Answers, switches model.Switches) bool {
	if section.IncludeIf == nil {
		return true
	}
	return conditions.Holds(*section.IncludeIf, section.ID, answers, switches)
}
