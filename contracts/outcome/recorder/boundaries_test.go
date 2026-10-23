package recorder_test

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/outcome/recorder"
	"onboard-kit/internal/model"
)

func TestWarningsAndRequiredQuestionsNeverBecomeNewVerificationGates(t *testing.T) {
	t.Parallel()
	const canonical = `{"id":"a","bootstrap":true,"timing":"anytime","requires":["b"],"body":"Use $account.","variables":[{"name":"account","type":"string","required":true,"audience":"onboardee"}]}`
	const digest = "e60ddd333faa4f9c76a6fe8fc8ad74af08419e17584ebc62feeeb1ecc9587c6f"
	sum := sha256.Sum256([]byte(canonical))
	if hex.EncodeToString(sum[:]) != digest {
		t.Fatal("independent required-question canonical is inconsistent")
	}
	plan := model.Plan{Sections: []model.Section{
		{ID: "a", Bootstrap: true, Requires: []string{"b"}, Body: "Use $account.", Variables: []model.Variable{{Name: "account", Required: true, Audience: model.AudienceOnboardee}}},
		{ID: "b", Bootstrap: true},
	}}
	record := model.Record{Outcomes: []model.Outcome{{Section: "b", Kind: model.OutcomeVerified, SectionHash: "older"}}}
	proposal := contract.Proposal{Section: "a", Kind: model.OutcomeVerified}
	shared := newRecorder()
	got, problems := shared.Record(plan, record, proposal, "2026-10-23T12:34:56Z")
	if len(got.Outcomes) != 2 || got.Outcomes[0].SectionHash != digest || len(problems) != 1 {
		t.Fatalf("warning or question blocked explicit outcome: %#v %#v", got, problems)
	}
	assertAccountGap(t, problems[0])
	record.Outcomes[0].Kind = model.OutcomeFailed
	got, problems = shared.Record(plan, record, proposal, "2026-10-23T12:34:56Z")
	if !reflect.DeepEqual(got, record) || len(problems) != 2 {
		t.Fatal("waiting refusal lost warning or changed record")
	}
	assertAccountGap(t, problems[0])
	want := model.Problem{Code: "outcome.prerequisite-unverified", Severity: model.SeverityError, Message: "Verify the required section before recording this section as verified.", Section: "a", Detail: map[string]string{"field": "requires", "required": "b"}}
	if !reflect.DeepEqual(problems[1], want) {
		t.Fatal("recorder refusal did not follow unchanged walker warning")
	}
}

func assertAccountGap(t *testing.T, p model.Problem) {
	t.Helper()
	if p.Code != "gap.unanswered" || p.Severity != model.SeverityWarning || p.Section != "a" || p.Question != "a.account" || p.Detail["audience"] != "onboardee" || strings.TrimSpace(p.Message) == "" {
		t.Fatalf("lost useful resolver warning: %#v", p)
	}
}

func TestOnlyDirectVerifiedEvidenceNotTransitiveHistory(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Sections: []model.Section{{ID: "access", Title: "Access", Summary: "Request access", Requires: []string{"a"}}, {ID: "a", Requires: []string{"z"}}, {ID: "z"}}}
	record := model.Record{Outcomes: []model.Outcome{{Section: "z", Kind: model.OutcomeFailed, Note: "Keep failed history"}, {Section: "a", Kind: model.OutcomeVerified, SectionHash: "old", Note: "Keep verified fact"}}}
	want := []model.Outcome{
		record.Outcomes[1],
		{Section: "access", Kind: model.OutcomeVerified, On: "2026-10-23T12:34:56Z", SectionHash: "7db6dbae189ea1177381c325d068dc9ec49fa17a61a6d2e2858db04acefd1978"},
		record.Outcomes[0],
	}
	got, problems := newRecorder().Record(plan, record, contract.Proposal{Section: "access", Kind: model.OutcomeVerified}, "2026-10-23T12:34:56Z")
	if len(problems) != 0 || !reflect.DeepEqual(got.Outcomes, want) {
		t.Fatal("historical required verification was recursively revoked")
	}
}

