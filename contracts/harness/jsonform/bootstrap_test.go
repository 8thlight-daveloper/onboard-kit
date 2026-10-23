package jsonform_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

func TestBootstrapReaderUsesIndependentTypedValues(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		input string
		want  bool
	}{{`{"id":"access","bootstrap":true}`, true}, {`{"id":"access","bootstrap":false}`, false}, {`{"id":"access"}`, false}} {
		got, err := jsonform.Section([]byte(c.input))
		if err != nil || !reflect.DeepEqual(got, model.Section{ID: "access", Bootstrap: c.want}) {
			t.Fatalf("section %#v, error %v; want marker %v", got, err, c.want)
		}
		plan, err := jsonform.Plan([]byte(`{"sections":[` + c.input + `]}`))
		if err != nil || len(plan.Sections) != 1 || plan.Sections[0].Bootstrap != c.want {
			t.Fatalf("plan marker mismatch: %#v, %v", plan, err)
		}
	}
	for _, input := range []string{`{"bootstrap":"true"}`, `{"bootstrap":1}`, `{"bootstrap":null}`, `{"bootstrap":[]}`} {
		if _, err := jsonform.Section([]byte(input)); err == nil {
			t.Fatal("fixture harness accepted a nonboolean bootstrap")
		}
	}
}

func TestBootstrapWriterUsesIndependentPrimitiveJSON(t *testing.T) {
	t.Parallel()
	for _, marker := range []bool{false, true} {
		input := model.Section{ID: "access", Bootstrap: marker}
		data, err := jsonform.SectionJSON(input)
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]any
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		want := map[string]any{"id": "access"}
		if marker {
			want["bootstrap"] = true
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %s; want %#v", data, want)
		}
	}
}
