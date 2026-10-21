package splitter_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type indexEntry struct {
	Name   string `json:"name"`
	File   string `json:"file"`
	Source string `json:"source"`
}

type caseFile struct {
	Plan     json.RawMessage `json:"plan"`
	Joined   json.RawMessage `json:"joined"`
	Previous json.RawMessage `json:"previous"`
	Expected json.RawMessage `json:"expected"`
}

type loadedCase struct {
	name     string
	plan     model.Plan
	joined   model.Plan
	previous model.Record
	expected model.Record
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}

func loadIndex(t *testing.T) []indexEntry {
	t.Helper()
	var index []indexEntry
	if err := json.Unmarshal(readFixture(t, "cases.json"), &index); err != nil {
		t.Fatalf("cases.json: %v", err)
	}
	return index
}

func readPlan(t *testing.T, name string, data json.RawMessage) model.Plan {
	t.Helper()
	plan, err := jsonform.Plan(data)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return plan
}

func readRecord(t *testing.T, name string, data json.RawMessage) model.Record {
	t.Helper()
	record, err := jsonform.Record(data)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return record
}

func loadCase(t *testing.T, entry indexEntry) loadedCase {
	t.Helper()
	var file caseFile
	if err := json.Unmarshal(readFixture(t, entry.File), &file); err != nil {
		t.Fatalf("%s: %v", entry.File, err)
	}
	return loadedCase{
		name:     entry.Name,
		plan:     readPlan(t, entry.File+" plan", file.Plan),
		joined:   readPlan(t, entry.File+" joined", file.Joined),
		previous: readRecord(t, entry.File+" previous", file.Previous),
		expected: readRecord(t, entry.File+" expected", file.Expected),
	}
}

func loadCases(t *testing.T) []loadedCase {
	t.Helper()
	var cases []loadedCase
	for _, entry := range loadIndex(t) {
		cases = append(cases, loadCase(t, entry))
	}
	return cases
}

func loadCaseAgain(t *testing.T, c loadedCase) loadedCase {
	t.Helper()
	for _, entry := range loadIndex(t) {
		if entry.Name == c.name {
			return loadCase(t, entry)
		}
	}
	t.Fatalf("case %s not found", c.name)
	return loadedCase{}
}

func recordText(t *testing.T, record model.Record) string {
	t.Helper()
	data, err := jsonform.RecordJSON(record)
	if err != nil {
		t.Fatalf("write record: %v", err)
	}
	return string(data)
}

func split(c loadedCase) model.Record {
	return newSplitter().Split(c.plan, c.joined, c.previous)
}

func sameValue(a, b model.Value) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case model.ValueKindText:
		return a.Text == b.Text
	case model.ValueKindFlag:
		return a.Flag == b.Flag
	case model.ValueKindList:
		return len(a.List) == len(b.List) && (len(a.List) == 0 || reflect.DeepEqual(a.List, b.List))
	}
	return true
}

type question struct {
	scope string
	name  string
}

func (q question) id() string { return q.scope + "." + q.name }

func answerMap(answers model.Answers) map[question]model.Value {
	flat := map[question]model.Value{}
	for name, value := range answers.Plan {
		if value.Kind != model.ValueKindNone {
			flat[question{"plan", name}] = value
		}
	}
	for id, values := range answers.Sections {
		for name, value := range values {
			if value.Kind != model.ValueKindNone {
				flat[question{id, name}] = value
			}
		}
	}
	return flat
}

func audiences(plan model.Plan) map[question]model.Audience {
	declared := map[question]model.Audience{}
	for _, v := range plan.Variables {
		if _, seen := declared[question{"plan", v.Name}]; !seen {
			declared[question{"plan", v.Name}] = v.Audience
		}
	}
	for _, s := range plan.Sections {
		for _, v := range s.Variables {
			if _, seen := declared[question{s.ID, v.Name}]; !seen {
				declared[question{s.ID, v.Name}] = v.Audience
			}
		}
	}
	return declared
}

func join(plan model.Plan, record model.Record) model.Plan {
	joined := plan
	joined.Answers = model.Answers{}
	for q, value := range answerMap(plan.Answers) {
		put(&joined.Answers, q, value)
	}
	for q, value := range answerMap(record.Answers) {
		put(&joined.Answers, q, value)
	}
	return joined
}

