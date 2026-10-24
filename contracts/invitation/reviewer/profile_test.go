package reviewer_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/invitation/reviewer"
	"onboard-kit/internal/model"
)

func TestValidSupportProjectsEveryKnownFieldOnly(t *testing.T) {
	r := newReviewer()
	plan := string(fixture(t, "minimal.plan.yaml")) + "directory:\n  contacts:\n    - id: helper\n      name: Helper\n      email: helper@company.example\n  teams:\n    - id: platform\n      name: Platform\n"
	in := inspectPlan(t, r, []byte(plan))
	raw := fixture(t, "support.preparation.yaml")
	local := contract.Local{Preparation: contract.Body{Present: true, Bytes: raw}}
	before := copied(local)
	result, problems := r.Review(in, local, contract.PlanChoiceUseIncoming)
	if len(problems) != 0 || result.Proposal == nil {
		t.Fatal("valid prepared support refused")
	}
	want := contract.Support{
		Available: true,
		Recipient: model.PreparationRecipient{
			Name: " Jules ", CompanyEmail: "jules@company.example", PlannedStartDate: "2026-11-02",
			Team: "platform", Manager: "helper", OnboardingOwner: "helper", Buddy: "helper",
			WorkLocation: "remote", TimeZone: "Europe/Paris",
		},
		Arrangements: []model.Arrangement{{
			ID: "support-a", Section: "setup", Kind: "support", Owner: "helper",
			Contacts: []string{"helper", "helper"}, Reference: "Read the deployment notes",
			TargetDate: "2026-11-03", Status: "arranged", Note: "Ask when blocked",
		}},
	}
	p := *result.Proposal
	if !reflect.DeepEqual(p.Review.Support, want) || !reflect.DeepEqual(local, before) || !bytes.Equal(p.Review.Artifacts[1].Bytes, raw) || !bytes.Equal(p.Review.InvitationFirstSteps, fixture(t, "first-steps.md")) || len(p.Review.Session.Blockers) != 1 {
		t.Error("support projection lost fields/repeats, changed source, or supplied a personal answer")
	}
	pub, problems := r.Begin(p, beginAt)
	if len(problems) != 0 {
		t.Fatal("valid support proposal not accepted")
	}
	pub.OriginalReview.Support.Arrangements[0].Contacts[0] = "changed"
	if !reflect.DeepEqual(p.Review.Support, want) {
		t.Error("support helper list aliases another result")
	}
}

func TestPreviousPlanComparisonUnavailableDoesNotBlockExplicitReplacement(t *testing.T) {
	r := newReviewer()
	base := string(fixture(t, "minimal.plan.yaml"))
	cases := []struct{ name, source, reason string }{
		{"foreign", strings.Replace(base, "id: plan-a", "id: plan-b", 1), "different-id"},
		{"empty-id", strings.Replace(base, "id: plan-a\n", "", 1), "empty-id"},
		{"empty-context", "schemaVersion: 1\nid: plan-a\n", "empty-context"},
		{"bad-authoring", strings.Replace(base, "audience: onboardee", "audience: strange", 1), "authoring"},
		{"bad-integrity", base + "directory:\n  roles:\n    - id: lead\n      name: Lead\n      holders:\n        - missing\n", "validation"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			local := contract.Local{Plan: contract.Body{Present: true, Bytes: []byte(c.source)}, Record: contract.Body{Present: true, Bytes: fixture(t, "retained.record.yaml")}}
			p := proposal(t, r, local)
			if p.Review.Comparison.Available || p.Review.Comparison.Reason != c.reason || p.Review.Artifacts[0].Action != contract.ActionReplace || !bytes.Equal(p.Evidence.Local.Plan.Bytes, []byte(c.source)) || !bytes.Equal(p.Review.Artifacts[0].Bytes, fixture(t, "minimal.plan.yaml")) {
				t.Error("previous context availability must not discard original preimage or prohibit explicit replacement")
			}
		})
	}
}

