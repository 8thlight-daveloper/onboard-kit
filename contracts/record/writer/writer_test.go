package writer_test

import (
	"reflect"
	"testing"

	contract "onboard-kit/contracts/record/writer"
	"onboard-kit/internal/model"
)

func text(value string) model.Value {
	return model.Value{Kind: model.ValueKindText, Text: value}
}

func sharedPlan() model.Plan {
	return model.Plan{
		SchemaVersion: 1,
		ID:            "plan-one",
		Variables:     []model.Variable{{Name: "organization", Type: model.VariableTypeString}},
		Answers: model.Answers{
			Plan:     map[string]model.Value{"organization": text("shared-only")},
			Sections: map[string]map[string]model.Value{"personal": {"choice": text("org-a")}},
		},
		Sections: []model.Section{{
			ID: "personal", Title: "Personal choices", Summary: "Use $choice.",
			Variables: []model.Variable{{
				Name: "choice", Description: "Your choice", Type: model.VariableTypeString,
				Audience: model.AudienceOnboardee, Default: text("default-example"), Example: "example-only",
			}},
		}},
	}
}

func heldRecord() model.Record {
	return model.Record{
		SchemaVersion: 1, Account: "newcomer", StartDate: "legacy-date",
		FirstOpenedAt: " held\nnot a timestamp\u2028\u2029 ",
		Plan:          model.PlanReference{ID: "plan-one", Hash: "held-hash", URL: "plans/shared.obk.yaml"},
		Answers: model.Answers{
			Plan: map[string]model.Value{},
			Sections: map[string]map[string]model.Value{
				"gone": {
					"hidden":   {Kind: model.ValueKindList, List: []string{"first", "", "third"}},
					"empty":    {Kind: model.ValueKindList, List: []string{}},
					"deferred": {Kind: model.ValueKindDeferred},
				},
				"empty-scope": {},
			},
		},
		Outcomes: []model.Outcome{
			{Section: "z-removed", Kind: model.OutcomeVerified, On: "legacy-on", SectionHash: "old", Note: "Keep removed", Differently: "Keep this"},
			{Section: "a-removed", Kind: model.OutcomeDeferred, On: "older", Note: "Keep other"},
		},
		UnknownFields: map[string]any{
			"future": map[string]any{
				"items":     []any{"held", map[string]any{"nested": true}},
				"emptyList": []any{}, "emptyMap": map[string]any{},
			},
		},
	}
}

func requireRecord(t *testing.T, got, want model.Record) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("record mismatch\ngot  %#v\nwant %#v", got, want)
	}
}

func requireNoProblems(t *testing.T, problems []model.Problem) {
	t.Helper()
	if len(problems) != 0 {
		t.Fatalf("unexpected problems: %#v", problems)
	}
}

func TestEqualPreseedBecomesOwned(t *testing.T) {
	plan, before := sharedPlan(), heldRecord()
	intent := contract.Intent{Operation: contract.OperationSubmit, Question: "personal.choice", Value: text("org-a")}
	want := heldRecord()
	want.Answers.Sections["personal"] = map[string]model.Value{"choice": text("org-a")}
	writer := newWriter()
	got, problems := writer.Write(plan, before, intent)
	requireNoProblems(t, problems)
	requireRecord(t, got, want)
	requireRecord(t, before, heldRecord())
	if !reflect.DeepEqual(plan, sharedPlan()) {
		t.Fatal("shared plan changed")
	}
	again, problems := writer.Write(plan, got, intent)
	requireNoProblems(t, problems)
	requireRecord(t, again, want)
	planB := sharedPlan()
	planB.Answers.Sections["personal"]["choice"] = text("org-b")
	follower, problems := writer.Write(planB, before, contract.Intent{Operation: contract.OperationFollowPlan, Question: "personal.choice"})
	requireNoProblems(t, problems)
	requireRecord(t, follower, heldRecord())
	owner, problems := writer.Write(planB, got, intent)
	requireNoProblems(t, problems)
	requireRecord(t, owner, want)
}

func TestFollowPlanRemovesOnlyExactOverride(t *testing.T) {
	for _, extra := range []bool{false, true} {
		t.Run(map[bool]string{false: "last-key", true: "other-key"}[extra], func(t *testing.T) {
			before, want := heldRecord(), heldRecord()
			before.Answers.Sections["personal"] = map[string]model.Value{"choice": text("org-a")}
			if extra {
				before.Answers.Sections["personal"]["hidden"] = text("keep")
				want.Answers.Sections["personal"] = map[string]model.Value{"hidden": text("keep")}
			}
			plan := sharedPlan()
			plan.Answers.Sections["personal"]["choice"] = text("org-b")
			got, problems := newWriter().Write(plan, before, contract.Intent{Operation: contract.OperationFollowPlan, Question: "personal.choice"})
			requireNoProblems(t, problems)
			requireRecord(t, got, want)
			if !reflect.DeepEqual(before.Answers.Sections["personal"]["choice"], text("org-a")) {
				t.Fatal("original override changed")
			}
		})
	}
}

