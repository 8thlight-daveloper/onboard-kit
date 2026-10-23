package assembler_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type bootstrapCase struct {
	Name, Source, Document string
	Plan                   json.RawMessage
	Problems               []model.Problem
}

func bootstrapCases(t *testing.T, file string, count int) []bootstrapCase {
	t.Helper()
	var cases []bootstrapCase
	if err := json.Unmarshal(readFixture(t, file), &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != count {
		t.Fatalf("%s: expected %d independent projection cases", file, count)
	}
	seen := map[string]bool{}
	for _, c := range cases {
		if c.Name == "" || c.Source == "" || seen[c.Name] {
			t.Fatal("projection fixture name/source missing or name repeated")
		}
		seen[c.Name] = true
	}
	return cases
}

func bootstrapPlan(t *testing.T, c bootstrapCase) model.Plan {
	t.Helper()
	plan, err := jsonform.Plan(c.Plan)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func bootstrapDetail(t *testing.T, detail map[string]string) map[string]string {
	t.Helper()
	if len(detail) == 0 {
		return nil
	}
	out := map[string]string{}
	for key, value := range detail {
		out[key] = value
	}
	if raw, ok := out["members"]; ok {
		var compact bytes.Buffer
		if err := json.Compact(&compact, []byte(raw)); err != nil || compact.String() != raw || !strings.HasPrefix(raw, "[") {
			t.Fatal("structural members must be compact valid JSON array")
		}
		var items []json.RawMessage
		if err := json.Unmarshal([]byte(raw), &items); err != nil {
			t.Fatal(err)
		}
		members := make([]string, len(items))
		for i, item := range items {
			if len(item) == 0 || item[0] != '"' {
				t.Fatal("structural members must contain JSON strings")
			}
			if err := json.Unmarshal(item, &members[i]); err != nil {
				t.Fatal(err)
			}
		}
		encoded, err := json.Marshal(members)
		if err != nil {
			t.Fatal(err)
		}
		out["members"] = string(encoded)
	}
	return out
}

func assertBootstrap(t *testing.T, c bootstrapCase, document string, problems []model.Problem) {
	t.Helper()
	if document != c.Document || len(problems) != len(c.Problems) {
		t.Fatalf("%s: fragment:\n%s\nwant:\n%s\nproblems %#v; want %#v", c.Name, document, c.Document, problems, c.Problems)
	}
	for i, want := range c.Problems {
		got := problems[i]
		if got.Code != want.Code || got.Severity != want.Severity || got.Question != want.Question || got.Section != want.Section || strings.TrimSpace(got.Message) == "" {
			t.Fatalf("%s: diagnostic %d got %#v, want %#v with nonblank message", c.Name, i, got, want)
		}
		if strings.HasPrefix(want.Code, "plan.") {
			if got.Message != want.Message || !reflect.DeepEqual(bootstrapDetail(t, got.Detail), bootstrapDetail(t, want.Detail)) {
				t.Fatalf("%s: structural diagnostic semantics changed: %#v; want %#v", c.Name, got, want)
			}
		} else {
			for key, value := range want.Detail {
				if got.Detail[key] != value {
					t.Fatalf("%s: gap detail %s got %q, want %q", c.Name, key, got.Detail[key], value)
				}
			}
		}
	}
	if document != "" && (!strings.HasSuffix(document, "\n") || strings.HasSuffix(document, "\n\n")) {
		t.Fatal("fragment must end with exactly one newline")
	}
}

func TestBootstrapProjection(t *testing.T) {
	for _, c := range bootstrapCases(t, "bootstrap.json", 25) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			plan := bootstrapPlan(t, c)
			before := bootstrapPlan(t, c)
			assembler := newAssembler()
			got, problems := assembler.AssembleBootstrap(plan)
			assertBootstrap(t, c, got, problems)
			if !reflect.DeepEqual(plan, before) {
				t.Fatal("projection changed input")
			}
			again, repeated := assembler.AssembleBootstrap(plan)
			if got != again || !reflect.DeepEqual(problems, repeated) {
				t.Fatal("projection nondeterministic on same instance")
			}
			permuted := bootstrapPlan(t, c)
			slices.Reverse(permuted.Sections)
			reordered, reorderedProblems := assembler.AssembleBootstrap(permuted)
			assertBootstrap(t, c, reordered, reorderedProblems)
		})
	}
}

