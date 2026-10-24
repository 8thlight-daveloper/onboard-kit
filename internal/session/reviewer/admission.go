package reviewer

import (
	"reflect"

	"onboard-kit/internal/model"
)

const threshold uint64 = 67108864
const exceeded uint64 = threshold + 1

func add(values ...uint64) uint64 {
	total := uint64(0)
	for _, value := range values {
		if value >= exceeded || total > exceeded-value {
			return exceeded
		}
		total += value
	}
	return total
}

func multiply(left, right uint64) uint64 {
	if left == 0 || right == 0 {
		return 0
	}
	if left > exceeded/right {
		return exceeded
	}
	return left * right
}

func maximum(values ...uint64) uint64 {
	result := uint64(0)
	for _, value := range values {
		if value > result {
			result = value
		}
	}
	return result
}

func count(length int) uint64    { return uint64(length) }
func length(value string) uint64 { return count(len(value)) }

type unknownData struct{ value any }

type meter struct {
	failed         bool
	nodes          uint64
	unknown        uint64
	entries        uint64
	variables      uint64
	steps          uint64
	checks         uint64
	clauses        uint64
	comparisons    uint64
	holders        uint64
	lists          uint64
	options        uint64
	conditionPairs uint64
}

func (meter *meter) increase(total *uint64, value, limit uint64) {
	*total = add(*total, value)
	if *total > limit {
		meter.failed = true
	}
}

func admitted(plan model.Plan, record model.Record, preparation *model.Preparation, source uint64, full bool) bool {
	meter := meter{}
	planWeight := meter.weight(plan, 1, false)
	recordWeight := meter.weight(record, 1, false)
	preparationWeight := meter.weight(preparation, 1, false)
	if meter.failed {
		return false
	}
	if !full {
		return true
	}
	base := add(source, 64, planWeight, recordWeight, preparationWeight)
	render, questions := expansion(plan, record)
	eligibility := multiply(16, multiply(meter.variables, meter.clauses))
	pairs := add(multiply(32, add(multiply(meter.variables, meter.variables), multiply(count(len(plan.Sections)), count(len(plan.Sections))))), multiply(8, meter.conditionPairs))
	return add(multiply(8, add(base, render, questions)), eligibility, pairs) <= threshold
}

