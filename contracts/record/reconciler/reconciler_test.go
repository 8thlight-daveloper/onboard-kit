package reconciler_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	contract "onboard-kit/contracts/record/reconciler"
	"onboard-kit/internal/model"
)

type indexEntry struct {
	Name   string `json:"name"`
	File   string `json:"file"`
	Source string `json:"source"`
}

type referenceFile struct {
	Hash string `json:"hash"`
	URL  string `json:"url"`
}

type expectedProblem struct {
	Code     string            `json:"code"`
	Severity string            `json:"severity"`
	Detail   map[string]string `json:"detail"`
}

type expectedReconciliation struct {
	NewQuestions     []string `json:"newQuestions"`
	RemovedQuestions []string `json:"removedQuestions"`
	AnswersToReview  []string `json:"answersToReview"`
	NewSections      []string `json:"newSections"`
	RemovedSections  []string `json:"removedSections"`
	ChangedOutcomes  []string `json:"changedOutcomes"`
}

type caseFile struct {
	Record    json.RawMessage `json:"record"`
	Previous  json.RawMessage `json:"previous"`
	Plan      json.RawMessage `json:"plan"`
	Reference referenceFile   `json:"reference"`
	Expected  struct {
		Record         json.RawMessage        `json:"record"`
		Reconciliation expectedReconciliation `json:"reconciliation"`
		Problems       []expectedProblem      `json:"problems"`
	} `json:"expected"`
}

type loadedCase struct {
	name           string
	record         model.Record
	previous       model.Plan
	plan           model.Plan
	reference      model.PlanReference
	expected       *model.Record
	reconciliation model.Reconciliation
	problems       []expectedProblem
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
	decoder := json.NewDecoder(bytes.NewReader(readFixture(t, entry.File)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil {
		t.Fatalf("%s: %v", entry.File, err)
	}
	c := loadedCase{
		name:      entry.Name,
		record:    readRecord(t, entry.File+" record", file.Record),
		previous:  readPlan(t, entry.File+" previous", file.Previous),
		plan:      readPlan(t, entry.File+" plan", file.Plan),
		reference: model.PlanReference{Hash: file.Reference.Hash, URL: file.Reference.URL},
		reconciliation: model.Reconciliation{
			NewQuestions:     file.Expected.Reconciliation.NewQuestions,
			RemovedQuestions: file.Expected.Reconciliation.RemovedQuestions,
			AnswersToReview:  file.Expected.Reconciliation.AnswersToReview,
			NewSections:      file.Expected.Reconciliation.NewSections,
			RemovedSections:  file.Expected.Reconciliation.RemovedSections,
			ChangedOutcomes:  file.Expected.Reconciliation.ChangedOutcomes,
		},
		problems: file.Expected.Problems,
	}
	if len(file.Expected.Record) > 0 {
		record := readRecord(t, entry.File+" expected", file.Expected.Record)
		c.expected = &record
	}
	return c
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

func normalizedAnswers(answers model.Answers) model.Answers {
	var out model.Answers
	for name, value := range answers.Plan {
		if value.Kind == model.ValueKindNone {
			continue
		}
		if out.Plan == nil {
			out.Plan = map[string]model.Value{}
		}
		out.Plan[name] = normalizedValue(value)
	}
	for id, values := range answers.Sections {
		for name, value := range values {
			if value.Kind == model.ValueKindNone {
				continue
			}
			if out.Sections == nil {
				out.Sections = map[string]map[string]model.Value{}
			}
			if out.Sections[id] == nil {
				out.Sections[id] = map[string]model.Value{}
			}
			out.Sections[id][name] = normalizedValue(value)
		}
	}
	return out
}

func normalizedValue(value model.Value) model.Value {
	switch value.Kind {
	case model.ValueKindText:
		return model.Value{Kind: value.Kind, Text: value.Text}
	case model.ValueKindFlag:
		return model.Value{Kind: value.Kind, Flag: value.Flag}
	case model.ValueKindList:
		return model.Value{Kind: value.Kind, List: append([]string{}, value.List...)}
	}
	return model.Value{Kind: value.Kind}
}

func recordText(t *testing.T, record model.Record) string {
	t.Helper()
	record.Answers = normalizedAnswers(record.Answers)
	data, err := jsonform.RecordJSON(record)
	if err != nil {
		t.Fatalf("write record: %v", err)
	}
	return string(data)
}

func reconciliationLists(r model.Reconciliation) [][]string {
	return [][]string{r.NewQuestions, r.RemovedQuestions, r.AnswersToReview, r.NewSections, r.RemovedSections, r.ChangedOutcomes}
}

func reconciliationText(r model.Reconciliation) string {
	names := []string{"newQuestions", "removedQuestions", "answersToReview", "newSections", "removedSections", "changedOutcomes"}
	var b strings.Builder
	for i, list := range reconciliationLists(r) {
		b.WriteString(names[i] + ": [" + strings.Join(list, ", ") + "]\n")
	}
	return b.String()
}

func reconciliationIsEmpty(r model.Reconciliation) bool {
	for _, list := range reconciliationLists(r) {
		if len(list) > 0 {
			return false
		}
	}
	return true
}

type outcome struct {
	record         string
	reconciliation string
	problems       []model.Problem
}

func run(reconciler contract.Reconciler, c loadedCase) (model.Record, model.Reconciliation, []model.Problem) {
	return reconciler.Reconcile(c.record, c.previous, c.plan, c.reference)
}

func outcomeOf(t *testing.T, reconciler contract.Reconciler, c loadedCase) outcome {
	t.Helper()
	record, reconciliation, problems := run(reconciler, c)
	if len(problems) == 0 {
		problems = nil
	}
	return outcome{record: recordText(t, record), reconciliation: reconciliationText(reconciliation), problems: problems}
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
			record, reconciliation, problems := run(newReconciler(), c)

			want := c.record
			if c.expected != nil {
				want = *c.expected
			}
			if got, wantText := recordText(t, record), recordText(t, want); got != wantText {
				t.Errorf("record\n got: %s\nwant: %s", got, wantText)
			}
			if record.Plan != want.Plan {
				t.Errorf("plan reference: got %+v, want %+v", record.Plan, want.Plan)
			}
			if got, wantText := reconciliationText(reconciliation), reconciliationText(c.reconciliation); got != wantText {
				t.Errorf("reconciliation\n got:\n%swant:\n%s", got, wantText)
			}
			checkProblems(t, problems, c.problems)
		})
	}
}

