package validator_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/preparation/validator"
	"onboard-kit/internal/model"
)

type fixture struct {
	Name        string            `json:"name"`
	Source      string            `json:"source"`
	Rule        string            `json:"rule"`
	Context     contract.Context  `json:"context"`
	Preparation model.Preparation `json:"preparation"`
	Problems    []expectedProblem `json:"problems"`
}

type expectedProblem struct {
	Code string `json:"code"`
	Path string `json:"path"`
}

func fixtures(t *testing.T) []fixture {
	t.Helper()
	var cases []fixture
	for _, path := range []string{"testdata/cases.json", "testdata/mutation-boundaries.json"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var group []fixture
		if err := json.Unmarshal(data, &group); err != nil {
			t.Fatal(err)
		}
		cases = append(cases, group...)
	}
	if len(cases) == 0 {
		t.Fatal("no fixtures")
	}
	names := map[string]bool{}
	for _, item := range cases {
		if item.Name == "" || names[item.Name] || item.Source == "" || item.Rule == "" {
			t.Fatalf("missing provenance or duplicate fixture: %q", item.Name)
		}
		names[item.Name] = true
	}
	return cases
}

func TestValidation(t *testing.T) {
	t.Parallel()
	for _, item := range fixtures(t) {
		t.Run(item.Name, func(t *testing.T) {
			t.Parallel()
			contextBefore := clone(item.Context)
			preparationBefore := clone(item.Preparation)
			got := newValidator().Validate(item.Context, item.Preparation)
			if !reflect.DeepEqual(item.Context, contextBefore) || !reflect.DeepEqual(item.Preparation, preparationBefore) {
				t.Error("validation changed an input")
			}
			if len(got) != len(item.Problems) {
				t.Fatalf("problem count: got %d, want %d; codes: %v", len(got), len(item.Problems), codes(got))
			}
			for i, want := range item.Problems {
				if got[i].Code != want.Code || got[i].Detail["path"] != want.Path {
					t.Errorf("problem %d: got %s at %s, want %s at %s", i, got[i].Code, got[i].Detail["path"], want.Code, want.Path)
				}
				if got[i].Severity != model.SeverityError {
					t.Errorf("problem %d is not an error", i)
				}
				if strings.TrimSpace(got[i].Message) == "" {
					t.Errorf("problem %d has no useful message", i)
				}
			}
			again := newValidator().Validate(item.Context, item.Preparation)
			if !reflect.DeepEqual(got, again) {
				t.Error("validation is not deterministic")
			}
		})
	}
}

func TestValidationDoesNotEchoInputValues(t *testing.T) {
	t.Parallel()
	for _, item := range fixtures(t) {
		if item.Name != "unknown-contact" {
			continue
		}
		item.Preparation.Recipient.Buddy = "private-value-do-not-echo"
		for _, problem := range newValidator().Validate(item.Context, item.Preparation) {
			if strings.Contains(problem.Message, "private-value-do-not-echo") {
				t.Error("diagnostic echoes an invalid contact value")
			}
		}
		return
	}
	t.Fatal("missing unknown-contact fixture")
}

func clone[T any](value T) T {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		panic(err)
	}
	return result
}

func codes(problems []model.Problem) []string {
	var result []string
	for _, problem := range problems {
		result = append(result, problem.Code+"@"+problem.Detail["path"])
	}
	return result
}
