package reviewer_test

import (
	"bytes"
	"encoding/json"
	"maps"
	"os"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	contract "onboard-kit/contracts/preparation/reviewer"
	"onboard-kit/internal/model"
)

type expectedProblem struct {
	Code     string            `json:"code"`
	Severity model.Severity    `json:"severity"`
	Message  string            `json:"message"`
	Question *string           `json:"question"`
	Section  *string           `json:"section"`
	Path     string            `json:"path"`
	Detail   map[string]string `json:"detail"`
}

type fixture struct {
	Name                string            `json:"name"`
	Source              string            `json:"source"`
	Context             contract.Context  `json:"context"`
	Plan                json.RawMessage   `json:"plan"`
	Preparation         json.RawMessage   `json:"preparation"`
	PreparationProblems []expectedProblem `json:"preparationProblems"`
	OnboarderBlockers   []string          `json:"onboarderBlockers"`
	Golden              string            `json:"golden"`
	FirstStepsProblems  []expectedProblem `json:"firstStepsProblems"`
}

func fixtures(t *testing.T) []fixture {
	t.Helper()
	data, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []fixture
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 15 {
		t.Fatal("expected fifteen independently sourced reviewer cases")
	}
	seen := make(map[string]bool)
	for _, c := range cases {
		if c.Name == "" || c.Source == "" || seen[c.Name] || c.PreparationProblems == nil || c.OnboarderBlockers == nil || c.FirstStepsProblems == nil {
			t.Fatal("missing provenance, explicit field oracle or duplicate fixture name")
		}
		if c.Golden == "" && len(c.PreparationProblems)+len(c.OnboarderBlockers)+len(c.FirstStepsProblems) == 0 {
			t.Fatal("meaningful fixture must distinguish zero stub")
		}
		seen[c.Name] = true
	}
	return cases
}

func inputValues(t *testing.T, c fixture) (contract.Context, model.Plan, model.Preparation) {
	t.Helper()
	var context contract.Context
	encoded, err := json.Marshal(c.Context)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, &context); err != nil {
		t.Fatal(err)
	}
	plan, err := jsonform.Plan(c.Plan)
	if err != nil {
		t.Fatal(err)
	}
	var preparation model.Preparation
	decoder := json.NewDecoder(bytes.NewReader(c.Preparation))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&preparation); err != nil {
		t.Fatal(err)
	}
	return context, plan, preparation
}

func assertProblems(t *testing.T, field string, got []model.Problem, want []expectedProblem) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s count got%d want%d: %#v", field, len(got), len(want), got)
	}
	for i, expected := range want {
		actual := got[i]
		if actual.Code != expected.Code || actual.Severity != expected.Severity || expected.Question != nil && actual.Question != *expected.Question || expected.Section != nil && actual.Section != *expected.Section {
			t.Fatalf("%s[%d] diagnostic identity/location differs: got%#v want%#v", field, i, actual, expected)
		}
		if strings.TrimSpace(actual.Message) == "" || expected.Message != "" && actual.Message != expected.Message {
			t.Fatalf("%s[%d] missing or incorrect dependency message: %#v", field, i, actual)
		}
		if strings.Contains(actual.Message, "PRIVATE") {
			t.Fatal("diagnostic echoed a private fixture value")
		}
		for _, value := range actual.Detail {
			if strings.Contains(value, "PRIVATE") {
				t.Fatal("diagnostic detail echoed a private fixture value")
			}
		}
		if expected.Path != "" {
			if actual.Detail["path"] != expected.Path {
				t.Fatalf("%s[%d] path got%q want%q", field, i, actual.Detail["path"], expected.Path)
			}
		} else if strings.HasPrefix(expected.Code, "gap.") {
			for key, value := range expected.Detail {
				if actual.Detail[key] != value {
					t.Fatalf("%s[%d] gap detail %s got%q want%q", field, i, key, actual.Detail[key], value)
				}
			}
		} else if !maps.Equal(actual.Detail, expected.Detail) {
			t.Fatalf("%s[%d] detail got%#v want%#v", field, i, actual.Detail, expected.Detail)
		}
	}
}

func assertReview(t *testing.T, c fixture, got contract.Review) {
	t.Helper()
	assertProblems(t, "preparationProblems", got.PreparationProblems, c.PreparationProblems)
	assertProblems(t, "firstStepsProblems", got.FirstStepsProblems, c.FirstStepsProblems)
	var blockers []string
	for _, blocker := range got.OnboarderBlockers {
		blockers = append(blockers, blocker.Question)
	}
	if !slices.Equal(blockers, c.OnboarderBlockers) {
		t.Fatalf("onboarder blockers got%q want%q", blockers, c.OnboarderBlockers)
	}
	want := ""
	if c.Golden != "" {
		data, err := os.ReadFile("testdata/" + c.Golden)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) == 0 {
			t.Fatal("successful guide oracle may not be empty")
		}
		want = string(data)
	}
	if got.FirstSteps != want {
		t.Fatalf("firstSteps literal bytes differ for %s\ngot%q\nwant%q", c.Name, got.FirstSteps, want)
	}
}

