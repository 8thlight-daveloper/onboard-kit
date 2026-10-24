package reviewer_test

import (
	"bytes"
	"maps"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/session/reviewer"
	"onboard-kit/internal/model"
)

func TestReleaseIdentityEqualityRejectsBothLexicalDirections(t *testing.T) {
	for _, c := range []struct {
		name    string
		change  func(*contract.Source)
		code    string
		message string
		path    string
	}{
		{"shared-before", func(s *contract.Source) {
			s.Plan = []byte(strings.Replace(accessPlan, "id: plan-one", "id: a-plan", 1))
		}, "session.plan", "The shared plan does not match this session.", "plan.id"},
		{"shared-after", func(s *contract.Source) {
			s.Plan = []byte(strings.Replace(accessPlan, "id: plan-one", "id: z-plan", 1))
		}, "session.plan", "The shared plan does not match this session.", "plan.id"},
		{"account-before", func(s *contract.Source) {
			s.Record = []byte(strings.Replace(recordBefore, "account: newcomer", "account: a-person", 1))
		}, "session.record-account", "The record account does not match this session.", "record.account"},
		{"account-after", func(s *contract.Source) {
			s.Record = []byte(strings.Replace(recordBefore, "account: newcomer", "account: z-person", 1))
		}, "session.record-account", "The record account does not match this session.", "record.account"},
		{"record-plan-before", func(s *contract.Source) {
			s.Record = []byte(strings.Replace(recordBefore, "id: plan-one", "id: a-plan", 1))
		}, "session.record-plan", "The record plan does not match this session.", "record.plan.id"},
		{"record-plan-after", func(s *contract.Source) {
			s.Record = []byte(strings.Replace(recordBefore, "id: plan-one", "id: z-plan", 1))
		}, "session.record-plan", "The record plan does not match this session.", "record.plan.id"},
	} {
		t.Run(c.name, func(t *testing.T) {
			input := source(accessPlan)
			c.change(&input)
			original := cloneSource(input)
			view, diagnostics := newReviewer().Inspect(input)
			if !reflect.DeepEqual(view, contract.Session{}) {
				t.Fatal("mismatching identity yielded a usable session")
			}
			requireOwnError(t, diagnostics, "identity", c.code, c.message, c.path)
			proposal, diagnostics := newReviewer().Propose(input, contract.Intent{Operation: contract.OperationOutcome, Section: "access", Kind: model.OutcomeFailed})
			if !reflect.DeepEqual(proposal, contract.Proposal{}) {
				t.Fatal("mismatching identity yielded a usable proposal")
			}
			requireOwnError(t, diagnostics, "identity", c.code, c.message, c.path)
			if !reflect.DeepEqual(input, original) {
				t.Fatal("identity refusal changed the supplied source")
			}
		})
	}
}

func TestReleaseInactiveIntentFieldsRequireExactZero(t *testing.T) {
	answer := contract.Intent{Operation: contract.OperationAnswer, Question: "personal.choice", Value: text("org-a")}
	follow := contract.Intent{Operation: contract.OperationFollowPlan, Question: "personal.choice"}
	attempt := contract.Intent{Operation: contract.OperationOutcome, Section: "personal", Kind: model.OutcomeFailed}
	for _, c := range []struct {
		name   string
		base   contract.Intent
		change func(*contract.Intent)
	}{
		{"answer-section", answer, func(i *contract.Intent) { i.Section = "personal" }},
		{"answer-kind", answer, func(i *contract.Intent) { i.Kind = model.OutcomeFailed }},
		{"answer-note", answer, func(i *contract.Intent) { i.Note = "held note" }},
		{"answer-differently", answer, func(i *contract.Intent) { i.Differently = "held alternative" }},
		{"follow-section", follow, func(i *contract.Intent) { i.Section = "personal" }},
		{"follow-kind", follow, func(i *contract.Intent) { i.Kind = model.OutcomeFailed }},
		{"follow-note", follow, func(i *contract.Intent) { i.Note = "held note" }},
		{"follow-differently", follow, func(i *contract.Intent) { i.Differently = "held alternative" }},
		{"follow-value-kind", follow, func(i *contract.Intent) { i.Value.Kind = model.ValueKindDeferred }},
		{"follow-value-text", follow, func(i *contract.Intent) { i.Value.Text = "held text" }},
		{"follow-value-flag", follow, func(i *contract.Intent) { i.Value.Flag = true }},
		{"follow-value-empty-list", follow, func(i *contract.Intent) { i.Value.List = []string{} }},
		{"follow-value-list", follow, func(i *contract.Intent) { i.Value.List = []string{"held item"} }},
		{"outcome-question", attempt, func(i *contract.Intent) { i.Question = "personal.choice" }},
		{"outcome-value-kind", attempt, func(i *contract.Intent) { i.Value.Kind = model.ValueKindDeferred }},
		{"outcome-value-text", attempt, func(i *contract.Intent) { i.Value.Text = "held text" }},
		{"outcome-value-flag", attempt, func(i *contract.Intent) { i.Value.Flag = true }},
		{"outcome-value-empty-list", attempt, func(i *contract.Intent) { i.Value.List = []string{} }},
		{"outcome-value-list", attempt, func(i *contract.Intent) { i.Value.List = []string{"held item"} }},
	} {
		t.Run(c.name, func(t *testing.T) {
			intent := c.base
			c.change(&intent)
			proposal, diagnostics := newReviewer().Propose(source(answerPlan), intent)
			if !reflect.DeepEqual(proposal, contract.Proposal{}) {
				t.Fatal("inactive field became an accepted second change")
			}
			requireOwnError(t, diagnostics, "intent", "session.intent", "Choose exactly one session change.", "")
		})
	}
	for _, intent := range []contract.Intent{answer, follow, attempt} {
		proposal, diagnostics := newReviewer().Propose(source(answerPlan), intent)
		success(t, diagnostics)
		if proposal.Before.Account != "newcomer" || proposal.After.Account != "newcomer" || proposal.NeedsEvent != (intent.Operation == contract.OperationOutcome) || !reflect.DeepEqual(proposal.Intent, intent) {
			t.Fatalf("exact-zero control did not reach the actual proposal: %#v", proposal)
		}
	}
}

