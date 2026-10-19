package encoder_test

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type expectedProblem struct {
	Code string `json:"code"`
	Path string `json:"path"`
}

type fixtureCase struct {
	Name     string            `json:"name"`
	Input    string            `json:"input"`
	Output   string            `json:"output"`
	Problems []expectedProblem `json:"problems"`
	Source   string            `json:"source"`
}

func loadCases(t *testing.T) []fixtureCase {
	t.Helper()
	data := readFixture(t, "cases.json")
	var cases []fixtureCase
	if err := json.Unmarshal(data, &cases); err != nil {
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

func TestEveryCaseCitesItsSource(t *testing.T) {
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
	}
}

func TestEncodesTheCanonicalFile(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Output == "" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got, problems := newEncoder().Encode(loadPlan(t, c.Input))
			if len(problems) != 0 {
				t.Fatalf("unexpected problems: %+v", problems)
			}
			want := readFixture(t, c.Output)
			if !bytes.Equal(got, want) {
				t.Errorf("bytes differ\n--- got ---\n%s\n--- want ---\n%s", got, want)
			}
		})
	}
}

func TestReportsProblemsAndWritesNothing(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Output != "" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got, problems := newEncoder().Encode(loadPlan(t, c.Input))
			if len(got) != 0 {
				t.Errorf("want no bytes, got %q", got)
			}
			var actual []expectedProblem
			for _, p := range problems {
				if p.Severity != model.SeverityError {
					t.Errorf("problem %s has severity %q, want error", p.Code, p.Severity)
				}
				actual = append(actual, expectedProblem{Code: p.Code, Path: p.Detail["path"]})
			}
			if !reflect.DeepEqual(actual, c.Problems) {
				t.Errorf("problems differ\n got: %+v\nwant: %+v", actual, c.Problems)
			}
		})
	}
}

func TestReportsAnUnsupportedUnknownFieldValue(t *testing.T) {
	plan := model.Plan{SchemaVersion: 1, UnknownFields: map[string]any{"x-bad": make(chan int)}}
	got, problems := newEncoder().Encode(plan)
	if len(got) != 0 {
		t.Errorf("want no bytes, got %q", got)
	}
	if len(problems) != 1 || problems[0].Code != "encode.unsupported-value" || problems[0].Detail["path"] != "x-bad" {
		t.Errorf("want one encode.unsupported-value at x-bad, got %+v", problems)
	}
}

func TestOutputIsDeterministic(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Output == "" {
			continue
		}
		plan := loadPlan(t, c.Input)
		first, _ := newEncoder().Encode(plan)
		for i := 0; i < 5; i++ {
			again, _ := newEncoder().Encode(plan)
			if !bytes.Equal(first, again) {
				t.Fatalf("%s: two calls gave different bytes", c.Name)
			}
		}
	}
}

func TestInputIsNotChanged(t *testing.T) {
	for _, c := range loadCases(t) {
		plan := loadPlan(t, c.Input)
		snapshot := loadPlan(t, c.Input)
		newEncoder().Encode(plan)
		if !reflect.DeepEqual(plan, snapshot) {
			t.Errorf("%s: the plan changed during Encode", c.Name)
		}
	}
}

func TestOutputIsWellFormed(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Output == "" {
			continue
		}
		got, _ := newEncoder().Encode(loadPlan(t, c.Input))
		text := string(got)
		if !strings.HasSuffix(text, "\n") || strings.HasSuffix(text, "\n\n") {
			t.Errorf("%s: must end in exactly one newline", c.Name)
		}
		if strings.Contains(text, "---") && strings.Contains("\n"+text, "\n---") {
			t.Errorf("%s: must not contain a document marker", c.Name)
		}
		for i, line := range strings.Split(text, "\n") {
			if line != strings.TrimRight(line, " \t") {
				t.Errorf("%s: line %d has trailing whitespace: %q", c.Name, i+1, line)
			}
			if strings.Contains(line, "\t") {
				t.Errorf("%s: line %d has a tab", c.Name, i+1)
			}
		}
	}
}

func TestWritesTheSupportedNumberKinds(t *testing.T) {
	cases := map[string]struct {
		value any
		want  string
	}{
		"int":           {int(5), "schemaVersion: 1\nx: 5\n"},
		"int64":         {int64(-7), "schemaVersion: 1\nx: -7\n"},
		"float64":       {1.5, "schemaVersion: 1\nx: 1.5\n"},
		"whole decimal": {float64(2), "schemaVersion: 1\nx: 2.0\n"},
		"ten":           {int(10), "schemaVersion: 1\nx: 10\n"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			plan := model.Plan{SchemaVersion: 1, UnknownFields: map[string]any{"x": c.value}}
			got, problems := newEncoder().Encode(plan)
			if len(problems) != 0 || string(got) != c.want {
				t.Errorf("got %q and %+v, want %q", got, problems, c.want)
			}
		})
	}
}