func (meter *meter) weight(value any, depth uint64, unknown bool) uint64 {
	if meter.failed {
		return exceeded
	}
	if wrapped, ok := value.(unknownData); ok {
		return meter.weight(wrapped.value, depth, true)
	}
	meter.increase(&meter.nodes, 1, 131072)
	if unknown {
		meter.increase(&meter.unknown, 1, 65536)
	}
	if depth > 128 {
		meter.failed = true
	}
	if meter.failed {
		return exceeded
	}
	if value == nil {
		return 64
	}
	var fields []any
	switch item := value.(type) {
	case model.Plan:
		if len(item.Sections) > 256 {
			meter.failed = true
		}
		fields = []any{item.SchemaVersion, item.ID, item.Provenance, item.Origin, item.URL, item.ResultsPath, item.Directory, item.Variables, item.Answers, item.Sections, unknownData{item.UnknownFields}}
	case model.Provenance:
		fields = []any{item.TemplateSet, item.GeneratedOn}
	case model.Origin:
		fields = []any{item.PlanID, item.Source, item.ContentHash, item.RefreshedOn}
	case model.Directory:
		if len(item.Contacts) > 2048 || len(item.Roles) > 2048 || len(item.Teams) > 2048 || len(item.Contacts)+len(item.Roles)+len(item.Teams) > 4096 {
			meter.failed = true
		}
		fields = []any{item.Contacts, item.Roles, item.Teams}
	case model.Contact:
		fields = []any{item.ID, item.Name, item.Email, item.Slack, item.SlackID}
	case model.Role:
		meter.holder(item.Holders)
		fields = []any{item.ID, item.Name, item.Holders}
	case model.Team:
		for _, holders := range item.Holders {
			meter.holder(holders)
		}
		fields = []any{item.ID, item.Name, item.Contact, item.Holders}
	case model.Answers:
		fields = []any{item.Plan, item.Sections}
	case model.Variable:
		meter.increase(&meter.variables, 1, 1024)
		if len(item.Options) > 1024 {
			meter.failed = true
		}
		meter.increase(&meter.options, count(len(item.Options)), 16384)
		fields = []any{item.Name, item.Description, item.Example, item.Type, item.Options, item.Pattern, item.PatternHint, item.Required, item.Multiple, item.Default, item.Audience, item.AskIf}
	case model.Value:
		if len(item.List) > 4096 {
			meter.failed = true
		}
		meter.increase(&meter.lists, count(len(item.List)), 32768)
		fields = []any{item.Kind, item.Text, item.Flag, item.List}
	case model.Condition:
		if len(item.All) > 256 {
			meter.failed = true
		}
		meter.increase(&meter.clauses, count(len(item.All)), 4096)
		meter.conditionPairs = add(meter.conditionPairs, multiply(count(len(item.All)), count(len(item.All))))
		fields = []any{item.All}
	case model.Clause:
		fields = []any{item.Question, item.Answer, item.Section, item.Enabled}
	case model.AnswerTest:
		if len(item.IsOneOf) > 256 || len(item.IsNotOneOf) > 256 {
			meter.failed = true
		}
		meter.increase(&meter.comparisons, add(count(len(item.IsOneOf)), count(len(item.IsNotOneOf))), 16384)
		fields = []any{item.Is, item.IsNot, item.IsOneOf, item.IsNotOneOf, item.IsSettled}
	case model.Section:
		fields = []any{item.ID, item.Title, item.Summary, item.Parent, item.Bootstrap, item.Owner, item.Timing, item.Enabled, item.Verification, item.Body, item.IncludeIf, item.After, item.Requires, item.Escalation, item.Variables, item.Steps, item.Source, unknownData{item.UnknownFields}}
	case model.Source:
		fields = []any{item.Template, item.Hash}
	case model.Step:
		meter.increase(&meter.steps, 1, 4096)
		fields = []any{item.ID, item.Name, item.Notes, item.Owner, item.Timing, item.Run, item.Open, item.Directory, item.Env, item.Escalation, item.Checks}
	case model.Check:
		meter.increase(&meter.checks, 1, 8192)
		fields = []any{item.Run, item.Ask, item.Env}
	case model.Record:
		if len(item.Outcomes) > 4096 {
			meter.failed = true
		}
		fields = []any{item.SchemaVersion, item.Account, item.StartDate, item.FirstOpenedAt, item.Plan, item.Answers, item.Outcomes, unknownData{item.UnknownFields}}
	case model.PlanReference:
		fields = []any{item.ID, item.Hash, item.URL}
	case model.Outcome:
		fields = []any{item.Section, item.Kind, item.On, item.SectionHash, item.Note, item.Differently}
	case model.Preparation:
		fields = []any{item.SchemaVersion, item.Kind, item.Account, item.PlanID, item.Recipient, item.Arrangements, unknownData{item.UnknownFields}}
	case model.PreparationRecipient:
		fields = []any{item.Name, item.CompanyEmail, item.PlannedStartDate, item.Team, item.Manager, item.OnboardingOwner, item.Buddy, item.WorkLocation, item.TimeZone}
	case model.Arrangement:
		fields = []any{item.ID, item.Section, item.Kind, item.Owner, item.Contacts, item.Reference, item.TargetDate, item.Status, item.Note}
	}
	if fields != nil {
		if unknown {
			meter.failed = true
			return exceeded
		}
		total := uint64(64)
		for _, field := range fields {
			total = add(total, meter.weight(field, depth+1, false))
		}
		return total
	}
	if unknown {
		switch value.(type) {
		case string, bool, int64, float64, []any, map[string]any:
		default:
			meter.failed = true
			return exceeded
		}
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.String:
		return add(64, count(reflected.Len()))
	case reflect.Bool:
		return 65
	case reflect.Int, reflect.Int64, reflect.Float64:
		return 72
	case reflect.Pointer:
		switch value.(type) {
		case *bool, *model.Condition, *model.Source, *model.Preparation:
		default:
			meter.failed = true
			return exceeded
		}
		if reflected.IsNil() {
			return 64
		}
		return add(64, meter.weight(reflected.Elem().Interface(), depth+1, unknown))
	case reflect.Slice:
		if !unknown {
			meter.increase(&meter.entries, count(reflected.Len()), 65536)
		}
		total := uint64(64)
		for index := 0; index < reflected.Len(); index++ {
			total = add(total, meter.weight(reflected.Index(index).Interface(), depth+1, unknown))
			if meter.failed {
				break
			}
		}
		return total
	case reflect.Map:
		if !unknown {
			meter.increase(&meter.entries, count(reflected.Len()), 65536)
		}
		total := uint64(64)
		iterator := reflected.MapRange()
		for iterator.Next() {
			total = add(total, meter.weight(iterator.Key().Interface(), depth+1, unknown), meter.weight(iterator.Value().Interface(), depth+1, unknown))
			if meter.failed {
				break
			}
		}
		return total
	default:
		meter.failed = true
		return exceeded
	}
}

func (meter *meter) holder(values []string) {
	if len(values) > 1024 {
		meter.failed = true
	}
	meter.increase(&meter.holders, count(len(values)), 8192)
}

func weight(value any) uint64 { meter := meter{}; return meter.weight(value, 1, false) }

func joinedLength(values []string, separator uint64) uint64 {
	result := uint64(0)
	for _, value := range values {
		result = add(result, length(value))
	}
	if len(values) > 1 {
		result = add(result, multiply(separator, count(len(values)-1)))
	}
	return result
}

func selectedCharge(value model.Value) uint64 {
	switch value.Kind {
	case model.ValueKindText:
		return add(64, length(value.Text))
	case model.ValueKindFlag:
		return 69
	case model.ValueKindList:
		return add(64, joinedLength(value.List, 0), multiply(16, count(len(value.List))))
	default:
		return 64
	}
}

func dollars(value string) uint64 {
	result := uint64(0)
	for index := 0; index < len(value); index++ {
		if value[index] == '$' {
			result++
		}
	}
	return result
}
