package recorder_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	contract "onboard-kit/contracts/outcome/recorder"
	"onboard-kit/internal/model"
)

type recordCase struct {
	Name, Source string
	Plan, Record json.RawMessage
	Proposal     contract.Proposal
	On           model.Timestamp
	Outcomes     json.RawMessage
	Problems     []model.Problem
}

type hashCase struct{ Name, Source, Canonical, Hash string }

func readFixture(t *testing.T, name string, target any) {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		t.Fatal(err)
	}
}

func cases(t *testing.T) []recordCase {
	t.Helper()
	var result []recordCase
	readFixture(t, "cases.json", &result)
	if len(result) != 15 {
		t.Fatal("expected fifteen independently specified recording cases")
	}
	for _, c := range result {
		if c.Name == "" || c.Source == "" || len(c.Outcomes) == 0 && len(c.Problems) == 0 {
			t.Fatal("fixture lacks a source or an observable output oracle")
		}
	}
	return result
}

func inputs(t *testing.T, c recordCase) (model.Plan, model.Record, model.Record) {
	t.Helper()
	plan, err := jsonform.Plan(c.Plan)
	if err != nil {
		t.Fatal(err)
	}
	record, err := jsonform.Record(c.Record)
	if err != nil {
		t.Fatal(err)
	}
	want, err := jsonform.Record(c.Record)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Outcomes) > 0 {
		expected, err := jsonform.Record(append(append([]byte(`{"outcomes":`), c.Outcomes...), '}'))
		if err != nil {
			t.Fatal(err)
		}
		want.Outcomes = expected.Outcomes
	}
	return plan, record, want
}

func equalProblems(a, b []model.Problem) bool {
	return len(a) == len(b) && (len(a) == 0 || reflect.DeepEqual(a, b))
}

func TestLiteralRecordingAndRefusalCases(t *testing.T) {
	t.Parallel()
	for _, c := range cases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			plan, record, want := inputs(t, c)
			beforePlan, beforeRecord, _ := inputs(t, c)
			got, problems := newRecorder().Record(plan, record, c.Proposal, c.On)
			if !reflect.DeepEqual(got, want) || !equalProblems(problems, c.Problems) {
				t.Fatalf("recording mismatch\ngot %#v\nwant %#v\nproblems %#v\nwant problems %#v", got, want, problems, c.Problems)
			}
			if !reflect.DeepEqual(plan, beforePlan) || !reflect.DeepEqual(record, beforeRecord) {
				t.Fatal("operation changed arguments")
			}
		})
	}
}

func TestIndependentCanonicalHashOracles(t *testing.T) {
	t.Parallel()
	var fixtures []hashCase
	readFixture(t, "hashes.json", &fixtures)
	if len(fixtures) != 3 {
		t.Fatal("expected three canonical hash oracles")
	}
	for _, c := range fixtures {
		if c.Name == "" || c.Source == "" || c.Canonical == "" {
			t.Fatal("hash oracle lacks source")
		}
		sum := sha256.Sum256([]byte(c.Canonical))
		if hex.EncodeToString(sum[:]) != c.Hash {
			t.Fatalf("independent canonical digest inconsistent: %s", c.Name)
		}
	}
}

func TestNewTimestampProfileAndExactPreservation(t *testing.T) {
	t.Parallel()
	var times struct {
		Source            string
		Accepted, Refused []model.Timestamp
	}
	readFixture(t, "timestamps.json", &times)
	if times.Source == "" || len(times.Accepted) != 6 || len(times.Refused) != 22 {
		t.Fatal("timestamp table lacks independent boundaries")
	}
	plan := model.Plan{Sections: []model.Section{{ID: "access", Title: "Access", Summary: "Request access"}}}
	proposal := contract.Proposal{Section: "access", Kind: model.OutcomeVerified}
	shared := newRecorder()
	for i, on := range times.Accepted {
		t.Run("accepted/"+strconv.Itoa(i), func(t *testing.T) {
			got, problems := shared.Record(plan, model.Record{}, proposal, on)
			if len(problems) != 0 || len(got.Outcomes) != 1 || got.Outcomes[0].On != on || got.FirstOpenedAt != "" || got.Outcomes[0].SectionHash != "6c0301d995d0a3b0dab95a641a087120c5266746e59c50d934c575e135ca2fdc" {
				t.Fatalf("valid exact timestamp changed or refused: %q %#v %#v", on, got, problems)
			}
		})
	}
	want := []model.Problem{{Code: "outcome.timestamp-invalid", Severity: model.SeverityError, Message: "Supply a valid UTC outcome timestamp.", Detail: map[string]string{"field": "on"}}}
	for i, on := range times.Refused {
		t.Run("refused/"+strconv.Itoa(i), func(t *testing.T) {
			got, problems := shared.Record(plan, model.Record{}, proposal, on)
			if !reflect.DeepEqual(got, model.Record{}) || !reflect.DeepEqual(problems, want) {
				t.Fatalf("invalid timestamp accepted or repaired: %q %#v %#v", on, got, problems)
			}
		})
	}
}