func releaseRequirePreparationProblem(t *testing.T, got contract.Diagnostic, code, path string) {
	t.Helper()
	if got.Stage != "preparation.validate" || got.Problem.Code != code || got.Problem.Severity != model.SeverityError || got.Problem.Message == "" || got.Problem.Question != "" || got.Problem.Section != "" || !maps.Equal(got.Problem.Detail, map[string]string{"path": path}) {
		t.Fatalf("preparation diagnostic lost its public meaning: %#v", got)
	}
}

func TestReleaseEveryStaleSupportPathAndMixedFatalBatch(t *testing.T) {
	for _, c := range []struct {
		name string
		tail string
		code string
		path string
	}{
		{"team", "  team: missing-team\n", "preparation.unknown-team", "recipient.team"},
		{"manager", "  manager: missing-contact\n", "preparation.unknown-contact", "recipient.manager"},
		{"onboarding-owner", "  onboardingOwner: missing-contact\n", "preparation.unknown-contact", "recipient.onboardingOwner"},
		{"buddy", "  buddy: missing-contact\n", "preparation.unknown-contact", "recipient.buddy"},
		{"section", "arrangements:\n  - id: help\n    section: missing-section\n    kind: support\n    status: pending\n", "preparation.unknown-section", "arrangements[0].section"},
		{"owner", "arrangements:\n  - id: help\n    section: access\n    kind: support\n    status: pending\n    owner: missing-contact\n", "preparation.unknown-contact", "arrangements[0].owner"},
		{"contact", "arrangements:\n  - id: help\n    section: access\n    kind: support\n    status: pending\n    contacts:\n      - missing-contact\n", "preparation.unknown-contact", "arrangements[0].contacts[0]"},
	} {
		t.Run(c.name, func(t *testing.T) {
			input := source(accessPlan)
			input.Preparation = []byte(preparation + c.tail)
			view, diagnostics := newReviewer().Inspect(input)
			success(t, diagnostics)
			if view.Account != "newcomer" || view.Support.Available || view.Support.Recipient != (model.PreparationRecipient{}) || view.Support.Arrangements != nil || len(view.Support.Problems) != 1 {
				t.Fatalf("one stale reference did not produce unavailable support: %#v", view)
			}
			releaseRequirePreparationProblem(t, view.Support.Problems[0], c.code, c.path)
			proposal, diagnostics := newReviewer().Propose(input, contract.Intent{Operation: contract.OperationOutcome, Section: "access", Kind: model.OutcomeFailed})
			success(t, diagnostics)
			if proposal.Before.Support.Available || proposal.After.Support.Available || len(proposal.Before.Support.Problems) != 1 || len(proposal.After.Support.Problems) != 1 || !reflect.DeepEqual(proposal.AfterRecord.Answers, heldRecord().Answers) {
				t.Fatal("a stale-support proposal repaired support or silently supplied an answer")
			}
			releaseRequirePreparationProblem(t, proposal.Before.Support.Problems[0], c.code, c.path)
			releaseRequirePreparationProblem(t, proposal.After.Support.Problems[0], c.code, c.path)
			input.Preparation = []byte(strings.Replace(string(input.Preparation), "  name: Example newcomer\n", "  name: Example newcomer\n  plannedStartDate: \"2026-02-30\"\n", 1))
			refused, diagnostics := newReviewer().Inspect(input)
			errors := fatal(diagnostics)
			if !reflect.DeepEqual(refused, contract.Session{}) || len(errors) != 2 {
				t.Fatalf("mixed stale/fatal batch was not refused in full: %#v %#v", refused, diagnostics)
			}
			releaseRequirePreparationProblem(t, errors[0], "preparation.date", "recipient.plannedStartDate")
			releaseRequirePreparationProblem(t, errors[1], c.code, c.path)
		})
	}
	for _, c := range []struct {
		name string
		tail string
		code string
		path string
	}{
		{"empty-contact", "arrangements:\n  - id: help\n    section: access\n    kind: support\n    status: pending\n    contacts: [\"\"]\n", "preparation.unknown-contact", "arrangements[0].contacts[0]"},
		{"malformed-contact", "  manager: Missing-Contact\n", "preparation.unknown-contact", "recipient.manager"},
		{"reserved-section", "arrangements:\n  - id: help\n    section: \"yes\"\n    kind: support\n    status: pending\n", "preparation.unknown-section", "arrangements[0].section"},
		{"missing-section", "arrangements:\n  - id: help\n    kind: support\n    status: pending\n", "preparation.unknown-section", "arrangements[0].section"},
		{"missing-arranged-owner", "arrangements:\n  - id: help\n    section: access\n    kind: support\n    status: arranged\n", "preparation.incomplete-arrangement", "arrangements[0].owner"},
	} {
		t.Run(c.name, func(t *testing.T) {
			input := source(accessPlan)
			input.Preparation = []byte(preparation + c.tail)
			view, diagnostics := newReviewer().Inspect(input)
			errors := fatal(diagnostics)
			if !reflect.DeepEqual(view, contract.Session{}) || len(errors) != 1 {
				t.Fatalf("excluded stale-support shape became usable: %#v %#v", view, diagnostics)
			}
			releaseRequirePreparationProblem(t, errors[0], c.code, c.path)
		})
	}
}

