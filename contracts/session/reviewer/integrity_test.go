package reviewer_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/session/reviewer"
	"onboard-kit/internal/model"
)

func TestFinalizeRederivesEverySuppliedProjection(t *testing.T) {
	for _, c := range []struct {
		name   string
		mutate func(*contract.Proposal)
	}{
		{"needs-event", func(p *contract.Proposal) { p.NeedsEvent = true }},
		{"before-record", func(p *contract.Proposal) { p.BeforeRecord.Account = "other" }},
		{"after-record", func(p *contract.Proposal) { p.AfterRecord.Answers.Sections["personal"]["choice"] = text("forged") }},
		{"before-view", func(p *contract.Proposal) { p.Before.Reference.Hash = "forged" }},
		{"after-view", func(p *contract.Proposal) { p.After.Activities[0].Section.Summary = "forged" }},
		{"intent-only", func(p *contract.Proposal) { p.Intent.Value = text("different") }},
		{"source-only", func(p *contract.Proposal) {
			p.Source.Record = []byte(strings.Replace(string(p.Source.Record), "firstOpenedAt: held-uninterpreted-event", "firstOpenedAt: different-held-event", 1))
		}},
		{"wrong-shape", func(p *contract.Proposal) { p.AfterRecord.Outcomes = make([]model.Outcome, 4097) }},
		{"cyclic-projection", func(p *contract.Proposal) {
			cycle := map[string]any{}
			cycle["self"] = cycle
			p.AfterRecord.UnknownFields["future"] = cycle
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			proposal, diagnostics := newReviewer().Propose(source(answerPlan), contract.Intent{Operation: contract.OperationAnswer, Question: "personal.choice", Value: text("org-a")})
			success(t, diagnostics)
			if proposal.Before.Account != "newcomer" || len(proposal.After.Activities) != 1 {
				t.Fatal("missing positive proposal before tamper")
			}
			sameRecord(t, proposal.BeforeRecord, heldRecord())
			wantAfter := heldRecord()
			wantAfter.Answers.Sections["personal"] = map[string]model.Value{"choice": text("org-a")}
			sameRecord(t, proposal.AfterRecord, wantAfter)
			c.mutate(&proposal)
			got, diagnostics := newReviewer().Finalize(proposal, "")
			if !reflect.DeepEqual(got, contract.Publication{}) {
				t.Fatalf("forged projection produced %#v", got)
			}
			requireOwnError(t, diagnostics, "proposal", "session.proposal", "Review this change again.", "")
		})
	}
}

func TestAnswerFinalizeRejectsEventAndPublicationBytesAreOwned(t *testing.T) {
	reviewer := newReviewer()
	proposal, diagnostics := reviewer.Propose(source(answerPlan), contract.Intent{Operation: contract.OperationAnswer, Question: "personal.choice", Value: text("org-a")})
	success(t, diagnostics)
	got, diagnostics := reviewer.Finalize(proposal, event)
	if !reflect.DeepEqual(got, contract.Publication{}) {
		t.Fatal("answer accepted an outcome event")
	}
	requireOwnError(t, diagnostics, "event", "session.event", "An answer change does not use an event timestamp.", "")
	first, diagnostics := reviewer.Finalize(proposal, "")
	success(t, diagnostics)
	if !bytes.Equal(first.Record, []byte(recordOwned)) {
		t.Fatalf("independent literal output\n%s", first.Record)
	}
	first.Record[0] = '!'
	second, diagnostics := reviewer.Finalize(proposal, "")
	success(t, diagnostics)
	if !bytes.Equal(second.Record, []byte(recordOwned)) || !bytes.Equal(proposal.Source.Record, []byte(recordBefore)) {
		t.Fatal("publication aliases source or a later result")
	}
}

const commandPlan = `schemaVersion: 1
id: plan-one
answers:
  sections:
    personal:
      command:
        - go
        - test
sections:
  - id: personal
    title: Personal command
    summary: Use $command.
    variables:
      - name: command
        type: command
        default:
          - suggested
          - run
        audience: onboardee
`

