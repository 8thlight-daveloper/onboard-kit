package scheduler_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type scheduleCase struct {
	Name, Source string
	Plan         json.RawMessage
	Sequence     *[]string
}

type structureErrorCase struct {
	Name, Source string
	Plan         json.RawMessage
	Problems     []model.Problem
}

func readCases(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile("testdata/" + path)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		t.Fatal(err)
	}
}

func scheduleCases(t *testing.T) []scheduleCase {
	t.Helper()
	var cases []scheduleCase
	readCases(t, "cases.json", &cases)
	if len(cases) != 26 {
		t.Fatal("expected twenty-six independently ordered schedule cases")
	}
	seen := map[string]bool{}
	for _, c := range cases {
		if c.Name == "" || c.Source == "" || c.Sequence == nil || seen[c.Name] {
			t.Fatal("schedule fixture name/source/order missing or name repeated")
		}
		seen[c.Name] = true
	}
	return cases
}

func errorCases(t *testing.T) []structureErrorCase {
	t.Helper()
	var cases []structureErrorCase
	readCases(t, "structure-errors.json", &cases)
	if len(cases) != 31 {
		t.Fatal("expected thirty-one literal independent structure-error oracles")
	}
	for _, c := range cases {
		if c.Name == "" || c.Source == "" || len(c.Problems) == 0 {
			t.Fatal("structure fixture lacks source or explicit error output")
		}
	}
	return cases
}

func readPlan(t *testing.T, raw json.RawMessage) model.Plan {
	t.Helper()
	plan, err := jsonform.Plan(raw)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func ids(sections []model.Section) []string {
	var out []string
	for _, section := range sections {
		out = append(out, section.ID)
	}
	return out
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
				t.Fatal(err)
			}
			members := make([]string, len(items))
			for j, item := range items {
				if len(item) == 0 || item[0] != '"' {
					t.Fatal("members must contain only JSON strings")
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

func assertSchedule(t *testing.T, c scheduleCase, plan model.Plan, got []model.Section, problems []model.Problem) {
	t.Helper()
	if len(problems) != 0 || !slices.Equal(ids(got), *c.Sequence) {
		t.Fatalf("%s: order %v, problems %+v; want hand-specified %v without problems", c.Name, ids(got), problems, *c.Sequence)
	}
	for _, section := range got {
		found := false
		for _, held := range plan.Sections {
			if section.ID == held.ID {
				if !reflect.DeepEqual(section, held) {
					t.Fatalf("%s: section %q differs from held definition", c.Name, section.ID)
				}
				found = true
				break
			}
		}
		if !found {
			t.Fatal("schedule contains a section absent from input")
		}
	}
}

func assertErrors(t *testing.T, c structureErrorCase, got []model.Section, problems []model.Problem) {
	t.Helper()
	if len(got) != 0 || !reflect.DeepEqual(normalizedProblems(t, problems), normalizedProblems(t, c.Problems)) {
		t.Fatalf("%s: must refuse all work and preserve complete structural diagnostics: sections %v, problems %#v, want %#v", c.Name, ids(got), problems, c.Problems)
	}
}

func TestSchedulesIndependentActionOrders(t *testing.T) {
	for _, c := range scheduleCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			plan := readPlan(t, c.Plan)
			got, problems := newScheduler().Schedule(plan)
			assertSchedule(t, c, plan, got, problems)
		})
	}
}

func TestStructureErrorsRefuseAllWorkBeforeActivityFiltering(t *testing.T) {
	for _, c := range errorCases(t) {
		for _, inactive := range []bool{false, true} {
			plan := readPlan(t, c.Plan)
			for _, section := range plan.Sections {
				if section.ID == "free-independent-action" {
					t.Fatal("independent sentinel would change a structural oracle")
				}
			}
			plan.Sections = append(plan.Sections, model.Section{ID: "free-independent-action"})
			if inactive {
				for i := range plan.Sections {
					off := false
					plan.Sections[i].Enabled = &off
					plan.Sections[i].IncludeIf = &model.Condition{All: []model.Clause{{Question: "plan.unanswered", Answer: model.AnswerTest{Is: model.Value{Kind: model.ValueKindText, Text: "yes"}}}}}
				}
			}
			got, problems := newScheduler().Schedule(plan)
			assertErrors(t, c, got, problems)
		}
	}
}

