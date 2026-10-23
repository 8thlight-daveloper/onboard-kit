package encoder_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

func TestBootstrapEncoding(t *testing.T) {
	t.Parallel()
	testSectionFieldEncoding(t, "bootstrap.json")
}

func testSectionFieldEncoding(t *testing.T, fixture string) {
	t.Helper()
	var cases []struct {
		Name, Source, Output string
		Input                json.RawMessage
		Problems             []expectedProblem
	}
	if err := json.Unmarshal(readFixture(t, fixture), &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("empty section field manifest")
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			if c.Source == "" || (c.Output == "") == (len(c.Problems) == 0) {
				t.Fatal("fixture needs source and exactly one expectation")
			}
			input, err := jsonform.Plan(c.Input)
			if err != nil {
				t.Fatal(err)
			}
			before, err := jsonform.Plan(c.Input)
			if err != nil {
				t.Fatal(err)
			}
			encoder := newEncoder()
			got, problems := encoder.Encode(input)
			if string(got) != c.Output || len(problems) != len(c.Problems) {
				t.Fatalf("bytes %q, problems %#v; want %q, %#v", got, problems, c.Output, c.Problems)
			}
			for i, p := range problems {
				if p.Code != c.Problems[i].Code || p.Detail["path"] != c.Problems[i].Path || p.Severity != model.SeverityError || strings.TrimSpace(p.Message) == "" {
					t.Fatalf("problem %#v; want %#v", p, c.Problems[i])
				}
			}
			if !reflect.DeepEqual(input, before) {
				t.Fatal("input changed")
			}
			again, repeated := encoder.Encode(input)
			if !bytes.Equal(got, again) || !reflect.DeepEqual(problems, repeated) {
				t.Fatal("nondeterministic")
			}
		})
	}
}

func TestBootstrapCollisionDoesNotInspectItsValue(t *testing.T) {
	t.Parallel()
	for _, marker := range []bool{false, true} {
		plan := model.Plan{SchemaVersion: 1, Sections: []model.Section{{ID: "access", Bootstrap: marker, UnknownFields: map[string]any{"bootstrap": make(chan int)}}}}
		data, problems := newEncoder().Encode(plan)
		if len(data) != 0 || len(problems) != 1 || problems[0].Code != "encode.unknown-field-collision" || problems[0].Detail["path"] != "sections[0].bootstrap" {
			t.Fatalf("collision must precede unsupported-value traversal: %q, %#v", data, problems)
		}
	}
}

func TestBootstrapStateAcrossCalls(t *testing.T) {
	t.Parallel()
	encoder := newEncoder()
	for _, marker := range []bool{true, false, true} {
		plan := model.Plan{SchemaVersion: 1, Sections: []model.Section{{ID: "access", Bootstrap: marker}}}
		want := "schemaVersion: 1\nsections:\n  - id: access\n"
		if marker {
			want += "    bootstrap: true\n"
		}
		got, problems := encoder.Encode(plan)
		if len(problems) != 0 || string(got) != want {
			t.Fatalf("marker retained across inputs: %q, %#v", got, problems)
		}
	}
}