func TestReleasePreviewPreservesEveryHistoricalSentinelProjection(t *testing.T) {
	const sentinel = "2000-01-01T00:00:00Z"
	input := source(accessPlan)
	input.Record = []byte(strings.Replace(strings.Replace(recordBefore, "on: legacy-on", "on: \""+sentinel+"\"", 1), "on: older", "on: \""+sentinel+"\"", 1))
	input.Record = []byte(strings.Replace(string(input.Record), "  - section: a-removed", "  - section: access\n    outcome: failed\n    on: \""+sentinel+"\"\n    sectionHash: old-target\n    note: Held target note\n    differently: Held target alternative\n  - section: a-removed", 1))
	proposal, diagnostics := newReviewer().Propose(input, contract.Intent{Operation: contract.OperationOutcome, Section: "access", Kind: model.OutcomeVerified})
	success(t, diagnostics)
	wantBefore := heldRecord()
	wantBefore.Outcomes[0].On = sentinel
	wantBefore.Outcomes[1].On = sentinel
	wantBefore.Outcomes = []model.Outcome{wantBefore.Outcomes[0], {Section: "access", Kind: model.OutcomeFailed, On: sentinel, SectionHash: "old-target", Note: "Held target note", Differently: "Held target alternative"}, wantBefore.Outcomes[1]}
	sameRecord(t, proposal.BeforeRecord, wantBefore)
	for _, id := range []string{"access", "a-removed", "z-removed"} {
		if outcome(t, proposal.Before, id).Outcome.On != sentinel {
			t.Fatalf("original event was blanked in Before for %s", id)
		}
	}
	wantAfter := heldRecord()
	wantAfter.Outcomes[0].On = sentinel
	wantAfter.Outcomes[1].On = sentinel
	wantAfter.Outcomes = []model.Outcome{wantAfter.Outcomes[1], {Section: "access", Kind: model.OutcomeVerified, SectionHash: accessHash}, wantAfter.Outcomes[0]}
	sameRecord(t, proposal.AfterRecord, wantAfter)
	if !proposal.NeedsEvent || outcome(t, proposal.After, "access").Outcome.On != "" || outcome(t, proposal.After, "a-removed").Outcome.On != sentinel || outcome(t, proposal.After, "z-removed").Outcome.On != sentinel {
		t.Fatal("preview blanking was not restricted to the newly proposed target")
	}
	publication, diagnostics := newReviewer().Finalize(proposal, event)
	success(t, diagnostics)
	wantBytes := strings.Replace(strings.Replace(recordVerified, "on: legacy-on", "on: \""+sentinel+"\"", 1), "on: older", "on: \""+sentinel+"\"", 1)
	if !publication.Changed || publication.Event != event || !bytes.Equal(publication.Record, []byte(wantBytes)) {
		t.Fatalf("final event overwrote historical sentinel text: %#v\n%s", publication, publication.Record)
	}
}

