package writer_test

import (
	"maps"
	"reflect"
	"testing"

	contract "onboard-kit/contracts/record/writer"
	"onboard-kit/internal/model"
)

func TestDelegatedRefusals(t *testing.T) {
	cases := []struct {
		name    string
		change  func(*model.Plan)
		value   model.Value
		codes   []string
		details []map[string]string
	}{
		{
			name: "wrong-kind", change: func(*model.Plan) {},
			value:   model.Value{Kind: model.ValueKindFlag, Flag: false},
			codes:   []string{"answer.wrong-kind"},
			details: []map[string]string{{"expected": "text", "found": "flag"}},
		},
		{
			name: "choice", change: func(plan *model.Plan) {
				plan.Sections[0].Variables[0].Type = model.VariableTypeChoice
				plan.Sections[0].Variables[0].Options = []string{"org-a", "other"}
				plan.Sections[0].Variables[0].Default = model.Value{}
			},
			value: text("ORG-A"), codes: []string{"answer.not-an-option"},
			details: []map[string]string{nil},
		},
		{
			name: "inactive", change: func(plan *model.Plan) {
				enabled := false
				plan.Sections[0].Enabled = &enabled
			},
			value: text("new"), codes: []string{"answer.inactive-question"},
			details: []map[string]string{nil},
		},
		{
			name: "complete-contact-list-batch", change: func(plan *model.Plan) {
				plan.Sections[0].Variables[0].Type = model.VariableTypeContact
				plan.Sections[0].Variables[0].Multiple = true
				plan.Sections[0].Variables[0].Default = model.Value{}
				delete(plan.Answers.Sections, "personal")
			},
			value:   model.Value{Kind: model.ValueKindList, List: []string{"missing-one", "", "missing-two"}},
			codes:   []string{"directory.unknown-id", "directory.unknown-id"},
			details: []map[string]string{{"kind": "contact", "item": "0"}, {"kind": "contact", "item": "2"}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			plan := sharedPlan()
			c.change(&plan)
			before := heldRecord()
			got, problems := newWriter().Write(plan, before, contract.Intent{Operation: contract.OperationSubmit, Question: "personal.choice", Value: c.value})
			requireRecord(t, got, model.Record{})
			requireRecord(t, before, heldRecord())
			if len(problems) != len(c.codes) {
				t.Fatalf("got %#v, want codes %#v", problems, c.codes)
			}
			for i, problem := range problems {
				if problem.Code != c.codes[i] || problem.Severity != model.SeverityError ||
					problem.Question != "personal.choice" || problem.Section != "personal" ||
					!maps.Equal(problem.Detail, c.details[i]) || problem.Message == "" {
					t.Fatalf("problem[%d] %#v", i, problem)
				}
			}
		})
	}
}

