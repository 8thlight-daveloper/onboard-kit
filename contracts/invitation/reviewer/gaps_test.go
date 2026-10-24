package reviewer_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/invitation/reviewer"
	"onboard-kit/internal/model"
)

func TestGapIncomingPlanIdentityPrecedesSeed(t *testing.T) {
	base := string(fixture(t, "minimal.plan.yaml"))
	for _, replacement := range []string{"", "id: \"\"\n"} {
		plan := []byte(strings.Replace(base, "id: plan-a\n", replacement, 1))
		input := archive(t, map[string][]byte{"plan.obk.yaml": plan})
		before := append([]byte(nil), input...)
		got, problems := newReviewer().Inspect(input)
		ownProblem(t, problems, "inspect.plan-identity", "invitation.plan-id", "plan.id", "Supply an invitation plan identity.")
		if !reflect.DeepEqual(got, contract.Incoming{}) || !bytes.Equal(input, before) {
			t.Error("missing plan identity must refuse without partial content or source changes")
		}
	}
}

func TestGapKnownMissingAndPresentRecordIsInconsistent(t *testing.T) {
	r := newReviewer()
	in := incoming(t, r)
	for _, body := range [][]byte{nil, fixture(t, "minimal.seed.yaml")} {
		for _, choice := range []contract.PlanChoice{contract.PlanChoiceUndecided, contract.PlanChoice("other")} {
			local := contract.Local{Record: contract.Body{Present: true, Bytes: body}, KnownRecordMissing: true}
			before := copied(local)
			inBefore := copied(in)
			got, problems := r.Review(in, local, choice)
			ownProblem(t, problems, "review.local", "invitation.local", "local.knownRecordMissing", "Supply consistent local presence observations.")
			if !reflect.DeepEqual(got, contract.ReviewResult{}) || !reflect.DeepEqual(local, before) || !reflect.DeepEqual(in, inBefore) {
				t.Error("inconsistent registry observations must refuse before choice or record decoding and preserve evidence")
			}
		}
	}
}

func TestGapSectionOwnedAnswersRefuseEvenInactive(t *testing.T) {
	base := string(fixture(t, "minimal.plan.yaml"))
	for _, disabled := range []bool{false, true} {
		for _, audience := range []string{"", "        audience: onboarder\n"} {
			plan := base + "      - name: org_url\n        description: Organization address\n" + audience
			if disabled {
				plan = strings.Replace(plan, "    bootstrap: true\n", "    bootstrap: true\n    enabled: false\n", 1)
			}
			r := newReviewer()
			in := inspectPlan(t, r, []byte(plan))
			for _, value := range []string{"false", "deferred", "\"\""} {
				raw := append([]byte(nil), in.Files.Record...)
				raw = append(raw, []byte("answers:\n  sections:\n    setup:\n      org_url: "+value+"\n")...)
				local := contract.Local{Record: contract.Body{Present: true, Bytes: raw}}
				before := copied(local)
				inBefore := copied(in)
				got, problems := r.Review(in, local, contract.PlanChoiceUseIncoming)
				ownProblem(t, problems, "review.record-ownership", "invitation.record-owned-answer", "answers.sections[0].values[0]", "Keep onboarder-owned answers out of the personal record.")
				if !reflect.DeepEqual(got, contract.ReviewResult{}) || !reflect.DeepEqual(local, before) || !reflect.DeepEqual(in, inBefore) {
					t.Error("section-owned keys must refuse intact regardless of activity, default audience or held value")
				}
			}
		}
	}
}

