package finder_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type fixtureCase struct {
	Name     string `json:"name"`
	Plan     string `json:"plan"`
	Expected string `json:"expected"`
	Source   string `json:"source"`
}

type expectedFile struct {
	Questions []expectedQuestion `json:"questions"`
	Blockers  []string           `json:"blockers"`
}

type expectedQuestion struct {
	ID              string               `json:"id"`
	Type            string               `json:"type"`
	Phase           string               `json:"phase"`
	Required        bool                 `json:"required"`
	Answer          json.RawMessage      `json:"answer"`
	DeclaredDefault json.RawMessage      `json:"declaredDefault"`
	Group           string               `json:"group"`
	Suggestions     []expectedSuggestion `json:"suggestions"`
}

type expectedSuggestion struct {
	Value  json.RawMessage `json:"value"`
	Source string          `json:"source"`
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

func loadPlan(t *testing.T, c fixtureCase) model.Plan {
	t.Helper()
	plan, err := jsonform.Plan(readFixture(t, c.Plan))
	if err != nil {
		t.Fatalf("%s: %v", c.Plan, err)
	}
	return plan
}

func loadExpected(t *testing.T, c fixtureCase) expectedFile {
	t.Helper()
	var expected expectedFile
	decoder := json.NewDecoder(strings.NewReader(string(readFixture(t, c.Expected))))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&expected); err != nil {
		t.Fatalf("%s: %v", c.Expected, err)
	}
	return expected
}

func valueOf(t *testing.T, raw json.RawMessage) model.Value {
	t.Helper()
	if len(raw) == 0 {
		return model.Value{}
	}
	value, err := jsonform.Value(raw)
	if err != nil {
		t.Fatalf("value %s: %v", raw, err)
	}
	return value
}

func sameValue(left, right model.Value) bool {
	return left.Kind == right.Kind && left.Text == right.Text && left.Flag == right.Flag && slices.Equal(left.List, right.List)
}

func describe(value model.Value) string {
	switch value.Kind {
	case model.ValueKindNone:
		return "none"
	case model.ValueKindText:
		return "text " + value.Text
	case model.ValueKindFlag:
		if value.Flag {
			return "flag true"
		}
		return "flag false"
	case model.ValueKindList:
		return "list " + strings.Join(value.List, "|")
	case model.ValueKindDeferred:
		return "deferred"
	}
	return "unknown kind"
}

func describeSuggestions(suggestions []model.Suggestion) []string {
	described := []string{}
	for _, suggestion := range suggestions {
		described = append(described, string(suggestion.Source)+": "+describe(suggestion.Value))
	}
	return described
}

func idsOf(questions []model.Question) []string {
	ids := []string{}
	for _, question := range questions {
		ids = append(ids, question.ID)
	}
	return ids
}

func blockerIDs(blockers []model.Blocker) []string {
	ids := []string{}
	for _, blocker := range blockers {
		ids = append(ids, blocker.Question)
	}
	return ids
}

func findVariable(plan model.Plan, id string) (model.Variable, bool) {
	scope, name, _ := strings.Cut(id, ".")
	variables := plan.Variables
	if scope != "plan" {
		variables = nil
		for _, section := range plan.Sections {
			if section.ID == scope {
				variables = section.Variables
			}
		}
	}
	for _, variable := range variables {
		if variable.Name == name {
			return variable, true
		}
	}
	return model.Variable{}, false
}

func questionsByID(questions []model.Question) map[string]model.Question {
	byID := map[string]model.Question{}
	for _, question := range questions {
		byID[question.ID] = question
	}
	return byID
}

func isSettledOrDeferred(value model.Value) bool {
	switch value.Kind {
	case model.ValueKindText:
		return value.Text != ""
	case model.ValueKindFlag, model.ValueKindDeferred:
		return true
	case model.ValueKindList:
		return slices.ContainsFunc(value.List, func(item string) bool { return item != "" })
	}
	return false
}

func TestEveryCaseCitesItsSourceAndNamesItsQuestions(t *testing.T) {
	names := map[string]bool{}
	someCaseAsksQuestions, someCaseBlocks := false, false
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
		if names[c.Name] {
			t.Errorf("case %s is listed twice", c.Name)
		}
		names[c.Name] = true
		plan := loadPlan(t, c)
		expected := loadExpected(t, c)
		someCaseAsksQuestions = someCaseAsksQuestions || len(expected.Questions) > 0
		someCaseBlocks = someCaseBlocks || len(expected.Blockers) > 0
		asked := map[string]bool{}
		for _, question := range expected.Questions {
			asked[question.ID] = true
			if _, found := findVariable(plan, question.ID); !found {
				t.Errorf("case %s expects %s, which its plan does not declare", c.Name, question.ID)
			}
		}
		for _, id := range expected.Blockers {
			if !asked[id] {
				t.Errorf("case %s expects blocker %s, which is not among its questions", c.Name, id)
			}
		}
	}
	if !someCaseAsksQuestions || !someCaseBlocks {
		t.Errorf("no case expects a question, or no case expects a blocker")
	}
}