func TestUnresolvedBootstrapRequiresHashRecheckedAtWrite(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Sections: []model.Section{
		{ID: "a", Bootstrap: true, Requires: []string{"b"}, Body: "Use $account.", Variables: []model.Variable{{Name: "account", Audience: model.AudienceOnboardee}}, Source: &model.Source{Hash: "wrong-baseline"}},
		{ID: "b", Bootstrap: true},
	}, Answers: model.Answers{Sections: map[string]map[string]model.Value{"a": {"account": {Kind: model.ValueKindText, Text: "accepted"}}}}}
	record := model.Record{Outcomes: []model.Outcome{{Section: "b", Kind: model.OutcomeVerified, SectionHash: "legacy", Note: "Keep prerequisite"}}}
	proposal := contract.Proposal{Section: "a", Kind: model.OutcomeVerified}
	shared := newRecorder()
	for _, value := range []string{"accepted", "changed"} {
		plan.Answers.Sections["a"]["account"] = model.Value{Kind: model.ValueKindText, Text: value}
		got, problems := shared.Record(plan, record, proposal, "2026-10-23T12:34:56Z")
		if len(problems) != 0 || len(got.Outcomes) != 2 || got.Outcomes[0].SectionHash != "ac608743e3475f9e217496379b65365ed93cbe7bf91c0261dd440773437abf5a" || !reflect.DeepEqual(got.Outcomes[1], record.Outcomes[0]) {
			t.Fatal("used resolved/baseline/old hash or rewrote prerequisite")
		}
	}
	record.Outcomes[0].Kind = model.OutcomeFailed
	got, problems := shared.Record(plan, record, proposal, "2026-10-23T12:34:56Z")
	if !reflect.DeepEqual(got, record) || len(problems) != 1 || problems[0].Code != "outcome.prerequisite-unverified" {
		t.Fatal("cached earlier prerequisite evidence")
	}
}

func changeOwned(value reflect.Value) {
	switch value.Kind() {
	case reflect.Pointer:
		if !value.IsNil() {
			changeOwned(value.Elem())
		}
	case reflect.Interface:
		if !value.IsNil() {
			copy := reflect.New(value.Elem().Type()).Elem()
			copy.Set(value.Elem())
			changeOwned(copy)
			value.Set(copy)
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if value.Field(i).CanSet() {
				changeOwned(value.Field(i))
			}
		}
	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			changeOwned(value.Index(i))
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			copy := reflect.New(value.Type().Elem()).Elem()
			copy.Set(value.MapIndex(key))
			changeOwned(copy)
			value.SetMapIndex(key, copy)
		}
	case reflect.String:
		value.SetString(value.String() + " changed")
	case reflect.Bool:
		value.SetBool(!value.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value.SetInt(value.Int() + 1)
	}
}

func TestOwnedRecordAndProblemStorageOnSuccessAndRefusal(t *testing.T) {
	t.Parallel()
	fixture := cases(t)[1]
	for _, refuse := range []bool{false, true} {
		plan, record, want := inputs(t, fixture)
		_, beforeRecord, _ := inputs(t, fixture)
		proposal := fixture.Proposal
		if refuse {
			proposal.Kind = "invalid"
			want = beforeRecord
		}
		shared := newRecorder()
		first, firstProblems := shared.Record(plan, record, proposal, fixture.On)
		held, heldProblems := shared.Record(plan, record, proposal, fixture.On)
		if !reflect.DeepEqual(first, want) || !reflect.DeepEqual(held, want) || refuse && (len(firstProblems) != 1 || len(heldProblems) != 1) {
			t.Fatal("ownership fixture operation failed")
		}
		changeOwned(reflect.ValueOf(&first).Elem())
		if !reflect.DeepEqual(record, beforeRecord) || !reflect.DeepEqual(held, want) {
			t.Fatal("returned nested maps/lists or outcome slots alias inputs or another result")
		}
		changeOwned(reflect.ValueOf(&record).Elem())
		if !reflect.DeepEqual(held, want) {
			t.Fatal("input mutation changed prior record result")
		}
		if refuse {
			firstProblems[0].Detail["field"] = "changed"
			if heldProblems[0].Detail["field"] != "outcome" {
				t.Fatal("problem Detail maps alias another result")
			}
		}
	}
}