func TestParentAfterGoalsAndEarlierActionsSupplyNoExtraGate(t *testing.T) {
	t.Parallel()
	const canonical = `{"id":"access","title":"Access","summary":"Request access","parent":"group","timing":"before-day-one","after":["later"]}`
	const digest = "60b8c096e4e874a4e27d9df92d2cf2b361c551b8928fa126fa3424abc54bd6c9"
	sum := sha256.Sum256([]byte(canonical))
	if hex.EncodeToString(sum[:]) != digest {
		t.Fatal("independent containment/order canonical is inconsistent")
	}
	plan := model.Plan{Sections: []model.Section{
		{ID: "access", Title: "Access", Summary: "Request access", Parent: "group", Timing: model.TimingBeforeDayOne, After: []string{"later"}},
		{ID: "group"},
		{ID: "later", Parent: "group", Timing: model.TimingBeforeDayOne},
	}}
	got, problems := newRecorder().Record(plan, model.Record{}, contract.Proposal{Section: "access", Kind: model.OutcomeVerified}, "2026-10-23T12:34:56Z")
	want := []model.Outcome{{Section: "access", Kind: model.OutcomeVerified, On: "2026-10-23T12:34:56Z", SectionHash: digest}}
	if len(problems) != 0 || !reflect.DeepEqual(got.Outcomes, want) {
		t.Fatal("invented parent, After, first-use or earlier-scheduled-activity gate")
	}
}

func TestAllGraphAndOutcomeAmbiguityProblemsPassThrough(t *testing.T) {
	t.Parallel()
	off := false
	plan := model.Plan{Sections: []model.Section{{ID: "a", Parent: "private-parent", Enabled: &off}, {ID: "b", Requires: []string{"b"}, Enabled: &off}}}
	record := model.Record{Outcomes: []model.Outcome{{Section: "gone", Kind: model.OutcomeVerified}, {Section: "gone", Kind: model.OutcomeFailed}}}
	want := []model.Problem{
		{Code: "plan.parent-unknown", Severity: model.SeverityError, Message: "Choose an existing section as parent, or remove parent.", Section: "a", Detail: map[string]string{"field": "parent"}},
		{Code: "plan.requires-self", Severity: model.SeverityError, Message: "Remove the section's reference to itself in requires.", Section: "b", Detail: map[string]string{"field": "requires", "item": "0"}},
	}
	shared := newRecorder()
	got, problems := shared.Record(plan, record, contract.Proposal{Section: "a", Kind: model.OutcomeFailed}, "2026-10-23T12:34:56Z")
	if !reflect.DeepEqual(got, record) || !reflect.DeepEqual(problems, want) {
		t.Fatal("lost full inactive graph diagnostics or chose a duplicate outcome")
	}
	plan = model.Plan{Sections: []model.Section{{ID: "a"}, {ID: "a"}}}
	want = []model.Problem{{Code: "plan.duplicate-section-id", Severity: model.SeverityError, Message: "Give each section a unique ID before checking relationships.", Section: "a", Detail: map[string]string{}}}
	got, problems = shared.Record(plan, record, contract.Proposal{Section: "a", Kind: model.OutcomeFailed}, "2026-10-23T12:34:56Z")
	if !reflect.DeepEqual(got, record) || !reflect.DeepEqual(problems, want) {
		t.Fatal("ambiguous section graph did not refuse before record duplicate lookup")
	}
}

func TestUnknownAndCaseDifferentTargetsNeverNormalize(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Sections: []model.Section{{ID: "access"}}}
	want := []model.Problem{{Code: "outcome.section-unknown", Severity: model.SeverityError, Message: "Choose a section declared in this plan.", Detail: map[string]string{"field": "section"}}}
	for _, target := range []string{"Access", "access ", " access", "private-note-like-target"} {
		got, problems := newRecorder().Record(plan, model.Record{}, contract.Proposal{Section: target, Kind: model.OutcomeFailed}, "2026-10-23T12:34:56Z")
		if !reflect.DeepEqual(got, model.Record{}) || !reflect.DeepEqual(problems, want) {
			t.Fatal("normalized or echoed an unknown target")
		}
	}
}

