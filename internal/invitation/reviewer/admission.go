package reviewer

import (
	"reflect"
	"strings"

	"onboard-kit/internal/model"
)

const admissionLimit uint64 = 67108864
const admissionExceeded uint64 = admissionLimit + 1

func add(values ...uint64) uint64 {
	var total uint64
	for _, value := range values {
		if value >= admissionExceeded || total > admissionExceeded-value {
			return admissionExceeded
		}
		total += value
	}
	return total
}

func multiply(left, right uint64) uint64 {
	if left == 0 || right == 0 {
		return 0
	}
	if left > admissionExceeded/right {
		return admissionExceeded
	}
	return left * right
}

func length(text string) uint64 { return uint64(len(text)) }
func count(size int) uint64     { return uint64(size) }
func maximum(left, right uint64) uint64 {
	if left > right {
		return left
	}
	return right
}
func separators(size int) uint64 {
	if size < 2 {
		return 0
	}
	return count(size - 1)
}

type forest struct {
	occurrences    int
	unknown        int
	entries        int
	variables      int
	steps          int
	checks         int
	clauses        int
	comparisons    int
	holders        int
	listItems      int
	options        int
	conditionPairs uint64
	valid          bool
}

func admit(plan model.Plan, record model.Record, preparation *model.Preparation, sourceBytes int, full bool) bool {
	meter := forest{valid: true}
	weight := add(meter.walk(reflect.ValueOf(plan), 1, false), meter.walk(reflect.ValueOf(record), 1, false), meter.walk(reflect.ValueOf(preparation), 1, false))
	if !meter.valid {
		return false
	}
	if !full {
		return true
	}
	base := add(count(sourceBytes), 64, weight)
	render := renderCharge(plan, record)
	questions := questionCharge(plan)
	eligibility := multiply(16, multiply(count(meter.variables), count(meter.clauses)))
	pairs := add(multiply(32, add(multiply(count(meter.variables), count(meter.variables)), multiply(count(len(plan.Sections)), count(len(plan.Sections))))), multiply(8, meter.conditionPairs))
	return add(multiply(8, add(base, render, questions)), eligibility, pairs) <= admissionLimit
}

func (meter *forest) walk(value reflect.Value, depth int, unknown bool) uint64 {
	if !meter.valid {
		return admissionExceeded
	}
	if value.Kind() == reflect.Interface && !value.IsNil() {
		return meter.walk(value.Elem(), depth, unknown)
	}
	meter.occurrences++
	if unknown {
		meter.unknown++
	}
	if depth > 128 || meter.occurrences > 131072 || meter.unknown > 65536 {
		meter.valid = false
		return admissionExceeded
	}
	weight := uint64(64)
	if !unknown {
		meter.cardinality(value)
		if !meter.valid {
			return admissionExceeded
		}
	}
	if unknown {
		switch value.Kind() {
		case reflect.Interface:
			if !value.IsNil() {
				meter.valid = false
			}
			return weight
		case reflect.String:
			if value.Type() != reflect.TypeFor[string]() {
				meter.valid = false
			}
		case reflect.Bool:
			if value.Type() != reflect.TypeFor[bool]() {
				meter.valid = false
			}
		case reflect.Int64:
			if value.Type() != reflect.TypeFor[int64]() {
				meter.valid = false
			}
		case reflect.Float64:
			if value.Type() != reflect.TypeFor[float64]() {
				meter.valid = false
			}
		case reflect.Map:
			if value.Type() != reflect.TypeFor[map[string]any]() {
				meter.valid = false
			}
		case reflect.Slice:
			if value.Type() != reflect.TypeFor[[]any]() {
				meter.valid = false
			}
		default:
			meter.valid = false
		}
		if !meter.valid {
			return admissionExceeded
		}
	}
	switch value.Kind() {
	case reflect.Interface:
		return weight
	case reflect.String:
		return add(weight, length(value.String()))
	case reflect.Bool:
		return add(weight, 1)
	case reflect.Int, reflect.Int64, reflect.Float64:
		return add(weight, 8)
	case reflect.Pointer:
		if !value.IsNil() {
			weight = add(weight, meter.walk(value.Elem(), depth+1, unknown))
		}
		return weight
	case reflect.Struct:
		fields := modelFields(value.Type().Name())
		if len(fields) != value.NumField() {
			meter.valid = false
			return admissionExceeded
		}
		for i, field := range fields {
			if value.Type().Field(i).Name != field {
				meter.valid = false
				return admissionExceeded
			}
			weight = add(weight, meter.walk(value.Field(i), depth+1, unknown || field == "UnknownFields"))
		}
		return weight
	case reflect.Map:
		if !unknown {
			meter.entries += value.Len()
			if meter.entries > 65536 {
				meter.valid = false
				return admissionExceeded
			}
		}
		entries := value.MapRange()
		for entries.Next() {
			weight = add(weight, meter.walk(entries.Key(), depth+1, unknown), meter.walk(entries.Value(), depth+1, unknown))
		}
		return weight
	case reflect.Slice:
		if !unknown {
			meter.entries += value.Len()
			if meter.entries > 65536 {
				meter.valid = false
				return admissionExceeded
			}
		}
		for i := 0; i < value.Len(); i++ {
			weight = add(weight, meter.walk(value.Index(i), depth+1, unknown))
		}
		return weight
	default:
		meter.valid = false
		return admissionExceeded
	}
}

