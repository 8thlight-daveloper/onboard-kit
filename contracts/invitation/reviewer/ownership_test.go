package reviewer_test

import (
	"bytes"
	"reflect"
	"sync"
	"testing"

	contract "onboard-kit/contracts/invitation/reviewer"
)

func copied[T any](value T) T {
	return copyValue(reflect.ValueOf(value)).Interface().(T)
}

func copyValue(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		n := reflect.New(v.Type()).Elem()
		n.Set(copyValue(v.Elem()))
		return n
	case reflect.Pointer:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		n := reflect.New(v.Type().Elem())
		n.Elem().Set(copyValue(v.Elem()))
		return n
	case reflect.Slice:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		n := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			n.Index(i).Set(copyValue(v.Index(i)))
		}
		return n
	case reflect.Map:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		n := reflect.MakeMapWithSize(v.Type(), v.Len())
		iter := v.MapRange()
		for iter.Next() {
			n.SetMapIndex(iter.Key(), copyValue(iter.Value()))
		}
		return n
	case reflect.Struct:
		n := reflect.New(v.Type()).Elem()
		for i := 0; i < v.NumField(); i++ {
			n.Field(i).Set(copyValue(v.Field(i)))
		}
		return n
	default:
		return v
	}
}

func TestIncomingCannotAuthorizeForgedProjection(t *testing.T) {
	r := newReviewer()
	base := incoming(t, r)
	cases := []struct {
		name  string
		alter func(*contract.Incoming)
	}{
		{"account", func(v *contract.Incoming) { v.Files.Account = "other" }},
		{"plan-id", func(v *contract.Incoming) { v.PlanID = "other" }},
		{"reference", func(v *contract.Incoming) { v.Reference.Hash = "other" }},
		{"plan-payload", func(v *contract.Incoming) { v.Files.Plan = append(v.Files.Plan, '\n') }},
		{"preparation-payload", func(v *contract.Incoming) { v.Files.Preparation = nil }},
		{"record-payload", func(v *contract.Incoming) { v.Files.Record = nil }},
		{"guide-payload", func(v *contract.Incoming) { v.Files.FirstSteps = nil }},
		{"record-path", func(v *contract.Incoming) { v.Paths.Record = "elsewhere" }},
		{"preparation-path", func(v *contract.Incoming) { v.Paths.Preparation = "elsewhere" }},
		{"warnings", func(v *contract.Incoming) { v.Warnings = []contract.Diagnostic{{Stage: "invented"}} }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := copied(base)
			c.alter(&v)
			before := copied(v)
			got, problems := r.Review(v, contract.Local{}, contract.PlanChoiceUndecided)
			ownProblem(t, problems, "review.incoming", "invitation.incoming", "incoming", "Use an incoming projection consistent with its archive.")
			if !reflect.DeepEqual(got, contract.ReviewResult{}) || !reflect.DeepEqual(v, before) {
				t.Error("forged input changed or returned partial proposal")
			}
		})
	}
}

