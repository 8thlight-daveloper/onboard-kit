package authoringvalidator_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"onboard-kit/internal/model"
)

type expectation struct {
	Code     string
	Path     string
	Question string
	Section  string
	Severity model.Severity
	Item     *string
}

type planCase struct {
	Name string
	Plan model.Plan
	Want []expectation
}

type defaultCase struct {
	Name      string
	Variable  model.Variable
	Directory model.Directory
	Want      []expectation
}

type predicateCase struct {
	Name      string
	Target    model.Variable
	Directory model.Directory
	Clauses   []model.Clause
	Want      []expectation
}

var repairMessages = map[string]string{
	"authoring.section-id":         "Use a nonreserved lowercase section slug.",
	"authoring.step-id":            "Use a lowercase step slug or omit the identifier.",
	"authoring.required-prose":     "Supply nonblank required section prose.",
	"authoring.variable-name":      "Use a nonreserved variable name with at least three characters.",
	"authoring.variable-type":      "Choose a supported variable type.",
	"authoring.variable-audience":  "Choose onboarder or onboardee audience.",
	"authoring.multiple":           "Allow multiple values only for contact or role variables.",
	"authoring.choice-options":     "Supply at least one nonempty choice option.",
	"authoring.pattern":            "Use a valid RE2 string pattern.",
	"authoring.default":            "Supply a default that fits its declaration.",
	"authoring.timing":             "Choose a supported timing goal.",
	"authoring.directory-id":       "Use a lowercase directory identifier.",
	"authoring.directory-name":     "Supply a nonblank directory display name.",
	"authoring.dollar":             "Escape the dollar sign or use a declared local variable.",
	"authoring.owner-type":         "Reference a contact or role variable in owner text.",
	"authoring.escalation":         "Use one local contact or role variable reference.",
	"authoring.environment-name":   "Use an uppercase environment variable name.",
	"authoring.check":              "Supply exactly one meaningful check operation.",
	"authoring.plan-ask-if":        "Remove askIf from the plan variable.",
	"authoring.condition-empty":    "Remove the empty condition or supply its clauses.",
	"authoring.clause":             "Supply exactly one question or section-switch clause.",
	"authoring.question-reference": "Reference an existing unambiguous declared question.",
	"authoring.section-reference":  "Reference an existing declared section switch.",
	"authoring.include-scope":      "Gate a section only on plan or ancestor questions.",
	"authoring.predicate":          "Supply a supported meaningful answer predicate.",
	"authoring.predicate-value":    "Use a settled scalar value accepted by the referenced question.",
	"authoring.predicate-list":     "Use at least two values without a boolean or all-options list test.",
	"authoring.contradiction":      "Remove mutually incompatible condition tests.",
	"authoring.condition-cycle":    "Remove the question eligibility dependency cycle.",
	"authoring.unused-onboardee":   "Use this onboardee question in section content or eligibility.",
}

func TestLiteralPlans(t *testing.T) {
	for _, item := range readCases[planCase](t, "testdata/cases.json") {
		t.Run(item.Name, func(t *testing.T) {
			assertProblems(t, newValidator().Validate(item.Plan), item.Want)
		})
	}
}

func TestLiteralDefaults(t *testing.T) {
	for _, item := range readCases[defaultCase](t, "testdata/defaults.json") {
		t.Run(item.Name, func(t *testing.T) {
			plan := model.Plan{Directory: item.Directory, Variables: []model.Variable{item.Variable}}
			assertProblems(t, newValidator().Validate(plan), item.Want)
		})
	}
}

func TestLiteralPredicates(t *testing.T) {
	for _, item := range readCases[predicateCase](t, "testdata/predicates.json") {
		t.Run(item.Name, func(t *testing.T) {
			plan := model.Plan{
				Directory: item.Directory,
				Variables: []model.Variable{item.Target},
				Sections: []model.Section{{
					ID: "access", Title: "Access", Summary: "Summary",
					Variables: []model.Variable{{Name: "trigger", AskIf: &model.Condition{All: item.Clauses}}},
				}},
			}
			assertProblems(t, newValidator().Validate(plan), item.Want)
		})
	}
}

func TestInputOwnershipAndRepeatedCalls(t *testing.T) {
	validator := newValidator()
	cases := readCases[planCase](t, "testdata/cases.json")
	for _, item := range cases {
		before := marshal(t, item.Plan)
		assertProblems(t, validator.Validate(item.Plan), item.Want)
		if after := marshal(t, item.Plan); !reflect.DeepEqual(after, before) {
			t.Fatalf("%s: input changed", item.Name)
		}
	}
	for index := len(cases) - 1; index >= 0; index-- {
		item := cases[index]
		assertProblems(t, validator.Validate(item.Plan), item.Want)
		first := validator.Validate(item.Plan)
		for problemIndex := range first {
			if first[problemIndex].Detail != nil {
				first[problemIndex].Detail["path"] = "changed"
				first[problemIndex].Detail["extra"] = "changed"
			}
			first[problemIndex].Message = "changed"
		}
		assertProblems(t, validator.Validate(item.Plan), item.Want)
	}
}

