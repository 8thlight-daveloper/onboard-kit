package reviewer_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/invitation/reviewer"
	"onboard-kit/internal/model"
)

func inspectPlan(t *testing.T, r contract.Reviewer, plan []byte) contract.Incoming {
	t.Helper()
	hash := sha256.Sum256(plan)
	seed := bytes.Replace(fixture(t, "minimal.seed.yaml"), []byte(planHash), []byte(hex.EncodeToString(hash[:])), 1)
	in, problems := r.Inspect(archive(t, map[string][]byte{"plan.obk.yaml": plan, "records/jules/record.obk.yaml": seed}))
	if len(problems) != 0 {
		t.Fatalf("constructed plan inspection: %d diagnostics", len(problems))
	}
	return in
}

func TestStaleLocalSupportIsRetainedButUnavailable(t *testing.T) {
	r := newReviewer()
	raw := fixture(t, "stale.preparation.yaml")
	result, problems := r.Review(incoming(t, r), contract.Local{Preparation: contract.Body{Present: true, Bytes: raw}}, contract.PlanChoiceUseIncoming)
	if len(problems) != 0 || result.Proposal == nil {
		t.Fatal("well-formed stale organizational references must not block the personal session")
	}
	v := result.Proposal.Review
	if v.Support.Available || !reflect.DeepEqual(v.Support.Recipient, model.PreparationRecipient{}) || v.Support.Arrangements != nil || !bytes.Equal(v.Artifacts[1].Bytes, raw) || v.Artifacts[1].Action != contract.ActionRetain || len(v.Session.Activities) != 1 {
		t.Error("stale source was not held intact separately from usable session/support")
	}
	want := []struct{ code, path string }{
		{"preparation.unknown-team", "recipient.team"},
		{"preparation.unknown-contact", "recipient.manager"},
		{"preparation.unknown-section", "arrangements[0].section"},
		{"preparation.unknown-contact", "arrangements[0].contacts[0]"},
		{"preparation.unknown-contact", "arrangements[0].contacts[1]"},
	}
	if len(v.PreparationProblems) != len(want) {
		t.Fatalf("stale diagnostics count %d", len(v.PreparationProblems))
	}
	for i, expected := range want {
		d := v.PreparationProblems[i]
		if d.Stage != "review.preparation-validation" || d.Display != "Ask your onboarding owner to repair unavailable local support." || d.Problem.Code != expected.code || d.Problem.Detail["path"] != expected.path || d.Problem.Severity != model.SeverityError || d.Problem.Message == "" {
			t.Errorf("stale diagnostic %d provenance/data differs", i)
		}
	}
	for _, altered := range []string{
		strings.Replace(string(raw), "removed-team", "Removed-Team", 1),
		strings.Replace(string(raw), "removed-section", "plan", 1),
		strings.Replace(string(raw), "status: pending", "status: verified", 1),
	} {
		got, errors := r.Review(incoming(t, r), contract.Local{Preparation: contract.Body{Present: true, Bytes: []byte(altered)}}, contract.PlanChoiceUseIncoming)
		if got.Proposal != nil || len(errors) == 0 {
			t.Error("malformed or unrelated preparation errors must not enter stale allowance")
		}
		for _, d := range errors {
			if d.Stage != "review.preparation-validation" || d.Problem.Severity != model.SeverityError {
				t.Error("fatal validation phase must retain its entire original diagnostic channel")
			}
		}
	}
	_, problems = r.Inspect(archive(t, map[string][]byte{"records/jules/preparation.obk.yaml": raw}))
	if len(problems) != 5 || problems[0].Stage != "inspect.preparation-validation" {
		t.Error("incoming preparation must not receive the retained-local stale exception")
	}
}

func TestCurrentOnboarderKeysRefuseBeforeJoin(t *testing.T) {
	base := string(fixture(t, "minimal.plan.yaml"))
	for _, audience := range []string{"", "    audience: onboarder\n"} {
		plan := []byte(base + "variables:\n  - name: org_url\n    description: Organization address\n" + audience)
		r := newReviewer()
		in := inspectPlan(t, r, plan)
		for _, value := range []string{"false", "deferred", "\"\""} {
			raw := string(in.Files.Record) + "answers:\n  plan:\n    org_url: " + value + "\n"
			result, problems := r.Review(in, contract.Local{Record: contract.Body{Present: true, Bytes: []byte(raw)}}, contract.PlanChoiceUseIncoming)
			ownProblem(t, problems, "review.record-ownership", "invitation.record-owned-answer", "answers.plan[0]", "Keep onboarder-owned answers out of the personal record.")
			if result.Proposal != nil {
				t.Error("joiner must never receive an unsupported organizational record key")
			}
		}
	}
}

