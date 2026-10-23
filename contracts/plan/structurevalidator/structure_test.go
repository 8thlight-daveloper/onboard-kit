package structurevalidator_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type structureCase struct {
	Name, Source string
	Plan         json.RawMessage
	Expected     []model.Problem
}

func structureCases(t *testing.T) []structureCase {
	t.Helper()
	data, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []structureCase
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 42 {
		t.Fatal("expected forty-two independently specified structural cases")
	}
	hardening, err := os.ReadFile("testdata/hardening.json")
	if err != nil {
		t.Fatal(err)
	}
	var additional []structureCase
	decoder = json.NewDecoder(bytes.NewReader(hardening))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&additional); err != nil {
		t.Fatal(err)
	}
	if len(additional) != 11 {
		t.Fatal("expected eleven independently specified hardening cases")
	}
	cases = append(cases, additional...)
	seen := map[string]bool{}
	for _, c := range cases {
		if c.Name == "" || c.Source == "" || seen[c.Name] {
			t.Fatal("fixture name/source missing or repeated")
		}
		seen[c.Name] = true
		for _, problem := range c.Expected {
			if problem.Code == "" || problem.Severity != model.SeverityError || strings.TrimSpace(problem.Message) == "" || problem.Question != "" {
				t.Fatalf("%s has an incomplete diagnostic oracle", c.Name)
			}
		}
	}
	return cases
}

func readStructurePlan(t *testing.T, c structureCase) model.Plan {
	t.Helper()
	plan, err := jsonform.Plan(c.Plan)
	if err != nil {
		t.Fatalf("%s: %v", c.Name, err)
	}
	return plan
}

func normalizedProblems(t *testing.T, problems []model.Problem) []model.Problem {
	t.Helper()
	if len(problems) == 0 {
		return nil
	}
	out := append([]model.Problem{}, problems...)
	for i := range out {
		if len(out[i].Detail) == 0 {
			out[i].Detail = nil
			continue
		}
		out[i].Detail = make(map[string]string, len(problems[i].Detail))
		for key, value := range problems[i].Detail {
			out[i].Detail[key] = value
		}
		if raw, ok := out[i].Detail["members"]; ok {
			var compact bytes.Buffer
			if err := json.Compact(&compact, []byte(raw)); err != nil || compact.String() != raw || !strings.HasPrefix(raw, "[") {
				t.Fatal("members must be a compact valid JSON array")
			}
			var items []json.RawMessage
			if err := json.Unmarshal([]byte(raw), &items); err != nil {
				t.Fatal("members must be a JSON array")
			}
			members := make([]string, len(items))
			for j, item := range items {
				if len(item) == 0 || item[0] != '"' {
					t.Fatal("every members item must be a JSON string")
				}
				if err := json.Unmarshal(item, &members[j]); err != nil {
					t.Fatal(err)
				}
			}
			encoded, err := json.Marshal(members)
			if err != nil {
				t.Fatal(err)
			}
			out[i].Detail["members"] = string(encoded)
		}
	}
	return out
}

func assertStructureProblems(t *testing.T, name string, got, want []model.Problem) {
	t.Helper()
	if !reflect.DeepEqual(normalizedProblems(t, got), normalizedProblems(t, want)) {
		t.Fatalf("%s: problems\n got: %#v\nwant: %#v", name, got, want)
	}
}

func TestStructureCases(t *testing.T) {
	for _, c := range structureCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			plan := readStructurePlan(t, c)
			assertStructureProblems(t, c.Name, newValidator().Validate(plan), c.Expected)
		})
	}
}

func TestStructureDeclarationPermutationsAndPurity(t *testing.T) {
	shared := newValidator()
	for _, c := range structureCases(t) {
		for _, order := range []string{"held", "reverse", "rotate"} {
			plan := readStructurePlan(t, c)
			if order == "reverse" {
				for i, j := 0, len(plan.Sections)-1; i < j; i, j = i+1, j-1 {
					plan.Sections[i], plan.Sections[j] = plan.Sections[j], plan.Sections[i]
				}
			}
			if order == "rotate" && len(plan.Sections) > 1 {
				plan.Sections = append(plan.Sections[1:], plan.Sections[0])
			}
			before, err := jsonform.JSON(plan)
			if err != nil {
				t.Fatal(err)
			}
			assertStructureProblems(t, c.Name+"/"+order, shared.Validate(plan), c.Expected)
			after, err := jsonform.JSON(plan)
			if err != nil || string(before) != string(after) {
				t.Fatalf("%s: validation changed held plan/list data", c.Name)
			}
		}
	}
}

