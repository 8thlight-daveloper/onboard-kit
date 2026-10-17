package decoder_test

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
	Code   string            `json:"code"`
	Path   string            `json:"path"`
	Detail map[string]string `json:"detail"`
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

func describe(plan model.Plan) string {
	written, err := jsonform.JSON(plan)
	if err != nil {
		return err.Error()
	}
	return string(written)
}

func actualProblems(problems []model.Problem) []expectedProblem {
	var actual []expectedProblem
	for _, p := range problems {
		detail := map[string]string{}
		for key, value := range p.Detail {
			if key != "path" {
				detail[key] = value
			}
		}
		if len(detail) == 0 {
			detail = nil
		}
		actual = append(actual, expectedProblem{Code: p.Code, Path: p.Detail["path"], Detail: detail})
	}
	return actual
}

func TestEveryCaseCitesItsSource(t *testing.T) {
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
	}
}

func TestEveryCaseExpectsAnOutputOrAProblem(t *testing.T) {
	for _, c := range loadCases(t) {
		if (c.Output == "") == (len(c.Problems) == 0) {
			t.Errorf("case %s needs an output or at least one problem, and not both", c.Name)
		}
	}
}

func TestDecodesTheFile(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Output == "" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got, problems := newDecoder().Decode(readFixture(t, c.Input))
			if len(problems) != 0 {
				t.Fatalf("unexpected problems: %+v", problems)
			}
			want := loadPlan(t, c.Output)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("plans differ\n--- got ---\n%s\n--- want ---\n%s", describe(got), describe(want))
			}
		})
	}
}

func TestReportsProblemsAndReturnsNoPlan(t *testing.T) {
	for _, c := range loadCases(t) {
		if c.Output != "" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got, problems := newDecoder().Decode(readFixture(t, c.Input))
			if !reflect.DeepEqual(got, model.Plan{}) {
				t.Errorf("want the zero plan, got\n%s", describe(got))
			}
			for _, p := range problems {
				if p.Severity != model.SeverityError {
					t.Errorf("problem %s has severity %q, want error", p.Code, p.Severity)
				}
				if strings.TrimSpace(p.Message) == "" {
					t.Errorf("problem %s at %q has no message", p.Code, p.Detail["path"])
				}
			}
			if actual := actualProblems(problems); !reflect.DeepEqual(actual, c.Problems) {
				t.Errorf("problems differ\n got: %+v\nwant: %+v", actual, c.Problems)
			}
		})
	}
}

func TestResultIsDeterministic(t *testing.T) {
	for _, c := range loadCases(t) {
		data := readFixture(t, c.Input)
		firstPlan, firstProblems := newDecoder().Decode(data)
		for i := 0; i < 5; i++ {
			plan, problems := newDecoder().Decode(data)
			if !reflect.DeepEqual(firstPlan, plan) || !reflect.DeepEqual(firstProblems, problems) {
				t.Fatalf("%s: two calls gave different results", c.Name)
			}
		}
	}
}

func TestInputIsNotChanged(t *testing.T) {
	byteOrderMark := []byte("\xef\xbb\xbf")
	for _, c := range loadCases(t) {
		file := readFixture(t, c.Input)
		for name, data := range map[string][]byte{"as is": file, "with a byte order mark": append(bytes.Clone(byteOrderMark), file...)} {
			snapshot := bytes.Clone(data)
			newDecoder().Decode(data)
			if !bytes.Equal(data, snapshot) {
				t.Errorf("%s %s: the bytes changed during Decode", c.Name, name)
			}
		}
	}
}

func TestResultsShareNothingBetweenCalls(t *testing.T) {
	data := readFixture(t, "../../testdata/unknown-fields.yaml")
	first, _ := newDecoder().Decode(data)
	list, isList := first.UnknownFields["x-list"].([]any)
	nested, isMap := first.UnknownFields["x-map"].(map[string]any)
	if !isList || !isMap {
		t.Fatalf("unknown fields hold the wrong kinds\n%s", describe(first))
	}
	list[0] = "changed"
	nested["k"] = "changed"
	second, _ := newDecoder().Decode(data)
	if !reflect.DeepEqual(second, loadPlan(t, "../../testdata/unknown-fields.plan.json")) {
		t.Errorf("a change to one result showed in the next\n%s", describe(second))
	}
}

func TestReadsWindowsLineEndings(t *testing.T) {
	data := "schemaVersion: 1\r\nsections:\r\n  - id: s\r\n    title: T\r\n    steps:\r\n      - id: st\r\n        name: St\r\n"
	got, problems := newDecoder().Decode([]byte(data))
	want := loadPlan(t, "windows-line-endings.plan.json")
	if len(problems) != 0 || !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v and\n%s\nwant\n%s", problems, describe(got), describe(want))
	}
}

func TestReadsAByteOrderMark(t *testing.T) {
	data := append([]byte("\xef\xbb\xbf"), readFixture(t, "../../testdata/minimal.yaml")...)
	got, problems := newDecoder().Decode(data)
	if len(problems) != 0 || !reflect.DeepEqual(got, loadPlan(t, "../../testdata/minimal.plan.json")) {
		t.Errorf("got %+v and\n%s", problems, describe(got))
	}
}
