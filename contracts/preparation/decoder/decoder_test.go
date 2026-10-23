package decoder_test

import (
	"bytes"
	"encoding/json"
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

type fixtureCase struct {
	Name     string             `json:"name"`
	Source   string             `json:"source"`
	Input    string             `json:"input"`
	Output   *model.Preparation `json:"output"`
	Problems []expectedProblem  `json:"problems"`
	Private  string             `json:"private"`
}

func fixtureCases(t *testing.T) []fixtureCase {
	t.Helper()
	data, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []fixtureCase
	reader := json.NewDecoder(bytes.NewReader(data))
	reader.UseNumber()
	if err := reader.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("fixture manifest must not be empty")
	}
	for index := range cases {
		if cases[index].Output != nil {
			for key, value := range cases[index].Output.UnknownFields {
				cases[index].Output.UnknownFields[key] = fixtureValue(t, value)
			}
		}
	}
	return cases
}

func fixtureValue(t *testing.T, value any) any {
	t.Helper()
	switch typed := value.(type) {
	case json.Number:
		if strings.ContainsAny(string(typed), ".eE") {
			result, err := typed.Float64()
			if err != nil {
				t.Fatal(err)
			}
			return result
		}
		result, err := typed.Int64()
		if err != nil {
			t.Fatal(err)
		}
		return result
	case []any:
		for index, item := range typed {
			typed[index] = fixtureValue(t, item)
		}
	case map[string]any:
		for key, item := range typed {
			typed[key] = fixtureValue(t, item)
		}
	}
	return value
}

func TestFixtureSourcesAndExpectations(t *testing.T) {
	seen := map[string]bool{}
	for _, fixture := range fixtureCases(t) {
		if seen[fixture.Name] || strings.TrimSpace(fixture.Name) == "" {
			t.Errorf("duplicate or missing fixture name %q", fixture.Name)
		}
		seen[fixture.Name] = true
		if strings.TrimSpace(fixture.Source) == "" {
			t.Errorf("%s has no source", fixture.Name)
		}
		if (fixture.Output == nil) == (len(fixture.Problems) == 0) {
			t.Errorf("%s must specify output or problems, not both", fixture.Name)
		}
	}
}

func TestDecode(t *testing.T) {
	for _, fixture := range fixtureCases(t) {
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			got, problems := newDecoder().Decode([]byte(fixture.Input))
			if fixture.Output != nil {
				if len(problems) != 0 || !reflect.DeepEqual(got, *fixture.Output) {
					t.Fatalf("got %#v and %+v; want %#v without problems", got, problems, *fixture.Output)
				}
				return
			}
			if !reflect.DeepEqual(got, model.Preparation{}) {
				t.Errorf("problem result is not zero: %#v", got)
			}
			var actual []expectedProblem
			for _, problem := range problems {
				actual = append(actual, expectedProblem{Code: problem.Code, Path: problem.Detail["path"]})
				if problem.Severity != model.SeverityError || strings.TrimSpace(problem.Message) == "" {
					t.Errorf("problem needs error severity and useful message: %+v", problem)
				}
				if strings.HasPrefix(problem.Code, "kind.") && !strings.Contains(problem.Message, "kind: preparation") {
					t.Errorf("kind problem must explain required marker: %+v", problem)
				}
				if fixture.Private != "" {
					if strings.Contains(problem.Message, fixture.Private) {
						t.Error("message echoes supplied personal value")
					}
					for _, value := range problem.Detail {
						if strings.Contains(value, fixture.Private) {
							t.Error("detail echoes supplied personal value")
						}
					}
				}
			}
			if !reflect.DeepEqual(actual, fixture.Problems) {
				t.Errorf("got problems %+v; want %+v", actual, fixture.Problems)
			}
		})
	}
}

func TestInputUnchangedAndDeterministic(t *testing.T) {
	decoder := newDecoder()
	for _, fixture := range fixtureCases(t) {
		for _, prefix := range []string{"", "\xef\xbb\xbf"} {
			data := []byte(prefix + fixture.Input)
			snapshot := bytes.Clone(data)
			first, firstProblems := decoder.Decode(data)
			for range 3 {
				next, nextProblems := decoder.Decode(data)
				if !reflect.DeepEqual(first, next) || !reflect.DeepEqual(firstProblems, nextProblems) {
					t.Errorf("%s results differ between calls", fixture.Name)
				}
			}
			if !bytes.Equal(data, snapshot) {
				t.Errorf("%s input changed", fixture.Name)
			}
		}
	}
}

func TestStyleVariants(t *testing.T) {
	for _, fixture := range fixtureCases(t) {
		if fixture.Output == nil {
			continue
		}
		for name, input := range map[string]string{
			"windows": strings.ReplaceAll(fixture.Input, "\n", "\r\n"),
			"marker":  "---\n" + fixture.Input,
			"bom":     "\xef\xbb\xbf" + fixture.Input,
		} {
			got, problems := newDecoder().Decode([]byte(input))
			if len(problems) != 0 || !reflect.DeepEqual(got, *fixture.Output) {
				t.Errorf("%s %s differs: %#v %+v", fixture.Name, name, got, problems)
			}
		}
	}
}

func TestResultsShareNothing(t *testing.T) {
	for _, fixture := range fixtureCases(t) {
		if fixture.Name != "unknown-values" && fixture.Name != "all-fields" {
			continue
		}
		decoder := newDecoder()
		first, problems := decoder.Decode([]byte(fixture.Input))
		if len(problems) != 0 || !reflect.DeepEqual(first, *fixture.Output) {
			t.Fatalf("%s baseline differs", fixture.Name)
		}
		if len(first.Arrangements) > 0 {
			first.Arrangements[0].ID = "changed"
			first.Arrangements[0].Contacts[0] = "changed"
		}
		if first.UnknownFields != nil {
			first.UnknownFields["x-map"].(map[string]any)["k"] = "changed"
			first.UnknownFields["x-list"].([]any)[0] = "changed"
		}
		next, nextProblems := decoder.Decode([]byte(fixture.Input))
		if len(nextProblems) != 0 || !reflect.DeepEqual(next, *fixture.Output) {
			t.Errorf("%s mutation leaked into another result", fixture.Name)
		}
	}
}

func TestInvalidUTF8(t *testing.T) {
	got, problems := newDecoder().Decode([]byte{0xff, 0xfe})
	if !reflect.DeepEqual(got, model.Preparation{}) || len(problems) != 1 || problems[0].Code != "file.not-yaml" {
		t.Errorf("invalid UTF-8 must be one file problem: %#v %+v", got, problems)
	}
}

func TestCanonicalPreparationFiles(t *testing.T) {
	data, err := os.ReadFile("../encoder/testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name   string
		Input  model.Preparation
		Output string
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range cases {
		if fixture.Input.SchemaVersion != 1 || fixture.Input.Kind != "preparation" || fixture.Output == "" {
			continue
		}
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			got, problems := newDecoder().Decode([]byte(fixture.Output))
			if len(problems) != 0 || !reflect.DeepEqual(got, fixture.Input) {
				t.Errorf("canonical file got %#v %+v; want %#v", got, problems, fixture.Input)
			}
		})
	}
}
