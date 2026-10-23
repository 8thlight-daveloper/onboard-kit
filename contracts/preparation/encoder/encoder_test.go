package encoder_test

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

type expectedProblem struct {
	Code string `json:"code"`
	Path string `json:"path"`
}

type fixture struct {
	Name     string            `json:"name"`
	Source   string            `json:"source"`
	Input    model.Preparation `json:"input"`
	Output   string            `json:"output"`
	Problems []expectedProblem `json:"problems"`
}

func TestFixtures(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []fixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("fixture manifest is empty")
	}
	names := make(map[string]bool, len(fixtures))
	for _, f := range fixtures {
		if f.Name == "" || names[f.Name] {
			t.Fatalf("empty or duplicate fixture name %q", f.Name)
		}
		names[f.Name] = true
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			if f.Source == "" || (f.Output == "" && len(f.Problems) == 0) {
				t.Fatal("fixture needs source and output or problems")
			}
			before, err := json.Marshal(f.Input)
			if err != nil {
				t.Fatal(err)
			}
			encoder := newEncoder()
			got, problems := encoder.Encode(f.Input)
			check(t, got, problems, f.Output, f.Problems)
			again, repeated := encoder.Encode(f.Input)
			if string(got) != string(again) || !reflect.DeepEqual(problems, repeated) {
				t.Fatal("not deterministic")
			}
			after, err := json.Marshal(f.Input)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatal("input changed")
			}
		})
	}
}

func check(t *testing.T, got []byte, problems []model.Problem, want string, expected []expectedProblem) {
	t.Helper()
	if string(got) != want {
		t.Errorf("bytes:\n%q\nwant:\n%q", got, want)
	}
	if len(problems) != len(expected) {
		t.Fatalf("problems: %#v; want %#v", problems, expected)
	}
	for i, p := range problems {
		if string(p.Code) != expected[i].Code || p.Detail["path"] != expected[i].Path || p.Severity != model.Severity("error") || strings.TrimSpace(p.Message) == "" {
			t.Errorf("problem %d: %#v; want %#v", i, p, expected[i])
		}
	}
}

func TestUnsupportedValues(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		value any
	}{
		{"integer8", int8(1)}, {"integer32", int32(1)}, {"unsigned", uint(1)},
		{"float32", float32(1)}, {"nan", math.NaN()}, {"infinity", math.Inf(1)},
		{"negative-infinity", math.Inf(-1)}, {"function", func() {}},
		{"non-string-map", map[int]any{1: "value"}}, {"pointer", new(string)},
		{"invalid-text", string([]byte{0xff})},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			input := model.Preparation{UnknownFields: map[string]any{"x": []any{c.value}}}
			data, problems := newEncoder().Encode(input)
			check(t, data, problems, "", []expectedProblem{{"encode.unsupported-value", "x[0]"}})
		})
	}
}

func TestIntegerKinds(t *testing.T) {
	t.Parallel()
	input := model.Preparation{UnknownFields: map[string]any{"a": int(-2), "b": int64(9223372036854775807), "c": float64(-2)}}
	data, problems := newEncoder().Encode(input)
	check(t, data, problems, "schemaVersion: 0\nkind: \"\"\na: -2\nb: 9223372036854775807\nc: -2.0\n", nil)
}

func TestCycles(t *testing.T) {
	t.Parallel()
	mapping := map[string]any{}
	mapping["self"] = mapping
	list := make([]any, 1)
	list[0] = list
	for name, value := range map[string]any{"map": mapping, "list": list} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			input := model.Preparation{UnknownFields: map[string]any{"x": value}}
			data, problems := newEncoder().Encode(input)
			path := "x.self"
			if name == "list" {
				path = "x[0]"
			}
			check(t, data, problems, "", []expectedProblem{{"encode.unsupported-value", path}})
		})
	}
}

func TestSharedValuesAreNotCycles(t *testing.T) {
	t.Parallel()
	shared := map[string]any{"value": "kept"}
	input := model.Preparation{UnknownFields: map[string]any{"a": shared, "b": shared}}
	data, problems := newEncoder().Encode(input)
	check(t, data, problems, "schemaVersion: 0\nkind: \"\"\na:\n  value: kept\nb:\n  value: kept\n", nil)
}

func TestSharedSlicePrefixIsNotACycle(t *testing.T) {
	t.Parallel()
	items := []any{"kept", nil}
	items[1] = items[:1]
	input := model.Preparation{UnknownFields: map[string]any{"items": items}}
	data, problems := newEncoder().Encode(input)
	check(t, data, problems, "schemaVersion: 0\nkind: \"\"\nitems:\n  - kept\n  - - kept\n", nil)
}

func TestProblemOrderAndCollisionShortCircuit(t *testing.T) {
	t.Parallel()
	input := model.Preparation{
		Recipient:     model.PreparationRecipient{Name: string([]byte{0xff})},
		Arrangements:  []model.Arrangement{{}, {Note: string([]byte{0xff})}},
		UnknownFields: map[string]any{"account": func() {}, "z": []any{func() {}, map[string]any{"b": func() {}, "a": func() {}}}},
	}
	data, problems := newEncoder().Encode(input)
	check(t, data, problems, "", []expectedProblem{
		{"encode.unsupported-value", "recipient.name"},
		{"encode.empty-item", "arrangements[0]"},
		{"encode.unsupported-value", "arrangements[1].note"},
		{"encode.unknown-field-collision", "account"},
		{"encode.unsupported-value", "z[0]"},
		{"encode.unsupported-value", "z[1].a"},
		{"encode.unsupported-value", "z[1].b"},
	})
}

func TestInvalidUTF8(t *testing.T) {
	t.Parallel()
	marker := "private-person@example.invalid"
	bad := marker + string([]byte{0xff})
	cases := []struct {
		name, path string
		input      model.Preparation
	}{
		{"kind", "kind", model.Preparation{Kind: bad}},
		{"account", "account", model.Preparation{Account: bad}},
		{"plan", "planId", model.Preparation{PlanID: bad}},
		{"contact", "arrangements[0].contacts[0]", model.Preparation{Arrangements: []model.Arrangement{{Contacts: []string{bad}}}}},
		{"key", bad, model.Preparation{UnknownFields: map[string]any{bad: func() {}}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			data, problems := newEncoder().Encode(c.input)
			check(t, data, problems, "", []expectedProblem{{"encode.unsupported-value", c.path}})
			for _, problem := range problems {
				if strings.Contains(problem.Message, marker) {
					t.Fatal("diagnostic message exposed supplied personal data")
				}
			}
		})
	}
}
