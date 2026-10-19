package evaluator

import "onboard-kit/internal/model"

func sectionClauseHolds(clause model.Clause, switches model.Switches) bool {
	return isSwitchedOn(clause.Section, switches) == clause.Enabled
}

func isSwitchedOn(section string, switches model.Switches) bool {
	on, found := switches[section]
	return !found || on
}