func TestSuggestionsNeverBecomeAnswers(t *testing.T) {
	plan := sharedPlan()
	delete(plan.Answers.Sections, "personal")
	before := heldRecord()
	got, problems := newWriter().Write(plan, before, contract.Intent{Operation: contract.OperationFollowPlan, Question: "personal.choice"})
	requireNoProblems(t, problems)
	requireRecord(t, got, heldRecord())
	for _, accepted := range []string{"default-example", "example-only"} {
		want := heldRecord()
		want.Answers.Sections["personal"] = map[string]model.Value{"choice": text(accepted)}
		got, problems := newWriter().Write(plan, before, contract.Intent{Operation: contract.OperationSubmit, Question: "personal.choice", Value: text(accepted)})
		requireNoProblems(t, problems)
		requireRecord(t, got, want)
	}
}

func TestSettledAndDeferredValues(t *testing.T) {
	cases := []struct {
		name  string
		kind  model.VariableType
		value model.Value
	}{
		{"spaces", model.VariableTypeString, text("  ")},
		{"false", model.VariableTypeBoolean, model.Value{Kind: model.ValueKindFlag, Flag: false}},
		{"true", model.VariableTypeBoolean, model.Value{Kind: model.ValueKindFlag, Flag: true}},
		{"ordered-list", model.VariableTypeCommand, model.Value{Kind: model.ValueKindList, List: []string{"go", "", "test", "go"}}},
		{"deferred", model.VariableTypeString, model.Value{Kind: model.ValueKindDeferred}},
		{"deferred-text", model.VariableTypeString, text("deferred")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			plan := sharedPlan()
			plan.Sections[0].Variables[0].Type = c.kind
			plan.Sections[0].Variables[0].Default = model.Value{}
			delete(plan.Answers.Sections, "personal")
			want := heldRecord()
			wantValue := c.value
			if c.name == "ordered-list" {
				wantValue = model.Value{Kind: model.ValueKindList, List: []string{"go", "", "test", "go"}}
			}
			want.Answers.Sections["personal"] = map[string]model.Value{"choice": wantValue}
			got, problems := newWriter().Write(plan, heldRecord(), contract.Intent{Operation: contract.OperationSubmit, Question: "personal.choice", Value: c.value})
			requireNoProblems(t, problems)
			requireRecord(t, got, want)
		})
	}
}

func ownProblem(code, question, message, path string) model.Problem {
	return model.Problem{Code: code, Severity: model.SeverityError, Question: question, Message: message, Detail: map[string]string{"path": path}}
}

func TestUnansweredNeedsExplicitIntent(t *testing.T) {
	for _, value := range []model.Value{
		{}, {Kind: model.ValueKindNone, Text: "inactive"}, text(""),
		{Kind: model.ValueKindList}, {Kind: model.ValueKindList, List: []string{}},
		{Kind: model.ValueKindList, List: []string{"", ""}},
	} {
		got, problems := newWriter().Write(sharedPlan(), heldRecord(), contract.Intent{Operation: contract.OperationSubmit, Question: "personal.choice", Value: value})
		requireRecord(t, got, model.Record{})
		want := []model.Problem{ownProblem("record.answer.empty", "personal.choice", "Choose an answer, defer, or explicitly follow the plan.", "intent.value")}
		if !reflect.DeepEqual(problems, want) {
			t.Fatalf("value %#v: got %#v, want %#v", value, problems, want)
		}
	}
}

func TestOwnRefusalPhaseOrder(t *testing.T) {
	cases := []struct {
		name                string
		intent              contract.Intent
		code, message, path string
	}{
		{"unknown-operation-first", contract.Intent{Operation: "SUBMIT", Question: "missing", Value: text("")}, "record.answer.operation", "Unsupported answer operation.", "intent.operation"},
		{"follow-value-before-question", contract.Intent{Operation: contract.OperationFollowPlan, Question: "missing", Value: text("held")}, "record.answer.follow-value", "Follow-plan must not carry an answer.", "intent.value"},
		{"inactive-follow-value", contract.Intent{Operation: contract.OperationFollowPlan, Question: "personal.choice", Value: model.Value{List: []string{}}}, "record.answer.follow-value", "Follow-plan must not carry an answer.", "intent.value"},
		{"unknown-before-empty", contract.Intent{Operation: contract.OperationSubmit, Question: "missing"}, "record.answer.question", "Choose one declared newcomer question.", "intent.question"},
		{"case-exact", contract.Intent{Operation: contract.OperationSubmit, Question: "Personal.choice", Value: text("new")}, "record.answer.question", "Choose one declared newcomer question.", "intent.question"},
		{"default-audience", contract.Intent{Operation: contract.OperationSubmit, Question: "plan.organization"}, "record.answer.audience", "Only newcomer answers can be changed here.", "intent.question"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, problems := newWriter().Write(sharedPlan(), heldRecord(), c.intent)
			requireRecord(t, got, model.Record{})
			want := []model.Problem{ownProblem(c.code, c.intent.Question, c.message, c.path)}
			if !reflect.DeepEqual(problems, want) {
				t.Fatalf("got %#v, want %#v", problems, want)
			}
		})
	}
}
