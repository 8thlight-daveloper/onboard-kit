package reconciler_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type definitionHashCase struct {
	Name, Source, Canonical, CurrentHash, StoredHash string
	Section                                          json.RawMessage
	Changed                                          bool
}

func definitionHashCases(t *testing.T) []definitionHashCase {
	t.Helper()
	var cases []definitionHashCase
	if err := json.Unmarshal(readFixture(t, "outcome-hashes.json"), &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 12 {
		t.Fatal("expected twelve independent definition-comparison cases")
	}
	for _, c := range cases {
		if c.Source == "" {
			t.Fatalf("%s has no observable fixture source", c.Name)
		}
		if string(c.Section) == "null" {
			if c.Canonical != "" || c.CurrentHash != "" {
				t.Fatal("removed section must not invent a current definition hash")
			}
			continue
		}
		sum := sha256.Sum256([]byte(c.Canonical))
		if c.Canonical == "" || hex.EncodeToString(sum[:]) != c.CurrentHash {
			t.Fatalf("%s: independently pinned canonical SHA-256 is wrong", c.Name)
		}
		if c.Changed != (c.StoredHash != "" && c.StoredHash != c.CurrentHash) {
			t.Fatalf("%s: expected mismatch contradicts literal digests", c.Name)
		}
	}
	return cases
}

func definitionPlan(t *testing.T, c definitionHashCase) model.Plan {
	t.Helper()
	plan := model.Plan{ID: "plan"}
	if string(c.Section) != "null" {
		section, err := jsonform.Section(c.Section)
		if err != nil {
			t.Fatal(err)
		}
		plan.Sections = []model.Section{section}
	}
	return plan
}

func definitionRecord(hash string, kind model.OutcomeKind) model.Record {
	return model.Record{
		SchemaVersion: 1, Account: "account", StartDate: "legacy date",
		FirstOpenedAt: " held first use\nnot a timestamp ",
		Plan:          model.PlanReference{ID: "plan", Hash: "old-plan-reference", URL: "old-location"},
		Answers:       model.Answers{Plan: map[string]model.Value{"hidden": {Kind: model.ValueKindList, List: []string{"held", "answer"}}}},
		Outcomes: []model.Outcome{{
			Section: "access", Kind: kind, On: "2026-10-23T12:00:00Z",
			SectionHash: hash, Note: "held note", Differently: "held workaround",
		}},
		UnknownFields: map[string]any{"nested": []any{"held", map[string]any{"count": int64(3)}}},
	}
}

func TestOutcomeDefinitionsUseActualHeldContentAndPreserveHistory(t *testing.T) {
	t.Parallel()
	shared := newReconciler()
	previous := model.Plan{ID: "plan", Sections: []model.Section{{ID: "access", Title: "Access", Summary: "Request access"}}}
	reference := model.PlanReference{ID: "ignored", Hash: "driver-whole-plan-hash", URL: "driver-location"}
	for _, c := range definitionHashCases(t) {
		for _, kind := range []model.OutcomeKind{model.OutcomeVerified, model.OutcomeFailed, model.OutcomeDeferred} {
			t.Run(c.Name+"/"+string(kind), func(t *testing.T) {
				plan := definitionPlan(t, c)
				beforePlan := definitionPlan(t, c)
				record := definitionRecord(c.StoredHash, kind)
				beforeRecord := definitionRecord(c.StoredHash, kind)
				want := definitionRecord(c.StoredHash, kind)
				want.Plan = model.PlanReference{ID: "plan", Hash: reference.Hash, URL: reference.URL}
				for call := 0; call < 2; call++ {
					got, report, problems := shared.Reconcile(record, previous, plan, reference)
					if len(problems) != 0 || !reflect.DeepEqual(got, want) {
						t.Fatalf("history/reference changed beyond reconciliation: got %#v, want %#v; problems %+v", got, want, problems)
					}
					var wantChanged []string
					if c.Changed {
						wantChanged = []string{"access"}
					}
					if !reflect.DeepEqual(report.ChangedOutcomes, wantChanged) && !(len(report.ChangedOutcomes) == 0 && len(wantChanged) == 0) {
						t.Fatalf("changed outcomes %v; want %v using held digest %q and current digest %q", report.ChangedOutcomes, wantChanged, c.StoredHash, c.CurrentHash)
					}
					got.Outcomes[0].SectionHash = "must not migrate input history"
					got.Answers.Plan["hidden"].List[0] = "changed result"
					got.UnknownFields["nested"].([]any)[0] = "changed result"
					if len(report.ChangedOutcomes) > 0 {
						report.ChangedOutcomes[0] = "changed result"
					}
					if !reflect.DeepEqual(record, beforeRecord) || !reflect.DeepEqual(plan, beforePlan) {
						t.Fatal("comparison or result mutation changed input history or section baseline")
					}
				}
			})
		}
	}
}

func TestOutcomeHashRefusalPreservesHistoryWithoutComparison(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"", "different"} {
		record := definitionRecord("old arbitrary unversioned hash", model.OutcomeVerified)
		before := definitionRecord("old arbitrary unversioned hash", model.OutcomeVerified)
		plan := model.Plan{ID: id, Sections: []model.Section{{ID: "access", Summary: "changed"}}}
		got, report, problems := newReconciler().Reconcile(record, plan, plan, model.PlanReference{Hash: "new"})
		code := "reconcile.different-plan"
		if id == "" {
			code = "reconcile.missing-plan-id"
		}
		if len(problems) != 1 || problems[0].Code != code || !reconciliationIsEmpty(report) || !reflect.DeepEqual(got, before) {
			t.Fatalf("refusal must preserve all history without a changed list: %#v, %+v, %+v", got, report, problems)
		}
		got.Outcomes[0].SectionHash = "changed result"
		if !reflect.DeepEqual(record, before) {
			t.Fatal("refusal shared outcome storage")
		}
	}
}