func checkProblems(t *testing.T, got []model.Problem, want []expectedProblem) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("problems: got %d %+v, want %d", len(got), got, len(want))
	}
	for i, w := range want {
		g := got[i]
		if g.Code != w.Code || string(g.Severity) != w.Severity {
			t.Errorf("problem %d: got %s %s, want %s %s", i, g.Code, g.Severity, w.Code, w.Severity)
		}
		if strings.TrimSpace(g.Message) == "" {
			t.Errorf("problem %d: the message is empty", i)
		}
		if g.Question != "" || g.Section != "" {
			t.Errorf("problem %d: question and section are empty for this component, got %q and %q", i, g.Question, g.Section)
		}
		if len(w.Detail) != len(g.Detail) {
			t.Errorf("problem %d: detail got %v, want %v", i, g.Detail, w.Detail)
			continue
		}
		for key, value := range w.Detail {
			if g.Detail[key] != value {
				t.Errorf("problem %d: detail %s got %q, want %q", i, key, g.Detail[key], value)
			}
		}
	}
}

func TestEveryListIsSortedByBytesWithoutRepeats(t *testing.T) {
	reconciler := newReconciler()
	for _, c := range loadCases(t) {
		_, reconciliation, _ := run(reconciler, c)
		for i, list := range reconciliationLists(reconciliation) {
			if !sort.StringsAreSorted(list) {
				t.Errorf("case %s: list %d is not sorted: %v", c.name, i, list)
			}
			for j := 1; j < len(list); j++ {
				if list[j] == list[j-1] {
					t.Errorf("case %s: list %d repeats %q", c.name, i, list[j])
				}
			}
		}
	}
}

func declaredOnboardeeQuestions(plan model.Plan) map[string]bool {
	declared := map[string]bool{}
	for _, variable := range plan.Variables {
		if variable.Audience == model.AudienceOnboardee {
			declared["plan."+variable.Name] = true
		}
	}
	for _, section := range plan.Sections {
		for _, variable := range section.Variables {
			if variable.Audience == model.AudienceOnboardee {
				declared[section.ID+"."+variable.Name] = true
			}
		}
	}
	return declared
}

func TestAnswersToReviewAreOnboardeeQuestionsOfThePlan(t *testing.T) {
	reconciler := newReconciler()
	for _, c := range loadCases(t) {
		_, reconciliation, _ := run(reconciler, c)
		declared := declaredOnboardeeQuestions(c.plan)
		for _, id := range reconciliation.AnswersToReview {
			if !declared[id] {
				t.Errorf("case %s: %s is listed for review and is not an onboardee question of the plan", c.name, id)
			}
		}
	}
}