func TestSharedValidatorConcurrentCalls(t *testing.T) {
	validator := newValidator()
	cases := readCases[planCase](t, "testdata/cases.json")
	var workers sync.WaitGroup
	for repeat := range 3 {
		for _, item := range cases {
			workers.Add(1)
			go func(item planCase, repeat int) {
				defer workers.Done()
				assertProblems(t, validator.Validate(item.Plan), item.Want)
			}(item, repeat)
		}
	}
	workers.Wait()
}

func TestSecretsMessagesStableAndPrivate(t *testing.T) {
	plan := model.Plan{
		Variables: []model.Variable{{Name: "default_value", Default: model.Value{Kind: model.ValueKindText, Text: "-----BEGIN PRIVATE KEY-----"}}},
		Sections:  []model.Section{{ID: "access", Title: "Access", Summary: "Summary", Steps: []model.Step{{Env: map[string]string{"PRIVATE_KEY": "-----BEGIN PRIVATE KEY-----"}}}}},
	}
	want := []expectation{
		{Code: "secret.private-key", Path: "sections[0].steps[0].env[0]", Section: "access"},
		{Code: "secret.private-key", Path: "variables[0].default", Question: "plan.default_value"},
	}
	got := newValidator().Validate(plan)
	assertProblems(t, got, want)
	if len(got) != 2 {
		return
	}
	if got[0].Message != got[1].Message {
		t.Fatal("same secret code changed its dependency message by location")
	}
}

func TestRejectedIdentifierPrivacy(t *testing.T) {
	plan := model.Plan{
		Variables: []model.Variable{{Name: "private rejected name"}},
		Sections:  []model.Section{{ID: "private rejected scope", Title: "Title", Summary: "Summary", Variables: []model.Variable{{Name: "valid_name", Type: "wrong"}}, Steps: []model.Step{{Env: map[string]string{"private rejected env": "safe"}}}}},
	}
	want := []expectation{
		{Code: "authoring.section-id", Path: "sections[0].id"},
		{Code: "authoring.environment-name", Path: "sections[0].steps[0].env[0]"},
		{Code: "authoring.variable-type", Path: "sections[0].variables[0].type"},
		{Code: "authoring.variable-name", Path: "variables[0].name"},
	}
	got := newValidator().Validate(plan)
	assertProblems(t, got, want)
	if strings.Contains(string(marshal(t, got)), "private rejected") {
		t.Fatal("diagnostics echoed a rejected identifier")
	}
}

func readCases[T any](t *testing.T, path string) []T {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cases []T
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatalf("unexpected trailing fixture data: %v", err)
	}
	if len(cases) == 0 {
		t.Fatal("empty fixture collection")
	}
	return cases
}

func marshal(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertProblems(t *testing.T, got []model.Problem, want []expectation) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("problem count: got %d, want %d; codes/paths: %s", len(got), len(want), safeProjection(got))
		return
	}
	for index, expected := range want {
		problem := got[index]
		severity := expected.Severity
		if severity == "" {
			severity = model.SeverityError
		}
		detail := map[string]string{"path": expected.Path}
		if expected.Item != nil {
			detail["item"] = *expected.Item
		}
		if problem.Code != expected.Code || problem.Severity != severity || problem.Question != expected.Question || problem.Section != expected.Section || !reflect.DeepEqual(problem.Detail, detail) {
			t.Errorf("problem %d differs from literal expectation for %s at %s", index, expected.Code, expected.Path)
		}
		if strings.HasPrefix(expected.Code, "secret.") {
			if strings.TrimSpace(problem.Message) == "" || strings.Contains(problem.Message, "-----BEGIN") || strings.Contains(problem.Message, "PRIVATE_KEY") {
				t.Errorf("problem %d: secret message is empty or exposes fixture content", index)
			}
		} else if expectedMessage, found := repairMessages[expected.Code]; !found || problem.Message != expectedMessage {
			t.Errorf("problem %d: incorrect repair message", index)
		}
	}
}

func safeProjection(problems []model.Problem) string {
	var projection []string
	for _, problem := range problems {
		projection = append(projection, problem.Code+":"+problem.Detail["path"])
	}
	return strings.Join(projection, ", ")
}
