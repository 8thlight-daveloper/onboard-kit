package reviewer

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	api "onboard-kit/contracts/invitation/reviewer"
	"onboard-kit/internal/model"
)

func fatal(problems []model.Problem) bool {
	for _, problem := range problems {
		if problem.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func wrap(stage string, problems []model.Problem) []api.Diagnostic {
	var diagnostics []api.Diagnostic
	for _, problem := range problems {
		diagnostics = append(diagnostics, api.Diagnostic{Stage: stage, Display: dependencyDisplay(stage), Problem: owned(problem)})
	}
	return diagnostics
}

func dependencyDisplay(stage string) string {
	switch stage {
	case "inspect.envelope":
		return "Ask your onboarding owner for a readable invitation archive."
	case "inspect.plan-decode":
		return "Ask your onboarding owner for a readable invitation plan."
	case "inspect.plan-authoring":
		return "Review the invitation plan declarations with its owner."
	case "inspect.plan-validation":
		return "Review invitation plan integrity and organizational answers."
	case "inspect.preparation-decode":
		return "Ask your onboarding owner for readable invitation preparation."
	case "inspect.record-decode":
		return "Ask your onboarding owner for a readable invitation record seed."
	case "inspect.locator":
		return "Ask your onboarding owner for valid record destinations."
	case "inspect.preparation-validation":
		return "Review the preparation against this invitation."
	case "previous.plan-decode":
		return "The previous plan copy cannot be read for comparison."
	case "previous.plan-authoring":
		return "Review the previous plan declarations before using them for comparison."
	case "previous.plan-validation":
		return "Review the previous plan integrity before using it for comparison."
	case "review.preparation-decode":
		return "Keep local preparation intact and request help reading it."
	case "review.preparation-validation":
		return "Ask your onboarding owner to repair unavailable local support."
	case "review.record-decode":
		return "Keep the local record intact and request help reading it."
	case "review.reconciliation":
		return "Keep the local record intact and resolve its plan binding."
	case "review.effective-validation":
		return "Review stored personal answers against the incoming plan."
	case "review.walker":
		return "Resolve the session's activity or history problem before opening."
	case "review.record-encode":
		return "Keep the local record intact and resolve its unwritable content."
	case "begin.record-encode":
		return "Resolve unwritable record content before beginning."
	}
	return ""
}

func own(stage, code, path string) []api.Diagnostic {
	message := ""
	switch code {
	case "plan-id":
		message = "Supply an invitation plan identity."
	case "seed":
		message = "Use an identity-only seed matching the received invitation."
	case "first-steps":
		message = "Supply nonempty UTF-8 invitation guidance."
	case "incoming":
		message = "Use an incoming projection consistent with its archive."
	case "local":
		message = "Supply consistent local presence observations."
	case "source-size":
		message = "Keep each local source within the supported byte limit."
	case "plan-choice":
		message = "Choose undecided or use-incoming explicitly."
	case "known-record-missing":
		message = "Restore the registered record instead of creating another seed."
	case "record-account":
		message = "Use the record for this invitation account."
	case "record-plan":
		message = "Use this plan's record or explicitly bind an unbound record."
	case "record-owned-answer":
		message = "Keep onboarder-owned answers out of the personal record."
	case "outcome-kind":
		message = "Use verified, failed or deferred recorded outcomes."
	case "history-time":
		message = "Preserve a real UTC time for interpreted history."
	case "proposal":
		message = "Use the complete proposal derived from its exact source evidence."
	case "event":
		message = "Supply one supported begin event only for a new record."
	case "output-size":
		message = "Keep the encoded record within the supported byte limit."
	case "resource":
		message = "Use a smaller onboarding plan or fewer repeated values."
	case "previous-resource":
		message = "Previous plan comparison is unavailable within the resource limit."
	}
	severity := model.SeverityError
	if code == "previous-resource" {
		severity = model.SeverityWarning
	}
	return []api.Diagnostic{{Stage: stage, Display: message, Problem: model.Problem{Code: "invitation." + code, Severity: severity, Message: message, Detail: map[string]string{"path": path}}}}
}

func resource(stage, path string) []api.Diagnostic { return own(stage, "resource", path) }

func localProblems(local api.Local) []api.Diagnostic {
	var diagnostics []api.Diagnostic
	for _, entry := range []struct {
		body api.Body
		role string
	}{{local.Plan, "plan"}, {local.Preparation, "preparation"}, {local.Record, "record"}} {
		if !entry.body.Present && len(entry.body.Bytes) > 0 {
			diagnostics = append(diagnostics, own("review.local", "local", "local."+entry.role)...)
		}
		if len(entry.body.Bytes) > 1048576 {
			diagnostics = append(diagnostics, own("review.local", "source-size", "local."+entry.role)...)
		}
	}
	if local.KnownRecordMissing && local.Record.Present {
		diagnostics = append(diagnostics, own("review.local", "local", "local.knownRecordMissing")...)
	}
	return diagnostics
}

func ownershipProblems(plan model.Plan, record model.Record) []api.Diagnostic {
	declarations := map[string]map[string]bool{"plan": {}}
	for _, variable := range plan.Variables {
		declarations["plan"][variable.Name] = variable.Audience == "" || variable.Audience == model.AudienceOnboarder
	}
	for _, section := range plan.Sections {
		declarations[section.ID] = map[string]bool{}
		for _, variable := range section.Variables {
			declarations[section.ID][variable.Name] = variable.Audience == "" || variable.Audience == model.AudienceOnboarder
		}
	}
	var diagnostics []api.Diagnostic
	for i, key := range sortedKeys(record.Answers.Plan) {
		if declarations["plan"][key] {
			diagnostics = append(diagnostics, own("review.record-ownership", "record-owned-answer", fmt.Sprintf("answers.plan[%d]", i))...)
		}
	}
	for i, scope := range sortedKeys(record.Answers.Sections) {
		for j, key := range sortedKeys(record.Answers.Sections[scope]) {
			if declarations[scope][key] {
				diagnostics = append(diagnostics, own("review.record-ownership", "record-owned-answer", fmt.Sprintf("answers.sections[%d].values[%d]", i, j))...)
			}
		}
	}
	return diagnostics
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func historyProblems(record model.Record) []api.Diagnostic {
	var diagnostics []api.Diagnostic
	if record.FirstOpenedAt != "" && !validTimestamp(string(record.FirstOpenedAt), false) {
		diagnostics = append(diagnostics, own("review.record-history", "history-time", "record.firstOpenedAt")...)
	}
	for i, outcome := range record.Outcomes {
		if outcome.Kind != model.OutcomeVerified && outcome.Kind != model.OutcomeFailed && outcome.Kind != model.OutcomeDeferred {
			diagnostics = append(diagnostics, own("review.record-history", "outcome-kind", fmt.Sprintf("record.outcomes[%d].kind", i))...)
		}
		if !validTimestamp(string(outcome.On), false) {
			diagnostics = append(diagnostics, own("review.record-history", "history-time", fmt.Sprintf("record.outcomes[%d].on", i))...)
		}
	}
	return diagnostics
}

func validTimestamp(text string, event bool) bool {
	if len(text) < 20 {
		return false
	}
	for _, index := range []int{0, 1, 2, 3, 5, 6, 8, 9, 11, 12, 14, 15, 17, 18} {
		if text[index] < '0' || text[index] > '9' {
			return false
		}
	}
	if text[4] != '-' || text[7] != '-' || text[10] != 'T' || text[13] != ':' || text[16] != ':' {
		return false
	}
	year, _ := strconv.Atoi(text[:4])
	month, _ := strconv.Atoi(text[5:7])
	day, _ := strconv.Atoi(text[8:10])
	hour, _ := strconv.Atoi(text[11:13])
	minute, _ := strconv.Atoi(text[14:16])
	second, _ := strconv.Atoi(text[17:19])
	if year == 0 || month < 1 || month > 12 || day < 1 || hour > 23 || minute > 59 || second > 59 {
		return false
	}
	days := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
		days[1] = 29
	}
	if day > days[month-1] {
		return false
	}
	index := 19
	if text[index] == '.' {
		start := index + 1
		index = start
		for index < len(text) && text[index] >= '0' && text[index] <= '9' {
			index++
		}
		if index == start || (event && index-start > 9) {
			return false
		}
	}
	zone := text[index:]
	return zone == "Z" || (!event && zone == "+00:00")
}

func referenceID(value string, section bool) bool {
	if value == "" {
		return false
	}
	for i, c := range []byte(value) {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || (i > 0 && c == '-')) {
			return false
		}
	}
	if section {
		switch value {
		case "plan", "yes", "no", "y", "n", "true", "false", "on", "off", "null":
			return false
		}
	}
	return true
}

func staleReference(preparation model.Preparation, problem model.Problem) bool {
	path := problem.Detail["path"]
	if problem.Code == "preparation.unknown-team" {
		return path == "recipient.team" && referenceID(preparation.Recipient.Team, false)
	}
	if problem.Code == "preparation.unknown-contact" {
		switch path {
		case "recipient.manager":
			return referenceID(preparation.Recipient.Manager, false)
		case "recipient.onboardingOwner":
			return referenceID(preparation.Recipient.OnboardingOwner, false)
		case "recipient.buddy":
			return referenceID(preparation.Recipient.Buddy, false)
		}
	}
	for i, arrangement := range preparation.Arrangements {
		prefix := fmt.Sprintf("arrangements[%d].", i)
		if problem.Code == "preparation.unknown-section" && path == prefix+"section" {
			return referenceID(arrangement.Section, true)
		}
		if problem.Code == "preparation.unknown-contact" {
			if path == prefix+"owner" {
				return referenceID(arrangement.Owner, false)
			}
			for j, contact := range arrangement.Contacts {
				if path == fmt.Sprintf("%scontacts[%d]", prefix, j) {
					return referenceID(contact, false)
				}
			}
		}
	}
	return false
}

func owned[T any](value T) T { return clone(reflect.ValueOf(value)).Interface().(T) }

func clone(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.New(value.Type()).Elem()
		result.Set(clone(value.Elem()))
		return result
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.New(value.Type().Elem())
		result.Elem().Set(clone(value.Elem()))
		return result
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.MakeMapWithSize(value.Type(), value.Len())
		entries := value.MapRange()
		for entries.Next() {
			result.SetMapIndex(clone(entries.Key()), clone(entries.Value()))
		}
		return result
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := 0; i < value.Len(); i++ {
			result.Index(i).Set(clone(value.Index(i)))
		}
		return result
	case reflect.Struct:
		result := reflect.New(value.Type()).Elem()
		for i := 0; i < value.NumField(); i++ {
			result.Field(i).Set(clone(value.Field(i)))
		}
		return result
	default:
		return value
	}
}

