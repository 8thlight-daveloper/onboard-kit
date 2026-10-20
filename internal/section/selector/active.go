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
	active := make([]model.Section, 0, len(plan.Sections))
	for _, section := range plan.Sections {
		if isActive(section, byID, passes) {
			active = append(active, section)
		}
	}
	return active
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