func TestChangingPrerequisiteDoesNotResetVerifiedDependent(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Sections: []model.Section{{ID: "access", Title: "Access", Summary: "Request access"}, {ID: "dependent", Requires: []string{"access"}}}}
	record := model.Record{Outcomes: []model.Outcome{
		{Section: "dependent", Kind: model.OutcomeVerified, On: "held-time", SectionHash: "held-hash", Note: "Keep successful work", Differently: "Keep exception"},
		{Section: "access", Kind: model.OutcomeVerified, On: "old", SectionHash: "old"},
	}}
	shared := newRecorder()
	got, problems := shared.Record(plan, record, contract.Proposal{Section: "access", Kind: model.OutcomeFailed, Note: "Access no longer works"}, "2026-10-23T12:34:56Z")
	want := []model.Outcome{
		{Section: "access", Kind: model.OutcomeFailed, On: "2026-10-23T12:34:56Z", SectionHash: "6c0301d995d0a3b0dab95a641a087120c5266746e59c50d934c575e135ca2fdc", Note: "Access no longer works"},
		record.Outcomes[0],
	}
	if len(problems) != 0 || !reflect.DeepEqual(got.Outcomes, want) {
		t.Fatal("changed prerequisite retroactively reset historical dependent")
	}
	again, problems := shared.Record(plan, got, contract.Proposal{Section: "dependent", Kind: model.OutcomeVerified}, "2026-10-23T12:35:00Z")
	if !reflect.DeepEqual(again, got) || len(problems) != 1 || problems[0].Code != "outcome.prerequisite-unverified" || problems[0].Detail["required"] != "access" {
		t.Fatal("historical dependent bypassed a new verification attempt")
	}
}

func TestSuccessfulAndRefusedCopiesKeepNilAndEmptyShapes(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Sections: []model.Section{{ID: "access", Title: "Access", Summary: "Request access"}}}
	for _, refuse := range []bool{false, true} {
		record := model.Record{Answers: model.Answers{Plan: map[string]model.Value{}, Sections: map[string]map[string]model.Value{"empty": {}}}, Outcomes: []model.Outcome{}, UnknownFields: map[string]any{"empty-list": []any{}, "empty-map": map[string]any{}}}
		proposal := contract.Proposal{Section: "access", Kind: model.OutcomeVerified}
		if refuse {
			proposal.Kind = "invalid"
		}
		got, problems := newRecorder().Record(plan, record, proposal, "2026-10-23T12:34:56Z")
		want := model.Record{Answers: model.Answers{Plan: map[string]model.Value{}, Sections: map[string]map[string]model.Value{"empty": {}}}, Outcomes: []model.Outcome{}, UnknownFields: map[string]any{"empty-list": []any{}, "empty-map": map[string]any{}}}
		if !refuse {
			want.Outcomes = []model.Outcome{{Section: "access", Kind: model.OutcomeVerified, On: "2026-10-23T12:34:56Z", SectionHash: "6c0301d995d0a3b0dab95a641a087120c5266746e59c50d934c575e135ca2fdc"}}
		}
		if !reflect.DeepEqual(got, want) || refuse && len(problems) != 1 || !refuse && len(problems) != 0 {
			t.Fatal("operation normalized unchanged empty collection shapes")
		}
		got.Answers.Sections["empty"]["caller"] = model.Value{Kind: model.ValueKindText, Text: "changed"}
		got.UnknownFields["empty-map"].(map[string]any)["caller"] = "changed"
		if len(record.Answers.Sections["empty"]) != 0 || len(record.UnknownFields["empty-map"].(map[string]any)) != 0 {
			t.Fatal("empty returned map aliases record")
		}
	}
}

func TestIdenticalOutcomeDuplicatesRemainAmbiguous(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Sections: []model.Section{{ID: "access"}}}
	record := model.Record{Outcomes: []model.Outcome{{Section: "access", Kind: model.OutcomeVerified, Note: "same"}, {Section: "access", Kind: model.OutcomeVerified, Note: "same"}}}
	want := []model.Problem{{Code: "walk.duplicate-outcome", Severity: model.SeverityError, Message: "Keep one recorded outcome for this section.", Section: "access", Detail: map[string]string{"field": "outcomes"}}}
	got, problems := newRecorder().Record(plan, record, contract.Proposal{Section: "access", Kind: model.OutcomeFailed}, "2026-10-23T12:34:56Z")
	if !reflect.DeepEqual(got, record) || !reflect.DeepEqual(problems, want) {
		t.Fatal("matching duplicate outcomes were silently merged")
	}
}