func TestReportsEveryUnsupportedNumberAndKind(t *testing.T) {
	cases := map[string]struct {
		value any
		path  string
	}{
		"int8":    {int8(1), "x"},
		"uint":    {uint(1), "x"},
		"float32": {float32(1.5), "x"},
		"NaN":     {math.NaN(), "x"},
		"+Inf":    {math.Inf(1), "x"},
		"-Inf":    {math.Inf(-1), "x"},
		"strings": {[]string{"a"}, "x"},
		"struct":  {struct{}{}, "x"},
		"in list": {[]any{"ok", int8(1)}, "x[1]"},
		"in map":  {map[string]any{"k": uint(1)}, "x.k"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			plan := model.Plan{SchemaVersion: 1, UnknownFields: map[string]any{"x": c.value}}
			got, problems := newEncoder().Encode(plan)
			if len(got) != 0 {
				t.Errorf("want no bytes, got %q", got)
			}
			if len(problems) != 1 || problems[0].Code != "encode.unsupported-value" || problems[0].Detail["path"] != c.path {
				t.Errorf("want one encode.unsupported-value at %s, got %+v", c.path, problems)
			}
		})
	}
}

func planWithClause(clause model.Clause) model.Plan {
	return model.Plan{
		SchemaVersion: 1,
		Sections: []model.Section{{
			ID: "s", Title: "S", Summary: "S",
			IncludeIf: &model.Condition{All: []model.Clause{clause}},
		}},
	}
}

func TestWritesASectionClauseWithoutQuestionOrAnswer(t *testing.T) {
	yes := true
	clause := model.Clause{
		Question: "q",
		Answer:   model.AnswerTest{Is: model.Value{Kind: model.ValueKindText, Text: "a"}, IsSettled: &yes},
		Section:  "other",
		Enabled:  false,
	}
	got, problems := newEncoder().Encode(planWithClause(clause))
	want := "schemaVersion: 1\nsections:\n  - id: s\n    title: S\n    summary: S\n    includeIf:\n      all:\n        - section: other\n          enabled: false\n"
	if len(problems) != 0 || string(got) != want {
		t.Errorf("got %q and %+v, want %q", got, problems, want)
	}
}

func TestWritesAQuestionClauseWithoutTheSwitch(t *testing.T) {
	clause := model.Clause{Question: "q", Answer: model.AnswerTest{Is: model.Value{Kind: model.ValueKindText, Text: "a"}}, Enabled: true}
	got, problems := newEncoder().Encode(planWithClause(clause))
	want := "schemaVersion: 1\nsections:\n  - id: s\n    title: S\n    summary: S\n    includeIf:\n      all:\n        - question: q\n          answer:\n            is: a\n"
	if len(problems) != 0 || string(got) != want {
		t.Errorf("got %q and %+v, want %q", got, problems, want)
	}
}

func TestReportsAClauseThatHoldsOnlyTheSwitchValue(t *testing.T) {
	got, problems := newEncoder().Encode(planWithClause(model.Clause{Enabled: true}))
	if len(got) != 0 {
		t.Errorf("want no bytes, got %q", got)
	}
	want := "sections[0].includeIf.all[0]"
	if len(problems) != 1 || problems[0].Code != "encode.empty-item" || problems[0].Detail["path"] != want {
		t.Errorf("want one encode.empty-item at %s, got %+v", want, problems)
	}
}

func TestWritesTheSwitchOnlyWhenTheSectionHoldsOne(t *testing.T) {
	yes, no := true, false
	cases := map[string]struct {
		enabled *bool
		want    string
	}{
		"absent": {nil, ""},
		"true":   {&yes, "    enabled: true\n"},
		"false":  {&no, "    enabled: false\n"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			plan := model.Plan{SchemaVersion: 1, Sections: []model.Section{{ID: "s", Title: "S", Summary: "S", Enabled: c.enabled}}}
			got, problems := newEncoder().Encode(plan)
			want := "schemaVersion: 1\nsections:\n  - id: s\n    title: S\n    summary: S\n" + c.want
			if len(problems) != 0 || string(got) != want {
				t.Errorf("got %q and %+v, want %q", got, problems, want)
			}
		})
	}
}