func modelFields(name string) []string {
	fields := ""
	switch name {
	case "Plan":
		fields = "SchemaVersion ID Provenance Origin URL ResultsPath Directory Variables Answers Sections UnknownFields"
	case "Provenance":
		fields = "TemplateSet GeneratedOn"
	case "Origin":
		fields = "PlanID Source ContentHash RefreshedOn"
	case "Directory":
		fields = "Contacts Roles Teams"
	case "Contact":
		fields = "ID Name Email Slack SlackID"
	case "Role":
		fields = "ID Name Holders"
	case "Team":
		fields = "ID Name Contact Holders"
	case "Answers":
		fields = "Plan Sections"
	case "Variable":
		fields = "Name Description Example Type Options Pattern PatternHint Required Multiple Default Audience AskIf"
	case "Value":
		fields = "Kind Text Flag List"
	case "Condition":
		fields = "All"
	case "Clause":
		fields = "Question Answer Section Enabled"
	case "AnswerTest":
		fields = "Is IsNot IsOneOf IsNotOneOf IsSettled"
	case "Section":
		fields = "ID Title Summary Parent Bootstrap Owner Timing Enabled Verification Body IncludeIf After Requires Escalation Variables Steps Source UnknownFields"
	case "Source":
		fields = "Template Hash"
	case "Step":
		fields = "ID Name Notes Owner Timing Run Open Directory Env Escalation Checks"
	case "Check":
		fields = "Run Ask Env"
	case "Record":
		fields = "SchemaVersion Account StartDate FirstOpenedAt Plan Answers Outcomes UnknownFields"
	case "PlanReference":
		fields = "ID Hash URL"
	case "Outcome":
		fields = "Section Kind On SectionHash Note Differently"
	case "Preparation":
		fields = "SchemaVersion Kind Account PlanID Recipient Arrangements UnknownFields"
	case "PreparationRecipient":
		fields = "Name CompanyEmail PlannedStartDate Team Manager OnboardingOwner Buddy WorkLocation TimeZone"
	case "Arrangement":
		fields = "ID Section Kind Owner Contacts Reference TargetDate Status Note"
	}
	return strings.Fields(fields)
}

