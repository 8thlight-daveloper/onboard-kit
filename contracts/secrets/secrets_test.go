package secrets_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type expectedProblem struct {
	Code string `json:"code"`
	Item *int   `json:"item"`
}

type fixtureCheck struct {
	Name     string            `json:"name"`
	Value    json.RawMessage   `json:"value"`
	Problems []expectedProblem `json:"problems"`
	Secret   string            `json:"secret"`
}

type fixtureCase struct {
	Name   string         `json:"name"`
	Checks []fixtureCheck `json:"checks"`
	Source string         `json:"source"`
}

const fragmentLength = 8

func loadCases(t *testing.T) []fixtureCase {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "cases.json"))
	if err != nil {
		t.Fatalf("read cases.json: %v", err)
	}
	data = bytes.ReplaceAll(data, []byte("\u00a4"), nil)
	var cases []fixtureCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("cases.json: %v", err)
	}
	return cases
}

func loadValue(t *testing.T, c fixtureCase, check fixtureCheck) model.Value {
	t.Helper()
	value, err := jsonform.Value(check.Value)
	if err != nil {
		t.Fatalf("%s, %s: %v", c.Name, check.Name, err)
	}
	return value
}

func TestEveryCaseCitesItsSource(t *testing.T) {
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
	}
}

func TestReportsTheExpectedProblems(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			for _, check := range c.Checks {
				value := loadValue(t, c, check)
				got := newChecker().Check(value)
				if len(got) != len(check.Problems) {
					t.Errorf("%s: got %d problems, want %d\nvalue: %s\nproblems: %+v", check.Name, len(got), len(check.Problems), check.Value, got)
					continue
				}
				for i, want := range check.Problems {
					assertProblem(t, check, got[i], want)
				}
			}
		})
	}
}

func assertProblem(t *testing.T, check fixtureCheck, got model.Problem, want expectedProblem) {
	t.Helper()
	if got.Code != want.Code {
		t.Errorf("%s: code %q, want %q", check.Name, got.Code, want.Code)
	}
	if got.Severity != model.SeverityError {
		t.Errorf("%s: severity %q, want error", check.Name, got.Severity)
	}
	if strings.TrimSpace(got.Message) == "" {
		t.Errorf("%s: the message is empty", check.Name)
	}
	if got.Question != "" || got.Section != "" {
		t.Errorf("%s: question %q and section %q must be empty", check.Name, got.Question, got.Section)
	}
	wantDetail := map[string]string{}
	if want.Item != nil {
		wantDetail["item"] = strconv.Itoa(*want.Item)
	}
	gotDetail := got.Detail
	if gotDetail == nil {
		gotDetail = map[string]string{}
	}
	if !reflect.DeepEqual(gotDetail, wantDetail) {
		t.Errorf("%s: detail %v, want %v", check.Name, got.Detail, wantDetail)
	}
	if check.Secret == "" {
		return
	}
	shown := got.Code + "\n" + got.Message
	for key, value := range got.Detail {
		shown += "\n" + key + "\n" + value
	}
	runes := []rune(check.Secret)
	for start := 0; start+fragmentLength <= len(runes); start++ {
		if fragment := string(runes[start : start+fragmentLength]); strings.Contains(shown, fragment) {
			t.Errorf("%s: the problem repeats part of the secret: %q", check.Name, fragment)
			return
		}
	}
}

func TestResultsDoNotDependOnEarlierCalls(t *testing.T) {
	t.Parallel()
	type call struct {
		value model.Value
		want  []expectedProblem
	}
	var calls []call
	for _, c := range loadCases(t) {
		for _, check := range c.Checks {
			calls = append(calls, call{loadValue(t, c, check), check.Problems})
		}
	}
	checker := newChecker()
	for pass := 0; pass < 2; pass++ {
		for i := range calls {
			index := i
			if pass == 1 {
				index = len(calls) - 1 - i
			}
			got := checker.Check(calls[index].value)
			if len(got) != len(calls[index].want) {
				t.Fatalf("call %d on pass %d: got %d problems, want %d", index, pass, len(got), len(calls[index].want))
			}
			for j, want := range calls[index].want {
				if got[j].Code != want.Code {
					t.Fatalf("call %d on pass %d: code %q, want %q", index, pass, got[j].Code, want.Code)
				}
			}
		}
	}
}

func TestSameValueGivesTheSameProblems(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		for _, check := range c.Checks {
			value := loadValue(t, c, check)
			first := newChecker().Check(value)
			second := newChecker().Check(value)
			if !reflect.DeepEqual(first, second) {
				t.Errorf("%s, %s: two calls differ\n%+v\n%+v", c.Name, check.Name, first, second)
			}
		}
	}
}

func TestTheValueIsNotChanged(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		for _, check := range c.Checks {
			value := loadValue(t, c, check)
			snapshot := loadValue(t, c, check)
			newChecker().Check(value)
			if !reflect.DeepEqual(value, snapshot) {
				t.Fatalf("%s, %s: the value was changed", c.Name, check.Name)
			}
		}
	}
}

func TestAValueOfNoKindHasNoProblems(t *testing.T) {
	t.Parallel()
	if got := newChecker().Check(model.Value{}); len(got) != 0 {
		t.Errorf("got %+v, want no problems", got)
	}
}

func TestOnlyTheKindOfTheValueIsRead(t *testing.T) {
	t.Parallel()
	secret := "ghp_" + "FAKEFAKEFAKEFAKEFAKEFAKEFAKEFAKEFAKE"
	for _, value := range []model.Value{
		{Kind: model.ValueKindFlag, Flag: true, Text: secret, List: []string{secret}},
		{Kind: model.ValueKindDeferred, Text: secret, List: []string{secret}},
		{Kind: model.ValueKindText, Text: "ask the team lead", List: []string{secret}},
		{Kind: model.ValueKindList, List: []string{"ask the team lead"}, Text: secret},
	} {
		if got := newChecker().Check(value); len(got) != 0 {
			t.Errorf("kind %v: got %+v, want no problems", value.Kind, got)
		}
	}
}

func TestOneCodeHasOneMessage(t *testing.T) {
	t.Parallel()
	messages := map[string]string{}
	for _, c := range loadCases(t) {
		for _, check := range c.Checks {
			for _, problem := range newChecker().Check(loadValue(t, c, check)) {
				if first, seen := messages[problem.Code]; seen && first != problem.Message {
					t.Fatalf("%s: code %s has two messages: %q and %q", c.Name, problem.Code, first, problem.Message)
				}
				messages[problem.Code] = problem.Message
			}
		}
	}
}
