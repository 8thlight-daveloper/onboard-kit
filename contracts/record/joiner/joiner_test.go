package joiner_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type fixtureCase struct {
	Name   string `json:"name"`
	Plan   string `json:"plan"`
	Record string `json:"record"`
	Joined string `json:"joined"`
	Source string `json:"source"`
}

func loadCases(t *testing.T) []fixtureCase {
	t.Helper()
	var cases []fixtureCase
	if err := json.Unmarshal(readFixture(t, "cases.json"), &cases); err != nil {
		t.Fatalf("cases.json: %v", err)
	}
	return cases
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}

func loadPlan(t *testing.T, name string) model.Plan {
	t.Helper()
	plan, err := jsonform.Plan(readFixture(t, name))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return plan
}

func loadRecord(t *testing.T, name string) model.Record {
	t.Helper()
	record, err := jsonform.Record(readFixture(t, name))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return record
}

func showPlan(t *testing.T, plan model.Plan) string {
	t.Helper()
	data, err := jsonform.JSON(plan)
	if err != nil {
		t.Fatalf("write plan: %v", err)
	}
	return string(data)
}

func requireSamePlan(t *testing.T, got, want model.Plan) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("plans differ\n--- got ---\n%s\n--- want ---\n%s", showPlan(t, got), showPlan(t, want))
	}
}

func memoryOf(root any) map[uintptr]bool {
	owned := map[uintptr]bool{}
	var walk func(v reflect.Value)
	walk = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Pointer:
			if v.IsNil() {
				return
			}
			owned[v.Pointer()] = true
			walk(v.Elem())
		case reflect.Map:
			if v.IsNil() {
				return
			}
			owned[v.Pointer()] = true
			iter := v.MapRange()
			for iter.Next() {
				walk(iter.Value())
			}
		case reflect.Slice:
			if v.Len() == 0 {
				return
			}
			owned[v.Pointer()] = true
			for i := 0; i < v.Len(); i++ {
				walk(v.Index(i))
			}
		case reflect.Interface:
			if !v.IsNil() {
				walk(v.Elem())
			}
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				walk(v.Field(i))
			}
		}
	}
	walk(reflect.ValueOf(root))
	return owned
}

func requireNothingShared(t *testing.T, result model.Plan, arguments ...any) {
	t.Helper()
	held := memoryOf(result)
	for _, argument := range arguments {
		for address := range memoryOf(argument) {
			if held[address] {
				t.Errorf("the result shares a map, list, or pointer with an argument")
				return
			}
		}
	}
}

func TestEveryCaseCitesItsSource(t *testing.T) {
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
	}
}

func TestJoinsEachCase(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got := newJoiner().Join(loadPlan(t, c.Plan), loadRecord(t, c.Record))
			requireSamePlan(t, got, loadPlan(t, c.Joined))
		})
	}
}

func TestJoiningTheResultAgainChangesNothing(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			record := loadRecord(t, c.Record)
			once := newJoiner().Join(loadPlan(t, c.Plan), record)
			twice := newJoiner().Join(once, record)
			requireSamePlan(t, twice, once)
		})
	}
}

func TestAnEmptyRecordGivesThePlanUnchanged(t *testing.T) {
	empty := map[string]model.Record{
		"zero":    {},
		"version": {SchemaVersion: 1},
		"empty maps": {SchemaVersion: 1, Answers: model.Answers{
			Plan:     map[string]model.Value{},
			Sections: map[string]map[string]model.Value{},
		}},
	}
	for _, c := range loadCases(t) {
		for name, record := range empty {
			t.Run(c.Name+"/"+name, func(t *testing.T) {
				plan := loadPlan(t, c.Plan)
				got := newJoiner().Join(plan, record)
				requireSamePlan(t, got, loadPlan(t, c.Plan))
				requireNothingShared(t, got, plan, record)
			})
		}
	}
}

func TestOnlyTheAnswersOfTheRecordAreUsed(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			record := loadRecord(t, c.Record)
			onlyAnswers := model.Record{Answers: record.Answers}
			withOthers := model.Record{
				SchemaVersion: 9,
				Account:       "someone-else",
				StartDate:     "2030-01-01",
				Plan:          model.PlanReference{ID: "another-plan", Hash: "ffffff", URL: "https://elsewhere.example/plan"},
				Answers:       record.Answers,
				Outcomes:      []model.Outcome{{Section: "access", Kind: model.OutcomeVerified, On: "2030-01-01T00:00:00Z"}},
				UnknownFields: map[string]any{"x-note": "text", "id": "record-id", "sections": []any{"x"}},
			}
			want := newJoiner().Join(loadPlan(t, c.Plan), onlyAnswers)
			got := newJoiner().Join(loadPlan(t, c.Plan), withOthers)
			requireSamePlan(t, got, want)
		})
	}
}

func TestTheResultSharesNothingWithTheArguments(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			plan := loadPlan(t, c.Plan)
			record := loadRecord(t, c.Record)
			got := newJoiner().Join(plan, record)
			requireNothingShared(t, got, plan, record)
		})
	}
}

func TestChangingTheResultLeavesTheArgumentsAlone(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			plan := loadPlan(t, c.Plan)
			record := loadRecord(t, c.Record)
			got := newJoiner().Join(plan, record)
			for _, values := range got.Answers.Plan {
				for i := range values.List {
					values.List[i] = "changed"
				}
			}
			for _, values := range got.Answers.Sections {
				for _, value := range values {
					for i := range value.List {
						value.List[i] = "changed"
					}
				}
			}
			for _, values := range got.Answers.Sections {
				for name := range values {
					delete(values, name)
				}
			}
			for name := range got.Answers.Plan {
				delete(got.Answers.Plan, name)
			}
			for key := range got.UnknownFields {
				delete(got.UnknownFields, key)
			}
			for i := range got.Sections {
				got.Sections[i].Title = "changed"
			}
			requireSamePlan(t, plan, loadPlan(t, c.Plan))
			requireSameRecord(t, record, loadRecord(t, c.Record))
		})
	}
}

func requireSameRecord(t *testing.T, got, want model.Record) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		gotJSON, _ := jsonform.RecordJSON(got)
		wantJSON, _ := jsonform.RecordJSON(want)
		t.Errorf("records differ\n--- got ---\n%s\n--- want ---\n%s", gotJSON, wantJSON)
	}
}

func TestTheArgumentsAreNotChanged(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			plan := loadPlan(t, c.Plan)
			record := loadRecord(t, c.Record)
			newJoiner().Join(plan, record)
			requireSamePlan(t, plan, loadPlan(t, c.Plan))
			requireSameRecord(t, record, loadRecord(t, c.Record))
		})
	}
}

func TestTheResultIsTheSameEveryTime(t *testing.T) {
	for _, c := range loadCases(t) {
		plan := loadPlan(t, c.Plan)
		record := loadRecord(t, c.Record)
		first := newJoiner().Join(plan, record)
		for i := 0; i < 5; i++ {
			requireSamePlan(t, newJoiner().Join(plan, record), first)
		}
	}
}
