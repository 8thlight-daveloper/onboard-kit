package selector_test

import (
	"encoding/json"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type requiresApplicabilityCase struct {
	Name, Source string
	Plan         json.RawMessage
	Active       *[]string
}

func requiresApplicabilityCases(t *testing.T) []requiresApplicabilityCase {
	t.Helper()
	var cases []requiresApplicabilityCase
	if err := json.Unmarshal(readFixture(t, "requires-applicability.json"), &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 36 {
		t.Fatal("expected thirty-six independent Requires applicability cases")
	}
	seen := map[string]bool{}
	for _, c := range cases {
		if c.Name == "" || c.Source == "" || c.Active == nil || seen[c.Name] {
			t.Fatal("Requires fixture name/source/active list missing or name repeated")
		}
		seen[c.Name] = true
	}
	return cases
}

func requiresApplicabilityPlan(t *testing.T, c requiresApplicabilityCase) model.Plan {
	t.Helper()
	plan, err := jsonform.Plan(c.Plan)
	if err != nil {
		t.Fatalf("%s: %v", c.Name, err)
	}
	return plan
}

func assertRequiresApplicability(t *testing.T, c requiresApplicabilityCase, plan model.Plan, got []model.Section) {
	t.Helper()
	if !slices.Equal(idsOf(got), *c.Active) {
		t.Fatalf("%s: selected reference IDs %v; want hand-specified %v", c.Name, idsOf(got), *c.Active)
	}
	for _, section := range got {
		want, found := findSection(plan, section.ID)
		if !found || !reflect.DeepEqual(section, want) {
			t.Fatalf("%s: selected section %q was not returned exactly as held", c.Name, section.ID)
		}
	}
}

func TestRequiresApplicabilityAndGroupedReferenceOrder(t *testing.T) {
	for _, c := range requiresApplicabilityCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			plan := requiresApplicabilityPlan(t, c)
			assertRequiresApplicability(t, c, plan, newSelector().Select(plan))
		})
	}
}

func TestRequiresApplicabilityPermutationAndInputPreservation(t *testing.T) {
	shared := newSelector()
	for _, c := range requiresApplicabilityCases(t) {
		for _, order := range []string{"held", "reverse", "rotate"} {
			plan := requiresApplicabilityPlan(t, c)
			if order == "reverse" {
				slices.Reverse(plan.Sections)
			}
			if order == "rotate" && len(plan.Sections) > 1 {
				plan.Sections = append(plan.Sections[1:], plan.Sections[0])
			}
			before, err := jsonform.JSON(plan)
			if err != nil {
				t.Fatal(err)
			}
			got := shared.Select(plan)
			assertRequiresApplicability(t, c, plan, got)
			if len(got) > 0 {
				got[0] = model.Section{ID: "changed result slot", Requires: []string{"replacement"}}
				slices.Reverse(got)
			}
			after, err := jsonform.JSON(plan)
			if err != nil || string(before) != string(after) {
				t.Fatalf("%s/%s: selection or result-slot mutation changed input", c.Name, order)
			}
		}
	}
}

func TestRequiresApplicabilityHasNoCallHistory(t *testing.T) {
	shared := newSelector()
	cases := requiresApplicabilityCases(t)
	for round := 0; round < 3; round++ {
		for i := range cases {
			index := i
			if round%2 == 1 {
				index = len(cases) - 1 - i
			}
			c := cases[index]
			plan := requiresApplicabilityPlan(t, c)
			before := requiresApplicabilityPlan(t, c)
			got := shared.Select(plan)
			assertRequiresApplicability(t, c, plan, got)
			if !reflect.DeepEqual(plan, before) {
				t.Fatalf("%s: selection changed held input lists or maps", c.Name)
			}
			if len(got) > 0 {
				got[0].Requires = []string{"replacement"}
				got[0].Bootstrap = !got[0].Bootstrap
			}
			assertRequiresApplicability(t, c, plan, shared.Select(plan))
		}
	}
}

func TestRequiresApplicabilityConcurrentSharedInstance(t *testing.T) {
	shared := newSelector()
	var wait sync.WaitGroup
	for round := 0; round < 2; round++ {
		for _, c := range requiresApplicabilityCases(t) {
			plan := requiresApplicabilityPlan(t, c)
			before := requiresApplicabilityPlan(t, c)
			wait.Add(1)
			go func() {
				defer wait.Done()
				got := shared.Select(plan)
				if !slices.Equal(idsOf(got), *c.Active) || !reflect.DeepEqual(plan, before) {
					t.Errorf("%s: concurrent shared selector gave IDs %v instead of %v or changed input", c.Name, idsOf(got), *c.Active)
				}
			}()
		}
	}
	wait.Wait()
}

func TestRequiresApplicabilityNeverReadsOutcomeShapedUnknownData(t *testing.T) {
	for _, c := range requiresApplicabilityCases(t) {
		for _, kind := range []string{"verified", "failed", "deferred", ""} {
			plan := requiresApplicabilityPlan(t, c)
			plan.UnknownFields = map[string]any{"outcomes": []any{map[string]any{"section": "b", "outcome": kind, "sectionHash": "held"}}}
			assertRequiresApplicability(t, c, plan, newSelector().Select(plan))
		}
	}
}

func TestRequiresApplicabilityPreservesNilAndEmptyHeldLists(t *testing.T) {
	for _, requires := range [][]string{nil, {}} {
		plan := model.Plan{Sections: []model.Section{{ID: "a", Requires: requires}}}
		got := newSelector().Select(plan)
		if !reflect.DeepEqual(got, []model.Section{{ID: "a", Requires: requires}}) {
			t.Fatal("selection normalized held nil/empty Requires representation")
		}
	}
}

func TestRequiresApplicabilityDefensiveCyclesFinish(t *testing.T) {
	plans := []model.Plan{}
	for _, c := range requiresApplicabilityCases(t) {
		plans = append(plans, requiresApplicabilityPlan(t, c))
	}
	done := make(chan struct{})
	go func() {
		shared := newSelector()
		for _, plan := range plans {
			shared.Select(plan)
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("unchecked Requires/Parent input did not terminate")
	}
}
