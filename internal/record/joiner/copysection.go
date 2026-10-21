package joiner

import "onboard-kit/internal/model"

func copySection(section model.Section) model.Section {
	section.Enabled = copyPointer(section.Enabled)
	section.IncludeIf = copyCondition(section.IncludeIf)
	section.After = copyEach(section.After, identity[string])
	section.Escalation = copyEach(section.Escalation, identity[string])
	section.Variables = copyEach(section.Variables, copyVariable)
	section.Steps = copyEach(section.Steps, copyStep)
	section.Source = copyPointer(section.Source)
	section.UnknownFields = copyUnknownFields(section.UnknownFields)
	return section
}

func copyStep(step model.Step) model.Step {
	step.Run = copyEach(step.Run, identity[string])
	step.Env = copyMap(step.Env, identity[string])
	step.Escalation = copyEach(step.Escalation, identity[string])
	step.Checks = copyEach(step.Checks, copyCheck)
	return step
}

func copyCheck(check model.Check) model.Check {
	check.Run = copyEach(check.Run, identity[string])
	return check
}
