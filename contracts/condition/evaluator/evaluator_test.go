package evaluator_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type fixtureCheck struct {
	Name      string          `json:"name"`
	Condition json.RawMessage `json:"condition"`
	Holds     bool            `json:"holds"`
}

type fixtureCase struct {
	Name    string         `json:"name"`
	Scope   string         `json:"scope"`
	Answers string         `json:"answers"`
	Checks  []fixtureCheck `json:"checks"`
	Source  string         `json:"source"`
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}

func loadCases(t *testing.T) []fixtureCase {
	t.Helper()
	var cases []fixtureCase
	if err := json.Unmarshal(readFixture(t, "cases.json"), &cases); err != nil {
		t.Fatalf("cases.json: %v", err)
	}
	return cases
}

func loadAnswers(t *testing.T, c fixtureCase) model.Answers {
	t.Helper()
	answers, err := jsonform.Answers(readFixture(t, c.Answers))
	if err != nil {
		t.Fatalf("%s: %v", c.Answers, err)
	}
	return answers
}

func loadCondition(t *testing.T, c fixtureCase, check fixtureCheck) model.Condition {
	t.Helper()
	condition, err := jsonform.Condition(check.Condition)
	if err != nil {
		t.Fatalf("%s, %s: %v", c.Name, check.Name, err)
	}
	return condition
}

func TestEveryCaseCitesItsSource(t *testing.T) {
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
	}
}

func TestDecidesWhetherTheConditionHolds(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			answers := loadAnswers(t, c)
			for _, check := range c.Checks {
				condition := loadCondition(t, c, check)
				got := newEvaluator().Holds(condition, c.Scope, answers)
				if got != check.Holds {
					t.Errorf("%s: got %v, want %v\ncondition: %s", check.Name, got, check.Holds, check.Condition)
				}
			}
		})
	}
}

func TestResultsDoNotDependOnEarlierCalls(t *testing.T) {
	t.Parallel()
	type call struct {
		condition model.Condition
		scope     string
		answers   model.Answers
		want      bool
	}
	var calls []call
	for _, c := range loadCases(t) {
		answers := loadAnswers(t, c)
		for _, check := range c.Checks {
			calls = append(calls, call{loadCondition(t, c, check), c.Scope, answers, check.Holds})
		}
	}
	evaluator := newEvaluator()
	for pass := 0; pass < 2; pass++ {
		for i := range calls {
			index := i
			if pass == 1 {
				index = len(calls) - 1 - i
			}
			got := evaluator.Holds(calls[index].condition, calls[index].scope, calls[index].answers)
			if got != calls[index].want {
				t.Fatalf("call %d on pass %d: got %v, want %v", index, pass, got, calls[index].want)
			}
		}
	}
}

func TestInputsAreNotChanged(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		answers := loadAnswers(t, c)
		snapshot := loadAnswers(t, c)
		for _, check := range c.Checks {
			condition := loadCondition(t, c, check)
			conditionSnapshot := loadCondition(t, c, check)
			newEvaluator().Holds(condition, c.Scope, answers)
			if !reflect.DeepEqual(answers, snapshot) {
				t.Fatalf("%s, %s: the answers were changed", c.Name, check.Name)
			}
			if !reflect.DeepEqual(condition, conditionSnapshot) {
				t.Fatalf("%s, %s: the condition was changed", c.Name, check.Name)
			}
		}
	}
}

func TestZeroAnswersMeanNothingIsSettled(t *testing.T) {
	t.Parallel()
	condition, err := jsonform.Condition([]byte(`{"all":[{"question":"plan.discipline","answer":{"isSettled":false}},{"question":"billing.plan_name","answer":{"isNot":"x"}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if newEvaluator().Holds(condition, "access", model.Answers{}) {
		t.Errorf("want false: isNot on an unanswered question is false")
	}
	onlyUnanswered, err := jsonform.Condition([]byte(`{"all":[{"question":"plan.discipline","answer":{"isSettled":false}},{"question":"billing.plan_name","answer":{"isSettled":false}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !newEvaluator().Holds(onlyUnanswered, "access", model.Answers{}) {
		t.Errorf("want true: nothing is settled when the answers are the zero value")
	}
}