func TestSourceIntentQuestionsAndUnknownTreesAreIndependentlyOwned(t *testing.T) {
	input := source(commandPlan)
	original := cloneSource(input)
	intent := contract.Intent{Operation: contract.OperationAnswer, Question: "personal.command", Value: model.Value{Kind: model.ValueKindList, List: []string{"echo", "chosen"}}}
	reviewer := newReviewer()
	proposal, diagnostics := reviewer.Propose(input, intent)
	success(t, diagnostics)
	if !reflect.DeepEqual(proposal.AfterRecord.Answers.Sections["personal"]["command"].List, []string{"echo", "chosen"}) {
		t.Fatal("missing independent list answer")
	}
	want := heldRecord()
	want.Answers.Sections["personal"] = map[string]model.Value{"command": {Kind: model.ValueKindList, List: []string{"echo", "chosen"}}}
	sameRecord(t, proposal.AfterRecord, want)
	if !reflect.DeepEqual(proposal.Source, original) || !reflect.DeepEqual(proposal.Intent, intent) {
		t.Fatal("missing source/intent evidence before mutation")
	}
	second, diagnostics := reviewer.Propose(input, intent)
	success(t, diagnostics)
	q := question(t, proposal.After, "personal.command")
	if !reflect.DeepEqual(q.Answer.List, []string{"echo", "chosen"}) || !reflect.DeepEqual(q.DeclaredDefault.List, []string{"suggested", "run"}) {
		t.Fatalf("missing nested question data before mutation %#v", q)
	}
	q.Answer.List[0] = "changed answer"
	q.DeclaredDefault.List[0] = "changed default"
	proposal.AfterRecord.UnknownFields["future"].(map[string]any)["items"].([]any)[1].(map[string]any)["nested"] = false
	proposal.Source.Plan[0] = '!'
	proposal.Intent.Value.List[0] = "changed intent"
	proposal.BeforeRecord.Outcomes[0].Note = "changed record"
	if !reflect.DeepEqual(input, original) || !reflect.DeepEqual(intent.Value.List, []string{"echo", "chosen"}) {
		t.Fatal("proposal result changed source or intent")
	}
	q = question(t, second.After, "personal.command")
	if !reflect.DeepEqual(q.Answer.List, []string{"echo", "chosen"}) || !reflect.DeepEqual(q.DeclaredDefault.List, []string{"suggested", "run"}) ||
		!reflect.DeepEqual(second.AfterRecord.UnknownFields, heldRecord().UnknownFields) || !bytes.Equal(second.Source.Plan, []byte(commandPlan)) {
		t.Fatal("another result shares mutable question/source/unknown storage")
	}
	input.Plan[0] = '!'
	input.Record[0] = '!'
	intent.Value.List[0] = "input changed"
	if !bytes.Equal(second.Source.Plan, []byte(commandPlan)) || !bytes.Equal(second.Source.Record, []byte(recordBefore)) ||
		!reflect.DeepEqual(second.Intent.Value.List, []string{"echo", "chosen"}) {
		t.Fatal("input changes reached prior proposal")
	}
}

func TestDiagnosticOwnershipAndRawReferenceHashFact(t *testing.T) {
	reviewer := newReviewer()
	bad := source(accessPlan)
	bad.Account = ""
	_, first := reviewer.Inspect(bad)
	requireOwnError(t, first, "context", "session.context", "Select an explicit valid account and plan identity.", "context")
	first[0].Problem.Detail["path"] = "changed"
	_, second := reviewer.Inspect(bad)
	requireOwnError(t, second, "context", "session.context", "Select an explicit valid account and plan identity.", "context")
	input := source(accessPlan)
	sum := sha256.Sum256([]byte(accessPlan))
	digest := hex.EncodeToString(sum[:])
	input.Record = []byte(strings.Replace(recordBefore, "hash: old-plan-hash", "hash: \""+digest+"\"", 1))
	session, diagnostics := reviewer.Inspect(input)
	success(t, diagnostics)
	if !session.PlanReferenceMatches || session.Reference.Hash != digest || session.Reference.URL != "plans/shared.obk.yaml" {
		t.Fatalf("raw plan digest fact/reference %#v", session)
	}
}