func TestSchedulePermutationPurityAndFreshResultSlots(t *testing.T) {
	shared := newScheduler()
	for _, c := range scheduleCases(t) {
		for _, order := range []string{"held", "reverse", "rotate"} {
			plan := readPlan(t, c.Plan)
			if order == "reverse" {
				slices.Reverse(plan.Sections)
			}
			if order == "rotate" && len(plan.Sections) > 1 {
				plan.Sections = append(plan.Sections[1:], plan.Sections[0])
			}
			before, err := jsonform.JSON(plan)
			if err != nil {
				t.Fatal(err)
			}
			got, problems := shared.Schedule(plan)
			assertSchedule(t, c, plan, got, problems)
			if len(got) > 0 {
				got[0] = model.Section{ID: "changed result slot"}
				slices.Reverse(got)
			}
			after, err := jsonform.JSON(plan)
			if err != nil || string(before) != string(after) {
				t.Fatalf("%s/%s: scheduling or result-slot mutation changed input", c.Name, order)
			}
		}
	}
}

func TestScheduleStateAndDiagnosticStorageAreIndependent(t *testing.T) {
	shared := newScheduler()
	valid := scheduleCases(t)
	invalid := errorCases(t)
	for round := 0; round < 2; round++ {
		for i := range invalid {
			index := i
			if round == 1 {
				index = len(invalid) - 1 - i
			}
			c := invalid[index]
			plan := readPlan(t, c.Plan)
			before := readPlan(t, c.Plan)
			got, problems := shared.Schedule(plan)
			assertErrors(t, c, got, problems)
			for i := range problems {
				problems[i].Message = "changed returned message"
				for key := range problems[i].Detail {
					problems[i].Detail[key] = "changed returned detail"
				}
			}
			if !reflect.DeepEqual(plan, before) {
				t.Fatal("failed schedule changed input")
			}
			got, problems = shared.Schedule(plan)
			assertErrors(t, c, got, problems)
			v := valid[index%len(valid)]
			vPlan := readPlan(t, v.Plan)
			got, problems = shared.Schedule(vPlan)
			assertSchedule(t, v, vPlan, got, problems)
		}
	}
}

func TestScheduleErrorsAreDeclarationPermutationInvariant(t *testing.T) {
	for _, c := range errorCases(t) {
		plan := readPlan(t, c.Plan)
		slices.Reverse(plan.Sections)
		got, problems := newScheduler().Schedule(plan)
		assertErrors(t, c, got, problems)
	}
}

func TestScheduleConcurrentSharedInstance(t *testing.T) {
	shared := newScheduler()
	var wait sync.WaitGroup
	for round := 0; round < 2; round++ {
		for _, c := range scheduleCases(t) {
			plan := readPlan(t, c.Plan)
			before := readPlan(t, c.Plan)
			wait.Add(1)
			go func() {
				defer wait.Done()
				got, problems := shared.Schedule(plan)
				if len(problems) != 0 || !slices.Equal(ids(got), *c.Sequence) || !reflect.DeepEqual(plan, before) {
					t.Errorf("%s: concurrent shared schedule changed input or gave %v and %+v instead of %v", c.Name, ids(got), problems, *c.Sequence)
				}
			}()
		}
		for _, c := range errorCases(t) {
			plan := readPlan(t, c.Plan)
			wait.Add(1)
			go func() {
				defer wait.Done()
				got, problems := shared.Schedule(plan)
				if len(got) != 0 || !reflect.DeepEqual(normalizedProblems(t, problems), normalizedProblems(t, c.Problems)) {
					t.Errorf("%s: concurrent shared schedule lost structural diagnostics", c.Name)
				}
			}()
		}
	}
	wait.Wait()
}

func TestScheduleDoesNotReadOutcomeShapedUnknownData(t *testing.T) {
	for _, c := range scheduleCases(t) {
		for _, kind := range []string{"verified", "failed", "deferred", ""} {
			plan := readPlan(t, c.Plan)
			plan.UnknownFields = map[string]any{"outcomes": []any{map[string]any{"section": "b", "outcome": kind}}}
			got, problems := newScheduler().Schedule(plan)
			assertSchedule(t, c, plan, got, problems)
		}
	}
}

func TestScheduleReturnsFullHeldDefinitions(t *testing.T) {
	for _, requires := range [][]string{nil, {}, {"b", "b"}} {
		off := false
		plan := model.Plan{Sections: []model.Section{
			{ID: "a", Title: "Held title", Summary: "Held summary", Bootstrap: true, Requires: requires, Source: &model.Source{Template: "held", Hash: "baseline"}, Variables: []model.Variable{{Name: "q"}}, UnknownFields: map[string]any{"held": []any{"value"}}},
			{ID: "b", Bootstrap: true},
			{ID: "c", Enabled: &off},
		}}
		want := []model.Section{plan.Sections[0], plan.Sections[1]}
		if len(requires) != 0 {
			want = []model.Section{plan.Sections[1], plan.Sections[0]}
		}
		got, problems := newScheduler().Schedule(plan)
		if len(problems) != 0 || !reflect.DeepEqual(got, want) {
			t.Fatalf("held definitions changed: got %#v and %+v, want %#v", got, problems, want)
		}
	}
}
