package reviewer_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/session/reviewer"
	"onboard-kit/internal/model"
)

func TestInspectActualSessionAndHistoricalData(t *testing.T) {
	input := source(accessPlan)
	original := cloneSource(input)
	session, diagnostics := newReviewer().Inspect(input)
	success(t, diagnostics)
	if session.Account != "newcomer" || session.PlanID != "plan-one" || session.Reference != heldRecord().Plan || session.PlanReferenceMatches {
		t.Fatalf("identity/reference %#v", session)
	}
	if len(session.Activities) != 1 || session.Activities[0].Section.ID != "access" ||
		session.Activities[0].Section.Summary != "Request access" || session.Activities[0].CurrentHash != accessHash || len(session.Activities[0].WaitingOn) != 0 {
		t.Fatalf("actual current activity %#v", session.Activities)
	}
	if len(session.Outcomes) != 2 || session.Outcomes[0].Outcome.Section != "a-removed" || session.Outcomes[1].Outcome.Section != "z-removed" {
		t.Fatalf("history order %#v", session.Outcomes)
	}
	for _, row := range session.Outcomes {
		if row.CurrentHash != "" || row.Comparison != "unknown" {
			t.Fatalf("removed history %#v", row)
		}
	}
	if !session.Support.Available || session.Support.Recipient.Name != "Example newcomer" || len(session.Support.Arrangements) != 0 || len(session.Support.Problems) != 0 {
		t.Fatalf("support %#v", session.Support)
	}
	if !reflect.DeepEqual(input, original) {
		t.Fatal("Inspect mutated source")
	}
}

func TestEqualPreseedOwnershipAndFollowPlan(t *testing.T) {
	input := source(answerPlan)
	intent := contract.Intent{Operation: contract.OperationAnswer, Question: "personal.choice", Value: text("org-a")}
	reviewer := newReviewer()
	proposal, diagnostics := reviewer.Propose(input, intent)
	success(t, diagnostics)
	want := heldRecord()
	want.Answers.Sections["personal"] = map[string]model.Value{"choice": text("org-a")}
	sameRecord(t, proposal.BeforeRecord, heldRecord())
	sameRecord(t, proposal.AfterRecord, want)
	if proposal.NeedsEvent || !reflect.DeepEqual(proposal.Source, input) || !reflect.DeepEqual(proposal.Intent, intent) {
		t.Fatalf("proposal evidence %#v", proposal)
	}
	publication, diagnostics := reviewer.Finalize(proposal, "")
	success(t, diagnostics)
	if !publication.Changed || publication.Event != "" || !bytes.Equal(publication.Record, []byte(recordOwned)) {
		t.Fatalf("publication %#v\n%s", publication, publication.Record)
	}
	planB := strings.Replace(answerPlan, "choice: org-a", "choice: org-b", 1)
	follower := source(planB)
	view, diagnostics := reviewer.Inspect(follower)
	success(t, diagnostics)
	if !reflect.DeepEqual(question(t, view, "personal.choice").Answer, text("org-b")) {
		t.Fatal("untouched follower did not use current shared answer")
	}
	owned := source(planB)
	owned.Record = []byte(recordOwned)
	view, diagnostics = reviewer.Inspect(owned)
	success(t, diagnostics)
	if !reflect.DeepEqual(question(t, view, "personal.choice").Answer, text("org-a")) {
		t.Fatal("explicit equal override did not remain owned")
	}
	proposal, diagnostics = reviewer.Propose(owned, contract.Intent{Operation: contract.OperationFollowPlan, Question: "personal.choice"})
	success(t, diagnostics)
	sameRecord(t, proposal.AfterRecord, heldRecord())
	if !reflect.DeepEqual(question(t, proposal.After, "personal.choice").Answer, text("org-b")) {
		t.Fatal("follow-plan did not reveal current shared answer")
	}
	publication, diagnostics = reviewer.Finalize(proposal, "")
	success(t, diagnostics)
	if !publication.Changed || publication.Event != "" || !bytes.Equal(publication.Record, []byte(recordFollower)) {
		t.Fatalf("follow-plan publication %#v\n%s", publication, publication.Record)
	}
}

func TestSuggestionsAndRawNoOp(t *testing.T) {
	planD := strings.Replace(answerPlan, "answers:\n  sections:\n    personal:\n      choice: org-a\n", "", 1)
	input := source(planD)
	input.Record = []byte("# deliberately retained formatting\n" + recordBefore)
	reviewer := newReviewer()
	session, diagnostics := reviewer.Inspect(input)
	success(t, diagnostics)
	q := question(t, session, "personal.choice")
	if !reflect.DeepEqual(q.Answer, model.Value{}) || !reflect.DeepEqual(q.DeclaredDefault, text("default-example")) || q.Example != "example-only" {
		t.Fatalf("suggestion was accepted or lost: %#v", q)
	}
	proposal, diagnostics := reviewer.Propose(input, contract.Intent{Operation: contract.OperationFollowPlan, Question: "personal.choice"})
	success(t, diagnostics)
	publication, diagnostics := reviewer.Finalize(proposal, "")
	success(t, diagnostics)
	if publication.Changed || publication.Event != "" || !bytes.Equal(publication.Record, input.Record) {
		t.Fatalf("no-op canonicalized or invented data: %#v", publication)
	}
	proposal, diagnostics = reviewer.Propose(input, contract.Intent{Operation: contract.OperationAnswer, Question: "personal.choice", Value: text("default-example")})
	success(t, diagnostics)
	if !reflect.DeepEqual(proposal.AfterRecord.Answers.Sections["personal"]["choice"], text("default-example")) {
		t.Fatal("explicit suggestion acceptance was not stored")
	}
}

func TestKnownEmptyCanonicalizationPreservesUnknownEmptyData(t *testing.T) {
	input := source(answerPlan)
	input.Record = []byte(`schemaVersion: 1
kind: record
account: newcomer
plan:
  id: plan-one
answers:
  plan: {}
  sections:
    empty-scope: {}
future:
  map: {}
  list: []
`)
	proposal, diagnostics := newReviewer().Propose(input, contract.Intent{Operation: contract.OperationAnswer, Question: "personal.choice", Value: text("org-a")})
	success(t, diagnostics)
	if proposal.AfterRecord.Answers.Plan == nil || proposal.AfterRecord.Answers.Sections["empty-scope"] == nil {
		t.Fatal("pure writer shape was normalized before publication")
	}
	publication, diagnostics := newReviewer().Finalize(proposal, "")
	success(t, diagnostics)
	want := `schemaVersion: 1
kind: record
account: newcomer
plan:
  id: plan-one
answers:
  sections:
    personal:
      choice: org-a
future:
  list: []
  map: {}
`
	if !publication.Changed || !bytes.Equal(publication.Record, []byte(want)) {
		t.Fatalf("known/unknown empty canonical output %#v\n%s", publication, publication.Record)
	}
}