func TestGapOwnershipRanksIncludeHiddenKeysAndScopes(t *testing.T) {
	plan := string(fixture(t, "minimal.plan.yaml")) + `  - id: alpha
    title: Alpha
    summary: Organizational context.
    enabled: false
    variables:
      - name: bravo_org
      - name: tango_org
        audience: onboarder
  - id: zeta
    title: Zeta
    summary: Organizational context.
    variables:
      - name: bravo_org
      - name: tango_org
        audience: onboarder
variables:
  - name: bravo_org
  - name: tango_org
    audience: onboarder
`
	r := newReviewer()
	in := inspectPlan(t, r, []byte(plan))
	raw := string(in.Files.Record) + `answers:
  plan:
    tango_org: deferred
    "z-private\u001b[31m": held
    "m-private\u001b[31m": held
    bravo_org: false
    "a-private\u001b[31m": held
  sections:
    zeta:
      tango_org: ""
      "z-private\u001b[31m": held
      "m-private\u001b[31m": held
      bravo_org: deferred
      "a-private\u001b[31m": held
    "zz-private\u001b[31m":
      bravo_org: held
    "middle-private\u001b[31m":
      tango_org: held
    alpha:
      tango_org: false
      "z-private\u001b[31m": held
      "m-private\u001b[31m": held
      bravo_org: ""
      "a-private\u001b[31m": held
    "a-private\u001b[31m":
      bravo_org: held
`
	local := contract.Local{Record: contract.Body{Present: true, Bytes: []byte(raw)}}
	before := copied(local)
	inBefore := copied(in)
	wantPaths := []string{
		"answers.plan[1]",
		"answers.plan[3]",
		"answers.sections[1].values[1]",
		"answers.sections[1].values[3]",
		"answers.sections[3].values[1]",
		"answers.sections[3].values[3]",
	}
	for attempt := 0; attempt < 24; attempt++ {
		got, problems := r.Review(in, local, contract.PlanChoiceUseIncoming)
		if len(problems) != len(wantPaths) {
			t.Fatalf("ownership diagnostic count=%d, want six independent fixed-rank problems", len(problems))
		}
		for i, path := range wantPaths {
			ownProblem(t, problems[i:i+1], "review.record-ownership", "invitation.record-owned-answer", path, "Keep onboarder-owned answers out of the personal record.")
		}
		if !reflect.DeepEqual(got, contract.ReviewResult{}) || !reflect.DeepEqual(local, before) || !reflect.DeepEqual(in, inBefore) {
			t.Error("ordered ownership refusal must preserve every held key and input byte")
		}
	}
}

func TestGapOwnershipScopesRemainSeparate(t *testing.T) {
	base := string(fixture(t, "minimal.plan.yaml"))
	for _, personalScope := range []string{"section", "plan"} {
		t.Run(personalScope, func(t *testing.T) {
			plan := base + "      - name: org_url\n"
			if personalScope == "section" {
				plan += "        audience: onboardee\nvariables:\n  - name: org_url\n"
				plan = strings.Replace(plan, "summary: Use $account_name.", "summary: Use $account_name and $org_url.", 1)
			} else {
				plan += "variables:\n  - name: org_url\n    audience: onboardee\n"
				plan = strings.Replace(plan, "    bootstrap: true\n", "    bootstrap: true\n    includeIf:\n      all:\n        - question: plan.org_url\n          answer:\n            isSettled: true\n", 1)
			}
			r := newReviewer()
			in := inspectPlan(t, r, []byte(plan))
			raw := string(in.Files.Record)
			if personalScope == "section" {
				raw += "answers:\n  sections:\n    setup:\n      org_url: supplied\n    hidden-scope:\n      org_url: false\n"
			} else {
				raw += "answers:\n  plan:\n    org_url: supplied\n  sections:\n    hidden-scope:\n      org_url: deferred\n"
			}
			local := contract.Local{Record: contract.Body{Present: true, Bytes: []byte(raw)}}
			before := copied(local)
			got, problems := r.Review(in, local, contract.PlanChoiceUseIncoming)
			if len(problems) != 0 || got.PlanChoiceNeeded || got.Proposal == nil {
				t.Fatalf("an answer owned by its exact declared scope, plus hidden scopes, must remain allowed: diagnostics=%d", len(problems))
			}
			if !reflect.DeepEqual(local, before) || !bytes.Equal(got.Proposal.Review.Artifacts[2].Bytes, local.Record.Bytes) {
				t.Error("valid separate-scope and hidden answers must be retained exactly")
			}
		})
	}
}

