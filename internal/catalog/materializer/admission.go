package materializer

import (
	"sort"
	"strconv"
	"unicode/utf8"

	contract "onboard-kit/contracts/catalog/materializer"
	"onboard-kit/internal/model"
)

type admission struct {
	phase           int
	entries         int
	bindings        int
	holders         int
	teamEntries     int
	options         int
	valueItems      int
	clauses         int
	comparisons     int
	textBytes       int
	collectionLimit bool
	textLimit       bool
	invalidText     bool
	invalidKind     string
}

func admit(request contract.Request) *contract.Diagnostic {
	for phase := 0; phase < 4; phase++ {
		guard := admission{phase: phase}
		guard.scan(request)
		if guard.collectionLimit || guard.textLimit {
			diagnostic := own("limit", "request", "request", "Reduce the selection or held creation data to the supported limits.")
			return &diagnostic
		}
		if guard.invalidText {
			diagnostic := own("text", "request", "request", "Use valid UTF8 creation data.")
			return &diagnostic
		}
		if guard.invalidKind != "" {
			diagnostic := own("value-kind", "request", guard.invalidKind, "Choose a supported value kind.")
			return &diagnostic
		}
	}
	return nil
}

func saturatedAdd(held, count, maximum int) int {
	if held > maximum || count > maximum-held {
		return maximum + 1
	}
	return held + count
}

func (guard *admission) collection(count, maximum int, combined *int, aggregate int) bool {
	guard.entries = saturatedAdd(guard.entries, count, 65536)
	if count > maximum || guard.entries > 65536 {
		guard.collectionLimit = true
	}
	if combined != nil {
		*combined = saturatedAdd(*combined, count, aggregate)
		if *combined > aggregate {
			guard.collectionLimit = true
		}
	}
	return !guard.collectionLimit
}

func (guard *admission) text(texts ...string) {
	for _, value := range texts {
		if guard.phase == 1 {
			guard.textBytes = saturatedAdd(guard.textBytes, len(value), 8388608)
			if len(value) > 1048576 || guard.textBytes > 8388608 {
				guard.textLimit = true
			}
		}
		if guard.phase == 2 && !utf8.ValidString(value) {
			guard.invalidText = true
		}
	}
}

func (guard *admission) identity(text string) {
	if guard.phase == 1 && len(text) > 4096 {
		guard.textLimit = true
	}
	guard.text(text)
}

func (guard *admission) scan(request contract.Request) {
	if !guard.collection(len(request.Instances), 256, nil, 0) ||
		!guard.collection(len(request.PlanVariables), 1024, nil, 0) ||
		!guard.collection(len(request.PlanBindings), 8192, &guard.bindings, 8192) {
		return
	}
	guard.identity(request.PlanID)
	guard.text(request.Provenance.TemplateSet, string(request.Provenance.GeneratedOn))
	if !guard.directory(request.Directory) {
		return
	}
	for i, variable := range request.PlanVariables {
		if !guard.variable(variable, "planVariables["+strconv.Itoa(i)+"]") {
			return
		}
	}
	if !guard.bindingList(request.PlanBindings, "planBindings") {
		return
	}
	for i, instance := range request.Instances {
		if !guard.collection(len(instance.Bindings), 8192, &guard.bindings, 8192) {
			return
		}
		guard.identity(instance.SourceID)
		guard.identity(instance.PermanentSectionID)
		if !guard.bindingList(instance.Bindings, instancePath(i)+".bindings") {
			return
		}
	}
}

func (guard *admission) directory(directory model.Directory) bool {
	combined := 0
	if !guard.collection(len(directory.Contacts), 2048, &combined, 4096) ||
		!guard.collection(len(directory.Roles), 2048, &combined, 4096) ||
		!guard.collection(len(directory.Teams), 2048, &combined, 4096) {
		return false
	}
	for _, contact := range directory.Contacts {
		guard.text(contact.ID, contact.Name, contact.Email, contact.Slack, contact.SlackID)
	}
	for _, role := range directory.Roles {
		if !guard.collection(len(role.Holders), 1024, &guard.holders, 8192) {
			return false
		}
		guard.text(role.ID, role.Name)
		guard.text(role.Holders...)
	}
	for _, team := range directory.Teams {
		if !guard.collection(len(team.Holders), 2048, &guard.teamEntries, 8192) {
			return false
		}
		guard.text(team.ID, team.Name, team.Contact)
		keys := make([]string, 0, len(team.Holders))
		for key := range team.Holders {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			holders := team.Holders[key]
			if !guard.collection(len(holders), 1024, &guard.holders, 8192) {
				return false
			}
			guard.text(key)
			guard.text(holders...)
		}
	}
	return true
}

func (guard *admission) variable(variable model.Variable, path string) bool {
	if !guard.collection(len(variable.Options), 1024, &guard.options, 16384) {
		return false
	}
	guard.text(variable.Name, variable.Description, variable.Example, string(variable.Type))
	guard.text(variable.Options...)
	guard.text(variable.Pattern, variable.PatternHint, string(variable.Audience))
	if !guard.value(variable.Default, path+".default") {
		return false
	}
	return guard.condition(variable.AskIf, path+".askIf")
}

func (guard *admission) value(value model.Value, path string) bool {
	if !guard.collection(len(value.List), 4096, &guard.valueItems, 32768) {
		return false
	}
	guard.text(value.Text)
	guard.text(value.List...)
	if guard.phase == 3 && guard.invalidKind == "" && (value.Kind < model.ValueKindNone || value.Kind > model.ValueKindDeferred) {
		guard.invalidKind = path
	}
	return true
}

func (guard *admission) condition(condition *model.Condition, path string) bool {
	if condition == nil {
		return true
	}
	if !guard.collection(len(condition.All), 256, &guard.clauses, 4096) {
		return false
	}
	for i, clause := range condition.All {
		clausePath := path + ".all[" + strconv.Itoa(i) + "]"
		guard.text(clause.Question, clause.Section)
		if !guard.value(clause.Answer.Is, clausePath+".answer.is") || !guard.value(clause.Answer.IsNot, clausePath+".answer.isNot") {
			return false
		}
		if !guard.comparison(clause.Answer.IsOneOf, clausePath+".answer.isOneOf") || !guard.comparison(clause.Answer.IsNotOneOf, clausePath+".answer.isNotOneOf") {
			return false
		}
	}
	return true
}

func (guard *admission) comparison(values []model.Value, path string) bool {
	if !guard.collection(len(values), 256, &guard.comparisons, 16384) {
		return false
	}
	for i, value := range values {
		if !guard.value(value, path+"["+strconv.Itoa(i)+"]") {
			return false
		}
	}
	return true
}

func (guard *admission) bindingList(bindings []contract.Binding, path string) bool {
	for i, binding := range bindings {
		guard.text(binding.Name)
		if !guard.value(binding.Value, bindingPath(path, i)+".value") {
			return false
		}
	}
	return true
}