func TestBootstrapRejectsFullStructureBeforeProjectionOrResolution(t *testing.T) {
	for _, c := range bootstrapCases(t, "bootstrap-structure-errors.json", 31) {
		for _, inactive := range []bool{false, true} {
			mode := "held"
			if inactive {
				mode = "all-off"
			}
			t.Run(c.Name+"/"+mode, func(t *testing.T) {
				t.Parallel()
				plan := bootstrapPlan(t, c)
				for _, section := range plan.Sections {
					if section.ID == "projection-free-marker" {
						t.Fatal("sentinel would change error oracle")
					}
				}
				plan.Sections = append(plan.Sections, model.Section{ID: "projection-free-marker", Title: "Must not partially render", Bootstrap: true, Summary: "Use $account.", Variables: []model.Variable{{Name: "account", Audience: model.AudienceOnboardee}}})
				if inactive {
					for i := range plan.Sections {
						off := false
						plan.Sections[i].Enabled = &off
					}
				}
				got, problems := newAssembler().AssembleBootstrap(plan)
				assertBootstrap(t, c, got, problems)
			})
		}
	}
}

func TestBootstrapDoesNotChangeFullAssembly(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Sections: []model.Section{{ID: "access", Title: "Public", Bootstrap: true}, {ID: "private", Title: "Private"}}}
	before, problems := newAssembler().Assemble(plan)
	plan.Sections[0].Bootstrap = false
	plan.Sections[1].Bootstrap = true
	after, repeated := newAssembler().Assemble(plan)
	if before != after || !reflect.DeepEqual(problems, repeated) {
		t.Fatal("bootstrap marker changed ordinary full assembly")
	}
}

func TestOrdinaryAssemblyKeepsItsOriginalGraphTolerance(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Provenance: model.Provenance{GeneratedOn: "2026-10-23", TemplateSet: "test"}, Sections: []model.Section{{ID: "access", Title: "Access", After: []string{"missing"}, Bootstrap: true}}}
	want := "# Onboarding\n\nGenerated 2026-10-23 from template set " + "`test`" + ". A person must review this before it is sent.\n\n## Gaps to fill\n\n| Gap | Who fills it | By when |\n| --- | --- | --- |\n\n## Access\n"
	got, problems := newAssembler().Assemble(plan)
	if got != want || len(problems) != 0 {
		t.Fatal("ordinary Assemble acquired bootstrap structural validation or formatting")
	}
}

func TestBootstrapStateAndReturnedProblemIsolation(t *testing.T) {
	assembler := newAssembler()
	cases := append(bootstrapCases(t, "bootstrap.json", 25), bootstrapCases(t, "bootstrap-structure-errors.json", 31)...)
	for pass := 0; pass < 2; pass++ {
		for i := range cases {
			index := i
			if pass == 1 {
				index = len(cases) - 1 - i
			}
			c := cases[index]
			plan := bootstrapPlan(t, c)
			before := bootstrapPlan(t, c)
			got, problems := assembler.AssembleBootstrap(plan)
			assertBootstrap(t, c, got, problems)
			for i := range problems {
				problems[i].Message = "changed returned message"
				for key := range problems[i].Detail {
					problems[i].Detail[key] = "changed returned detail"
				}
			}
			if !reflect.DeepEqual(plan, before) {
				t.Fatal("projection or returned diagnostics changed input")
			}
			again, repeated := assembler.AssembleBootstrap(plan)
			assertBootstrap(t, c, again, repeated)
		}
	}
}

func TestBootstrapConcurrentSharedInstance(t *testing.T) {
	assembler := newAssembler()
	cases := append(bootstrapCases(t, "bootstrap.json", 25), bootstrapCases(t, "bootstrap-structure-errors.json", 31)...)
	var wait sync.WaitGroup
	for _, c := range cases {
		plan := bootstrapPlan(t, c)
		before := bootstrapPlan(t, c)
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, problems := assembler.AssembleBootstrap(plan)
			assertBootstrap(t, c, got, problems)
			if !reflect.DeepEqual(plan, before) {
				t.Errorf("%s: concurrent projection changed input", c.Name)
			}
		}()
	}
	wait.Wait()
}
