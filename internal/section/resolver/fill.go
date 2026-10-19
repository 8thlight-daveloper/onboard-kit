package resolver

import "onboard-kit/internal/model"

func fillSection(section model.Section, filler *filler) model.Section {
	filled := copySection(section)
	filled.Title = filler.fill(section.Title)
	filled.Summary = filler.fill(section.Summary)
	filled.Owner = filler.fill(section.Owner)
	filled.Verification = filler.fill(section.Verification)
	filled.Body = filler.fill(section.Body)
	for index, step := range section.Steps {
		filled.Steps[index] = fillStep(step, filled.Steps[index], filler)
	}
	return filled
}

func fillStep(step model.Step, copied model.Step, filler *filler) model.Step {
	copied.Name = filler.fill(step.Name)
	copied.Notes = filler.fill(step.Notes)
	copied.Owner = filler.fill(step.Owner)
	copied.Run = filler.fillCommands(step.Run)
	copied.Open = filler.fill(step.Open)
	copied.Directory = filler.fill(step.Directory)
	copied.Env = filler.fillValues(step.Env)
	for index, check := range step.Checks {
		copied.Checks[index].Run = filler.fillCommands(check.Run)
		copied.Checks[index].Ask = filler.fill(check.Ask)
	}
	return copied
}