func TestGapAvailablePreviousPlanSuppliesAllComparisonLists(t *testing.T) {
	previous := string(fixture(t, "minimal.plan.yaml")) + `  - id: retired
    title: Retired
    summary: Previous context.
variables:
  - name: old_org
`
	plan := string(fixture(t, "choices.plan.yaml")) + `  - id: fresh
    title: Fresh
    summary: Incoming context.
`
	r := newReviewer()
	in := inspectPlan(t, r, []byte(plan))
	raw := string(in.Files.Record) + `answers:
  sections:
    setup:
      flavors: red
      account_name: held
outcomes:
  - section: setup
    outcome: verified
    on: "2026-10-23T00:00:00Z"
    sectionHash: obsolete
  - section: retired
    outcome: deferred
    on: "2026-10-22T00:00:00Z"
    sectionHash: obsolete
`
	local := contract.Local{
		Plan:   contract.Body{Present: true, Bytes: []byte(previous)},
		Record: contract.Body{Present: true, Bytes: []byte(raw)},
	}
	before := copied(local)
	got, problems := r.Review(in, local, contract.PlanChoiceUseIncoming)
	if len(problems) != 0 || got.PlanChoiceNeeded || got.Proposal == nil {
		t.Fatalf("valid previous context and current choice reconciliation must produce a complete proposal: diagnostics=%d", len(problems))
	}
	want := model.Reconciliation{
		NewQuestions:     []string{"setup.flavors", "setup.tools"},
		RemovedQuestions: []string{"plan.old_org", "setup.account_name"},
		AnswersToReview:  []string{"setup.flavors"},
		NewSections:      []string{"fresh"},
		RemovedSections:  []string{"retired"},
		ChangedOutcomes:  []string{"setup"},
	}
	view := got.Proposal.Review
	if view.Comparison.Reason != "available" || !view.Comparison.Available || len(view.Comparison.Problems) != 0 || !reflect.DeepEqual(view.Reconciliation, want) {
		t.Error("actual reconciliation must use validated matching previous definitions and preserve all six literal lists")
	}
	if !reflect.DeepEqual(local, before) || !bytes.Equal(got.Proposal.Evidence.Local.Plan.Bytes, local.Plan.Bytes) || !bytes.Equal(got.Proposal.Evidence.Local.Record.Bytes, local.Record.Bytes) {
		t.Error("comparison and reconciliation must preserve exact previous and record source evidence")
	}
}

func TestGapWalkerReceivesJoinedPersonalAnswer(t *testing.T) {
	r := newReviewer()
	in := incoming(t, r)
	inBefore := copied(in)
	raw := string(in.Files.Record) + "answers:\n  sections:\n    setup:\n      account_name: jules\n"
	local := contract.Local{Record: contract.Body{Present: true, Bytes: []byte(raw)}}
	before := copied(local)
	got, problems := r.Review(in, local, contract.PlanChoiceUseIncoming)
	if len(problems) != 0 || got.PlanChoiceNeeded || got.Proposal == nil {
		t.Fatalf("a valid personal answer must open without a publication event: diagnostics=%d", len(problems))
	}
	view := got.Proposal.Review
	if len(view.Session.Activities) != 1 || len(view.Session.Questions) != 1 {
		t.Fatal("joined session must retain its one activity and question")
	}
	activity := view.Session.Activities[0]
	question := view.Session.Questions[0]
	if activity.Section.ID != "setup" || activity.Section.Summary != "Use jules." || activity.CurrentHash != sectionHash {
		t.Error("walker must resolve held personal answers while hashing the unchanged original definition")
	}
	if question.ID != "setup.account_name" || !reflect.DeepEqual(question.Answer, model.Value{Kind: model.ValueKindText, Text: "jules"}) || len(view.Session.Blockers) != 0 || len(view.SessionProblems) != 0 {
		t.Error("walker questions must contain the actual joined text without an unanswered blocker or gap")
	}
	if !reflect.DeepEqual(in, inBefore) || !reflect.DeepEqual(local, before) || view.NeedsFirstOpenedAt || !bytes.Equal(view.Artifacts[2].Bytes, local.Record.Bytes) {
		t.Error("joining must preserve invitation/record sources and an existing record must not become new")
	}
}
