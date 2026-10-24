package reviewer_test

import (
	"bytes"
	"reflect"
	"testing"

	contract "onboard-kit/contracts/invitation/reviewer"
	"onboard-kit/internal/model"
)

func TestGapBeginComparesEveryWarningAndArrangementChannel(t *testing.T) {
	r := newReviewer()
	base := proposal(t, r, contract.Local{})
	baseBefore := copied(base)
	if len(base.Review.PreparationProblems) != 0 || len(base.Review.IncomingWarnings) != 0 || len(base.Review.EffectivePlanWarnings) != 0 || len(base.Review.Support.Arrangements) != 0 {
		t.Fatal("literal baseline must have empty warning and arrangement channels")
	}
	cases := []struct {
		name  string
		alter func(*contract.Proposal)
	}{
		{"preparation-problems", func(v *contract.Proposal) {
			v.Review.PreparationProblems = []contract.Diagnostic{{
				Stage:   "review.preparation-validation",
				Display: "Ask your onboarding owner to repair unavailable local support.",
				Problem: model.Problem{Code: "preparation.unknown-team", Severity: model.SeverityError, Message: "Untrusted cached diagnostic.", Detail: map[string]string{"path": "recipient.team"}},
			}}
		}},
		{"incoming-warnings", func(v *contract.Proposal) {
			v.Review.IncomingWarnings = []contract.Diagnostic{{
				Stage:   "inspect.plan-authoring",
				Display: "Review the invitation plan declarations with its owner.",
				Problem: model.Problem{Code: "authoring.condition-empty", Severity: model.SeverityWarning, Message: "Untrusted cached diagnostic.", Detail: map[string]string{"path": "sections[0].includeIf"}},
			}}
		}},
		{"effective-plan-warnings", func(v *contract.Proposal) {
			v.Review.EffectivePlanWarnings = []contract.Diagnostic{{
				Stage:   "review.effective-validation",
				Display: "Review stored personal answers against the incoming plan.",
				Problem: model.Problem{Code: "answer.unknown-question", Severity: model.SeverityWarning, Message: "Untrusted cached diagnostic.", Detail: map[string]string{"path": "answers.plan"}},
			}}
		}},
		{"support-arrangements", func(v *contract.Proposal) {
			v.Review.Support.Arrangements = []model.Arrangement{{ID: "invented", Section: "setup", Kind: "support", Status: "pending", Note: "Untrusted cached support."}}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for _, event := range []model.Timestamp{beginAt, "invalid-event"} {
				value := copied(base)
				c.alter(&value)
				before := copied(value)
				got, problems := newReviewer().Begin(value, event)
				ownProblem(t, problems, "begin.proposal", "invitation.proposal", "proposal", "Use the complete proposal derived from its exact source evidence.")
				if !reflect.DeepEqual(got, contract.Publication{}) || !reflect.DeepEqual(value, before) || !reflect.DeepEqual(base, baseBefore) {
					t.Error("a forged warning/support projection must refuse before the event and preserve all proposal storage")
				}
			}
		})
	}
	got, problems := newReviewer().Begin(copied(base), beginAt)
	if len(problems) != 0 || len(got.Artifacts) != 3 || !bytes.Equal(got.Artifacts[2].Bytes, fixture(t, "begun.record.yaml")) || !reflect.DeepEqual(got.OriginalReview, base.Review) {
		t.Error("unchanged complete proposal must remain accepted on a fresh reviewer")
	}
}

func TestGapWalkerPreservesWaitingAndEscalationOnEveryActivity(t *testing.T) {
	plan := []byte(`schemaVersion: 1
id: plan-a
directory:
  contacts:
    - id: helper
      name: Support
      slack: "#help"
answers:
  sections:
    access:
      helper: helper
    deploy:
      helper: helper
sections:
  - id: deploy
    title: Deploy
    summary: Deploy the application.
    requires:
      - access
    escalation:
      - $helper
      - $helper
    variables:
      - name: helper
        type: contact
  - id: access
    title: Access
    summary: Obtain access.
    escalation:
      - $helper
    variables:
      - name: helper
        type: contact
`)
	r := newReviewer()
	in := inspectPlan(t, r, plan)
	before := copied(in)
	result, problems := r.Review(in, contract.Local{}, contract.PlanChoiceUseIncoming)
	if len(problems) != 0 || result.PlanChoiceNeeded || result.Proposal == nil {
		t.Fatalf("valid prerequisite and support declarations must produce a complete session: diagnostics=%d", len(problems))
	}
	view := result.Proposal.Review
	if len(view.Artifacts) != 3 {
		t.Fatal("complete review must preserve all three fixed artifact roles")
	}
	if len(view.Session.Activities) != 2 {
		t.Fatal("waiting work must remain in the complete activity view")
	}
	access := view.Session.Activities[0]
	deploy := view.Session.Activities[1]
	if access.Section.ID != "access" || deploy.Section.ID != "deploy" {
		t.Error("the actual prerequisite order must override reversed declaration order")
	}
	if len(access.WaitingOn) != 0 || !reflect.DeepEqual(deploy.WaitingOn, []string{"access"}) || !reflect.DeepEqual(deploy.Section.Requires, []string{"access"}) {
		t.Error("absent verification must leave deploy waiting directly on access without hiding or rewriting Requires")
	}
	if !reflect.DeepEqual(access.Escalation, []string{"Support, #help"}) || !reflect.DeepEqual(deploy.Escalation, []string{"Support, #help", "Support, #help"}) {
		t.Error("rendered support must appear on ready and waiting rows, preserving ordinary repeats")
	}
	if !reflect.DeepEqual(access.Section.Escalation, []string{"$helper"}) || !reflect.DeepEqual(deploy.Section.Escalation, []string{"$helper", "$helper"}) {
		t.Error("rendered escalation must not replace the held section references")
	}
	if len(view.Session.Outcomes) != 0 || len(view.Session.Blockers) != 0 || len(view.SessionProblems) != 0 || len(view.Session.Questions) != 2 {
		t.Fatal("support and dependency declarations must not manufacture verification, missing answers or hidden questions")
	}
	for _, question := range view.Session.Questions {
		if !reflect.DeepEqual(question.Answer, model.Value{Kind: model.ValueKindText, Text: "helper"}) {
			t.Error("organizational answers must remain their held IDs rather than rendered contact text")
		}
	}
	if !reflect.DeepEqual(in, before) || !bytes.Equal(result.Proposal.Evidence.Incoming.Files.Plan, plan) || !bytes.Equal(result.Proposal.Evidence.Incoming.Files.Record, in.Files.Record) || !view.NeedsFirstOpenedAt || view.Artifacts[2].Action != contract.ActionCreate || view.Artifacts[2].Bytes != nil {
		t.Error("inspection must preserve source answers and the empty seed without fabricated record writes")
	}
}
