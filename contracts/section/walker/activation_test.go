package walker_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	contract "onboard-kit/contracts/section/walker"
	"onboard-kit/internal/model"
)

type questionExpectation struct {
	ID              string
	Audience        model.Audience
	Answer          json.RawMessage
	DeclaredDefault json.RawMessage
}

type questionOrderCase struct {
	Source    string
	Plan      json.RawMessage
	Sequence  []string
	Questions []questionExpectation
	Blockers  []string
}

type structuralCase struct {
	Source   string
	Plan     json.RawMessage
	Problems []model.Problem
}

func readActivationFixture(t *testing.T, name string, target any) {
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

func optionalValue(t *testing.T, data json.RawMessage) model.Value {
	t.Helper()
	if len(data) == 0 {
		return model.Value{}
	}
	value, err := jsonform.Value(data)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestRichQuestionOrderRemainsFullFinderOrder(t *testing.T) {
	t.Parallel()
	var c questionOrderCase
	readActivationFixture(t, "question-order.json", &c)
	if strings.TrimSpace(c.Source) == "" || len(c.Questions) != 7 {
		t.Fatal("missing literal question oracle")
	}
	plan, err := jsonform.Plan(c.Plan)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		view, problems := newWalker().Walk(plan, model.Record{})
		if len(problems) != 0 {
			t.Fatalf("unexpected problems: %#v", problems)
		}
		var sequence []string
		for _, row := range view.Sections {
			sequence = append(sequence, row.Section.ID)
		}
		if !slices.Equal(sequence, c.Sequence) {
			t.Fatalf("action sequence: got %q want %q", sequence, c.Sequence)
		}
		if len(view.Questions) != len(c.Questions) {
			t.Fatalf("questions: got %#v want %#v", view.Questions, c.Questions)
		}
		for i, want := range c.Questions {
			got := view.Questions[i]
			if got.ID != want.ID || got.Audience != want.Audience || !reflect.DeepEqual(got.Answer, optionalValue(t, want.Answer)) || !reflect.DeepEqual(got.DeclaredDefault, optionalValue(t, want.DeclaredDefault)) || got.Group == "" {
				t.Fatalf("question %d: got %#v want %#v", i, got, want)
			}
		}
		if view.Questions[1].Group != view.Questions[3].Group {
			t.Fatal("lost same-name/type cross-audience group")
		}
		var blockers []string
		for _, blocker := range view.Blockers {
			blockers = append(blockers, blocker.Question)
		}
		if !slices.Equal(blockers, c.Blockers) {
			t.Fatalf("blockers: got %q want %q", blockers, c.Blockers)
		}
		if len(view.Sections) != 2 || !slices.Equal(view.Sections[1].WaitingOn, []string{"b"}) {
			t.Fatal("waiting activity questions or prerequisite evidence was conflated")
		}
		slices.Reverse(plan.Sections)
	}
}

func TestHashUsesUnresolvedFullBootstrapRequiresDefinition(t *testing.T) {
	t.Parallel()
	const canonical = `{"id":"a","bootstrap":true,"timing":"anytime","requires":["b"],"body":"Use $account.","variables":[{"name":"account","type":"string","audience":"onboardee"}]}`
	const digest = "ac608743e3475f9e217496379b65365ed93cbe7bf91c0261dd440773437abf5a"
	sum := sha256.Sum256([]byte(canonical))
	if hex.EncodeToString(sum[:]) != digest {
		t.Fatal("independent unresolved canonical oracle is inconsistent")
	}
	plan := model.Plan{
		Sections: []model.Section{
			{ID: "a", Bootstrap: true, Requires: []string{"b"}, Body: "Use $account.", Variables: []model.Variable{{Name: "account", Audience: model.AudienceOnboardee}}, Source: &model.Source{Template: "base", Hash: "old"}},
			{ID: "b", Bootstrap: true},
		},
		Answers: model.Answers{Sections: map[string]map[string]model.Value{"a": {"account": {Kind: model.ValueKindText, Text: "accepted"}}}},
	}
	record := model.Record{Outcomes: []model.Outcome{{Section: "a", Kind: model.OutcomeVerified, SectionHash: digest}, {Section: "b", Kind: model.OutcomeVerified}}}
	shared := newWalker()
	for _, answer := range []string{"accepted", "changed"} {
		plan.Answers.Sections["a"]["account"] = model.Value{Kind: model.ValueKindText, Text: answer}
		view, problems := shared.Walk(plan, record)
		if len(problems) != 0 || len(view.Outcomes) != 2 {
			t.Fatalf("missing resolved definition view: %#v %#v", view, problems)
		}
		row := rowFor(t, view, "a")
		if row.Section.Body != "Use "+answer+"." || row.CurrentHash != digest || len(row.WaitingOn) != 0 {
			t.Fatal("hashed resolved text, dropped metadata, or added a freshness gate")
		}
		if view.Outcomes[0].Outcome.Section != "a" || view.Outcomes[0].CurrentHash != digest || view.Outcomes[0].Comparison != contract.ComparisonSame {
			t.Fatal("history comparison used a different definition basis")
		}
	}
}

func TestMultipleStructuralDiagnosticsPassThroughExactly(t *testing.T) {
	t.Parallel()
	var c structuralCase
	readActivationFixture(t, "structural-errors.json", &c)
	if strings.TrimSpace(c.Source) == "" || len(c.Problems) != 7 {
		t.Fatal("missing literal structural oracle")
	}
	plan, err := jsonform.Plan(c.Plan)
	if err != nil {
		t.Fatal(err)
	}
	shared := newWalker()
	for range 2 {
		view, problems := shared.Walk(plan, model.Record{})
		if !reflect.DeepEqual(problems, c.Problems) || len(view.Sections)+len(view.Outcomes)+len(view.Questions)+len(view.Blockers) != 0 {
			t.Fatalf("structural diagnostics altered or partial view returned: %#v %#v", view, problems)
		}
		slices.Reverse(plan.Sections)
	}
}

func TestStructuralAndGapProblemMapsAreOwned(t *testing.T) {
	t.Parallel()
	var c structuralCase
	readActivationFixture(t, "structural-errors.json", &c)
	plan, err := jsonform.Plan(c.Plan)
	if err != nil {
		t.Fatal(err)
	}
	shared := newWalker()
	_, first := shared.Walk(plan, model.Record{})
	_, held := shared.Walk(plan, model.Record{})
	if !reflect.DeepEqual(first, c.Problems) || !reflect.DeepEqual(held, c.Problems) {
		t.Fatal("missing literal diagnostics")
	}
	first[0].Detail["field"] = "changed"
	first[0].Detail["added"] = "caller edit"
	first[0].Code = "caller.changed"
	if !reflect.DeepEqual(held, c.Problems) {
		t.Fatal("structural result maps or slots alias another result")
	}
	_, again := shared.Walk(plan, model.Record{})
	if !reflect.DeepEqual(again, c.Problems) {
		t.Fatal("caller edits persisted in walker diagnostics")
	}
	gapPlan := model.Plan{Sections: []model.Section{{ID: "gap", Body: "Ask $owner.", Variables: []model.Variable{{Name: "owner"}}}}}
	_, gapFirst := shared.Walk(gapPlan, model.Record{})
	_, gapHeld := shared.Walk(gapPlan, model.Record{})
	if len(gapFirst) != 1 || len(gapHeld) != 1 || gapHeld[0].Detail["audience"] != "onboarder" {
		t.Fatal("missing literal resolver diagnostic")
	}
	gapFirst[0].Detail["audience"] = "changed"
	gapFirst[0].Detail["added"] = "caller edit"
	if gapHeld[0].Detail["audience"] != "onboarder" || gapHeld[0].Detail["added"] != "" {
		t.Fatal("resolver problem map aliases another result")
	}
	_, gapAgain := shared.Walk(gapPlan, model.Record{})
	if len(gapAgain) != 1 || gapAgain[0].Detail["audience"] != "onboarder" || gapAgain[0].Detail["added"] != "" {
		t.Fatal("caller edits persisted in resolver diagnostics")
	}
}

func isolationPlan(t *testing.T) model.Plan {
	t.Helper()
	const raw = `{
		"variables":[{"name":"ready"},{"name":"command","type":"command"}],
		"answers":{"plan":{"ready":"yes","command":["echo","ok"]}},
		"sections":[{"id":"b"},{"id":"a","enabled":true,"requires":["b","b"],"after":["b"],"body":"Literal instructions.","escalation":["Ask support","Ask support"],
		"source":{"template":"original","hash":"baseline"},
		"includeIf":{"all":[{"question":"plan.ready","answer":{"isOneOf":["yes","maybe"]}},{"question":"plan.ready","answer":{"isNotOneOf":["no"]}},{"question":"plan.command","answer":{"isSettled":true}},{"question":"plan.ready","answer":{"is":"yes","isNot":"no"}},{"section":"b","enabled":true}]},
		"variables":[{"name":"picker","type":"choice","options":["one","two"],"default":"one","askIf":{"all":[{"question":"plan.ready","answer":{"isSettled":true}}]}},{"name":"command","type":"command","default":["echo","ok"],"askIf":{"all":[{"question":"plan.command","answer":{"isSettled":true}}]}}],
		"steps":[{"id":"read","name":"Read","run":["echo","ok"],"escalation":["Ask another helper"],"env":{"MODE":"quiet"},"checks":[{"run":["test","ok"]}]}],
		"unknownFields":{"extra":{"list":["held",{"inside":"value"}]}}}]
	}`
	plan, err := jsonform.Plan([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func rowFor(t *testing.T, view contract.View, id string) *contract.SectionView {
	t.Helper()
	for i := range view.Sections {
		if view.Sections[i].Section.ID == id {
			return &view.Sections[i]
		}
	}
	t.Fatalf("missing action %q", id)
	return nil
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

func TestResolvedSectionInnerStorageAndOwnedViewLists(t *testing.T) {
	t.Parallel()
	plan := isolationPlan(t)
	wantPlan := isolationPlan(t)
	shared := newWalker()
	first, firstProblems := shared.Walk(plan, model.Record{})
	held, heldProblems := shared.Walk(plan, model.Record{})
	if len(firstProblems)+len(heldProblems) != 0 || len(first.Sections) != 2 || len(held.Sections) != 2 {
		t.Fatal("missing valid ownership fixture outputs")
	}
	firstRow := rowFor(t, first, "a")
	heldRow := rowFor(t, held, "a")
	if !reflect.DeepEqual(firstRow.Section, wantPlan.Sections[1]) || !reflect.DeepEqual(heldRow.Section, wantPlan.Sections[1]) || !slices.Equal(firstRow.WaitingOn, []string{"b"}) || !slices.Equal(firstRow.Escalation, []string{"Ask support", "Ask support"}) {
		t.Fatal("initial literal section/waiting/escalation data wrong")
	}
	changeOwned(reflect.ValueOf(&firstRow.Section).Elem())
	firstRow.WaitingOn[0] = "changed"
	firstRow.Escalation[0] = "changed"
	first.Sections[0].CurrentHash = "changed"
	if !reflect.DeepEqual(plan, wantPlan) {
		t.Fatal("editing returned nested section, waiting or escalation changed input")
	}
	if !reflect.DeepEqual(heldRow.Section, wantPlan.Sections[1]) || !slices.Equal(heldRow.WaitingOn, []string{"b"}) || !slices.Equal(heldRow.Escalation, []string{"Ask support", "Ask support"}) || !validDigest(held.Sections[0].CurrentHash) {
		t.Fatal("returned section/value lists alias another result")
	}
	changeOwned(reflect.ValueOf(&plan.Sections[1]).Elem())
	if !reflect.DeepEqual(heldRow.Section, wantPlan.Sections[1]) || !slices.Equal(heldRow.WaitingOn, []string{"b"}) || !slices.Equal(heldRow.Escalation, []string{"Ask support", "Ask support"}) {
		t.Fatal("editing input nested section changed prior result")
	}
}

func TestOneWalkerSupportsConcurrentIndependentInputs(t *testing.T) {
	t.Parallel()
	type call struct {
		fixture      walkCase
		plan         model.Plan
		record       model.Record
		beforePlan   model.Plan
		beforeRecord model.Record
		wantProblems []model.Problem
	}
	type result struct {
		fixture              walkCase
		view                 contract.View
		problems             []model.Problem
		wantProblems         []model.Problem
		plan, beforePlan     model.Plan
		record, beforeRecord model.Record
	}
	var calls []call
	for _, c := range fixtureCases(t) {
		plan, record := inputs(t, c)
		beforePlan, beforeRecord := inputs(t, c)
		calls = append(calls, call{fixture: c, plan: plan, record: record, beforePlan: beforePlan, beforeRecord: beforeRecord})
	}
	var structural structuralCase
	readActivationFixture(t, "structural-errors.json", &structural)
	badPlan, err := jsonform.Plan(structural.Plan)
	if err != nil {
		t.Fatal(err)
	}
	beforeBadPlan, err := jsonform.Plan(structural.Plan)
	if err != nil {
		t.Fatal(err)
	}
	calls = append(calls, call{fixture: walkCase{Name: "parallel-structural-refusal", Waiting: map[string][]string{}}, plan: badPlan, beforePlan: beforeBadPlan, wantProblems: structural.Problems})
	shared := newWalker()
	results := make(chan result, len(calls)*3)
	var workers sync.WaitGroup
	for _, c := range calls {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 3 {
				view, problems := shared.Walk(c.plan, c.record)
				results <- result{fixture: c.fixture, view: view, problems: problems, wantProblems: c.wantProblems, plan: c.plan, beforePlan: c.beforePlan, record: c.record, beforeRecord: c.beforeRecord}
			}
		}()
	}
	workers.Wait()
	close(results)
	for result := range results {
		if len(result.problems) != len(result.wantProblems) || len(result.wantProblems) > 0 && !reflect.DeepEqual(result.problems, result.wantProblems) || !reflect.DeepEqual(result.plan, result.beforePlan) || !reflect.DeepEqual(result.record, result.beforeRecord) {
			t.Fatalf("concurrent call %s changed inputs or failed: %#v", result.fixture.Name, result.problems)
		}
		if len(result.wantProblems) > 0 && len(result.view.Sections)+len(result.view.Outcomes)+len(result.view.Questions)+len(result.view.Blockers) != 0 {
			t.Fatal("concurrent refusal returned partial view")
		}
		var sequence []string
		for _, row := range result.view.Sections {
			sequence = append(sequence, row.Section.ID)
			want, ok := result.fixture.Waiting[row.Section.ID]
			if !ok || !slices.Equal(row.WaitingOn, want) || !validDigest(row.CurrentHash) {
				t.Fatalf("concurrent action row wrong for %s: %#v", result.fixture.Name, row)
			}
		}
		if !slices.Equal(sequence, result.fixture.Sequence) || len(result.view.Outcomes) != len(result.fixture.History) {
			t.Fatalf("shared state contaminated %s", result.fixture.Name)
		}
		for i, want := range result.fixture.History {
			row := result.view.Outcomes[i]
			if row.Outcome.Section != want.Section || row.Comparison != want.Comparison {
				t.Fatalf("shared history contaminated %s", result.fixture.Name)
			}
			for _, held := range result.record.Outcomes {
				if held.Section == want.Section && !reflect.DeepEqual(row.Outcome, held) {
					t.Fatal("concurrent call rewrote held outcome")
				}
			}
		}
	}
}
