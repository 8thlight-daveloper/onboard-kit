package reviewer

import "onboard-kit/internal/model"

func directoryCharge(directory model.Directory) uint64 {
	contact, team, vacant, holders := uint64(0), uint64(0), uint64(0), uint64(0)
	for _, value := range directory.Contacts {
		contact = maximum(contact, add(length(value.Name), 2, length(value.Email), length(value.Slack)))
	}
	for _, value := range directory.Teams {
		team = maximum(team, length(value.Name))
		for _, list := range value.Holders {
			holders = maximum(holders, count(len(list)))
		}
	}
	for _, value := range directory.Roles {
		vacant = maximum(vacant, add(length(value.Name), 9))
		holders = maximum(holders, count(len(value.Holders)))
	}
	separators := uint64(0)
	if holders > 0 {
		separators = multiply(2, holders-1)
	}
	role := maximum(vacant, add(multiply(holders, contact), separators))
	return maximum(contact, team, role)
}

func sectionCharge(section model.Section, answers map[string]model.Value, directory uint64) uint64 {
	substitution := uint64(16)
	listItems := uint64(0)
	hasDirectory := false
	for _, variable := range section.Variables {
		substitution = maximum(substitution, add(16, length(variable.Name)))
		if variable.Type == model.VariableTypeContact || variable.Type == model.VariableTypeRole || variable.Type == model.VariableTypeTeam {
			hasDirectory = true
		}
	}
	for _, value := range answers {
		switch value.Kind {
		case model.ValueKindText:
			substitution = maximum(substitution, length(value.Text))
		case model.ValueKindFlag:
			substitution = maximum(substitution, 5)
		case model.ValueKindList:
			substitution = maximum(substitution, joinedLength(value.List, 1))
			listItems = maximum(listItems, count(len(value.List)))
		}
	}
	if hasDirectory {
		substitution = maximum(substitution, directory)
		if listItems > 0 {
			substitution = maximum(substitution, add(multiply(listItems, directory), multiply(2, listItems-1)))
		}
	}
	textExtra, runExtra, escalation, waiting := uint64(0), uint64(0), uint64(0), uint64(0)
	text := func(value string) { textExtra = add(textExtra, multiply(dollars(value), substitution)) }
	run := func(values []string) {
		for _, value := range values {
			text(value)
			runExtra = add(runExtra, multiply(64, multiply(dollars(value), maximum(1, listItems))))
		}
	}
	text(section.Title)
	text(section.Summary)
	text(section.Owner)
	text(section.Verification)
	text(section.Body)
	for _, step := range section.Steps {
		text(step.Name)
		text(step.Notes)
		text(step.Owner)
		text(step.Open)
		text(step.Directory)
		run(step.Run)
		for _, value := range step.Env {
			text(value)
		}
		for _, check := range step.Checks {
			run(check.Run)
			text(check.Ask)
		}
	}
	for _, value := range section.Escalation {
		escalation = add(escalation, 64, length(value), multiply(dollars(value), substitution))
	}
	for _, value := range section.Requires {
		waiting = add(waiting, 64, length(value))
	}
	return add(weight(section), textExtra, runExtra, escalation, waiting, 192)
}

func expansion(plan model.Plan, record model.Record) (uint64, uint64) {
	directory := directoryCharge(plan.Directory)
	render := uint64(0)
	for _, section := range plan.Sections {
		render = add(render, sectionCharge(section, plan.Answers.Sections[section.ID], directory))
	}
	for _, outcome := range record.Outcomes {
		render = add(render, weight(outcome), 200)
	}
	pools := make(map[string]uint64)
	declarations := func(variables []model.Variable) {
		for _, variable := range variables {
			pools[variable.Name] = add(pools[variable.Name], 64, selectedCharge(variable.Default), 7)
		}
	}
	declarations(plan.Variables)
	for _, section := range plan.Sections {
		declarations(section.Variables)
	}
	given := func(values map[string]model.Value) {
		for name, value := range values {
			pools[name] = add(pools[name], 64, selectedCharge(value), 5)
		}
	}
	given(plan.Answers.Plan)
	for _, values := range plan.Answers.Sections {
		given(values)
	}
	choices := make(map[model.VariableType]uint64)
	for _, value := range plan.Directory.Contacts {
		choices[model.VariableTypeContact] = add(choices[model.VariableTypeContact], 64, length(value.ID), length(value.Name))
	}
	for _, value := range plan.Directory.Roles {
		choices[model.VariableTypeRole] = add(choices[model.VariableTypeRole], 64, length(value.ID), length(value.Name))
	}
	for _, value := range plan.Directory.Teams {
		choices[model.VariableTypeTeam] = add(choices[model.VariableTypeTeam], 64, length(value.ID), length(value.Name))
	}
	questions := uint64(0)
	candidates := func(scope string, variables []model.Variable) {
		for _, variable := range variables {
			identity := add(length(scope), 1, length(variable.Name))
			typeBytes := length(string(variable.Type))
			if variable.Type == "" {
				typeBytes = 6
			}
			audienceBytes := length(string(variable.Audience))
			if variable.Audience == "" {
				audienceBytes = 9
			}
			options := add(joinedLength(variable.Options, 0), multiply(64, count(len(variable.Options))))
			group := add(128, length(variable.Name), typeBytes, options, 1)
			current := plan.Answers.Plan[variable.Name]
			if scope != "plan" {
				current = plan.Answers.Sections[scope][variable.Name]
			}
			candidate := add(256, identity, length(scope), length(variable.Name), length(variable.Description), length(variable.Example), typeBytes, audienceBytes, options, group, selectedCharge(current), selectedCharge(variable.Default), choices[variable.Type], pools[variable.Name], 64, identity)
			if scope != "plan" {
				candidate = add(candidate, multiply(2, add(256, identity, length(scope), length(variable.Name))))
			}
			questions = add(questions, candidate)
		}
	}
	candidates("plan", plan.Variables)
	for _, section := range plan.Sections {
		candidates(section.ID, section.Variables)
	}
	return render, questions
}