func put(answers *model.Answers, q question, value model.Value) {
	if q.scope == "plan" {
		if answers.Plan == nil {
			answers.Plan = map[string]model.Value{}
		}
		answers.Plan[q.name] = value
		return
	}
	if answers.Sections == nil {
		answers.Sections = map[string]map[string]model.Value{}
	}
	if answers.Sections[q.scope] == nil {
		answers.Sections[q.scope] = map[string]model.Value{}
	}
	answers.Sections[q.scope][q.name] = value
}

func sameQuestions(a, b model.Plan) bool {
	return reflect.DeepEqual(audiences(a), audiences(b))
}

func TestEveryCaseCitesItsSource(t *testing.T) {
	for _, entry := range loadIndex(t) {
		if strings.TrimSpace(entry.Source) == "" {
			t.Errorf("case %s has no source", entry.Name)
		}
		if _, err := os.Stat(filepath.Join("testdata", entry.File)); err != nil {
			t.Errorf("case %s: %v", entry.Name, err)
		}
	}
}

func TestCases(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := split(c)
			if g, w := recordText(t, got), recordText(t, c.expected); g != w {
				t.Errorf("record\n got: %s\nwant: %s", g, w)
			}
		})
	}
}

func TestNoEmptyMapIsInvented(t *testing.T) {
	check := func(name string, record model.Record) {
		if record.Answers.Plan != nil && len(record.Answers.Plan) == 0 {
			t.Errorf("%s: an empty plan answer map", name)
		}
		if record.Answers.Sections != nil && len(record.Answers.Sections) == 0 {
			t.Errorf("%s: an empty section map", name)
		}
		for id, values := range record.Answers.Sections {
			if len(values) == 0 {
				t.Errorf("%s: the section %s has an empty answer map", name, id)
			}
		}
	}
	for _, c := range loadCases(t) {
		check(c.name, split(c))
	}
	emptied := model.Record{
		SchemaVersion: 1,
		Answers: model.Answers{
			Plan:     map[string]model.Value{"a": {Kind: model.ValueKindText, Text: "x"}},
			Sections: map[string]map[string]model.Value{"s": {"b": {Kind: model.ValueKindText, Text: "y"}}, "t": {}},
		},
	}
	check("emptied", newSplitter().Split(model.Plan{}, model.Plan{}, emptied))
	got := newSplitter().Split(model.Plan{}, model.Plan{}, emptied)
	if got.Answers.Plan != nil || got.Answers.Sections != nil {
		t.Errorf("nothing left must give nil maps, got %+v", got.Answers)
	}
}

func TestAbsentStaysAbsent(t *testing.T) {
	got := newSplitter().Split(model.Plan{}, model.Plan{}, model.Record{SchemaVersion: 1, Account: "ada"})
	if got.Outcomes != nil || got.UnknownFields != nil || got.Answers.Plan != nil || got.Answers.Sections != nil {
		t.Errorf("absent parts must stay nil, got %+v", got)
	}
	if got.SchemaVersion != 1 || got.Account != "ada" {
		t.Errorf("the other fields must be copied, got %+v", got)
	}
}

func TestAnEmptyValueIsNoAnswer(t *testing.T) {
	plan := model.Plan{Variables: []model.Variable{{Name: "team", Audience: model.AudienceOnboardee}}}
	joined := plan
	joined.Answers = model.Answers{
		Plan:     map[string]model.Value{"team": {}, "old": {}},
		Sections: map[string]map[string]model.Value{"gone": {"q": {}}},
	}
	previous := model.Record{SchemaVersion: 1, Answers: model.Answers{
		Plan:     map[string]model.Value{"team": {Kind: model.ValueKindText, Text: "data"}, "old": {Kind: model.ValueKindText, Text: "x"}},
		Sections: map[string]map[string]model.Value{"gone": {"q": {Kind: model.ValueKindText, Text: "y"}}},
	}}
	got := newSplitter().Split(plan, joined, previous)
	if got.Answers.Plan != nil || got.Answers.Sections != nil {
		t.Errorf("a value of kind none is a cleared answer, got %+v", got.Answers)
	}
}

