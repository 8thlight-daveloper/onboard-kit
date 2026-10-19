package selector

import "onboard-kit/internal/model"

func isSwitchedOn(section model.Section) bool {
	return section.Enabled == nil || *section.Enabled
}

func collectSwitches(sections []model.Section) model.Switches {
	switches := make(model.Switches, len(sections))
	for _, section := range sections {
		switches[section.ID] = isSwitchedOn(section)
	}
	return switches
}