func TestPostJoinPatternFailureRemainsFatal(t *testing.T) {
	r := newReviewer()
	plan := strings.Replace(string(fixture(t, "minimal.plan.yaml")), "        required: true\n", "        pattern: '^[a-z]+$'\n        required: true\n", 1)
	in := inspectPlan(t, r, []byte(plan))
	raw := string(in.Files.Record) + "answers:\n  sections:\n    setup:\n      account_name: BAD\n"
	got, problems := r.Review(in, contract.Local{Record: contract.Body{Present: true, Bytes: []byte(raw)}}, contract.PlanChoiceUseIncoming)
	if got.Proposal != nil || len(problems) != 1 {
		t.Fatalf("expected post-join pattern refusal, diagnostics=%d", len(problems))
	}
	d := problems[0]
	if d.Stage != "review.effective-validation" || d.Display != "Review stored personal answers against the incoming plan." || d.Problem.Code != "answer.pattern-mismatch" || d.Problem.Question != "setup.account_name" || d.Problem.Section != "setup" || d.Problem.Severity != model.SeverityError {
		t.Error("actual pattern failure must remain unchanged, not silently reconciled away")
	}
}

func TestCompleteOutcomeHistoryAndDuplicateRefusal(t *testing.T) {
	r := newReviewer()
	seed := string(fixture(t, "minimal.seed.yaml"))
	history := "outcomes:\n  - section: setup\n    outcome: failed\n    on: \"2026-10-22T10:00:00.123456789012+00:00\"\n    sectionHash: legacy\n    note: Held note\n    differently: Try again\n  - section: absent\n    outcome: deferred\n    on: \"2026-10-21T00:00:00Z\"\n    sectionHash: gone\n"
	raw := []byte(seed + history)
	p := proposal(t, r, contract.Local{Record: contract.Body{Present: true, Bytes: raw}})
	want := []contract.Outcome{
		{Outcome: model.Outcome{Section: "absent", Kind: model.OutcomeDeferred, On: "2026-10-21T00:00:00Z", SectionHash: "gone"}, Comparison: "unknown"},
		{Outcome: model.Outcome{Section: "setup", Kind: model.OutcomeFailed, On: "2026-10-22T10:00:00.123456789012+00:00", SectionHash: "legacy", Note: "Held note", Differently: "Try again"}, CurrentHash: sectionHash, Comparison: "different"},
	}
	if !reflect.DeepEqual(p.Review.Session.Outcomes, want) || !reflect.DeepEqual(p.Review.Reconciliation.ChangedOutcomes, []string{"setup"}) || p.Review.Artifacts[2].Action != contract.ActionRetain || !bytes.Equal(p.Review.Artifacts[2].Bytes, raw) || len(p.Review.Session.Activities) != 1 {
		t.Error("history must be complete and sorted only in derived view; mismatch must not rewrite or hide work")
	}
	duplicate := []byte(seed + history + "  - section: absent\n    outcome: verified\n    on: \"2026-10-23T00:00:00Z\"\n")
	got, problems := r.Review(incoming(t, r), contract.Local{Record: contract.Body{Present: true, Bytes: duplicate}}, contract.PlanChoiceUseIncoming)
	if got.Proposal != nil || len(problems) != 1 {
		t.Fatalf("duplicate held outcome must refuse, got %d diagnostics", len(problems))
	}
	d := problems[0]
	wantProblem := model.Problem{Code: "walk.duplicate-outcome", Severity: model.SeverityError, Message: "Keep one recorded outcome for this section.", Section: "absent", Detail: map[string]string{"field": "outcomes"}}
	if d.Stage != "review.walker" || d.Display != "Resolve the session's activity or history problem before opening." || !reflect.DeepEqual(d.Problem, wantProblem) {
		t.Error("actual duplicate outcome problem must be preserved without winner/pruning")
	}
}

func TestHistoryGuardCollectsInHeldFieldOrder(t *testing.T) {
	r := newReviewer()
	raw := string(fixture(t, "minimal.seed.yaml")) + "firstOpenedAt: not-a-time\noutcomes:\n  - section: setup\n    outcome: completed\n    on: yesterday\n  - section: setup\n    outcome: verified\n"
	got, problems := r.Review(incoming(t, r), contract.Local{Record: contract.Body{Present: true, Bytes: []byte(raw)}}, contract.PlanChoiceUseIncoming)
	if got.Proposal != nil || len(problems) != 4 {
		t.Fatalf("history phase must collect four defects before duplicate walker phase: got %d", len(problems))
	}
	paths := []string{"record.firstOpenedAt", "record.outcomes[0].kind", "record.outcomes[0].on", "record.outcomes[1].on"}
	for i, path := range paths {
		code, message := "invitation.history-time", "Preserve a real UTC time for interpreted history."
		if i == 1 {
			code, message = "invitation.outcome-kind", "Use verified, failed or deferred recorded outcomes."
		}
		ownProblem(t, problems[i:i+1], "review.record-history", code, path, message)
	}
}