func TestBeginReDerivesAllCachedProjection(t *testing.T) {
	r := newReviewer()
	base := proposal(t, r, contract.Local{})
	cases := []struct {
		name  string
		alter func(*contract.Proposal)
	}{
		{"account", func(v *contract.Proposal) { v.Review.Account = "other" }},
		{"identity", func(v *contract.Proposal) { v.Review.PlanID = "other" }},
		{"paths", func(v *contract.Proposal) { v.Review.Paths.Record = "other" }},
		{"artifact-order", func(v *contract.Proposal) {
			v.Review.Artifacts[0], v.Review.Artifacts[1] = v.Review.Artifacts[1], v.Review.Artifacts[0]
		}},
		{"artifact-action", func(v *contract.Proposal) { v.Review.Artifacts[2].Action = contract.ActionRetain }},
		{"artifact-bytes", func(v *contract.Proposal) { v.Review.Artifacts[2].Bytes = []byte("forged") }},
		{"event-needed", func(v *contract.Proposal) { v.Review.NeedsFirstOpenedAt = false }},
		{"binding", func(v *contract.Proposal) { v.Review.BindsUnboundRecord = true }},
		{"comparison", func(v *contract.Proposal) { v.Review.Comparison.Available = true }},
		{"reconciliation", func(v *contract.Proposal) { v.Review.Reconciliation.RemovedSections = []string{"setup"} }},
		{"support", func(v *contract.Proposal) { v.Review.Support.Available = false }},
		{"recipient", func(v *contract.Proposal) { v.Review.Support.Recipient.Name = "forged" }},
		{"activities", func(v *contract.Proposal) { v.Review.Session.Activities = nil }},
		{"section", func(v *contract.Proposal) { v.Review.Session.Activities[0].Section.Summary = "forged" }},
		{"section-hash", func(v *contract.Proposal) { v.Review.Session.Activities[0].CurrentHash = "forged" }},
		{"question", func(v *contract.Proposal) { v.Review.Session.Questions[0].Required = false }},
		{"blocker", func(v *contract.Proposal) { v.Review.Session.Blockers = nil }},
		{"outcome", func(v *contract.Proposal) { v.Review.Session.Outcomes = []contract.Outcome{{Comparison: "same"}} }},
		{"raw-diagnostic", func(v *contract.Proposal) { v.Review.SessionProblems[0].Problem.Message = "forged" }},
		{"diagnostic-map", func(v *contract.Proposal) { v.Review.SessionProblems[0].Problem.Detail["audience"] = "forged" }},
		{"safe-display", func(v *contract.Proposal) { v.Review.SessionProblems[0].Display = "forged" }},
		{"guide", func(v *contract.Proposal) { v.Review.InvitationFirstSteps = nil }},
		{"local-evidence", func(v *contract.Proposal) { v.Evidence.Local.KnownRecordMissing = true }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := copied(base)
			c.alter(&v)
			before := copied(v)
			got, problems := r.Begin(v, "invalid-event")
			if c.name == "local-evidence" {
				ownProblem(t, problems, "review.known-record", "invitation.known-record-missing", "local.record", "Restore the registered record instead of creating another seed.")
			} else {
				ownProblem(t, problems, "begin.proposal", "invitation.proposal", "proposal", "Use the complete proposal derived from its exact source evidence.")
			}
			if !reflect.DeepEqual(got, contract.Publication{}) || !reflect.DeepEqual(v, before) {
				t.Error("forged proposal changed or returned partial publication")
			}
		})
	}
	got, problems := newReviewer().Begin(copied(base), beginAt)
	if len(problems) != 0 || !bytes.Equal(got.Artifacts[2].Bytes, fixture(t, "begun.record.yaml")) {
		t.Error("consistent plain reconstructed proposal must work on a fresh instance")
	}
}

func TestDeepOwnershipAndReusedConcurrentInstance(t *testing.T) {
	r := newReviewer()
	raw := archive(t, nil)
	rawBefore := copied(raw)
	in, problems := r.Inspect(raw)
	if len(problems) != 0 {
		t.Fatal("baseline inspect refused")
	}
	inBefore := copied(in)
	local := contract.Local{Record: contract.Body{Present: true, Bytes: fixture(t, "retained.record.yaml")}}
	localBefore := copied(local)
	result, problems := r.Review(in, local, contract.PlanChoiceUseIncoming)
	if len(problems) != 0 || result.Proposal == nil {
		t.Fatal("baseline review refused")
	}
	p := *result.Proposal
	pBefore := copied(p)
	pub, problems := r.Begin(p, "")
	if len(problems) != 0 {
		t.Fatal("baseline begin refused")
	}
	if !reflect.DeepEqual(raw, rawBefore) || !reflect.DeepEqual(in, inBefore) || !reflect.DeepEqual(local, localBefore) || !reflect.DeepEqual(p, pBefore) {
		t.Error("operation changed deep input or nil/empty shape")
	}
	pub.Artifacts[0].Bytes[0] ^= 1
	pub.OriginalReview.Artifacts[2].Bytes[0] ^= 1
	pub.OriginalReview.SessionProblems[0].Problem.Detail["audience"] = "changed"
	pub.OriginalReview.Session.Activities[0].Section.Variables[0].Description = "changed"
	if !reflect.DeepEqual(p, pBefore) || !reflect.DeepEqual(in, inBefore) || !reflect.DeepEqual(local, localBefore) || !bytes.Equal(pub.Artifacts[2].Bytes, fixture(t, "retained.record.yaml")) {
		t.Error("publication owns neither inputs nor independent final artifacts")
	}
	want, problems := r.Begin(p, "")
	if len(problems) != 0 {
		t.Fatal("later call affected by earlier result mutation")
	}
	var wg sync.WaitGroup
	for range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, errors := r.Begin(p, "")
			if len(errors) != 0 || !reflect.DeepEqual(got, want) {
				t.Error("concurrent reuse differs from sequential result")
			}
		}()
	}
	wg.Wait()
}
