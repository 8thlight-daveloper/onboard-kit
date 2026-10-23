package validator_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

var structureCodes = []string{
	"plan.duplicate-section-id",
	"plan.parent-unknown", "plan.parent-self", "plan.parent-cycle",
	"plan.requires-unknown", "plan.requires-self", "plan.requires-cycle",
	"plan.bootstrap-requires-nonbootstrap",
	"plan.after-unknown", "plan.after-self", "plan.after-cycle",
	"plan.bootstrap-after-nonbootstrap", "plan.ordering-conflict",
}

type delegatedStructureCase struct {
	Name, Source string
	Plan         json.RawMessage
	Expected     []model.Problem
}

func delegatedStructureCases(t *testing.T) []delegatedStructureCase {
	t.Helper()
	var cases []delegatedStructureCase
	if err := json.Unmarshal(readFixture(t, "structure-delegation.json"), &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 53 {
		t.Fatal("expected all fifty-three independent structural oracles")
	}
	codes := map[string]bool{}
	for _, c := range cases {
		if c.Name == "" || c.Source == "" {
			t.Fatal("delegated structural case lacks source/name")
		}
		for _, problem := range c.Expected {
			if !slices.Contains(structureCodes, problem.Code) {
				t.Fatal("literal oracle contains an unjudged structural code")
			}
			codes[problem.Code] = true
		}
	}
	for _, code := range structureCodes {
		if !codes[code] || !slices.Contains(judgedPlanCodes, code) {
			t.Fatalf("structural code %s lacks an oracle or wrapper visibility", code)
		}
	}
	return cases
}

func delegatedStructurePlan(t *testing.T, c delegatedStructureCase) model.Plan {
	t.Helper()
	plan, err := jsonform.Plan(c.Plan)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func normalizedStructureProblems(t *testing.T, problems []model.Problem) []model.Problem {
	t.Helper()
	var out []model.Problem
	for _, problem := range problems {
		if !slices.Contains(structureCodes, problem.Code) {
			continue
		}
		var detail map[string]string
		if len(problem.Detail) > 0 {
			detail = map[string]string{}
			for key, value := range problem.Detail {
				detail[key] = value
			}
		}
		problem.Detail = detail
		if raw, ok := detail["members"]; ok {
			var compact bytes.Buffer
			if err := json.Compact(&compact, []byte(raw)); err != nil || compact.String() != raw || !strings.HasPrefix(raw, "[") {
				t.Fatal("structural members must be a compact valid JSON array")
			}
			var items []json.RawMessage
			if err := json.Unmarshal([]byte(raw), &items); err != nil {
				t.Fatal(err)
			}
			members := make([]string, len(items))
			for i, item := range items {
				if len(item) == 0 || item[0] != '"' {
					t.Fatal("structural members must contain JSON strings")
				}
				if err := json.Unmarshal(item, &members[i]); err != nil {
					t.Fatal(err)
				}
			}
			encoded, err := json.Marshal(members)
			if err != nil {
				t.Fatal(err)
			}
			problem.Detail["members"] = string(encoded)
		}
		out = append(out, problem)
	}
	return out
}

func assertDelegatedStructure(t *testing.T, c delegatedStructureCase, got []model.Problem) {
	t.Helper()
	if !reflect.DeepEqual(normalizedStructureProblems(t, got), normalizedStructureProblems(t, c.Expected)) {
		t.Fatalf("%s: delegated structural output got %#v, want exact independent %#v", c.Name, got, c.Expected)
	}
}

func TestValidateDelegatesAllStructureOnTheOriginalPlan(t *testing.T) {
	for _, c := range delegatedStructureCases(t) {
		for _, inactive := range []bool{false, true} {
			mode := "held"
			if inactive {
				mode = "all-off"
			}
			t.Run(c.Name+"/"+mode, func(t *testing.T) {
				t.Parallel()
				plan := delegatedStructurePlan(t, c)
				if inactive {
					for i := range plan.Sections {
						off := false
						plan.Sections[i].Enabled = &off
						plan.Sections[i].IncludeIf = &model.Condition{All: []model.Clause{{Question: "plan.unanswered", Answer: model.AnswerTest{Is: model.Value{Kind: model.ValueKindText, Text: "yes"}}}}}
					}
				}
				assertDelegatedStructure(t, c, newValidator().Validate(plan))
			})
		}
	}
}

func TestStructuralProblemsComposeWithoutGlobalShortCircuit(t *testing.T) {
	secret := "ghp_" + "FAKEFAKEFAKEFAKEFAKEFAKEFAKEFAKEFAKE"
	for _, duplicates := range []bool{false, true} {
		plan := model.Plan{
			Directory: model.Directory{Contacts: []model.Contact{{ID: "unreachable", Name: "Contact"}}},
			Variables: []model.Variable{{Name: "flag", Type: model.VariableTypeBoolean}, {Name: "owner"}, {Name: "owner"}},
			Answers:   model.Answers{Plan: map[string]model.Value{"flag": {Kind: model.ValueKindText, Text: secret}}},
			Sections:  []model.Section{{ID: "invalid", Parent: "missing"}},
		}
		want := []expectedProblem{
			{Code: "directory.contact-unreachable", Severity: "warning", Detail: map[string]string{"contact": "unreachable"}},
			{Code: "plan.duplicate-variable-name", Severity: "error", Detail: map[string]string{"scope": "plan", "name": "owner"}},
			{Code: "plan.parent-unknown", Severity: "error", Section: "invalid", Detail: map[string]string{"field": "parent"}},
			{Code: "answer.wrong-kind", Severity: "error", Question: "plan.flag", Detail: map[string]string{"expected": "flag", "found": "text"}},
			{Code: "secret.token-prefix", Severity: "error", Question: "plan.flag"},
		}
		if duplicates {
			plan.Sections = append(plan.Sections, model.Section{ID: "dup"}, model.Section{ID: "dup"})
			want = []expectedProblem{
				{Code: "directory.contact-unreachable", Severity: "warning", Detail: map[string]string{"contact": "unreachable"}},
				{Code: "plan.duplicate-section-id", Severity: "error", Section: "dup"},
				{Code: "plan.duplicate-variable-name", Severity: "error", Detail: map[string]string{"scope": "plan", "name": "owner"}},
				{Code: "answer.wrong-kind", Severity: "error", Question: "plan.flag", Detail: map[string]string{"expected": "flag", "found": "text"}},
				{Code: "secret.token-prefix", Severity: "error", Question: "plan.flag"},
			}
		}
		validator := newValidator()
		assertProblems(t, validator.Validate(plan), want, []string{"FAKEFAKE"})
		assertProblems(t, validator.CheckValue(plan, "plan.flag", model.Value{Kind: model.ValueKindFlag, Flag: true}), nil, nil)
		assertProblems(t, validator.CheckValue(plan, "plan.flag", model.Value{Kind: model.ValueKindText, Text: "not a flag"}), []expectedProblem{{Code: "answer.wrong-kind", Severity: "error", Question: "plan.flag", Detail: map[string]string{"expected": "flag", "found": "text"}}}, nil)
		assertProblems(t, validator.CheckValue(plan, "plan.owner", model.Value{Kind: model.ValueKindText, Text: secret}), []expectedProblem{{Code: "answer.ambiguous-question", Severity: "error", Question: "plan.owner"}}, []string{"FAKEFAKE"})
	}
}

func TestValidRequiresAndLegacyGoalsGiveNoInventedGraphErrors(t *testing.T) {
	for _, sections := range [][]model.Section{
		{{ID: "first", Title: "First", Summary: "First activity", Timing: model.TimingWeekOne, After: []string{"later"}}, {ID: "later", Title: "Later", Summary: "Later activity", Timing: model.TimingDays31To90, Requires: []string{"first"}}},
		{{ID: "parent", Title: "Parent", Summary: "Parent activity", Requires: []string{"child"}}, {ID: "child", Title: "Child", Summary: "Child activity", Parent: "parent"}},
	} {
		plan := model.Plan{SchemaVersion: 1, ID: "6f1c2a40-3b1e-4c1a-9d52-0a1b2c3d4e5f", Provenance: model.Provenance{TemplateSet: "test", GeneratedOn: "2026-10-23"}, Sections: sections}
		if got := newValidator().Validate(plan); len(got) != 0 {
			t.Fatalf("valid relationships/goals acquired unsupported diagnostics: %#v", got)
		}
	}
}

func TestRequiresInactiveQuestionsStillCheckStoredSecretsButNotKind(t *testing.T) {
	secret := "ghp_" + "FAKEFAKEFAKEFAKEFAKEFAKEFAKEFAKEFAKE"
	off := false
	plan := model.Plan{Sections: []model.Section{{ID: "dependent", Requires: []string{"target"}, Variables: []model.Variable{{Name: "flag", Type: model.VariableTypeBoolean}}}, {ID: "target", Enabled: &off}}, Answers: model.Answers{Sections: map[string]map[string]model.Value{"dependent": {"flag": {Kind: model.ValueKindText, Text: secret}}}}}
	validator := newValidator()
	assertProblems(t, validator.Validate(plan), []expectedProblem{{Code: "secret.token-prefix", Severity: "error", Question: "dependent.flag", Section: "dependent"}}, []string{"FAKEFAKE"})
	assertProblems(t, validator.CheckValue(plan, "dependent.flag", model.Value{Kind: model.ValueKindFlag, Flag: true}), []expectedProblem{{Code: "answer.inactive-question", Severity: "error", Question: "dependent.flag", Section: "dependent"}}, nil)
	plan.Sections[1].Enabled = nil
	assertProblems(t, validator.Validate(plan), []expectedProblem{
		{Code: "answer.wrong-kind", Severity: "error", Question: "dependent.flag", Section: "dependent", Detail: map[string]string{"expected": "flag", "found": "text"}},
		{Code: "secret.token-prefix", Severity: "error", Question: "dependent.flag", Section: "dependent"},
	}, []string{"FAKEFAKE"})
}

func TestDelegatedStructureStatePurityAndPermutations(t *testing.T) {
	shared := newValidator()
	cases := delegatedStructureCases(t)
	for round := 0; round < 2; round++ {
		for i := range cases {
			index := i
			if round == 1 {
				index = len(cases) - 1 - i
			}
			c := cases[index]
			plan := delegatedStructurePlan(t, c)
			before := delegatedStructurePlan(t, c)
			got := shared.Validate(plan)
			assertDelegatedStructure(t, c, got)
			for i := range got {
				got[i].Message = "changed returned problem"
				for key := range got[i].Detail {
					got[i].Detail[key] = "changed returned detail"
				}
			}
			if !reflect.DeepEqual(plan, before) {
				t.Fatal("delegation changed input")
			}
			assertDelegatedStructure(t, c, shared.Validate(plan))
			slices.Reverse(plan.Sections)
			assertDelegatedStructure(t, c, shared.Validate(plan))
		}
	}
}

func TestDelegatedStructureConcurrentSharedValidator(t *testing.T) {
	shared := newValidator()
	var wait sync.WaitGroup
	for _, c := range delegatedStructureCases(t) {
		plan := delegatedStructurePlan(t, c)
		before := delegatedStructurePlan(t, c)
		wait.Add(1)
		go func() {
			defer wait.Done()
			got := shared.Validate(plan)
			if !reflect.DeepEqual(normalizedStructureProblems(t, got), normalizedStructureProblems(t, c.Expected)) || !reflect.DeepEqual(plan, before) {
				t.Errorf("%s: concurrent structural delegation lost diagnostics or changed input", c.Name)
			}
		}()
	}
	wait.Wait()
}