func (meter *forest) cardinality(value reflect.Value) {
	switch item := value.Interface().(type) {
	case model.Plan:
		meter.valid = meter.valid && len(item.Sections) <= 256
	case model.Directory:
		meter.valid = meter.valid && len(item.Contacts) <= 2048 && len(item.Roles) <= 2048 && len(item.Teams) <= 2048 && len(item.Contacts)+len(item.Roles)+len(item.Teams) <= 4096
	case model.Variable:
		meter.variables++
		meter.options += len(item.Options)
		meter.valid = meter.valid && meter.variables <= 1024 && len(item.Options) <= 1024 && meter.options <= 16384
	case model.Step:
		meter.steps++
		meter.valid = meter.valid && meter.steps <= 4096
	case model.Check:
		meter.checks++
		meter.valid = meter.valid && meter.checks <= 8192
	case model.Condition:
		meter.clauses += len(item.All)
		meter.conditionPairs = add(meter.conditionPairs, multiply(count(len(item.All)), count(len(item.All))))
		meter.valid = meter.valid && len(item.All) <= 256 && meter.clauses <= 4096
	case model.AnswerTest:
		meter.comparisons += len(item.IsOneOf) + len(item.IsNotOneOf)
		meter.valid = meter.valid && len(item.IsOneOf) <= 256 && len(item.IsNotOneOf) <= 256 && meter.comparisons <= 16384
	case model.Role:
		meter.holderList(item.Holders)
	case model.Team:
		for _, holders := range item.Holders {
			meter.holderList(holders)
		}
	case model.Value:
		meter.listItems += len(item.List)
		meter.valid = meter.valid && len(item.List) <= 4096 && meter.listItems <= 32768
	case model.Record:
		meter.valid = meter.valid && len(item.Outcomes) <= 4096
	}
}

func (meter *forest) holderList(values []string) {
	meter.holders += len(values)
	meter.valid = meter.valid && len(values) <= 1024 && meter.holders <= 8192
}

func modelWeight(value any) uint64 {
	meter := forest{valid: true}
	return meter.walk(reflect.ValueOf(value), 1, false)
}

func joinedLength(values []string, separator uint64) uint64 {
	total := multiply(separator, separators(len(values)))
	for _, value := range values {
		total = add(total, length(value))
	}
	return total
}

func directoryCharge(directory model.Directory) uint64 {
	var contact, team, vacant, holders uint64
	for _, item := range directory.Contacts {
		contact = maximum(contact, add(length(item.Name), 2, length(item.Email), length(item.Slack)))
	}
	for _, item := range directory.Roles {
		vacant = maximum(vacant, add(length(item.Name), 9))
		holders = maximum(holders, count(len(item.Holders)))
	}
	for _, item := range directory.Teams {
		team = maximum(team, length(item.Name))
		for _, list := range item.Holders {
			holders = maximum(holders, count(len(list)))
		}
	}
	var holderSeparators uint64
	if holders > 0 {
		holderSeparators = holders - 1
	}
	role := maximum(vacant, add(multiply(holders, contact), multiply(2, holderSeparators)))
	return maximum(contact, maximum(team, role))
}