func normalized(value contract.Review) contract.Review {
	if len(value.PreparationProblems) == 0 {
		value.PreparationProblems = nil
	}
	if len(value.OnboarderBlockers) == 0 {
		value.OnboarderBlockers = nil
	}
	if len(value.FirstStepsProblems) == 0 {
		value.FirstStepsProblems = nil
	}
	return value
}

func TestFourIndependentResultFields(t *testing.T) {
	t.Parallel()
	for _, c := range fixtures(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			context, plan, preparation := inputValues(t, c)
			wantContext, wantPlan, wantPreparation := inputValues(t, c)
			got := newReviewer().Review(context, plan, preparation)
			assertReview(t, c, got)
			if !reflect.DeepEqual(context, wantContext) || !reflect.DeepEqual(plan, wantPlan) || !reflect.DeepEqual(preparation, wantPreparation) {
				t.Fatal("review modified input or nested backing storage")
			}
		})
	}
}

func TestSharedReviewerRepeatAndReverseCaseOrder(t *testing.T) {
	t.Parallel()
	cases := fixtures(t)
	shared := newReviewer()
	for round := range 3 {
		if round == 1 {
			slices.Reverse(cases)
		}
		for _, c := range cases {
			context, plan, preparation := inputValues(t, c)
			first := shared.Review(context, plan, preparation)
			again := shared.Review(context, plan, preparation)
			assertReview(t, c, first)
			assertReview(t, c, again)
			if !reflect.DeepEqual(normalized(first), normalized(again)) {
				t.Fatal("repeated reviewer call changed dependency result")
			}
		}
	}
}

func namedFixture(t *testing.T, name string) fixture {
	t.Helper()
	for _, c := range fixtures(t) {
		if c.Name == name {
			return c
		}
	}
	t.Fatal("missing named independent fixture")
	return fixture{}
}

func TestOwnedBlockerAndFirstStepsDiagnosticStorage(t *testing.T) {
	t.Parallel()
	c := namedFixture(t, "guide-help-errors-preserve-independent-blockers-and-order")
	context, plan, preparation := inputValues(t, c)
	shared := newReviewer()
	first := shared.Review(context, plan, preparation)
	held := shared.Review(context, plan, preparation)
	assertReview(t, c, first)
	assertReview(t, c, held)
	first.OnboarderBlockers[0].Question = "caller edit"
	first.FirstStepsProblems[0].Code = "caller.edit"
	first.FirstStepsProblems[0].Detail["item"] = "caller edit"
	first.FirstStepsProblems[0].Detail["added"] = "caller edit"
	assertReview(t, c, held)
	context.GeneralHelp[0].ID = "changed later input"
	plan.Variables[0].Name = "changed later input"
	preparation.UnknownFields = map[string]any{"caller": "edit"}
	assertReview(t, c, held)
	freshContext, freshPlan, freshPreparation := inputValues(t, c)
	assertReview(t, c, shared.Review(freshContext, freshPlan, freshPreparation))
}

func TestOwnedPreparationProblemStorage(t *testing.T) {
	t.Parallel()
	c := namedFixture(t, "invalid-arrangement-visible-without-repair-or-later-output")
	context, plan, preparation := inputValues(t, c)
	shared := newReviewer()
	first := shared.Review(context, plan, preparation)
	held := shared.Review(context, plan, preparation)
	assertReview(t, c, first)
	assertReview(t, c, held)
	first.PreparationProblems[0].Code = "caller.edit"
	first.PreparationProblems[0].Detail["path"] = "caller edit"
	first.PreparationProblems[0].Detail["added"] = "caller edit"
	assertReview(t, c, held)
	preparation.Arrangements[0].Note = "later input edit"
	assertReview(t, c, held)
	freshContext, freshPlan, freshPreparation := inputValues(t, c)
	assertReview(t, c, shared.Review(freshContext, freshPlan, freshPreparation))
}

func TestConcurrentValidAndRefusedReview(t *testing.T) {
	t.Parallel()
	shared := newReviewer()
	cases := fixtures(t)
	var group sync.WaitGroup
	for range 3 {
		for _, c := range cases {
			group.Add(1)
			go func() {
				defer group.Done()
				context, plan, preparation := inputValues(t, c)
				wantContext, wantPlan, wantPreparation := inputValues(t, c)
				got := shared.Review(context, plan, preparation)
				assertReview(t, c, got)
				if !reflect.DeepEqual(context, wantContext) || !reflect.DeepEqual(plan, wantPlan) || !reflect.DeepEqual(preparation, wantPreparation) {
					t.Error("concurrent review mutated its inputs")
				}
			}()
		}
	}
	group.Wait()
}