func TestArgumentsAreNotChanged(t *testing.T) {
	reconciler := newReconciler()
	for _, c := range loadCases(t) {
		run(reconciler, c)
		fresh := loadCaseAgain(t, c)
		if !reflect.DeepEqual(c.record, fresh.record) {
			t.Errorf("case %s: record changed", c.name)
		}
		if !reflect.DeepEqual(c.previous, fresh.previous) {
			t.Errorf("case %s: previous changed", c.name)
		}
		if !reflect.DeepEqual(c.plan, fresh.plan) {
			t.Errorf("case %s: plan changed", c.name)
		}
		if c.reference != fresh.reference {
			t.Errorf("case %s: reference changed", c.name)
		}
	}
}

func scramble(value any) {
	switch x := value.(type) {
	case map[string]any:
		if x == nil {
			return
		}
		for key, item := range x {
			scramble(item)
			x[key] = "changed"
		}
		x["added"] = "changed"
	case []any:
		for i := range x {
			scramble(x[i])
			x[i] = "changed"
		}
	}
}

func scrambleResult(record model.Record, reconciliation model.Reconciliation) {
	for _, value := range record.Answers.Plan {
		for i := range value.List {
			value.List[i] = "changed"
		}
	}
	for _, values := range record.Answers.Sections {
		for _, value := range values {
			for i := range value.List {
				value.List[i] = "changed"
			}
		}
		for name := range values {
			values[name] = model.Value{Kind: model.ValueKindText, Text: "changed"}
		}
	}
	for name := range record.Answers.Plan {
		record.Answers.Plan[name] = model.Value{Kind: model.ValueKindText, Text: "changed"}
	}
	record.Answers.Plan["added"] = model.Value{Kind: model.ValueKindText, Text: "changed"}
	for i := range record.Outcomes {
		record.Outcomes[i] = model.Outcome{Section: "changed", Note: "changed"}
	}
	scramble(record.UnknownFields)
	for _, list := range reconciliationLists(reconciliation) {
		for i := range list {
			list[i] = "changed"
		}
	}
}

func TestChangingTheResultDoesNotChangeTheArguments(t *testing.T) {
	reconciler := newReconciler()
	for _, c := range loadCases(t) {
		record, reconciliation, _ := run(reconciler, c)
		if record.Answers.Plan == nil {
			record.Answers.Plan = map[string]model.Value{}
		}
		scrambleResult(record, reconciliation)
		fresh := loadCaseAgain(t, c)
		if !reflect.DeepEqual(c.record, fresh.record) {
			t.Errorf("case %s: the record changed after the result was changed", c.name)
		}
		if !reflect.DeepEqual(c.previous, fresh.previous) || !reflect.DeepEqual(c.plan, fresh.plan) {
			t.Errorf("case %s: a plan changed after the result was changed", c.name)
		}
	}
}

func TestSameCallSameResultOnAnyReconciler(t *testing.T) {
	cases := loadCases(t)
	shared := newReconciler()
	forward := map[string]outcome{}
	for _, c := range cases {
		forward[c.name] = outcomeOf(t, shared, c)
	}
	for i := len(cases) - 1; i >= 0; i-- {
		c := cases[i]
		for label, reconciler := range map[string]contract.Reconciler{"shared, reverse order": shared, "new": newReconciler()} {
			again := outcomeOf(t, reconciler, c)
			if !reflect.DeepEqual(forward[c.name], again) {
				t.Errorf("case %s differs on a %s reconciler", c.name, label)
			}
		}
	}
}

func TestCallsAtTheSameTimeAgree(t *testing.T) {
	cases := loadCases(t)
	shared := newReconciler()
	want := map[string]outcome{}
	for _, c := range cases {
		want[c.name] = outcomeOf(t, newReconciler(), c)
	}
	var wait sync.WaitGroup
	for round := 0; round < 4; round++ {
		for _, c := range cases {
			wait.Add(1)
			go func() {
				defer wait.Done()
				record, reconciliation, problems := run(shared, c)
				record.Answers = normalizedAnswers(record.Answers)
				data, err := jsonform.RecordJSON(record)
				if err != nil {
					t.Errorf("case %s: %v", c.name, err)
					return
				}
				if len(problems) == 0 {
					problems = nil
				}
				got := outcome{record: string(data), reconciliation: reconciliationText(reconciliation), problems: problems}
				if !reflect.DeepEqual(want[c.name], got) {
					t.Errorf("case %s: a call at the same time gave another result", c.name)
				}
			}()
		}
	}
	wait.Wait()
}

