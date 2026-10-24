package reviewer_test

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/invitation/reviewer"
	"onboard-kit/internal/model"
)

func remainingHolderPlan(total, split int, override bool) []byte {
	var source strings.Builder
	source.WriteString("schemaVersion: 1\nid: plan-a\ndirectory:\n  contacts:\n    - id: c\n      name: C\n      email: c@x\n  roles:\n")
	groups := (total + split - 1) / split
	for group := 0; group < groups; group++ {
		fmt.Fprintf(&source, "    - id: r%d\n      name: R\n", group)
		if !override {
			source.WriteString("      holders:\n")
			count := min(split, total-group*split)
			for i := 0; i < count; i++ {
				source.WriteString("        - c\n")
			}
		}
	}
	if override {
		source.WriteString("  teams:\n    - id: team\n      name: T\n      holders:\n")
		for group := 0; group < groups; group++ {
			fmt.Fprintf(&source, "        r%d:\n", group)
			count := min(split, total-group*split)
			for i := 0; i < count; i++ {
				source.WriteString("          - c\n")
			}
		}
	}
	return []byte(source.String())
}

func remainingVariableListPlan(total, split int, options bool) []byte {
	var source strings.Builder
	source.WriteString("schemaVersion: 1\nid: plan-a\nvariables:\n")
	groups := (total + split - 1) / split
	for group := 0; group < groups; group++ {
		fmt.Fprintf(&source, "  - name: item%03d\n", group)
		if options {
			source.WriteString("    type: choice\n    options:\n")
		} else {
			source.WriteString("    type: command\n    default:\n")
		}
		count := min(split, total-group*split)
		for i := 0; i < count; i++ {
			source.WriteString("      - x\n")
		}
	}
	return []byte(source.String())
}

func TestAdmissionRemainingHolderCaps(t *testing.T) {
	for _, override := range []bool{false, true} {
		name := "roles"
		if override {
			name = "team-overrides"
		}
		t.Run(name, func(t *testing.T) {
			for _, totalLimit := range []bool{false, true} {
				limit := 1024
				if totalLimit {
					limit = 8192
				}
				for _, count := range []int{limit, limit + 1} {
					split := count
					if totalLimit {
						split = 1024
					}
					input := gapAdmissionArchive(t, remainingHolderPlan(count, split, override))
					got, problems := newReviewer().Inspect(input)
					if count == limit {
						if len(problems) != 0 || got.PlanID != "plan-a" {
							t.Fatalf("inclusive repeated-holder boundary must remain accepted: diagnostics=%d", len(problems))
						}
					} else {
						resourceProblem(t, problems, "inspect.plan-admission", "plan")
						if !reflect.DeepEqual(got, contract.Incoming{}) {
							t.Error("holder cap refusal returned partial content")
						}
					}
				}
			}
		})
	}
}

func TestAdmissionRemainingValueAndOptionsCaps(t *testing.T) {
	cases := []struct {
		name         string
		limit, split int
		options      bool
	}{
		{"value-list", 4096, 0, false},
		{"value-list-total", 32768, 4096, false},
		{"options", 1024, 0, true},
		{"options-total", 16384, 1024, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, count := range []int{c.limit, c.limit + 1} {
				split := c.split
				if split == 0 {
					split = count
				}
				input := gapAdmissionArchive(t, remainingVariableListPlan(count, split, c.options))
				got, problems := newReviewer().Inspect(input)
				if count == c.limit {
					if len(problems) != 0 || got.PlanID != "plan-a" {
						t.Fatalf("inclusive ordinary repeated-list boundary must remain accepted: diagnostics=%d", len(problems))
					}
				} else {
					resourceProblem(t, problems, "inspect.plan-admission", "plan")
					if !reflect.DeepEqual(got, contract.Incoming{}) {
						t.Error("list cap refusal returned partial content")
					}
				}
			}
		})
	}
}

func remainingClausePlan(total, split int) []byte {
	var source strings.Builder
	source.WriteString("schemaVersion: 1\nid: plan-a\nvariables:\n")
	groups := (total + split - 1) / split
	for group := 0; group < groups; group++ {
		fmt.Fprintf(&source, "  - name: gate%03d\n    askIf:\n      all:\n", group)
		count := min(split, total-group*split)
		for i := 0; i < count; i++ {
			source.WriteString("        - section: absent\n          enabled: true\n")
		}
	}
	return []byte(source.String())
}

func remainingPredicatePlan(total, split int, negative bool, alternate bool) []byte {
	var source strings.Builder
	source.WriteString("schemaVersion: 1\nid: plan-a\nvariables:\n  - name: gate000\n    askIf:\n      all:\n")
	groups := (total + split - 1) / split
	for group := 0; group < groups; group++ {
		field := "isOneOf"
		if negative || alternate && group%2 == 1 {
			field = "isNotOneOf"
		}
		fmt.Fprintf(&source, "        - question: plan.gate000\n          answer:\n            %s:\n", field)
		count := min(split, total-group*split)
		for i := 0; i < count; i++ {
			source.WriteString("              - x\n")
		}
	}
	return []byte(source.String())
}