func TestHeldEnumsAreExactWithoutNormalizingHistory(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Sections: []model.Section{{ID: "access"}}}
	want := []model.Problem{{Code: "outcome.held-kind-invalid", Severity: model.SeverityError, Message: "Repair the unknown recorded outcome kind before recording another outcome.", Section: "gone", Detail: map[string]string{"field": "outcomes"}}}
	for _, kind := range []model.OutcomeKind{"", "Verified", " verified", "verified ", "FAILED", "defer"} {
		record := model.Record{Outcomes: []model.Outcome{{Section: "gone", Kind: kind, On: "held-on", SectionHash: "held-hash", Note: "Keep exact unknown history"}}}
		got, problems := newRecorder().Record(plan, record, contract.Proposal{Section: "access", Kind: model.OutcomeFailed}, "2026-10-23T12:34:56Z")
		if !reflect.DeepEqual(got, record) || !reflect.DeepEqual(problems, want) {
			t.Fatal("unknown held kind normalized, ignored or lost")
		}
	}
}

func typedUnknownRecord() model.Record {
	return model.Record{UnknownFields: map[string]any{
		"int":        int(7),
		"int64":      int64(9),
		"float64":    float64(3.5),
		"nil":        nil,
		"nil-list":   []any(nil),
		"empty-list": []any{},
		"nil-map":    map[string]any(nil),
		"empty-map":  map[string]any{},
		"nested": map[string]any{
			"list":       []any{int(11), int64(13), float64(2.5), nil, map[string]any{"value": int64(17)}},
			"nil-list":   []any(nil),
			"empty-list": []any{},
			"nil-map":    map[string]any(nil),
			"empty-map":  map[string]any{},
		},
	}}
}

func TestTypedUnknownNumbersAndCollectionShapesRemainOwned(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Sections: []model.Section{{ID: "access", Title: "Access", Summary: "Request access"}}}
	for _, mode := range []string{"success", "refusal"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			record := typedUnknownRecord()
			before := typedUnknownRecord()
			want := typedUnknownRecord()
			proposal := contract.Proposal{Section: "access", Kind: model.OutcomeVerified}
			var wantProblems []model.Problem
			if mode == "refusal" {
				proposal.Kind = "invalid"
				wantProblems = []model.Problem{{Code: "outcome.kind-invalid", Severity: model.SeverityError, Message: "Choose verified, failed, or deferred for the new outcome.", Detail: map[string]string{"field": "outcome"}}}
			} else {
				want.Outcomes = []model.Outcome{{Section: "access", Kind: model.OutcomeVerified, On: "2026-10-23T12:34:56Z", SectionHash: "6c0301d995d0a3b0dab95a641a087120c5266746e59c50d934c575e135ca2fdc"}}
			}
			shared := newRecorder()
			got, problems := shared.Record(plan, record, proposal, "2026-10-23T12:34:56Z")
			held, heldProblems := shared.Record(plan, record, proposal, "2026-10-23T12:34:56Z")
			if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(held, want) || !equalProblems(problems, wantProblems) || !equalProblems(heldProblems, wantProblems) {
				t.Fatal("typed unknown numbers or nil/empty shapes changed during copying")
			}
			got.UnknownFields["int"] = int(70)
			got.UnknownFields["int64"] = int64(90)
			got.UnknownFields["float64"] = float64(35)
			got.UnknownFields["empty-map"].(map[string]any)["caller"] = "changed"
			nested := got.UnknownFields["nested"].(map[string]any)
			list := nested["list"].([]any)
			list[0] = int(110)
			list[1] = int64(130)
			list[2] = float64(25)
			list[4].(map[string]any)["value"] = int64(170)
			nested["empty-map"].(map[string]any)["caller"] = "changed"
			nested["empty-list"] = append(nested["empty-list"].([]any), int64(23))
			if !reflect.DeepEqual(record, before) || !reflect.DeepEqual(held, want) {
				t.Fatal("typed result unknown maps/lists alias input or another result")
			}
			record.UnknownFields["nested"].(map[string]any)["list"].([]any)[4].(map[string]any)["value"] = int64(999)
			record.UnknownFields["empty-map"].(map[string]any)["input"] = "changed"
			if !reflect.DeepEqual(held, want) {
				t.Fatal("typed input unknown data aliases prior copied record")
			}
		})
	}
}
