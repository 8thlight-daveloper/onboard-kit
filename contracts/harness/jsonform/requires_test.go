package jsonform_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

func TestRequiresReaderPreservesIndependentTypedLists(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		input string
		want  []string
	}{
		{`{"id":"access"}`, nil},
		{`{"id":"access","requires":[]}`, nil},
		{`{"id":"access","requires":["b"," a","b",""]}`, []string{"b", " a", "b", ""}},
	} {
		want := model.Section{ID: "access", Requires: c.want}
		got, err := jsonform.Section([]byte(c.input))
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("section %#v, error %v; want %#v", got, err, want)
		}
		plan, err := jsonform.Plan([]byte(`{"sections":[` + c.input + `]}`))
		if err != nil || !reflect.DeepEqual(plan.Sections, []model.Section{want}) {
			t.Fatalf("plan sections %#v, error %v; want %#v", plan.Sections, err, want)
		}
	}
	for _, input := range []string{`{"requires":null}`, `{"requires":"a"}`, `{"requires":true}`, `{"requires":{}}`, `{"requires":[1]}`, `{"requires":[null]}`} {
		if _, err := jsonform.Section([]byte(input)); err == nil {
			t.Fatalf("fixture harness accepted invalid list %s", input)
		}
	}
}

func TestRequiresWriterPreservesListsWithoutHashNormalization(t *testing.T) {
	t.Parallel()
	for _, requires := range [][]string{nil, {}, {"b", " a", "b", ""}} {
		input := model.Section{ID: "access", After: []string{"legacy"}, Requires: requires, Owner: "lead"}
		var before []string
		if requires != nil {
			before = append([]string{}, requires...)
		}
		data, err := jsonform.SectionJSON(input)
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]any
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		want := map[string]any{"id": "access", "after": []any{"legacy"}, "owner": "lead"}
		if len(requires) != 0 {
			want["requires"] = []any{"b", " a", "b", ""}
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %s; want %#v", data, want)
		}
		planData, err := jsonform.JSON(model.Plan{Sections: []model.Section{input}})
		if err != nil {
			t.Fatal(err)
		}
		var plan map[string]any
		if err := json.Unmarshal(planData, &plan); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(plan["sections"], []any{want}) {
			t.Fatalf("plan writer dropped or changed Requires: %s", planData)
		}
		if !reflect.DeepEqual(input.Requires, before) || !reflect.DeepEqual(input.After, []string{"legacy"}) {
			t.Fatal("fixture writer changed an input list")
		}
	}
}