func TestActualDefinitionMismatchesAreSortedUniqueWithoutPruningHistory(t *testing.T) {
	t.Parallel()
	const currentA = "40d7a66b3b1c2314222c84584010cbfbd0d451561531c1a56872c273a6065a45"
	const currentZ = "63dd780ec644019c5ed60eb7f11684172e4d7ded4e32d404b9d801cf45063baf"
	for canonical, want := range map[string]string{`{"id":"a","timing":"anytime"}`: currentA, `{"id":"z","timing":"anytime"}`: currentZ} {
		sum := sha256.Sum256([]byte(canonical))
		if hex.EncodeToString(sum[:]) != want {
			t.Fatal("independent sorted-unique definition digest is wrong")
		}
	}
	plan := model.Plan{ID: "plan", Sections: []model.Section{{ID: "z"}, {ID: "a"}}}
	record := definitionRecord("unused", model.OutcomeVerified)
	record.Outcomes = []model.Outcome{
		{Section: "z", Kind: model.OutcomeFailed, SectionHash: "legacy-z", Note: "first"},
		{Section: "a", Kind: model.OutcomeVerified, SectionHash: "legacy-a", Differently: "workaround"},
		{Section: "z", Kind: model.OutcomeDeferred, SectionHash: "another-z", Note: "unchecked duplicate retained"},
		{Section: "removed", Kind: model.OutcomeVerified, SectionHash: "legacy-removed"},
		{Section: "a", Kind: model.OutcomeVerified, SectionHash: currentA},
	}
	before, err := jsonform.RecordJSON(record)
	if err != nil {
		t.Fatal(err)
	}
	got, report, problems := newReconciler().Reconcile(record, plan, plan, record.Plan)
	if len(problems) != 0 || !reflect.DeepEqual(report.ChangedOutcomes, []string{"a", "z"}) {
		t.Fatalf("want sorted unique actual mismatches [a z], got %+v and %+v", report, problems)
	}
	want := record
	if !reflect.DeepEqual(got, want) {
		t.Fatal("comparison changed held unchecked duplicate outcomes or other history")
	}
	got.Outcomes[0].Note = "changed result"
	after, err := jsonform.RecordJSON(record)
	if err != nil || string(after) != string(before) {
		t.Fatal("result shares held outcome storage")
	}
}

func TestExistingMatchingOutcomeFixturesHaveIndependentCanonicalDigests(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		file, sectionID, canonical, digest string
	}{
		{"an-outcome-with-the-current-hash-is-not-listed.case.json", "g", `{"id":"g","title":"G","summary":"About g","timing":"anytime"}`, "2b2902a102a62918a84a8896888f076646a48044f4454b1c3edc6de35b46b682"},
		{"nothing-changed.case.json", "s", `{"id":"s","title":"S","summary":"About s","timing":"anytime","variables":[{"name":"three","type":"string","audience":"onboardee"}]}`, "7f7c445badee8b0a4bb61c9a35742fd146e9f5d2674e9d745f8a6aa957adc9ac"},
	} {
		sum := sha256.Sum256([]byte(c.canonical))
		if hex.EncodeToString(sum[:]) != c.digest {
			t.Fatal("independent matching fixture SHA-256 is wrong")
		}
		var file caseFile
		if err := json.Unmarshal(readFixture(t, c.file), &file); err != nil {
			t.Fatal(err)
		}
		plan := readPlan(t, c.file, file.Plan)
		canonical, err := jsonform.Section([]byte(c.canonical))
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, section := range plan.Sections {
			if section.ID != c.sectionID {
				continue
			}
			section.Source = nil
			if section.Timing == "" {
				section.Timing = model.DefaultTiming
			}
			for i := range section.Variables {
				if section.Variables[i].Type == "" {
					section.Variables[i].Type = model.DefaultVariableType
				}
			}
			if !reflect.DeepEqual(section, canonical) {
				t.Fatalf("%s no longer has its hand-pinned canonical definition", c.file)
			}
			found = true
		}
		if !found {
			t.Fatalf("%s missing pinned section %s", c.file, c.sectionID)
		}
		for _, raw := range []json.RawMessage{file.Record, file.Expected.Record} {
			record := readRecord(t, c.file, raw)
			matched := false
			for _, outcome := range record.Outcomes {
				if outcome.Section == c.sectionID {
					if outcome.SectionHash != c.digest {
						t.Fatalf("%s does not preserve its independently pinned matching digest", c.file)
					}
					matched = true
				}
			}
			if !matched {
				t.Fatal("missing matching outcome control")
			}
		}
	}
}
