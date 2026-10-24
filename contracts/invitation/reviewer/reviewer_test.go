package reviewer_test

import (
	"archive/zip"
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/invitation/reviewer"
	"onboard-kit/internal/model"
)

const planHash = "00d6473bd698e40caca271f61af7b87288f0c864924008bd24375d6b19b3c6a4"
const sectionHash = "5631a3d34e963b37019b9ae53b12907c3bdef3d17c6642f1462784d178c3f4ad"
const beginAt model.Timestamp = "2026-10-23T12:34:56.123Z"

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func archive(t *testing.T, changes map[string][]byte) []byte {
	t.Helper()
	roles := []struct{ path, fixture string }{
		{"plan.obk.yaml", "minimal.plan.yaml"},
		{"records/jules/preparation.obk.yaml", "minimal.preparation.yaml"},
		{"records/jules/record.obk.yaml", "minimal.seed.yaml"},
		{"first-steps.md", "first-steps.md"},
	}
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, role := range roles {
		b, ok := changes[role.path]
		if !ok {
			b = fixture(t, role.fixture)
		}
		entry, err := w.Create(role.path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(b); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func incoming(t *testing.T, r contract.Reviewer) contract.Incoming {
	t.Helper()
	in, problems := r.Inspect(archive(t, nil))
	if len(problems) != 0 || in.PlanID != "plan-a" {
		t.Fatalf("baseline Inspect: %d diagnostics, identity %q", len(problems), in.PlanID)
	}
	return in
}

func proposal(t *testing.T, r contract.Reviewer, local contract.Local) contract.Proposal {
	t.Helper()
	result, problems := r.Review(incoming(t, r), local, contract.PlanChoiceUseIncoming)
	if len(problems) != 0 || result.PlanChoiceNeeded || result.Proposal == nil {
		t.Fatalf("baseline Review: %d diagnostics, choice=%v, proposal=%v", len(problems), result.PlanChoiceNeeded, result.Proposal != nil)
	}
	return *result.Proposal
}

func ownProblem(t *testing.T, got []contract.Diagnostic, stage, code, path, message string) {
	t.Helper()
	want := []contract.Diagnostic{{Stage: stage, Display: message, Problem: model.Problem{
		Code: code, Severity: model.SeverityError, Message: message, Detail: map[string]string{"path": path},
	}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("diagnostic differs: count=%d expected %s/%s at %s", len(got), stage, code, path)
	}
}

func TestLiteralNewSession(t *testing.T) {
	r := newReviewer()
	in := incoming(t, r)
	if in.Reference != (model.PlanReference{ID: "plan-a", Hash: planHash}) || in.Files.Account != "jules" {
		t.Fatal("exact raw plan identity/reference lost")
	}
	p := proposal(t, r, contract.Local{})
	v := p.Review
	if v.Account != "jules" || v.PlanID != "plan-a" || !v.NeedsFirstOpenedAt || v.BindsUnboundRecord || !v.Support.Available {
		t.Error("new-session identity/event/support facts differ")
	}
	if len(v.Artifacts) != 3 {
		t.Fatalf("artifact count %d", len(v.Artifacts))
	}
	wantArtifacts := []contract.Artifact{
		{Role: contract.RolePlan, Action: contract.ActionInstall, Bytes: fixture(t, "minimal.plan.yaml")},
		{Role: contract.RolePreparation, Action: contract.ActionInstall, Bytes: fixture(t, "minimal.preparation.yaml")},
		{Role: contract.RoleRecord, Action: contract.ActionCreate},
	}
	if !reflect.DeepEqual(v.Artifacts, wantArtifacts) {
		t.Error("new-session ordered literal artifacts differ")
	}
	if len(v.Session.Activities) != 1 || len(v.Session.Questions) != 1 || len(v.Session.Blockers) != 1 {
		t.Fatal("complete record-backed walker session missing")
	}
	a := v.Session.Activities[0]
	if a.CurrentHash != sectionHash || a.Section.ID != "setup" || a.Section.Summary != "Use <account_name>." || len(a.WaitingOn) != 0 || len(a.Escalation) != 0 {
		t.Error("actual activity/hash/rendered context differs")
	}
	q := v.Session.Questions[0]
	if q.ID != "setup.account_name" || q.Scope != "setup" || q.Name != "account_name" || q.Description != "Which account?" || string(q.Type) != "string" || string(q.Audience) != "onboardee" || !q.Required || q.Group == "" || q.Answer.Kind != model.ValueKindNone || q.DeclaredDefault.Kind != model.ValueKindNone {
		t.Error("actual question projection differs")
	}
	if v.Session.Blockers[0].Question != q.ID || len(v.SessionProblems) != 1 {
		t.Fatal("required answer blocker/problem missing")
	}
	d := v.SessionProblems[0]
	if d.Stage != "review.walker" || d.Display != "Resolve the session's activity or history problem before opening." || d.Problem.Code != "gap.unanswered" || string(d.Problem.Severity) != "warning" || d.Problem.Question != q.ID || d.Problem.Section != "setup" || d.Problem.Message == "" || d.Problem.Detail["audience"] != "onboardee" {
		t.Error("walker diagnostic provenance or data lost")
	}
	pub, problems := r.Begin(p, beginAt)
	if len(problems) != 0 || pub.FirstOpenedAt != beginAt || len(pub.Artifacts) != 3 {
		t.Fatalf("Begin: diagnostics=%d artifact count=%d", len(problems), len(pub.Artifacts))
	}
	if !bytes.Equal(pub.Artifacts[2].Bytes, fixture(t, "begun.record.yaml")) || !reflect.DeepEqual(pub.OriginalReview, p.Review) {
		t.Error("literal first event bytes or original review differ")
	}
}

func TestChoiceAndPhysicalObservationPriority(t *testing.T) {
	r := newReviewer()
	in := incoming(t, r)
	local := contract.Local{Plan: contract.Body{Present: true, Bytes: []byte("not YAML: [")}}
	got, problems := r.Review(in, local, contract.PlanChoiceUndecided)
	if len(problems) != 0 || !got.PlanChoiceNeeded || got.Proposal != nil {
		t.Error("different local bytes must require a decision before local decoding")
	}
	got, problems = r.Review(in, local, contract.PlanChoiceUseIncoming)
	if len(problems) != 0 || got.Proposal == nil || got.Proposal.Review.Comparison.Available || got.Proposal.Review.Comparison.Reason != "decode" || len(got.Proposal.Review.Comparison.Problems) == 0 {
		t.Error("explicit replacement should retain unreadable previous comparison diagnostics")
	}
	_, problems = r.Review(in, contract.Local{KnownRecordMissing: true}, contract.PlanChoice("other"))
	ownProblem(t, problems, "review.choice", "invitation.plan-choice", "choice", "Choose undecided or use-incoming explicitly.")
	_, problems = r.Review(in, contract.Local{KnownRecordMissing: true}, contract.PlanChoiceUndecided)
	ownProblem(t, problems, "review.known-record", "invitation.known-record-missing", "local.record", "Restore the registered record instead of creating another seed.")
	_, problems = r.Review(in, contract.Local{Record: contract.Body{Bytes: []byte("private")}}, contract.PlanChoice("other"))
	ownProblem(t, problems, "review.local", "invitation.local", "local.record", "Supply consistent local presence observations.")
}

func TestRetainExactExistingBytesAndNoEvent(t *testing.T) {
	r := newReviewer()
	raw := fixture(t, "retained.record.yaml")
	p := proposal(t, r, contract.Local{Record: contract.Body{Present: true, Bytes: raw}})
	if p.Review.NeedsFirstOpenedAt || p.Review.Artifacts[2].Action != contract.ActionRetain || !bytes.Equal(p.Review.Artifacts[2].Bytes, raw) {
		t.Fatal("semantic no-op must retain exact commented source")
	}
	pub, problems := r.Begin(p, "")
	if len(problems) != 0 || pub.FirstOpenedAt != "" || !bytes.Equal(pub.Artifacts[2].Bytes, raw) {
		t.Error("existing unknown first-open time must remain unknown")
	}
	_, problems = r.Begin(p, beginAt)
	ownProblem(t, problems, "begin.event", "invitation.event", "event", "Supply one supported begin event only for a new record.")
}

func TestNoOpDoesNotEncodeExpandedUnknownString(t *testing.T) {
	r := newReviewer()
	seed := fixture(t, "minimal.seed.yaml")
	for _, changed := range []bool{false, true} {
		t.Run(map[bool]string{false: "retain", true: "reconcile"}[changed], func(t *testing.T) {
			raw := string(seed)
			if changed {
				raw = strings.Replace(raw, planHash, "old", 1)
			}
			raw += "x-blob: \"" + strings.Repeat(`\0`, 300000) + "\"\n"
			if len(raw) >= 1048576 || 4*300000 <= 1048576 {
				t.Fatal("independent escape expansion fixture does not straddle limits")
			}
			result, problems := r.Review(incoming(t, r), contract.Local{Record: contract.Body{Present: true, Bytes: []byte(raw)}}, contract.PlanChoiceUseIncoming)
			if changed {
				ownProblem(t, problems, "review.record-size", "invitation.output-size", "record", "Keep the encoded record within the supported byte limit.")
				if result.Proposal != nil {
					t.Error("oversized canonical changed record returned a proposal")
				}
				return
			}
			if len(problems) != 0 || result.Proposal == nil || result.Proposal.Review.Artifacts[2].Action != contract.ActionRetain || !bytes.Equal(result.Proposal.Review.Artifacts[2].Bytes, []byte(raw)) {
				t.Error("no-op must not introduce canonical encoder/size gate")
			}
		})
	}
}
