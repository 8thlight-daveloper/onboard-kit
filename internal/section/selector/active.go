package selector

import (
	"onboard-kit/internal/condition/evaluator"
	"onboard-kit/internal/model"
)

func activeSections(plan model.Plan) []model.Section {
	conditions := evaluator.New()
	switches := collectSwitches(plan.Sections)
	active := make([]model.Section, 0, len(plan.Sections))
	for _, section := range plan.Sections {
		if isSwitchedOn(section) && includeIfPasses(conditions, section, plan.Answers, switches) {
			active = append(active, section)
		}
	}
	return active
}

func includeIfPasses(conditions evaluator.Evaluator, section model.Section, answers model.Answers, switches model.Switches) bool {
	if section.IncludeIf == nil {
		return true
	}
	return conditions.Holds(*section.IncludeIf, section.ID, answers, switches)
}