func remainingAskIfRefusal(t *testing.T, got []contract.Diagnostic, count int) {
	t.Helper()
	indices := []int{0}
	if count == 16 {
		indices = []int{0, 10, 11, 12, 13, 14, 15, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	}
	if count != len(indices) {
		t.Fatal("unexpected independent authoring oracle size")
	}
	want := make([]contract.Diagnostic, 0, len(indices))
	for _, index := range indices {
		want = append(want, contract.Diagnostic{
			Stage:   "inspect.plan-authoring",
			Display: "Review the invitation plan declarations with its owner.",
			Problem: model.Problem{
				Code: "authoring.plan-ask-if", Severity: model.SeverityError,
				Message:  "Remove askIf from the plan variable.",
				Question: fmt.Sprintf("plan.gate%03d", index),
				Detail:   map[string]string{"path": fmt.Sprintf("variables[%d].askIf", index)},
			},
		})
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("at-cap input must reach its exact later authoring refusal, diagnostics=%d", len(got))
	}
}

func TestAdmissionRemainingConditionAndPredicateCaps(t *testing.T) {
	cases := []struct {
		name                    string
		limit, positiveProblems int
		build                   func(int) []byte
	}{
		{"condition-list", 256, 1, func(n int) []byte { return remainingClausePlan(n, n) }},
		{"clauses-total", 4096, 16, func(n int) []byte { return remainingClausePlan(n, 256) }},
		{"one-of-list", 256, 1, func(n int) []byte { return remainingPredicatePlan(n, n, false, false) }},
		{"not-one-of-list", 256, 1, func(n int) []byte { return remainingPredicatePlan(n, n, true, false) }},
		{"predicate-elements-total", 16384, 1, func(n int) []byte { return remainingPredicatePlan(n, 256, false, true) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, count := range []int{c.limit, c.limit + 1} {
				input := gapAdmissionArchive(t, c.build(count))
				got, problems := newReviewer().Inspect(input)
				if count == c.limit {
					remainingAskIfRefusal(t, problems, c.positiveProblems)
				} else {
					resourceProblem(t, problems, "inspect.plan-admission", "plan")
				}
				if !reflect.DeepEqual(got, contract.Incoming{}) {
					t.Error("guard or subsequent authoring refusal returned partial incoming")
				}
			}
		})
	}
}

func TestAdmissionRemainingOccurrenceCapAtCheapStage(t *testing.T) {
	r := newReviewer()
	in := incoming(t, r)
	for _, helperCount := range []int{1, 2} {
		var source strings.Builder
		source.WriteString("schemaVersion: 1\nkind: preparation\naccount: other\nplanId: plan-a\narrangements:\n  - contacts: [x")
		if helperCount == 2 {
			source.WriteString(", x")
		}
		source.WriteString("]\n")
		for i := 1; i < 13098; i++ {
			source.WriteString("  - {}\n")
		}
		local := contract.Local{Preparation: contract.Body{Present: true, Bytes: []byte(source.String())}}
		before := copied(local)
		got, problems := r.Review(in, local, contract.PlanChoiceUseIncoming)
		if helperCount == 1 {
			if len(problems) != 1 || problems[0].Stage != "review.preparation-validation" || problems[0].Display != "Ask your onboarding owner to repair unavailable local support." || problems[0].Problem.Code != "preparation.account-mismatch" || problems[0].Problem.Severity != model.SeverityError || problems[0].Problem.Detail["path"] != "account" || problems[0].Problem.Message == "" {
				t.Errorf("exact occurrence cap must pass the cheap guard and reach only later identity refusal, diagnostics=%d", len(problems))
			}
		} else {
			resourceProblem(t, problems, "review.preparation-admission", "local.preparation")
		}
		if !reflect.DeepEqual(got, contract.ReviewResult{}) || !reflect.DeepEqual(local, before) {
			t.Error("occurrence boundary refusal must preserve source and return no proposal")
		}
	}
}

func TestAdmissionRemainingKnownEntryCap(t *testing.T) {
	plan := append(append([]byte(nil), fixture(t, "minimal.plan.yaml")...), []byte("directory:\n  contacts:\n    - id: c\n      name: C\n      email: c@x\n")...)
	r := newReviewer()
	in := inspectPlan(t, r, plan)
	for _, count := range []int{65532, 65533} {
		var source strings.Builder
		source.WriteString("schemaVersion: 1\nkind: preparation\naccount: jules\nplanId: plan-a\narrangements:\n  - id: support-a\n    section: setup\n    kind: support\n    status: pending\n    contacts:\n")
		for i := 0; i < count; i++ {
			source.WriteString("      - c\n")
		}
		body := []byte(source.String())
		if len(body) > 700000 {
			t.Fatal("independent known-entry source unexpectedly grew")
		}
		local := contract.Local{Preparation: contract.Body{Present: true, Bytes: body}}
		before := copied(local)
		got, problems := r.Review(in, local, contract.PlanChoiceUseIncoming)
		if count == 65532 {
			if len(problems) != 0 || got.PlanChoiceNeeded || got.Proposal == nil || !got.Proposal.Review.Support.Available || len(got.Proposal.Review.Support.Arrangements) != 1 || len(got.Proposal.Review.Artifacts) != 3 {
				t.Fatal("inclusive known-entry cap must remain a complete valid support/session review")
			}
			if len(got.Proposal.Review.Support.Arrangements[0].Contacts) != 65532 || !bytes.Equal(got.Proposal.Review.Artifacts[1].Bytes, body) {
				t.Error("valid repeated helpers must remain complete and byte-exact")
			}
		} else {
			resourceProblem(t, problems, "review.preparation-admission", "local.preparation")
			if !reflect.DeepEqual(got, contract.ReviewResult{}) {
				t.Error("over-cap known entries returned a partial review")
			}
		}
		if !reflect.DeepEqual(local, before) {
			t.Error("known-entry admission mutated held support source")
		}
	}
}
