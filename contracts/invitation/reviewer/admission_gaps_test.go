package reviewer_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/invitation/reviewer"
	"onboard-kit/internal/model"
)

func gapAdmissionArchive(t *testing.T, plan []byte) []byte {
	t.Helper()
	hash := sha256.Sum256(plan)
	seed := bytes.Replace(fixture(t, "minimal.seed.yaml"), []byte(planHash), []byte(hex.EncodeToString(hash[:])), 1)
	return archive(t, map[string][]byte{"plan.obk.yaml": plan, "records/jules/record.obk.yaml": seed})
}

func TestAdmissionGapEffectiveUsesJoinedPlan(t *testing.T) {
	plan := bytes.Replace(fixture(t, "minimal.plan.yaml"), []byte("    bootstrap: true\n"), []byte("    bootstrap: true\n    body: \""+strings.Repeat("$", 1000)+"\"\n"), 1)
	r := newReviewer()
	in := inspectPlan(t, r, plan)
	before := copied(in)
	for _, length := range []int{1000, 10000} {
		answer := strings.Repeat("a", length)
		raw := string(in.Files.Record) + "answers:\n  sections:\n    setup:\n      account_name: " + answer + "\n"
		local := contract.Local{Record: contract.Body{Present: true, Bytes: []byte(raw)}}
		localBefore := copied(local)
		got, problems := r.Review(in, local, contract.PlanChoiceUseIncoming)
		if length == 10000 {
			resourceProblem(t, problems, "review.effective-admission", "effectivePlan")
			if !reflect.DeepEqual(got, contract.ReviewResult{}) {
				t.Error("joined-answer fanout refusal must not return a partial session")
			}
		} else {
			if len(problems) != 0 || got.Proposal == nil || got.PlanChoiceNeeded || len(got.Proposal.Review.Session.Activities) != 1 || len(got.Proposal.Review.Session.Questions) != 1 {
				t.Fatal("smaller joined answer must remain an admitted complete session")
			}
			view := got.Proposal.Review.Session
			if view.Activities[0].Section.Body != strings.Repeat("$", 500) || view.Activities[0].Section.Summary != "Use "+answer+"." || !reflect.DeepEqual(view.Questions[0].Answer, model.Value{Kind: model.ValueKindText, Text: answer}) || len(view.Blockers) != 0 {
				t.Error("conservative charging must preserve actual small escaped-dollar rendering and the personal answer")
			}
		}
		if !reflect.DeepEqual(in, before) || !reflect.DeepEqual(local, localBefore) {
			t.Error("resource admission must not change joined-answer source evidence")
		}
	}
}

func gapEffectiveRootBoundary(t *testing.T, extraLines int) {
	t.Helper()
	plan := paddedAdmissionPlan(t, 440000, 10599)
	raw := admissionArchive(t, plan)
	if len(raw) != 451468 {
		t.Fatal("independent B-minus480 archive length changed")
	}
	r := newReviewer()
	in, problems := r.Inspect(raw)
	if len(problems) != 0 {
		t.Fatal("incoming B-minus480 must remain admitted")
	}
	preparation := fixture(t, "admission.preparation.yaml")
	if len(preparation) != 60 {
		t.Fatal("literal retained preparation must have exactly60 source bytes")
	}
	for _, extra := range []int{0, extraLines} {
		body := append(append([]byte(nil), preparation...), bytes.Repeat([]byte{'\n'}, extra)...)
		local := contract.Local{Preparation: contract.Body{Present: true, Bytes: body}}
		before := copied(local)
		inBefore := copied(in)
		got, problems := r.Review(in, local, contract.PlanChoiceUseIncoming)
		if extra == 0 {
			if len(problems) != 0 || got.Proposal == nil || got.PlanChoiceNeeded || len(got.Proposal.Review.Artifacts) != 3 || got.Proposal.Review.Artifacts[1].Action != contract.ActionRetain || !bytes.Equal(got.Proposal.Review.Artifacts[1].Bytes, body) {
				t.Fatal("unchanged retained60-byte preparation must reach inclusive effective B")
			}
		} else {
			resourceProblem(t, problems, "review.effective-admission", "effectivePlan")
			if !reflect.DeepEqual(got, contract.ReviewResult{}) {
				t.Error("source-only excess must refuse effective admission without a proposal")
			}
		}
		if !reflect.DeepEqual(local, before) || !reflect.DeepEqual(in, inBefore) {
			t.Error("retained preparation source bytes must remain exact on both sides of B")
		}
	}
}

func TestAdmissionGapEffectiveKeepsProposedRecordRoot(t *testing.T) {
	gapEffectiveRootBoundary(t, 1)
}

func TestAdmissionGapEffectiveKeepsSelectedPreparationRoot(t *testing.T) {
	gapEffectiveRootBoundary(t, 100)
}