func TestExactScopesAndAudiences(t *testing.T) {
	for _, question := range []string{"plan.choice", "aaa.choice", "zzz.choice"} {
		t.Run(question, func(t *testing.T) {
			plan := sharedPlan()
			plan.Variables = []model.Variable{{Name: "choice", Type: model.VariableTypeString, Audience: model.AudienceOnboardee}}
			plan.Answers = model.Answers{
				Plan: map[string]model.Value{"choice": text("plan-seed")},
				Sections: map[string]map[string]model.Value{
					"aaa": {"choice": text("aaa-seed")}, "zzz": {"choice": text("zzz-seed")},
				},
			}
			settled := true
			plan.Sections = []model.Section{
				{ID: "zzz", Title: "Later scope", Summary: "Use $choice.", Variables: []model.Variable{{Name: "choice", Audience: model.AudienceOnboardee}}, IncludeIf: &model.Condition{All: []model.Clause{{Question: "plan.choice", Answer: model.AnswerTest{IsSettled: &settled}}}}},
				{ID: "aaa", Title: "Earlier scope", Summary: "Use $choice.", Variables: []model.Variable{{Name: "choice", Audience: model.AudienceOnboardee}}, IncludeIf: &model.Condition{All: []model.Clause{{Question: "plan.choice", Answer: model.AnswerTest{IsSettled: &settled}}}}},
			}
			want := heldRecord()
			switch question {
			case "plan.choice":
				want.Answers.Plan["choice"] = text("owned")
			case "aaa.choice":
				want.Answers.Sections["aaa"] = map[string]model.Value{"choice": text("owned")}
			case "zzz.choice":
				want.Answers.Sections["zzz"] = map[string]model.Value{"choice": text("owned")}
			}
			got, problems := newWriter().Write(plan, heldRecord(), contract.Intent{Operation: contract.OperationSubmit, Question: question, Value: text("owned")})
			requireNoProblems(t, problems)
			requireRecord(t, got, want)
		})
	}
	for _, audience := range []model.Audience{"", model.AudienceOnboarder, "future-audience"} {
		plan := sharedPlan()
		plan.Sections[0].Variables[0].Audience = audience
		for _, operation := range []contract.Operation{contract.OperationSubmit, contract.OperationFollowPlan} {
			intent := contract.Intent{Operation: operation, Question: "personal.choice"}
			if operation == contract.OperationSubmit {
				intent.Value = text("owned")
			}
			got, problems := newWriter().Write(plan, heldRecord(), intent)
			requireRecord(t, got, model.Record{})
			want := []model.Problem{ownProblem("record.answer.audience", "personal.choice", "Only newcomer answers can be changed here.", "intent.question")}
			if !reflect.DeepEqual(problems, want) {
				t.Fatalf("audience %q operation %q: %#v", audience, operation, problems)
			}
		}
	}
}

func TestAmbiguousTargetIsNotChosen(t *testing.T) {
	for _, duplicateSection := range []bool{false, true} {
		plan := sharedPlan()
		if duplicateSection {
			plan.Sections = append(plan.Sections, sharedPlan().Sections[0])
		} else {
			plan.Sections[0].Variables = append(plan.Sections[0].Variables, sharedPlan().Sections[0].Variables[0])
		}
		got, problems := newWriter().Write(plan, heldRecord(), contract.Intent{Operation: contract.OperationSubmit, Question: "personal.choice", Value: text("owned")})
		requireRecord(t, got, model.Record{})
		want := []model.Problem{ownProblem("record.answer.question", "personal.choice", "Choose one declared newcomer question.", "intent.question")}
		if !reflect.DeepEqual(problems, want) {
			t.Fatalf("duplicate section %v: %#v", duplicateSection, problems)
		}
	}
}

func TestActivityUsesHeldRecordAnswer(t *testing.T) {
	for _, heldGate := range []bool{false, true} {
		plan := sharedPlan()
		plan.Variables = append(plan.Variables, model.Variable{Name: "gate", Type: model.VariableTypeBoolean, Audience: model.AudienceOnboardee})
		plan.Answers.Plan["gate"] = model.Value{Kind: model.ValueKindFlag, Flag: !heldGate}
		plan.Sections[0].IncludeIf = &model.Condition{All: []model.Clause{{Question: "plan.gate", Answer: model.AnswerTest{Is: model.Value{Kind: model.ValueKindFlag, Flag: true}}}}}
		before := heldRecord()
		before.Answers.Plan["gate"] = model.Value{Kind: model.ValueKindFlag, Flag: heldGate}
		got, problems := newWriter().Write(plan, before, contract.Intent{Operation: contract.OperationSubmit, Question: "personal.choice", Value: text("owned")})
		if heldGate {
			want := heldRecord()
			want.Answers.Plan["gate"] = model.Value{Kind: model.ValueKindFlag, Flag: true}
			want.Answers.Sections["personal"] = map[string]model.Value{"choice": text("owned")}
			requireNoProblems(t, problems)
			requireRecord(t, got, want)
		} else {
			requireRecord(t, got, model.Record{})
			if len(problems) != 1 || problems[0].Code != "answer.inactive-question" || problems[0].Severity != model.SeverityError ||
				problems[0].Question != "personal.choice" || problems[0].Section != "personal" || len(problems[0].Detail) != 0 || problems[0].Message == "" {
				t.Fatalf("inactive delegated problems: %#v", problems)
			}
		}
	}
}
