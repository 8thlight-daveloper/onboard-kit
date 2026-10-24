package reviewer

import (
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	contract "onboard-kit/contracts/session/reviewer"
	"onboard-kit/internal/model"
)

func own(stage, code, message, path string) []contract.Diagnostic {
	problem := model.Problem{Code: "session." + code, Severity: model.SeverityError, Message: message}
	if path != "" {
		problem.Detail = map[string]string{"path": path}
	}
	return []contract.Diagnostic{{Stage: stage, Problem: problem}}
}

func appendProblems(diagnostics []contract.Diagnostic, stage string, problems []model.Problem) []contract.Diagnostic {
	for _, problem := range problems {
		diagnostics = append(diagnostics, contract.Diagnostic{Stage: stage, Problem: clone(problem)})
	}
	return diagnostics
}

func fatal(diagnostics []contract.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func appendResource(diagnostics []contract.Diagnostic, stage, path string) []contract.Diagnostic {
	return append(diagnostics, own(stage, "resource", "Use a smaller onboarding session or fewer repeated values.", path)...)
}

func boundedText(value string, maximum int) bool {
	return len(value) <= maximum && utf8.ValidString(value)
}

func sourceAdmission(source contract.Source) []contract.Diagnostic {
	for _, body := range []struct {
		body []byte
		path string
	}{{source.Plan, "plan"}, {source.Preparation, "preparation"}, {source.Record, "record"}} {
		if len(body.body) > 1048576 {
			return appendResource(nil, "source.admission", body.path)
		}
	}
	if !boundedText(source.Account, 4096) {
		return appendResource(nil, "source.admission", "account")
	}
	if !boundedText(source.PlanID, 4096) {
		return appendResource(nil, "source.admission", "planId")
	}
	return nil
}

func sourceSize(source contract.Source) uint64 {
	return uint64(len(source.Plan)) + uint64(len(source.Preparation)) + uint64(len(source.Record))
}

func intentAdmission(intent contract.Intent) []contract.Diagnostic {
	total := uint64(0)
	for _, item := range []struct {
		text  string
		limit int
	}{{string(intent.Operation), 4096}, {intent.Question, 4096}, {intent.Section, 4096}, {string(intent.Kind), 4096}, {intent.Note, 65536}, {intent.Differently, 65536}, {intent.Value.Text, 65536}} {
		if !boundedText(item.text, item.limit) {
			return appendResource(nil, "intent.admission", "intent")
		}
		total += uint64(len(item.text))
	}
	if len(intent.Value.List) > 4096 {
		return appendResource(nil, "intent.admission", "intent")
	}
	for _, item := range intent.Value.List {
		if !boundedText(item, 65536) {
			return appendResource(nil, "intent.admission", "intent")
		}
		total += uint64(len(item))
		if total > 262144 {
			return appendResource(nil, "intent.admission", "intent")
		}
	}
	if total > 262144 {
		return appendResource(nil, "intent.admission", "intent")
	}
	return nil
}

func accountValid(account string) bool {
	if len(account) == 0 {
		return false
	}
	for index, value := range []byte(account) {
		alphanumeric := value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
		if !alphanumeric && (index == 0 || value != '.' && value != '_' && value != '-') {
			return false
		}
	}
	return true
}

func validIntent(intent contract.Intent) bool {
	if intent.Value.Kind < model.ValueKindNone || intent.Value.Kind > model.ValueKindDeferred {
		return false
	}
	emptyValue := intent.Value.Kind == model.ValueKindNone && intent.Value.Text == "" && !intent.Value.Flag && intent.Value.List == nil
	switch intent.Operation {
	case contract.OperationAnswer:
		return intent.Section == "" && intent.Kind == "" && intent.Note == "" && intent.Differently == ""
	case contract.OperationFollowPlan:
		return emptyValue && intent.Section == "" && intent.Kind == "" && intent.Note == "" && intent.Differently == ""
	case contract.OperationOutcome:
		return emptyValue && intent.Question == ""
	default:
		return false
	}
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func indexed(name string, index int) string { return name + "[" + strconv.Itoa(index) + "]" }

func ownership(plan model.Plan, record model.Record) []contract.Diagnostic {
	var result []contract.Diagnostic
	declared := func(variables []model.Variable, name string) bool {
		for _, variable := range variables {
			if variable.Name == name {
				return variable.Audience != model.AudienceOnboardee
			}
		}
		return false
	}
	for index, name := range sortedKeys(record.Answers.Plan) {
		if declared(plan.Variables, name) {
			result = append(result, own("record.ownership", "record-owned-answer", "Keep preparation answers out of the newcomer record.", indexed("answers.plan", index))...)
		}
	}
	for sectionIndex, scope := range sortedKeys(record.Answers.Sections) {
		var variables []model.Variable
		for _, section := range plan.Sections {
			if section.ID == scope {
				variables = section.Variables
				break
			}
		}
		for valueIndex, name := range sortedKeys(record.Answers.Sections[scope]) {
			if declared(variables, name) {
				result = append(result, own("record.ownership", "record-owned-answer", "Keep preparation answers out of the newcomer record.", indexed("answers.sections", sectionIndex)+"."+indexed("values", valueIndex))...)
			}
		}
	}
	return result
}

func referenceValid(value string) bool {
	if value == "" {
		return false
	}
	for index, value := range []byte(value) {
		if value >= 'a' && value <= 'z' || value >= '0' && value <= '9' {
			continue
		}
		if index == 0 || value != '-' {
			return false
		}
	}
	return true
}

func staleSupport(preparation model.Preparation, problems []model.Problem) bool {
	if len(problems) == 0 {
		return false
	}
	allowed := make(map[string]struct {
		value string
		code  string
	})
	allowed["recipient.team"] = struct {
		value string
		code  string
	}{preparation.Recipient.Team, "preparation.unknown-team"}
	for path, value := range map[string]string{"recipient.manager": preparation.Recipient.Manager, "recipient.onboardingOwner": preparation.Recipient.OnboardingOwner, "recipient.buddy": preparation.Recipient.Buddy} {
		allowed[path] = struct {
			value string
			code  string
		}{value, "preparation.unknown-contact"}
	}
	for index, value := range preparation.Arrangements {
		prefix := indexed("arrangements", index)
		allowed[prefix+".section"] = struct {
			value string
			code  string
		}{value.Section, "preparation.unknown-section"}
		allowed[prefix+".owner"] = struct {
			value string
			code  string
		}{value.Owner, "preparation.unknown-contact"}
		for contactIndex, contact := range value.Contacts {
			allowed[prefix+"."+indexed("contacts", contactIndex)] = struct {
				value string
				code  string
			}{contact, "preparation.unknown-contact"}
		}
	}
	for _, problem := range problems {
		if problem.Severity != model.SeverityError || problem.Question != "" || problem.Section != "" || len(problem.Detail) != 1 {
			return false
		}
		path, hasPath := problem.Detail["path"]
		value, found := allowed[path]
		if !hasPath || !found || value.code != problem.Code || !referenceValid(value.value) {
			return false
		}
		if strings.HasSuffix(path, ".section") {
			switch value.value {
			case "plan", "yes", "no", "y", "n", "true", "false", "on", "off", "null":
				return false
			}
		}
	}
	return true
}