func TestRefusalChangesNothing(t *testing.T) {
	reconciler := newReconciler()
	refused := 0
	for _, c := range loadCases(t) {
		if c.expected != nil {
			continue
		}
		refused++
		record, reconciliation, problems := run(reconciler, c)
		if recordText(t, record) != recordText(t, c.record) || record.Plan != c.record.Plan {
			t.Errorf("case %s: a refused reconcile returns the record exactly as it was", c.name)
		}
		if !reflect.DeepEqual(record, c.record) {
			t.Errorf("case %s: a refused reconcile returns a record equal to the argument in every part, an absent map staying absent", c.name)
		}
		if !reconciliationIsEmpty(reconciliation) {
			t.Errorf("case %s: a refused reconcile has an empty reconciliation, got %s", c.name, reconciliationText(reconciliation))
		}
		if len(problems) != 1 {
			t.Errorf("case %s: a refusal gives exactly one problem, got %d", c.name, len(problems))
		}
	}
	if refused == 0 {
		t.Fatal("no case refuses")
	}
}

func TestReconcilingAgainWithTheSamePlanChangesOnlyTheReference(t *testing.T) {
	reconciler := newReconciler()
	later := model.PlanReference{Hash: "hash-later", URL: "https://plans.example.test/moved.plan.obk.yaml"}
	for _, c := range loadCases(t) {
		if c.expected == nil {
			continue
		}
		first, firstReconciliation, _ := run(reconciler, c)
		again, reconciliation, problems := reconciler.Reconcile(first, c.plan, c.plan, later)
		if len(problems) != 0 {
			t.Errorf("case %s: problems on the second reconcile: %+v", c.name, problems)
		}
		want := first
		want.Plan = model.PlanReference{ID: c.plan.ID, Hash: later.Hash, URL: later.URL}
		if recordText(t, again) != recordText(t, want) || again.Plan != want.Plan {
			t.Errorf("case %s: the second reconcile changed more than the reference\n got: %s\nwant: %s", c.name, recordText(t, again), recordText(t, want))
		}
		wantReconciliation := model.Reconciliation{ChangedOutcomes: firstReconciliation.ChangedOutcomes}
		if got, wantText := reconciliationText(reconciliation), reconciliationText(wantReconciliation); got != wantText {
			t.Errorf("case %s: the second reconcile lists more than the changed outcomes\n got:\n%swant:\n%s", c.name, got, wantText)
		}
	}
}

func TestResultHoldsNoEntryOfNoKindAndNoEmptySectionMap(t *testing.T) {
	reconciler := newReconciler()
	for _, c := range loadCases(t) {
		record, _, _ := run(reconciler, c)
		for name, value := range record.Answers.Plan {
			if value.Kind == model.ValueKindNone {
				t.Errorf("case %s: plan answer %s has no kind", c.name, name)
			}
		}
		for id, values := range record.Answers.Sections {
			if len(values) == 0 {
				t.Errorf("case %s: section %s holds an empty map", c.name, id)
			}
			for name, value := range values {
				if value.Kind == model.ValueKindNone {
					t.Errorf("case %s: answer %s.%s has no kind", c.name, id, name)
				}
			}
		}
	}
}

func TestRecordWithNoPlanAnswersGivesResultWithNoPlanAnswers(t *testing.T) {
	reconciler := newReconciler()
	checked := 0
	for _, c := range loadCases(t) {
		if c.expected == nil || c.record.Answers.Plan != nil {
			continue
		}
		checked++
		record, _, _ := run(reconciler, c)
		if record.Answers.Plan != nil {
			t.Errorf("case %s: the record holds no plan answer map and the result holds one", c.name)
		}
	}
	if checked == 0 {
		t.Fatal("no case has a record with no plan answers")
	}
}

func TestEveryCaseGivesOneResultWhateverTheMapOrder(t *testing.T) {
	reconciler := newReconciler()
	for _, c := range loadCases(t) {
		first := outcomeOf(t, reconciler, c)
		for round := 0; round < 30; round++ {
			if again := outcomeOf(t, reconciler, c); !reflect.DeepEqual(first, again) {
				t.Errorf("case %s: the result changed between calls", c.name)
				break
			}
		}
	}
}