func renderCharge(plan model.Plan, record model.Record) uint64 {
	directory := directoryCharge(plan.Directory)
	var render uint64
	for _, section := range plan.Sections {
		marker := uint64(16)
		var listSize uint64
		hasDirectory := false
		for _, variable := range section.Variables {
			marker = maximum(marker, add(16, length(variable.Name)))
			if variable.Type == model.VariableTypeContact || variable.Type == model.VariableTypeRole || variable.Type == model.VariableTypeTeam {
				hasDirectory = true
			}
		}
		for _, value := range plan.Answers.Sections[section.ID] {
			switch value.Kind {
			case model.ValueKindText:
				marker = maximum(marker, length(value.Text))
			case model.ValueKindFlag:
				marker = maximum(marker, 5)
			case model.ValueKindList:
				marker = maximum(marker, joinedLength(value.List, 1))
				listSize = maximum(listSize, count(len(value.List)))
			}
		}
		if hasDirectory {
			marker = maximum(marker, directory)
			if listSize > 0 {
				marker = maximum(marker, add(multiply(listSize, directory), multiply(2, listSize-1)))
			}
		}
		textExtra := uint64(0)
		runExtra := uint64(0)
		text := func(value string) { textExtra = add(textExtra, multiply(count(strings.Count(value, "$")), marker)) }
		run := func(values []string) {
			for _, value := range values {
				text(value)
				runExtra = add(runExtra, multiply(64, multiply(count(strings.Count(value, "$")), maximum(1, listSize))))
			}
		}
		for _, value := range []string{section.Title, section.Summary, section.Owner, section.Verification, section.Body} {
			text(value)
		}
		for _, step := range section.Steps {
			for _, value := range []string{step.Name, step.Notes, step.Owner, step.Open, step.Directory} {
				text(value)
			}
			run(step.Run)
			for _, value := range step.Env {
				text(value)
			}
			for _, check := range step.Checks {
				run(check.Run)
				text(check.Ask)
			}
		}
		escalation := uint64(0)
		waiting := uint64(0)
		for _, value := range section.Escalation {
			escalation = add(escalation, 64, length(value), multiply(count(strings.Count(value, "$")), marker))
		}
		for _, value := range section.Requires {
			waiting = add(waiting, 64, length(value))
		}
		render = add(render, modelWeight(section), textExtra, runExtra, escalation, waiting, 192)
	}
	for _, outcome := range record.Outcomes {
		render = add(render, modelWeight(outcome), 200)
	}
	return render
}

func valueCharge(value model.Value) uint64 {
	switch value.Kind {
	case model.ValueKindText:
		return add(64, length(value.Text))
	case model.ValueKindFlag:
		return 69
	case model.ValueKindList:
		return add(64, joinedLength(value.List, 0), multiply(16, count(len(value.List))))
	}
	return 64
}

func questionCharge(plan model.Plan) uint64 {
	pool := map[string]uint64{}
	addDeclarations := func(variables []model.Variable) {
		for _, variable := range variables {
			pool[variable.Name] = add(pool[variable.Name], 64, valueCharge(variable.Default), 7)
		}
	}
	addDeclarations(plan.Variables)
	for _, section := range plan.Sections {
		addDeclarations(section.Variables)
	}
	addAnswers := func(values map[string]model.Value) {
		for name, value := range values {
			pool[name] = add(pool[name], 64, valueCharge(value), 5)
		}
	}
	addAnswers(plan.Answers.Plan)
	for _, values := range plan.Answers.Sections {
		addAnswers(values)
	}
	choices := map[model.VariableType]uint64{}
	for _, item := range plan.Directory.Contacts {
		choices[model.VariableTypeContact] = add(choices[model.VariableTypeContact], 64, length(item.ID), length(item.Name))
	}
	for _, item := range plan.Directory.Roles {
		choices[model.VariableTypeRole] = add(choices[model.VariableTypeRole], 64, length(item.ID), length(item.Name))
	}
	for _, item := range plan.Directory.Teams {
		choices[model.VariableTypeTeam] = add(choices[model.VariableTypeTeam], 64, length(item.ID), length(item.Name))
	}
	var questions uint64
	candidates := func(scope string, variables []model.Variable) {
		for _, variable := range variables {
			identity := add(length(scope), 1, length(variable.Name))
			typeBytes := length(string(variable.Type))
			if typeBytes == 0 {
				typeBytes = 6
			}
			audienceBytes := length(string(variable.Audience))
			if audienceBytes == 0 {
				audienceBytes = 9
			}
			options := add(joinedLength(variable.Options, 0), multiply(64, count(len(variable.Options))))
			group := add(128, length(variable.Name), typeBytes, options, 1)
			current := plan.Answers.Sections[scope][variable.Name]
			if scope == "plan" {
				current = plan.Answers.Plan[variable.Name]
			}
			candidate := add(256, identity, length(scope), length(variable.Name), length(variable.Description), length(variable.Example), typeBytes, audienceBytes, options, group, valueCharge(current), valueCharge(variable.Default), choices[variable.Type], pool[variable.Name], 64, identity)
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
	return questions
}