func TestAHiddenAnswerThePreviousRecordDidNotHoldIsDropped(t *testing.T) {
	joined := model.Plan{Answers: model.Answers{Plan: map[string]model.Value{"old": {Kind: model.ValueKindFlag, Flag: true}}}}
	got := newSplitter().Split(model.Plan{}, joined, model.Record{})
	if got.Answers.Plan != nil {
		t.Errorf("a hidden answer the previous record did not hold must be dropped, got %+v", got.Answers)
	}
}

func TestEveryStoredAnswerIsOnboardeeOrHeldAndHidden(t *testing.T) {
	for _, c := range loadCases(t) {
		declared := audiences(c.joined)
		held := answerMap(c.previous.Answers)
		offered := answerMap(c.joined.Answers)
		pre := answerMap(c.plan.Answers)
		for q, value := range answerMap(split(c).Answers) {
			joinedValue, ok := offered[q]
			if !ok {
				t.Errorf("%s: %s is in the record and not in the joined plan", c.name, q.id())
				continue
			}
			if !sameValue(value, joinedValue) {
				t.Errorf("%s: %s is %+v, the joined plan has %+v", c.name, q.id(), value, joinedValue)
			}
			audience, isDeclared := declared[q]
			_, wasHeld := held[q]
			switch {
			case isDeclared && audience != model.AudienceOnboardee:
				t.Errorf("%s: %s is an onboarder answer", c.name, q.id())
			case isDeclared:
				if preValue, seeded := pre[q]; seeded && sameValue(preValue, value) && !wasHeld {
					t.Errorf("%s: %s equals the plan and was not held", c.name, q.id())
				}
			case !wasHeld:
				t.Errorf("%s: %s is a hidden answer the previous record did not hold", c.name, q.id())
			}
		}
	}
}

func TestEveryAnswerThatShouldBeStoredIsStored(t *testing.T) {
	for _, c := range loadCases(t) {
		declared := audiences(c.joined)
		held := answerMap(c.previous.Answers)
		pre := answerMap(c.plan.Answers)
		stored := answerMap(split(c).Answers)
		for q, value := range answerMap(c.joined.Answers) {
			audience, isDeclared := declared[q]
			_, wasHeld := held[q]
			preValue, seeded := pre[q]
			want := wasHeld
			if isDeclared {
				want = audience == model.AudienceOnboardee && (wasHeld || !seeded || !sameValue(preValue, value))
			}
			if _, ok := stored[q]; ok != want {
				t.Errorf("%s: %s stored is %v, want %v", c.name, q.id(), ok, want)
			}
		}
	}
}

func TestJoiningTheResultGivesBackTheOnboardeeAnswers(t *testing.T) {
	checked := 0
	for _, c := range loadCases(t) {
		if !sameQuestions(c.plan, c.joined) {
			continue
		}
		checked++
		rejoined := answerMap(join(c.plan, split(c)).Answers)
		offered := answerMap(c.joined.Answers)
		pre := answerMap(c.plan.Answers)
		for q, audience := range audiences(c.plan) {
			got, gotOK := rejoined[q]
			if audience != model.AudienceOnboardee {
				want, wantOK := pre[q]
				if gotOK != wantOK || (gotOK && !sameValue(got, want)) {
					t.Errorf("%s: onboarder answer %s is %+v, the plan has %+v", c.name, q.id(), got, want)
				}
				continue
			}
			want, wantOK := offered[q]
			if !wantOK {
				want, wantOK = pre[q]
			}
			if gotOK != wantOK || (gotOK && !sameValue(got, want)) {
				t.Errorf("%s: %s is %+v after the round trip, want %+v", c.name, q.id(), got, want)
			}
		}
	}
	if checked < 10 {
		t.Errorf("only %d cases were round tripped", checked)
	}
}

func TestSplittingAJoinedRecordAgainChangesNothing(t *testing.T) {
	for _, c := range loadCases(t) {
		if !sameQuestions(c.plan, c.joined) {
			continue
		}
		first := split(c)
		again := newSplitter().Split(c.plan, join(c.plan, first), first)
		if g, w := recordText(t, again), recordText(t, first); g != w {
			t.Errorf("%s: splitting again changed the record\n got: %s\nwant: %s", c.name, g, w)
		}
	}
}

func scribbleValue(value model.Value) {
	for i := range value.List {
		value.List[i] = "changed"
	}
}

