package instantiator

import (
	"sort"

	"onboard-kit/internal/model"
	"onboard-kit/internal/section/hasher"
)

func instantiateSections(templates []model.Section) []model.Section {
	if templates == nil {
		return nil
	}
	sections := copySlice(templates, copySection)
	sort.SliceStable(sections, func(first, second int) bool {
		return sections[first].ID < sections[second].ID
	})
	return sections
}

func copySection(template model.Section) model.Section {
	copied := template
	copied.IncludeIf = copyCondition(template.IncludeIf)
	copied.After = copyStrings(template.After)
	copied.Requires = copyStrings(template.Requires)
	copied.Escalation = copyStrings(template.Escalation)
	copied.Variables = copyVariables(template.Variables)
	copied.Steps = copySlice(template.Steps, copyStep)
	copied.UnknownFields = copyUnknownMap(template.UnknownFields)
	copied.Source = &model.Source{Template: template.ID, Hash: hasher.New().Hash(template)}
	return copied
}

func copyStep(step model.Step) model.Step {
	copied := step
	copied.Run = copyStrings(step.Run)
	copied.Env = copyStringMap(step.Env)
	copied.Escalation = copyStrings(step.Escalation)
	copied.Checks = copySlice(step.Checks, copyCheck)
	return copied
}

func copyCheck(check model.Check) model.Check {
	check.Run = copyStrings(check.Run)
	return check
}