func TestReleaseFinalizeChecksIsolatedPublicProjectionFields(t *testing.T) {
	for _, side := range []string{"before", "after"} {
		for _, c := range []struct {
			name   string
			change func(*contract.Session)
		}{
			{"account", func(s *contract.Session) { s.Account = "other" }},
			{"plan-id", func(s *contract.Session) { s.PlanID = "other" }},
			{"reference-id", func(s *contract.Session) { s.Reference.ID = "other" }},
			{"reference-url", func(s *contract.Session) { s.Reference.URL = "other" }},
			{"reference-match", func(s *contract.Session) { s.PlanReferenceMatches = !s.PlanReferenceMatches }},
			{"activity-hash", func(s *contract.Session) { s.Activities[0].CurrentHash = "forged" }},
			{"activity-waiting", func(s *contract.Session) { s.Activities[0].WaitingOn = []string{"forged"} }},
			{"activity-escalation", func(s *contract.Session) { s.Activities[0].Escalation = []string{"forged"} }},
			{"outcome-hash", func(s *contract.Session) { s.Outcomes[0].CurrentHash = "forged" }},
			{"outcome-comparison", func(s *contract.Session) { s.Outcomes[0].Comparison = "same" }},
			{"outcome-on", func(s *contract.Session) { s.Outcomes[0].Outcome.On = "forged" }},
			{"question-description", func(s *contract.Session) { s.Questions[0].Description = "forged" }},
			{"question-answer-inactive-field", func(s *contract.Session) { s.Questions[0].Answer.Flag = true }},
			{"question-default-inactive-list", func(s *contract.Session) { s.Questions[0].DeclaredDefault.List = []string{} }},
			{"blocker", func(s *contract.Session) { s.Blockers = []model.Blocker{{Question: "forged"}} }},
			{"support-availability", func(s *contract.Session) { s.Support.Available = false }},
			{"support-recipient", func(s *contract.Session) { s.Support.Recipient.Name = "forged" }},
			{"support-arrangement", func(s *contract.Session) { s.Support.Arrangements = []model.Arrangement{{ID: "forged"}} }},
			{"support-problem", func(s *contract.Session) { s.Support.Problems = []contract.Diagnostic{{Stage: "forged"}} }},
		} {
			t.Run(side+"/"+c.name, func(t *testing.T) {
				proposal, diagnostics := newReviewer().Propose(source(answerPlan), contract.Intent{Operation: contract.OperationAnswer, Question: "personal.choice", Value: text("org-a")})
				success(t, diagnostics)
				if proposal.Before.Account != "newcomer" || proposal.After.Account != "newcomer" || len(proposal.Before.Activities) != 1 || len(proposal.After.Activities) != 1 || len(proposal.Before.Outcomes) != 2 || len(proposal.After.Outcomes) != 2 || len(proposal.Before.Questions) != 1 || len(proposal.After.Questions) != 1 {
					t.Fatal("missing actual projections before the one-field tamper")
				}
				if side == "before" {
					c.change(&proposal.Before)
				} else {
					c.change(&proposal.After)
				}
				publication, diagnostics := newReviewer().Finalize(proposal, "")
				if !reflect.DeepEqual(publication, contract.Publication{}) {
					t.Fatal("isolated forged projection produced a publication")
				}
				requireOwnError(t, diagnostics, "proposal", "session.proposal", "Review this change again.", "")
			})
		}
	}
	proposal, diagnostics := newReviewer().Propose(source(answerPlan), contract.Intent{Operation: contract.OperationAnswer, Question: "personal.choice", Value: text("org-a")})
	success(t, diagnostics)
	if proposal.Before.Account != "newcomer" || proposal.After.Account != "newcomer" {
		t.Fatal("missing positive proposal for permitted top-level empty-slice controls")
	}
	proposal.Before.Blockers = []model.Blocker{}
	proposal.After.Blockers = nil
	publication, diagnostics := newReviewer().Finalize(proposal, "")
	success(t, diagnostics)
	if !publication.Changed || !bytes.Equal(publication.Record, []byte(recordOwned)) {
		t.Fatal("permitted derived top-level empty slice shape prevented the actual save")
	}
}