func scribbleAnswers(answers model.Answers) {
	for name, value := range answers.Plan {
		scribbleValue(value)
		answers.Plan[name] = model.Value{Kind: model.ValueKindText, Text: "changed"}
	}
	for _, values := range answers.Sections {
		for name, value := range values {
			scribbleValue(value)
			values[name] = model.Value{Kind: model.ValueKindText, Text: "changed"}
		}
		values["added"] = model.Value{Kind: model.ValueKindFlag, Flag: true}
	}
	if answers.Plan != nil {
		answers.Plan["added"] = model.Value{Kind: model.ValueKindFlag, Flag: true}
	}
}

func scribbleAny(value any) {
	switch x := value.(type) {
	case map[string]any:
		if x == nil {
			return
		}
		for key, item := range x {
			scribbleAny(item)
			x[key] = "changed"
		}
		x["added"] = "changed"
	case []any:
		for i := range x {
			scribbleAny(x[i])
			x[i] = "changed"
		}
	}
}

func scribbleRecord(record model.Record) {
	scribbleAnswers(record.Answers)
	for i := range record.Outcomes {
		record.Outcomes[i] = model.Outcome{Section: "changed", Note: "changed"}
	}
	scribbleAny(record.UnknownFields)
}

func TestArgumentsAreNotChanged(t *testing.T) {
	for _, c := range loadCases(t) {
		fresh := loadCaseAgain(t, c)
		split(c)
		if !reflect.DeepEqual(c.plan, fresh.plan) {
			t.Errorf("%s: plan changed", c.name)
		}
		if !reflect.DeepEqual(c.joined, fresh.joined) {
			t.Errorf("%s: joined changed", c.name)
		}
		if !reflect.DeepEqual(c.previous, fresh.previous) {
			t.Errorf("%s: previous changed", c.name)
		}
	}
}

func TestChangingTheResultDoesNotChangeTheArguments(t *testing.T) {
	for _, c := range loadCases(t) {
		scribbleRecord(split(c))
		fresh := loadCaseAgain(t, c)
		if !reflect.DeepEqual(c.plan, fresh.plan) || !reflect.DeepEqual(c.joined, fresh.joined) || !reflect.DeepEqual(c.previous, fresh.previous) {
			t.Errorf("%s: an argument changed after the result was changed", c.name)
		}
	}
}

func TestChangingTheArgumentsDoesNotChangeTheResult(t *testing.T) {
	for _, c := range loadCases(t) {
		got := split(c)
		before := recordText(t, got)
		scribbleAnswers(c.joined.Answers)
		scribbleAnswers(c.plan.Answers)
		scribbleRecord(c.previous)
		if after := recordText(t, got); after != before {
			t.Errorf("%s: the result changed after the arguments were changed\n got: %s\nwant: %s", c.name, after, before)
		}
	}
}

func TestSameCallSameResultOnAnySplitter(t *testing.T) {
	cases := loadCases(t)
	shared := newSplitter()
	forward := map[string]string{}
	for _, c := range cases {
		forward[c.name] = recordText(t, shared.Split(c.plan, c.joined, c.previous))
	}
	for i := len(cases) - 1; i >= 0; i-- {
		c := cases[i]
		if again := recordText(t, shared.Split(c.plan, c.joined, c.previous)); again != forward[c.name] {
			t.Errorf("case %s differs on the shared splitter in reverse order", c.name)
		}
		if again := recordText(t, newSplitter().Split(c.plan, c.joined, c.previous)); again != forward[c.name] {
			t.Errorf("case %s differs on a new splitter", c.name)
		}
	}
}

func TestCallsAtTheSameTimeAgree(t *testing.T) {
	cases := loadCases(t)
	shared := newSplitter()
	want := map[string]string{}
	for _, c := range cases {
		want[c.name] = recordText(t, shared.Split(c.plan, c.joined, c.previous))
	}
	var wg sync.WaitGroup
	for round := 0; round < 4; round++ {
		for _, c := range cases {
			wg.Add(1)
			go func() {
				defer wg.Done()
				got := shared.Split(c.plan, c.joined, c.previous)
				data, err := jsonform.RecordJSON(got)
				if err != nil || string(data) != want[c.name] {
					t.Errorf("case %s differs when called at the same time", c.name)
				}
			}()
		}
	}
	wg.Wait()
}