func TestStructureStateAndProblemStorageAreIndependent(t *testing.T) {
	shared := newValidator()
	cases := structureCases(t)
	for round := 0; round < 2; round++ {
		for i := range cases {
			index := i
			if round == 1 {
				index = len(cases) - 1 - i
			}
			c := cases[index]
			plan := readStructurePlan(t, c)
			before := readStructurePlan(t, c)
			got := shared.Validate(plan)
			assertStructureProblems(t, c.Name, got, c.Expected)
			for i := range got {
				got[i].Code = "changed returned problem"
				if got[i].Detail != nil {
					for key := range got[i].Detail {
						got[i].Detail[key] = "changed returned detail"
					}
					got[i].Detail["added"] = "changed returned detail"
				}
			}
			if !reflect.DeepEqual(plan, before) {
				t.Fatal("changing returned diagnostic storage changed an argument")
			}
			assertStructureProblems(t, c.Name+"/after result mutation", shared.Validate(plan), c.Expected)
		}
	}
}

func TestStructureIgnoresActivityAnswersAndUnrelatedContent(t *testing.T) {
	for _, c := range structureCases(t) {
		plan := readStructurePlan(t, c)
		plan.ID = "ignored-plan-id"
		plan.Answers = model.Answers{Plan: map[string]model.Value{"secret-answer-do-not-echo": {Kind: model.ValueKindText, Text: "private-content-do-not-echo"}}}
		plan.UnknownFields = map[string]any{"private-content-do-not-echo": []any{"private-content-do-not-echo"}}
		for i := range plan.Sections {
			section := &plan.Sections[i]
			off := false
			section.Enabled = &off
			section.IncludeIf = &model.Condition{All: []model.Clause{{Question: "plan.missing", Answer: model.AnswerTest{Is: model.Value{Kind: model.ValueKindText, Text: "private-content-do-not-echo"}}}}}
			section.Title = "private-content-do-not-echo"
			section.Body = "private-content-do-not-echo"
			section.Source = &model.Source{Template: "private-content-do-not-echo", Hash: "private-content-do-not-echo"}
			section.UnknownFields = map[string]any{"private-content-do-not-echo": true}
		}
		got := newValidator().Validate(plan)
		assertStructureProblems(t, c.Name, got, c.Expected)
		encoded, err := json.Marshal(got)
		if err != nil {
			t.Fatal(err)
		}
		for _, private := range []string{"private-content-do-not-echo", "secret-answer-do-not-echo", "secret-parent-do-not-echo", "secret-target-do-not-echo", "second-private-target", "secret-after-do-not-echo"} {
			if strings.Contains(string(encoded), private) {
				t.Fatalf("%s: diagnostic exposed unrelated content or an unknown reference", c.Name)
			}
		}
	}
}

func TestConcurrentStructureCallsShareNoState(t *testing.T) {
	cases := structureCases(t)
	shared := newValidator()
	var wait sync.WaitGroup
	for round := 0; round < 3; round++ {
		for _, c := range cases {
			plan := readStructurePlan(t, c)
			before := readStructurePlan(t, c)
			wait.Add(1)
			go func() {
				defer wait.Done()
				got := shared.Validate(plan)
				if !reflect.DeepEqual(normalizedProblems(t, got), normalizedProblems(t, c.Expected)) {
					t.Errorf("%s: concurrent shared-instance problems got %#v, want %#v", c.Name, got, c.Expected)
				}
				if !reflect.DeepEqual(plan, before) {
					t.Errorf("%s: concurrent shared-instance call changed input", c.Name)
				}
			}()
		}
	}
	wait.Wait()
}

func TestMembersEncodingIsDeterministicWithinAnImplementation(t *testing.T) {
	shared := newValidator()
	for _, c := range structureCases(t) {
		plan := readStructurePlan(t, c)
		first := shared.Validate(plan)
		assertStructureProblems(t, c.Name, first, c.Expected)
		before, err := json.Marshal(first)
		if err != nil {
			t.Fatal(err)
		}
		for call := 0; call < 3; call++ {
			again := shared.Validate(plan)
			after, err := json.Marshal(again)
			if err != nil || string(before) != string(after) {
				t.Fatalf("%s: diagnostic encoding changed between identical calls", c.Name)
			}
		}
	}
}
