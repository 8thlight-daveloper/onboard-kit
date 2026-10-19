package selector_test

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
	Name   string    `json:"name"`
	Plan   string    `json:"plan"`
	Active *[]string `json:"active"`
	Source string    `json:"source"`
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

func idsOf(sections []model.Section) []string {
	ids := []string{}
	for _, section := range sections {
		ids = append(ids, section.ID)
	}
	return ids
}

func findSection(plan model.Plan, id string) (model.Section, bool) {
	for _, section := range plan.Sections {
		if section.ID == id {
			return section, true
		}
	}
	return model.Section{}, false
}

func TestEveryCaseCitesItsSourceAndNamesItsResult(t *testing.T) {
	names := map[string]bool{}
	someCaseSelectsSections := false
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
		if c.Active == nil {
			t.Errorf("case %s has no active list", c.Name)
			continue
		}
		if len(*c.Active) > 0 {
			someCaseSelectsSections = true
		}
		if names[c.Name] {
			t.Errorf("case %s is listed twice", c.Name)
		}
		names[c.Name] = true
		plan := loadPlan(t, c)
		for _, id := range *c.Active {
			if _, found := findSection(plan, id); !found {
				t.Errorf("case %s expects %s, which its plan does not hold", c.Name, id)
			}
		}
	}
	if !someCaseSelectsSections {
		t.Errorf("no case expects an active section")
	}
}

func TestSelectsTheActiveSectionsInOrder(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			plan := loadPlan(t, c)
			got := newSelector().Select(plan)
			if !slices.Equal(idsOf(got), *c.Active) {
				t.Errorf("got %v, want %v", idsOf(got), *c.Active)
			}
		})
	}
}

func TestReturnsEachSectionAsHeldInThePlan(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			plan := loadPlan(t, c)
			for _, got := range newSelector().Select(plan) {
				want, found := findSection(plan, got.ID)
				if !found {
					t.Fatalf("returned %s, which the plan does not hold", got.ID)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("section %s differs from the plan's\ngot:  %#v\nwant: %#v", got.ID, got, want)
				}
			}
		})
	}
}

func TestOrderDoesNotDependOnTheOrderOfTheSectionsInThePlan(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		plan := loadPlan(t, c)
		count := len(plan.Sections)
		for turn := 0; turn < count; turn++ {
			rotated := loadPlan(t, c)
			rotated.Sections = append(slices.Clone(plan.Sections[turn:]), plan.Sections[:turn]...)
			if got := idsOf(newSelector().Select(rotated)); !slices.Equal(got, *c.Active) {
				t.Errorf("%s, sections rotated by %d: got %v, want %v", c.Name, turn, got, *c.Active)
			}
		}
		reversed := loadPlan(t, c)
		slices.Reverse(reversed.Sections)
		if got := idsOf(newSelector().Select(reversed)); !slices.Equal(got, *c.Active) {
			t.Errorf("%s, sections reversed: got %v, want %v", c.Name, got, *c.Active)
		}
	}
}

func TestResultsDoNotDependOnEarlierCalls(t *testing.T) {
	t.Parallel()
	cases := loadCases(t)
	selector := newSelector()
	for pass := 0; pass < 2; pass++ {
		for i := range cases {
			index := i
			if pass == 1 {
				index = len(cases) - 1 - i
			}
			got := idsOf(selector.Select(loadPlan(t, cases[index])))
			if !slices.Equal(got, *cases[index].Active) {
				t.Fatalf("%s on pass %d: got %v, want %v", cases[index].Name, pass, got, *cases[index].Active)
			}
		}
	}
}

func TestPlanIsNotChanged(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		plan := loadPlan(t, c)
		snapshot := loadPlan(t, c)
		newSelector().Select(plan)
		if !reflect.DeepEqual(plan, snapshot) {
			t.Fatalf("%s: the plan was changed", c.Name)
		}
	}
}

func TestResultListIsNotThePlansList(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		if len(*c.Active) < 2 {
			continue
		}
		plan := loadPlan(t, c)
		snapshot := loadPlan(t, c)
		got := newSelector().Select(plan)
		slices.Reverse(got)
		got[0] = model.Section{ID: "changed"}
		if !reflect.DeepEqual(plan, snapshot) {
			t.Fatalf("%s: changing the result changed the plan", c.Name)
		}
	}
}

func TestPlanWithNoSectionsHasNoActiveSection(t *testing.T) {
	t.Parallel()
	if got := newSelector().Select(model.Plan{}); len(got) != 0 {
		t.Errorf("got %v, want no sections", idsOf(got))
	}
}
