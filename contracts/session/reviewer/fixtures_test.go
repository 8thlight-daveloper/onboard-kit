package reviewer_test

import (
	"reflect"
	"testing"

	contract "onboard-kit/contracts/session/reviewer"
	"onboard-kit/internal/model"
)

const accessPlan = `schemaVersion: 1
id: plan-one
sections:
  - id: access
    title: Access
    summary: Request access
`

const answerPlan = `schemaVersion: 1
id: plan-one
answers:
  sections:
    personal:
      choice: org-a
sections:
  - id: personal
    title: Personal choices
    summary: Use $choice.
    variables:
      - name: choice
        description: Your choice
        type: string
        default: default-example
        example: example-only
        audience: onboardee
`

const preparation = `schemaVersion: 1
kind: preparation
account: newcomer
planId: plan-one
recipient:
  name: Example newcomer
`

const recordBefore = `schemaVersion: 1
kind: record
account: newcomer
startDate: legacy-date
firstOpenedAt: held-uninterpreted-event
plan:
  id: plan-one
  hash: old-plan-hash
  url: plans/shared.obk.yaml
answers:
  sections:
    gone:
      hidden: deferred
outcomes:
  - section: z-removed
    outcome: verified
    on: legacy-on
    sectionHash: old
    note: Keep removed
    differently: Keep this
  - section: a-removed
    outcome: deferred
    on: older
    note: Keep other
future:
  items:
    - held
    - nested: true
  emptyList: []
  emptyMap: {}
`

const recordOwned = `schemaVersion: 1
kind: record
account: newcomer
startDate: legacy-date
firstOpenedAt: held-uninterpreted-event
plan:
  id: plan-one
  hash: old-plan-hash
  url: plans/shared.obk.yaml
answers:
  sections:
    gone:
      hidden: deferred
    personal:
      choice: org-a
outcomes:
  - section: z-removed
    outcome: verified
    on: legacy-on
    sectionHash: old
    note: Keep removed
    differently: Keep this
  - section: a-removed
    outcome: deferred
    on: older
    note: Keep other
future:
  emptyList: []
  emptyMap: {}
  items:
    - held
    - nested: true
`

const recordFollower = `schemaVersion: 1
kind: record
account: newcomer
startDate: legacy-date
firstOpenedAt: held-uninterpreted-event
plan:
  id: plan-one
  hash: old-plan-hash
  url: plans/shared.obk.yaml
answers:
  sections:
    gone:
      hidden: deferred
outcomes:
  - section: z-removed
    outcome: verified
    on: legacy-on
    sectionHash: old
    note: Keep removed
    differently: Keep this
  - section: a-removed
    outcome: deferred
    on: older
    note: Keep other
future:
  emptyList: []
  emptyMap: {}
  items:
    - held
    - nested: true
`

const accessHash = "6c0301d995d0a3b0dab95a641a087120c5266746e59c50d934c575e135ca2fdc"
const accessRequiresHash = "7db6dbae189ea1177381c325d068dc9ec49fa17a61a6d2e2858db04acefd1978"
const event = model.Timestamp("2026-10-25T10:11:12.3400+00:00")

const recordVerified = `schemaVersion: 1
kind: record
account: newcomer
startDate: legacy-date
firstOpenedAt: held-uninterpreted-event
plan:
  id: plan-one
  hash: old-plan-hash
  url: plans/shared.obk.yaml
answers:
  sections:
    gone:
      hidden: deferred
outcomes:
  - section: a-removed
    outcome: deferred
    on: older
    note: Keep other
  - section: access
    outcome: verified
    on: "2026-10-25T10:11:12.3400+00:00"
    sectionHash: "6c0301d995d0a3b0dab95a641a087120c5266746e59c50d934c575e135ca2fdc"
  - section: z-removed
    outcome: verified
    on: legacy-on
    sectionHash: old
    note: Keep removed
    differently: Keep this
future:
  emptyList: []
  emptyMap: {}
  items:
    - held
    - nested: true
`

func source(plan string) contract.Source {
	return contract.Source{Plan: []byte(plan), Preparation: []byte(preparation), Record: []byte(recordBefore), Account: "newcomer", PlanID: "plan-one"}
}

func cloneSource(input contract.Source) contract.Source {
	return contract.Source{
		Plan: append([]byte(nil), input.Plan...), Preparation: append([]byte(nil), input.Preparation...),
		Record: append([]byte(nil), input.Record...), Account: input.Account, PlanID: input.PlanID,
	}
}

func text(value string) model.Value {
	return model.Value{Kind: model.ValueKindText, Text: value}
}

func heldRecord() model.Record {
	return model.Record{
		SchemaVersion: 1, Account: "newcomer", StartDate: "legacy-date", FirstOpenedAt: "held-uninterpreted-event",
		Plan:    model.PlanReference{ID: "plan-one", Hash: "old-plan-hash", URL: "plans/shared.obk.yaml"},
		Answers: model.Answers{Sections: map[string]map[string]model.Value{"gone": {"hidden": {Kind: model.ValueKindDeferred}}}},
		Outcomes: []model.Outcome{
			{Section: "z-removed", Kind: model.OutcomeVerified, On: "legacy-on", SectionHash: "old", Note: "Keep removed", Differently: "Keep this"},
			{Section: "a-removed", Kind: model.OutcomeDeferred, On: "older", Note: "Keep other"},
		},
		UnknownFields: map[string]any{"future": map[string]any{
			"items": []any{"held", map[string]any{"nested": true}}, "emptyList": []any{}, "emptyMap": map[string]any{},
		}},
	}
}

func success(t *testing.T, diagnostics []contract.Diagnostic) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Stage == "" || diagnostic.Problem.Message == "" || diagnostic.Problem.Severity != model.SeverityWarning {
			t.Fatalf("unexpected diagnostic: %#v", diagnostics)
		}
	}
}

func sameRecord(t *testing.T, got, want model.Record) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("record mismatch\ngot %#v\nwant %#v", got, want)
	}
}

func question(t *testing.T, session contract.Session, id string) model.Question {
	t.Helper()
	for _, item := range session.Questions {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("missing question %q: %#v", id, session.Questions)
	return model.Question{}
}

func outcome(t *testing.T, session contract.Session, id string) contract.HeldOutcome {
	t.Helper()
	for _, item := range session.Outcomes {
		if item.Outcome.Section == id {
			return item
		}
	}
	t.Fatalf("missing outcome %q: %#v", id, session.Outcomes)
	return contract.HeldOutcome{}
}
