package editor

import (
	"onboard-kit/internal/model"
)

const unknownSectionCode = "switch.unknown-section"

func (Editor) SetSectionEnabled(plan model.Plan, section string, enabled bool) (model.Plan, []model.Problem) {
	plan = copyEditableParts(plan)
	index := sectionIndex(plan.Sections, section)
	if index < 0 || section == planScope {
		return plan, []model.Problem{unknownSection(section)}
	}
	if enabled {
		plan.Sections[index].Enabled = nil
	} else {
		off := false
		plan.Sections[index].Enabled = &off
	}
	return plan, []model.Problem{}
}

func sectionIndex(sections []model.Section, id string) int {
	for index, section := range sections {
		if section.ID == id {
			return index
		}
	}
	return -1
}

func unknownSection(section string) model.Problem {
	return model.Problem{
		Code:     unknownSectionCode,
		Severity: model.SeverityError,
		Message:  "This plan has no section with that id.",
		Section:  section,
	}
}