func TestAsksTheActiveQuestionsInOrder(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			plan := loadPlan(t, c)
			want := idsOf(nil)
			for _, question := range loadExpected(t, c).Questions {
				want = append(want, question.ID)
			}
			got := idsOf(newFinder().Questions(plan))
			if !slices.Equal(got, want) {
				t.Errorf("got  %v\nwant %v", got, want)
			}
		})
	}
}

func TestEachQuestionCarriesItsVariable(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			plan := loadPlan(t, c)
			expected := loadExpected(t, c)
			got := questionsByID(newFinder().Questions(plan))
			for _, want := range expected.Questions {
				question, found := got[want.ID]
				if !found {
					t.Errorf("%s: not returned", want.ID)
					continue
				}
				variable, _ := findVariable(plan, want.ID)
				scope, name, _ := strings.Cut(want.ID, ".")
				wantType := model.VariableType(want.Type)
				if want.Type == "" {
					wantType = model.DefaultVariableType
				}
				wantPhase := model.Phase(want.Phase)
				if want.Phase == "" {
					wantPhase = model.DefaultPhase
				}
				switch {
				case question.Scope != scope || question.Name != name:
					t.Errorf("%s: scope and name are %q and %q", want.ID, question.Scope, question.Name)
				case question.Description != variable.Description:
					t.Errorf("%s: description is %q, want %q", want.ID, question.Description, variable.Description)
				case question.Type != wantType:
					t.Errorf("%s: type is %q, want %q", want.ID, question.Type, wantType)
				case question.Phase != wantPhase:
					t.Errorf("%s: phase is %q, want %q", want.ID, question.Phase, wantPhase)
				case question.Required != want.Required || question.Required != variable.Required:
					t.Errorf("%s: required is %v, want %v", want.ID, question.Required, want.Required)
				case question.Multiple != variable.Multiple:
					t.Errorf("%s: multiple is %v, want %v", want.ID, question.Multiple, variable.Multiple)
				case !slices.Equal(question.Options, variable.Options):
					t.Errorf("%s: options are %v, want %v", want.ID, question.Options, variable.Options)
				}
			}
		})
	}
}

func TestAnswersAndDeclaredDefaultsAreReturnedAsStored(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got := questionsByID(newFinder().Questions(loadPlan(t, c)))
			for _, want := range loadExpected(t, c).Questions {
				question, found := got[want.ID]
				if !found {
					t.Errorf("%s: not returned", want.ID)
					continue
				}
				if wantAnswer := valueOf(t, want.Answer); !sameValue(question.Answer, wantAnswer) {
					t.Errorf("%s: answer is %s, want %s", want.ID, describe(question.Answer), describe(wantAnswer))
				}
				if wantDefault := valueOf(t, want.DeclaredDefault); !sameValue(question.DeclaredDefault, wantDefault) {
					t.Errorf("%s: declared default is %s, want %s", want.ID, describe(question.DeclaredDefault), describe(wantDefault))
				}
			}
		})
	}
}

func TestOffersTheSuggestionsInOrder(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got := questionsByID(newFinder().Questions(loadPlan(t, c)))
			for _, want := range loadExpected(t, c).Questions {
				question, found := got[want.ID]
				if !found {
					t.Errorf("%s: not returned", want.ID)
					continue
				}
				wantSuggestions := []model.Suggestion{}
				for _, suggestion := range want.Suggestions {
					wantSuggestions = append(wantSuggestions, model.Suggestion{
						Value:  valueOf(t, suggestion.Value),
						Source: model.SuggestionSource(suggestion.Source),
					})
				}
				same := len(question.Suggestions) == len(wantSuggestions)
				for i := 0; same && i < len(wantSuggestions); i++ {
					same = sameValue(question.Suggestions[i].Value, wantSuggestions[i].Value) && question.Suggestions[i].Source == wantSuggestions[i].Source
				}
				if !same {
					t.Errorf("%s:\ngot  %v\nwant %v", want.ID, describeSuggestions(question.Suggestions), describeSuggestions(wantSuggestions))
				}
			}
		})
	}
}

func TestGroupsQuestionsByNameTypeOptionsAndMultiple(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got := questionsByID(newFinder().Questions(loadPlan(t, c)))
			labelOfKey := map[string]string{}
			keyOfLabel := map[string]string{}
			for _, want := range loadExpected(t, c).Questions {
				question, found := got[want.ID]
				if !found {
					t.Errorf("%s: not returned", want.ID)
					continue
				}
				label := want.Group
				if label == "" {
					label = want.ID
				}
				if question.Group == "" {
					t.Errorf("%s: no group key", want.ID)
					continue
				}
				if other, seen := labelOfKey[question.Group]; seen && other != label {
					t.Errorf("%s: shares a group key with a question of group %q, want group %q", want.ID, other, label)
				}
				if other, seen := keyOfLabel[label]; seen && other != question.Group {
					t.Errorf("%s: has a different group key from the other questions of group %q", want.ID, label)
				}
				labelOfKey[question.Group] = label
				keyOfLabel[label] = question.Group
			}
		})
	}
}

