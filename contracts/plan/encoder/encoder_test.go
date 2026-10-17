package encoder_test

import (
	"bytes"
	"encoding/json"
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