func gapCardinalityPlan(kind string, count int) []byte {
	var source strings.Builder
	source.WriteString("schemaVersion: 1\nid: plan-a\n")
	if kind == "variables" {
		source.WriteString("variables:\n")
		for i := 0; i < count; i++ {
			fmt.Fprintf(&source, "  - name: var%04d\n", i)
		}
		return []byte(source.String())
	}
	if kind == "sections" {
		source.WriteString("sections:\n")
		for i := 0; i < count; i++ {
			fmt.Fprintf(&source, "  - id: s%04d\n    title: S\n    summary: S\n", i)
		}
		return []byte(source.String())
	}
	if kind == "steps" || kind == "checks" {
		source.WriteString("sections:\n  - id: task\n    title: T\n    summary: T\n    steps:\n")
		if kind == "steps" {
			for i := 0; i < count; i++ {
				source.WriteString("      - {}\n")
			}
		} else {
			source.WriteString("      - checks:\n")
			for i := 0; i < count; i++ {
				source.WriteString("          - ask: A\n")
			}
		}
		return []byte(source.String())
	}
	source.WriteString("directory:\n")
	contacts, roles, teams := 0, 0, 0
	switch kind {
	case "contacts":
		contacts = count
	case "roles":
		roles = count
	case "teams":
		teams = count
	case "combined-directory":
		contacts, roles, teams = 2048, 2048, count-4096
	}
	if contacts > 0 {
		source.WriteString("  contacts:\n")
		for i := 0; i < contacts; i++ {
			fmt.Fprintf(&source, "    - id: c%04d\n      name: C\n      email: c@x\n", i)
		}
	}
	if roles > 0 {
		source.WriteString("  roles:\n")
		for i := 0; i < roles; i++ {
			fmt.Fprintf(&source, "    - id: r%04d\n      name: R\n", i)
		}
	}
	if teams > 0 {
		source.WriteString("  teams:\n")
		for i := 0; i < teams; i++ {
			fmt.Fprintf(&source, "    - id: t%04d\n      name: T\n", i)
		}
	}
	return []byte(source.String())
}

func TestAdmissionGapSimpleCardinalityCaps(t *testing.T) {
	cases := []struct {
		name  string
		limit int
	}{
		{"sections", 256},
		{"variables", 1024},
		{"steps", 4096},
		{"checks", 8192},
		{"contacts", 2048},
		{"roles", 2048},
		{"teams", 2048},
		{"combined-directory", 4096},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, count := range []int{c.limit, c.limit + 1} {
				plan := gapCardinalityPlan(c.name, count)
				if len(plan) > 300000 {
					t.Fatal("independent small-source cap witness unexpectedly grew")
				}
				input := gapAdmissionArchive(t, plan)
				before := append([]byte(nil), input...)
				got, problems := newReviewer().Inspect(input)
				if count == c.limit {
					if len(problems) != 0 || got.PlanID != "plan-a" {
						t.Fatalf("inclusive cardinality boundary must remain admitted: diagnostics=%d", len(problems))
					}
				} else {
					resourceProblem(t, problems, "inspect.plan-admission", "plan")
					if !reflect.DeepEqual(got, contract.Incoming{}) {
						t.Error("over-cap declaration must not yield partial incoming content")
					}
				}
				if !bytes.Equal(input, before) {
					t.Error("cardinality refusal must preserve exact archive input")
				}
			}
		})
	}
}

func TestAdmissionGapOutcomeCardinalityBeforeHistory(t *testing.T) {
	r := newReviewer()
	in := incoming(t, r)
	for _, count := range []int{4096, 4097} {
		var source strings.Builder
		source.Write(in.Files.Record)
		source.WriteString("outcomes:\n")
		for i := 0; i < count; i++ {
			fmt.Fprintf(&source, "  - section: h%04d\n    outcome: deferred\n    on: \"2026-10-23T00:00:00Z\"\n", i)
		}
		body := []byte(source.String())
		if len(body) > 400000 {
			t.Fatal("independent historical cap source unexpectedly grew")
		}
		local := contract.Local{Record: contract.Body{Present: true, Bytes: body}}
		before := copied(local)
		got, problems := r.Review(in, local, contract.PlanChoiceUseIncoming)
		if count == 4096 {
			if len(problems) != 0 || got.Proposal == nil || got.PlanChoiceNeeded || len(got.Proposal.Review.Session.Outcomes) != 4096 {
				t.Fatal("inclusive held-outcome boundary must remain a complete preserved history")
			}
		} else {
			resourceProblem(t, problems, "review.record-admission", "local.record")
			if !reflect.DeepEqual(got, contract.ReviewResult{}) {
				t.Error("over-cap history must refuse before interpretation or partial session projection")
			}
		}
		if !reflect.DeepEqual(local, before) {
			t.Error("outcome admission must preserve original held history source")
		}
	}
}
