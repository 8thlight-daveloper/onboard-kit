package reviewer_test

import (
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/session/reviewer"
	"onboard-kit/internal/model"
)

func TestEveryKnownSupportFieldAndHelperRepeatIsOnlySupport(t *testing.T) {
	input := source(accessPlan)
	input.Plan = []byte(strings.Replace(accessPlan, "sections:\n", `directory:
  contacts:
    - id: lead
      name: Example lead
      email: lead@example.invalid
    - id: buddy
      name: Example buddy
      email: buddy@example.invalid
  teams:
    - id: team-one
      name: Team one
sections:
`, 1))
	input.Preparation = []byte(`schemaVersion: 1
kind: preparation
account: newcomer
planId: plan-one
recipient:
  name: Example newcomer
  companyEmail: newcomer@example.invalid
  plannedStartDate: "2026-11-02"
  team: team-one
  manager: lead
  onboardingOwner: lead
  buddy: buddy
  workLocation: remote
  timeZone: Example/Zone
arrangements:
  - id: access-support
    section: access
    kind: support
    owner: lead
    contacts:
      - buddy
      - buddy
    reference: Example booking
    targetDate: "2026-11-05"
    status: arranged
    note: Example support note
future: not-a-support-field
`)
	session, diagnostics := newReviewer().Inspect(input)
	success(t, diagnostics)
	wantRecipient := model.PreparationRecipient{
		Name: "Example newcomer", CompanyEmail: "newcomer@example.invalid", PlannedStartDate: "2026-11-02",
		Team: "team-one", Manager: "lead", OnboardingOwner: "lead", Buddy: "buddy", WorkLocation: "remote", TimeZone: "Example/Zone",
	}
	wantArrangements := []model.Arrangement{{
		ID: "access-support", Section: "access", Kind: "support", Owner: "lead", Contacts: []string{"buddy", "buddy"},
		Reference: "Example booking", TargetDate: "2026-11-05", Status: "arranged", Note: "Example support note",
	}}
	if !session.Support.Available || !reflect.DeepEqual(session.Support.Recipient, wantRecipient) || !reflect.DeepEqual(session.Support.Arrangements, wantArrangements) || len(session.Support.Problems) != 0 || len(session.Outcomes) != 2 {
		t.Fatalf("support was dropped, normalized or turned into an outcome %#v", session)
	}
	proposal, diagnostics := newReviewer().Propose(input, contract.Intent{Operation: contract.OperationOutcome, Section: "access", Kind: model.OutcomeFailed})
	success(t, diagnostics)
	if !reflect.DeepEqual(proposal.AfterRecord.Answers, heldRecord().Answers) || proposal.AfterRecord.FirstOpenedAt != "held-uninterpreted-event" {
		t.Fatal("preparation promoted an answer or event")
	}
	session.Support.Arrangements[0].Contacts[0] = "changed"
	if !reflect.DeepEqual(proposal.Before.Support.Arrangements, wantArrangements) {
		t.Fatal("support result aliases another call")
	}
}

func TestBothAudiencesAndPreparationBlockersRemainVisible(t *testing.T) {
	plan := strings.Replace(answerPlan, "summary: Use $choice.", "summary: Use $choice and $manager.", 1)
	plan += "      - name: manager\n        type: string\n        required: true\n        audience: onboarder\n"
	input := source(plan)
	session, diagnostics := newReviewer().Inspect(input)
	success(t, diagnostics)
	if len(session.Questions) != 2 || question(t, session, "personal.choice").Audience != model.AudienceOnboardee ||
		question(t, session, "personal.manager").Audience != model.AudienceOnboarder || len(session.Blockers) != 1 || session.Blockers[0].Question != "personal.manager" {
		t.Fatalf("audience/question/blocker view filtered %#v", session)
	}
	proposal, diagnostics := newReviewer().Propose(input, contract.Intent{Operation: contract.OperationOutcome, Section: "personal", Kind: model.OutcomeVerified})
	success(t, diagnostics)
	if !proposal.NeedsEvent || len(proposal.After.Blockers) != 1 || proposal.After.Blockers[0].Question != "personal.manager" {
		t.Fatal("preparation blocker became an invented outcome gate")
	}
}