func equalIncoming(left, right api.Incoming) bool {
	return equalProjection(reflect.ValueOf(left), reflect.ValueOf(right), "Incoming")
}

func equalProposal(left, right api.Proposal) bool {
	return equalProjection(reflect.ValueOf(left), reflect.ValueOf(right), "Proposal")
}

func equalProjection(left, right reflect.Value, path string) bool {
	if left.Type() != right.Type() {
		return false
	}
	switch left.Kind() {
	case reflect.Slice:
		if left.Len() != right.Len() {
			return false
		}
		outer := false
		for _, suffix := range []string{"Incoming.Warnings", "Review.Artifacts", "Review.PreparationProblems", "Review.IncomingWarnings", "Review.EffectivePlanWarnings", "Review.SessionProblems", "Comparison.Problems", "Reconciliation.NewQuestions", "Reconciliation.RemovedQuestions", "Reconciliation.AnswersToReview", "Reconciliation.NewSections", "Reconciliation.RemovedSections", "Reconciliation.ChangedOutcomes", "Session.Activities", "Session.Outcomes", "Session.Questions", "Session.Blockers"} {
			if strings.HasSuffix(path, suffix) {
				outer = true
				break
			}
		}
		if left.Type().Elem().Kind() == reflect.Uint8 {
			outer = true
		}
		if !outer && left.IsNil() != right.IsNil() {
			return false
		}
		for i := 0; i < left.Len(); i++ {
			if !equalProjection(left.Index(i), right.Index(i), path+"[]") {
				return false
			}
		}
		return true
	case reflect.Struct:
		for i := 0; i < left.NumField(); i++ {
			field := left.Type().Field(i)
			child := path + "." + field.Name
			if field.Type.Kind() == reflect.Struct {
				child = path + "." + field.Type.Name()
			}
			if !equalProjection(left.Field(i), right.Field(i), child) {
				return false
			}
		}
		return true
	case reflect.Pointer, reflect.Interface:
		if left.IsNil() || right.IsNil() {
			return left.IsNil() == right.IsNil()
		}
		return equalProjection(left.Elem(), right.Elem(), path)
	case reflect.Map:
		if left.IsNil() != right.IsNil() || left.Len() != right.Len() {
			return false
		}
		entries := left.MapRange()
		for entries.Next() {
			other := right.MapIndex(entries.Key())
			if !other.IsValid() || !equalProjection(entries.Value(), other, path+"{}") {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(left.Interface(), right.Interface())
	}
}