func TestUnboundRecordUsesActualBindAndPreservesLegacyData(t *testing.T) {
	r := newReviewer()
	raw := []byte("schemaVersion: 1\nkind: record\naccount: jules\nstartDate: long ago\nnotes: held\n")
	p := proposal(t, r, contract.Local{Record: contract.Body{Present: true, Bytes: raw}})
	if p.Review.Artifacts[2].Action != contract.ActionBind || !p.Review.BindsUnboundRecord || p.Review.NeedsFirstOpenedAt {
		t.Fatal("unbound existing record requires explicit derived binding, not first-use reseeding")
	}
	want := []byte("schemaVersion: 1\nkind: record\naccount: jules\nstartDate: long ago\nplan:\n  id: plan-a\n  hash: \"" + planHash + "\"\nnotes: held\n")
	if !bytes.Equal(p.Review.Artifacts[2].Bytes, want) {
		t.Error("handwritten bound record bytes differ")
	}
	pub, errors := r.Begin(p, "")
	if len(errors) != 0 || !bytes.Equal(pub.Artifacts[2].Bytes, want) || pub.FirstOpenedAt != "" {
		t.Error("binding must not invent first use or reinterpret legacy startDate")
	}
}

func TestChoicesReconciliationAndDeepQuestionOwnership(t *testing.T) {
	r := newReviewer()
	in := inspectPlan(t, r, fixture(t, "choices.plan.yaml"))
	if in.Reference.Hash != "cb1f0be905bb0c1d742c2dea3c314786d5cf7f3199bbadf89b55c742e2e6b787" {
		t.Fatal("independent raw choices-plan digest differs")
	}
	raw := string(in.Files.Record) + "answers:\n  sections:\n    setup:\n      flavors: red\n    removed:\n      hidden: old\n"
	local := contract.Local{Plan: contract.Body{Present: true, Bytes: in.Files.Plan}, Record: contract.Body{Present: true, Bytes: []byte(raw)}}
	result, problems := r.Review(in, local, contract.PlanChoiceUndecided)
	if len(problems) != 0 || result.Proposal == nil {
		t.Fatal("invalid current choice must be removed by actual reconciliation, hidden answer retained")
	}
	p := *result.Proposal
	if !reflect.DeepEqual(p.Review.Reconciliation.AnswersToReview, []string{"setup.flavors"}) || len(p.Review.Reconciliation.NewQuestions) != 0 || p.Review.Artifacts[2].Action != contract.ActionReconcile {
		t.Error("actual reconciliation lists/action differ")
	}
	if !bytes.Equal(p.Review.Artifacts[2].Bytes, fixture(t, "choices.reconciled.record.yaml")) {
		t.Error("literal reconciliation must remove invalid choice but preserve hidden scope")
	}
	if len(p.Review.Session.Questions) != 2 {
		t.Fatal("question projection missing")
	}
	q := p.Review.Session.Questions[0]
	if !reflect.DeepEqual(q.Options, []string{"blue", "green"}) || q.Multiple || q.Answer.Kind != model.ValueKindNone || !reflect.DeepEqual(q.DeclaredDefault, model.Value{Kind: model.ValueKindText, Text: "blue"}) || len(q.Suggestions) != 1 || !reflect.DeepEqual(q.Suggestions[0].Value, q.DeclaredDefault) {
		t.Fatal("actual finder default/list/suggestion projection differs")
	}
	command := p.Review.Session.Questions[1]
	if command.Name != "tools" || string(command.Type) != "command" || command.Answer.Kind != model.ValueKindNone || !reflect.DeepEqual(command.DeclaredDefault, model.Value{Kind: model.ValueKindList, List: []string{"echo", "hello"}}) || len(command.Suggestions) != 1 || !reflect.DeepEqual(command.Suggestions[0].Value, command.DeclaredDefault) {
		t.Fatal("actual command list default/suggestion projection differs")
	}
	before := copied(p)
	pub, errors := r.Begin(p, "")
	if len(errors) != 0 {
		t.Fatal("accepted choices proposal refused")
	}
	pub.OriginalReview.Session.Questions[0].Options[0] = "changed"
	pub.OriginalReview.Session.Questions[1].DeclaredDefault.List[0] = "changed"
	pub.OriginalReview.Session.Questions[1].Suggestions[0].Value.List[0] = "changed"
	pub.OriginalReview.Session.Activities[0].Section.Variables[1].Default.List[0] = "changed"
	custom := pub.OriginalReview.Session.Activities[0].Section.UnknownFields["custom"].(map[string]any)
	custom["rows"].([]any)[0] = "changed"
	if !reflect.DeepEqual(p, before) {
		t.Error("deep question/default/suggestion/unknown storage aliases the accepted proposal")
	}
	if command.DeclaredDefault.List[0] != "echo" || command.Suggestions[0].Value.List[0] != "echo" {
		t.Error("derived question list branches alias a different returned result")
	}
}