func TestSameInstanceStateAndConcurrentCalls(t *testing.T) {
	t.Parallel()
	type request struct {
		fixture      recordCase
		plan         model.Plan
		record, want model.Record
	}
	type result struct {
		request  request
		record   model.Record
		problems []model.Problem
	}
	var requests []request
	for _, c := range cases(t) {
		plan, record, want := inputs(t, c)
		requests = append(requests, request{fixture: c, plan: plan, record: record, want: want})
	}
	shared := newRecorder()
	results := make(chan result, len(requests)*2)
	var workers sync.WaitGroup
	for _, request := range requests {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 2 {
				got, problems := shared.Record(request.plan, request.record, request.fixture.Proposal, request.fixture.On)
				results <- result{request: request, record: got, problems: problems}
			}
		}()
	}
	workers.Wait()
	close(results)
	for result := range results {
		if !reflect.DeepEqual(result.record, result.request.want) || !equalProblems(result.problems, result.request.fixture.Problems) {
			t.Fatalf("concurrent state affected %s", result.request.fixture.Name)
		}
	}
	for i := len(requests) - 1; i >= 0; i-- {
		request := requests[i]
		got, problems := shared.Record(request.plan, request.record, request.fixture.Proposal, request.fixture.On)
		if !reflect.DeepEqual(got, request.want) || !equalProblems(problems, request.fixture.Problems) {
			t.Fatal("earlier operation affected later output")
		}
	}
}

func TestExactEnumsAndUnchangedCollectionShapes(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Sections: []model.Section{{ID: "access", Title: "Access", Summary: "Request access"}}}
	for _, kind := range []model.OutcomeKind{"", "Verified", " verified", "verified ", "FAILED", "defer"} {
		record := model.Record{Answers: model.Answers{Plan: map[string]model.Value{}, Sections: map[string]map[string]model.Value{"empty": {}}}, Outcomes: []model.Outcome{}, UnknownFields: map[string]any{"empty-list": []any{}, "empty-map": map[string]any{}}}
		got, problems := newRecorder().Record(plan, record, contract.Proposal{Section: "access", Kind: kind}, "2026-10-23T12:34:56Z")
		if !reflect.DeepEqual(got, record) || len(problems) != 1 || problems[0].Code != "outcome.kind-invalid" {
			t.Fatal("bad enum normalized or refusal changed nil/empty shapes")
		}
	}
}

func TestWaitingDiagnosticsAreDistinctFirstHeldOrder(t *testing.T) {
	t.Parallel()
	plan := model.Plan{Sections: []model.Section{{ID: "target", Requires: []string{"c", "a", "c", "b", "a"}}, {ID: "a"}, {ID: "b"}, {ID: "c"}}}
	record := model.Record{Outcomes: []model.Outcome{{Section: "b", Kind: model.OutcomeVerified}}}
	got, problems := newRecorder().Record(plan, record, contract.Proposal{Section: "target", Kind: model.OutcomeVerified}, "2026-10-23T12:34:56Z")
	want := []model.Problem{
		{Code: "outcome.prerequisite-unverified", Severity: model.SeverityError, Message: "Verify the required section before recording this section as verified.", Section: "target", Detail: map[string]string{"field": "requires", "required": "c"}},
		{Code: "outcome.prerequisite-unverified", Severity: model.SeverityError, Message: "Verify the required section before recording this section as verified.", Section: "target", Detail: map[string]string{"field": "requires", "required": "a"}},
	}
	if !reflect.DeepEqual(got, record) || !reflect.DeepEqual(problems, want) {
		t.Fatal("waiting obligations sorted, repeated, dropped or bypassed")
	}
	if !slices.Equal(plan.Sections[0].Requires, []string{"c", "a", "c", "b", "a"}) {
		t.Fatal("mutated authored Requires")
	}
	for _, problem := range problems {
		if strings.TrimSpace(problem.Message) == "" {
			t.Fatal("unhelpful diagnostic")
		}
	}
}