func TestBlockersAreTheActiveRequiredQuestionsWithNoAnswerInOrder(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			want := loadExpected(t, c).Blockers
			got := blockerIDs(newFinder().Blockers(loadPlan(t, c)))
			if !slices.Equal(got, append([]string{}, want...)) {
				t.Errorf("got  %v\nwant %v", got, want)
			}
		})
	}
}

func TestBlockersAgreeWithTheQuestions(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		plan := loadPlan(t, c)
		finder := newFinder()
		want := []string{}
		for _, question := range finder.Questions(plan) {
			if question.Required && question.Phase == model.PhasePlanning && !isSettledOrDeferred(question.Answer) {
				want = append(want, question.ID)
			}
		}
		if got := blockerIDs(finder.Blockers(plan)); !slices.Equal(got, want) {
			t.Errorf("%s: blockers %v, but the required questions with no answer are %v", c.Name, got, want)
		}
	}
}

func TestPlanIsNotChanged(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		plan := loadPlan(t, c)
		snapshot := loadPlan(t, c)
		finder := newFinder()
		finder.Questions(plan)
		finder.Blockers(plan)
		if !reflect.DeepEqual(plan, snapshot) {
			t.Fatalf("%s: the plan was changed", c.Name)
		}
	}
}

func TestResultsDoNotDependOnEarlierCallsOrOnRepeats(t *testing.T) {
	t.Parallel()
	cases := loadCases(t)
	finder := newFinder()
	type result struct {
		questions []model.Question
		blockers  []model.Blocker
	}
	first := map[string]result{}
	for _, c := range cases {
		plan := loadPlan(t, c)
		first[c.Name] = result{newFinder().Questions(plan), newFinder().Blockers(plan)}
	}
	for pass := 0; pass < 2; pass++ {
		for i := range cases {
			index := i
			if pass == 1 {
				index = len(cases) - 1 - i
			}
			c := cases[index]
			for repeat := 0; repeat < 10; repeat++ {
				plan := loadPlan(t, c)
				got := result{finder.Questions(plan), finder.Blockers(plan)}
				if !reflect.DeepEqual(got, first[c.Name]) {
					t.Fatalf("%s on pass %d, repeat %d: the result changed", c.Name, pass, repeat)
				}
			}
		}
	}
}

func TestResultDoesNotDependOnTheOrderOfTheSectionsInThePlan(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		plan := loadPlan(t, c)
		wantQuestions := newFinder().Questions(plan)
		wantBlockers := newFinder().Blockers(plan)
		count := len(plan.Sections)
		variants := []model.Plan{}
		for turn := 1; turn < count; turn++ {
			rotated := loadPlan(t, c)
			rotated.Sections = append(slices.Clone(plan.Sections[turn:]), plan.Sections[:turn]...)
			variants = append(variants, rotated)
		}
		reversed := loadPlan(t, c)
		slices.Reverse(reversed.Sections)
		variants = append(variants, reversed)
		for _, variant := range variants {
			if got := newFinder().Questions(variant); !reflect.DeepEqual(got, wantQuestions) {
				t.Errorf("%s: sections reordered, questions %v, want %v", c.Name, idsOf(got), idsOf(wantQuestions))
			}
			if got := newFinder().Blockers(variant); !reflect.DeepEqual(got, wantBlockers) {
				t.Errorf("%s: sections reordered, blockers %v, want %v", c.Name, blockerIDs(got), blockerIDs(wantBlockers))
			}
		}
	}
}

func TestResultListsAreNotThePlansOrTheNextCalls(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		plan := loadPlan(t, c)
		snapshot := loadPlan(t, c)
		finder := newFinder()
		questions := finder.Questions(plan)
		blockers := finder.Blockers(plan)
		wantIDs := idsOf(questions)
		wantBlockers := blockerIDs(blockers)
		slices.Reverse(questions)
		slices.Reverse(blockers)
		if len(questions) > 0 {
			questions[0] = model.Question{ID: "changed"}
		}
		if len(blockers) > 0 {
			blockers[0] = model.Blocker{Question: "changed"}
		}
		if !reflect.DeepEqual(plan, snapshot) {
			t.Fatalf("%s: changing the results changed the plan", c.Name)
		}
		if !slices.Equal(idsOf(finder.Questions(plan)), wantIDs) || !slices.Equal(blockerIDs(finder.Blockers(plan)), wantBlockers) {
			t.Fatalf("%s: changing the results changed the next call", c.Name)
		}
	}
}

func TestPlanWithNothingHasNoQuestionsAndNoBlockers(t *testing.T) {
	t.Parallel()
	finder := newFinder()
	if got := finder.Questions(model.Plan{}); len(got) != 0 {
		t.Errorf("got questions %v, want none", idsOf(got))
	}
	if got := finder.Blockers(model.Plan{}); len(got) != 0 {
		t.Errorf("got blockers %v, want none", blockerIDs(got))
	}
}
