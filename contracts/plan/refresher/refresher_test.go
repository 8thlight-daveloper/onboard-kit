package refresher_test

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
	contract "onboard-kit/contracts/plan/refresher"
	"onboard-kit/internal/model"
)

type indexEntry struct {
	Name   string `json:"name"`
	File   string `json:"file"`
	Source string `json:"source"`
}

type originFile struct {
	PlanID      string `json:"planId"`
	Source      string `json:"source"`
	ContentHash string `json:"contentHash"`
	RefreshedOn string `json:"refreshedOn"`
}

type expectedProblem struct {
	Code     string            `json:"code"`
	Severity string            `json:"severity"`
	Detail   map[string]string `json:"detail"`
}

type expectedReport struct {
	NewQuestions     []string `json:"newQuestions"`
	RemovedQuestions []string `json:"removedQuestions"`
	AnswersTaken     []string `json:"answersTaken"`
	AnswersKept      []string `json:"answersKept"`
	AnswersToReview  []string `json:"answersToReview"`
	NewSections      []string `json:"newSections"`
	RemovedSections  []string `json:"removedSections"`
}

type caseFile struct {
	Base     json.RawMessage `json:"base"`
	Theirs   json.RawMessage `json:"theirs"`
	Ours     json.RawMessage `json:"ours"`
	Origin   originFile      `json:"origin"`
	Expected struct {
		Plan     json.RawMessage   `json:"plan"`
		Report   expectedReport    `json:"report"`
		Problems []expectedProblem `json:"problems"`
	} `json:"expected"`
}

type loadedCase struct {
	name     string
	base     model.Plan
	theirs   model.Plan
	ours     model.Plan
	origin   model.Origin
	expected *model.Plan
	report   contract.Report
	problems []expectedProblem
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

func loadCase(t *testing.T, entry indexEntry) loadedCase {
	t.Helper()
	var file caseFile
	decoder := json.NewDecoder(bytes.NewReader(readFixture(t, entry.File)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil {
		t.Fatalf("%s: %v", entry.File, err)
	}
	c := loadedCase{
		name:   entry.Name,
		base:   readPlan(t, entry.File+" base", file.Base),
		theirs: readPlan(t, entry.File+" theirs", file.Theirs),
		ours:   readPlan(t, entry.File+" ours", file.Ours),
		origin: model.Origin{
			PlanID:      file.Origin.PlanID,
			Source:      file.Origin.Source,
			ContentHash: file.Origin.ContentHash,
			RefreshedOn: model.Date(file.Origin.RefreshedOn),
		},
		report: contract.Report{
			NewQuestions:     file.Expected.Report.NewQuestions,
			RemovedQuestions: file.Expected.Report.RemovedQuestions,
			AnswersTaken:     file.Expected.Report.AnswersTaken,
			AnswersKept:      file.Expected.Report.AnswersKept,
			AnswersToReview:  file.Expected.Report.AnswersToReview,
			NewSections:      file.Expected.Report.NewSections,
			RemovedSections:  file.Expected.Report.RemovedSections,
		},
		problems: file.Expected.Problems,
	}
	if len(file.Expected.Plan) > 0 {
		plan := readPlan(t, entry.File+" expected", file.Expected.Plan)
		c.expected = &plan
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
		list := append([]string{}, value.List...)
		return model.Value{Kind: value.Kind, List: list}
	}
	return model.Value{Kind: value.Kind}
}

func planText(t *testing.T, plan model.Plan) string {
	t.Helper()
	plan.Answers = normalizedAnswers(plan.Answers)
	data, err := jsonform.JSON(plan)
	if err != nil {
		t.Fatalf("write plan: %v", err)
	}
	return string(data)
}

func reportText(report contract.Report) string {
	var b strings.Builder
	for _, list := range []struct {
		name  string
		items []string
	}{
		{"newQuestions", report.NewQuestions},
		{"removedQuestions", report.RemovedQuestions},
		{"answersTaken", report.AnswersTaken},
		{"answersKept", report.AnswersKept},
		{"answersToReview", report.AnswersToReview},
		{"newSections", report.NewSections},
		{"removedSections", report.RemovedSections},
	} {
		b.WriteString(list.name + ": [" + strings.Join(list.items, ", ") + "]\n")
	}
	return b.String()
}

func reportLists(report contract.Report) [][]string {
	return [][]string{report.NewQuestions, report.RemovedQuestions, report.AnswersTaken, report.AnswersKept, report.AnswersToReview, report.NewSections, report.RemovedSections}
}

func reportIsEmpty(report contract.Report) bool {
	for _, list := range reportLists(report) {
		if len(list) > 0 {
			return false
		}
	}
	return true
}

type outcome struct {
	plan     string
	report   string
	problems []model.Problem
}

func run(t *testing.T, refresher contract.Refresher, c loadedCase) (model.Plan, contract.Report, []model.Problem) {
	t.Helper()
	return refresher.Refresh(c.base, c.theirs, c.ours, c.origin)
}

func outcomeOf(t *testing.T, refresher contract.Refresher, c loadedCase) outcome {
	t.Helper()
	plan, report, problems := run(t, refresher, c)
	if len(problems) == 0 {
		problems = nil
	}
	return outcome{plan: planText(t, plan), report: reportText(report), problems: problems}
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
			plan, report, problems := run(t, newRefresher(), c)

			want := c.ours
			if c.expected != nil {
				want = *c.expected
			}
			if got, wantText := planText(t, plan), planText(t, want); got != wantText {
				t.Errorf("plan\n got: %s\nwant: %s", got, wantText)
			}
			if !reflect.DeepEqual(plan.Origin, want.Origin) {
				t.Errorf("origin: got %+v, want %+v", plan.Origin, want.Origin)
			}
			if plan.ID != want.ID || plan.SchemaVersion != want.SchemaVersion || plan.Provenance != want.Provenance {
				t.Errorf("identity: got %q %d %+v, want %q %d %+v", plan.ID, plan.SchemaVersion, plan.Provenance, want.ID, want.SchemaVersion, want.Provenance)
			}
			if got, wantText := reportText(report), reportText(c.report); got != wantText {
				t.Errorf("report\n got:\n%swant:\n%s", got, wantText)
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

func TestReportListsAreSortedByBytesWithoutRepeats(t *testing.T) {
	refresher := newRefresher()
	for _, c := range loadCases(t) {
		_, report, _ := run(t, refresher, c)
		for i, list := range reportLists(report) {
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

func TestAQuestionIsInOneOfTheAnswerLists(t *testing.T) {
	refresher := newRefresher()
	for _, c := range loadCases(t) {
		_, report, _ := run(t, refresher, c)
		seen := map[string]string{}
		for name, list := range map[string][]string{"taken": report.AnswersTaken, "kept": report.AnswersKept, "review": report.AnswersToReview} {
			for _, id := range list {
				if other, ok := seen[id]; ok {
					t.Errorf("case %s: %s is in %s and in %s", c.name, id, other, name)
				}
				seen[id] = name
			}
		}
	}
}

func TestArgumentsAreNotChanged(t *testing.T) {
	refresher := newRefresher()
	for _, c := range loadCases(t) {
		fresh := loadCaseAgain(t, c)
		run(t, refresher, c)
		for name, pair := range map[string][2]model.Plan{"base": {c.base, fresh.base}, "theirs": {c.theirs, fresh.theirs}, "ours": {c.ours, fresh.ours}} {
			if !reflect.DeepEqual(pair[0], pair[1]) {
				t.Errorf("case %s: %s changed", c.name, name)
			}
		}
		if !reflect.DeepEqual(c.origin, fresh.origin) {
			t.Errorf("case %s: origin changed", c.name)
		}
	}
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

func TestChangingTheResultDoesNotChangeTheArguments(t *testing.T) {
	refresher := newRefresher()
	for _, c := range loadCases(t) {
		plan, _, _ := run(t, refresher, c)
		for name, values := range plan.Answers.Plan {
			for i := range values.List {
				values.List[i] = "changed"
			}
			plan.Answers.Plan[name] = model.Value{Kind: model.ValueKindText, Text: "changed"}
		}
		for _, values := range plan.Answers.Sections {
			for name, value := range values {
				for i := range value.List {
					value.List[i] = "changed"
				}
				values[name] = model.Value{Kind: model.ValueKindText, Text: "changed"}
			}
		}
		plan.Answers.Plan = nil
		plan.Answers.Sections = nil
		fresh := loadCaseAgain(t, c)
		for name, pair := range map[string][2]model.Plan{"base": {c.base, fresh.base}, "theirs": {c.theirs, fresh.theirs}, "ours": {c.ours, fresh.ours}} {
			if !reflect.DeepEqual(pair[0], pair[1]) {
				t.Errorf("case %s: %s changed after the result was changed", c.name, name)
			}
		}
	}
}

func TestSameCallSameResultOnAnyRefresher(t *testing.T) {
	cases := loadCases(t)
	shared := newRefresher()
	forward := map[string]outcome{}
	for _, c := range cases {
		forward[c.name] = outcomeOf(t, shared, c)
	}
	for i := len(cases) - 1; i >= 0; i-- {
		c := cases[i]
		for label, refresher := range map[string]contract.Refresher{"shared, reverse order": shared, "new": newRefresher()} {
			again := outcomeOf(t, refresher, c)
			if !reflect.DeepEqual(forward[c.name], again) {
				t.Errorf("case %s differs on a %s refresher", c.name, label)
			}
		}
	}
}

func TestCallsAtTheSameTimeAgree(t *testing.T) {
	cases := loadCases(t)
	shared := newRefresher()
	want := map[string]outcome{}
	for _, c := range cases {
		want[c.name] = outcomeOf(t, newRefresher(), c)
	}
	var wait sync.WaitGroup
	for round := 0; round < 4; round++ {
		for _, c := range cases {
			wait.Add(1)
			go func() {
				defer wait.Done()
				plan, report, problems := shared.Refresh(c.base, c.theirs, c.ours, c.origin)
				plan.Answers = normalizedAnswers(plan.Answers)
				data, err := jsonform.JSON(plan)
				if err != nil {
					t.Errorf("case %s: %v", c.name, err)
					return
				}
				if len(problems) == 0 {
					problems = nil
				}
				got := outcome{plan: string(data), report: reportText(report), problems: problems}
				if !reflect.DeepEqual(want[c.name], got) {
					t.Errorf("case %s: a call at the same time gave another result", c.name)
				}
			}()
		}
	}
	wait.Wait()
}

func TestRefusalChangesNothing(t *testing.T) {
	refresher := newRefresher()
	refused := 0
	for _, c := range loadCases(t) {
		if c.expected != nil {
			continue
		}
		refused++
		plan, report, problems := run(t, refresher, c)
		if !reflect.DeepEqual(plan, c.ours) {
			t.Errorf("case %s: a refused refresh must return ours exactly", c.name)
		}
		if !reportIsEmpty(report) {
			t.Errorf("case %s: a refused refresh has an empty report, got %s", c.name, reportText(report))
		}
		if len(problems) == 0 {
			t.Errorf("case %s: a refusal gives a problem", c.name)
		}
	}
	if refused == 0 {
		t.Fatal("no case refuses")
	}
}

func TestRefreshingAgainWithTheSameOrgPlanChangesOnlyTheOrigin(t *testing.T) {
	refresher := newRefresher()
	later := model.Origin{Source: "https://plans.example.test/moved.org.obk.yaml", ContentHash: "hash-later", RefreshedOn: "2026-11-01"}
	for _, c := range loadCases(t) {
		if c.expected == nil {
			continue
		}
		first, _, _ := run(t, refresher, c)
		again, report, problems := refresher.Refresh(c.theirs, c.theirs, first, later)
		if len(problems) != 0 {
			t.Errorf("case %s: problems on the second refresh: %+v", c.name, problems)
		}
		want := first
		want.Origin = model.Origin{PlanID: c.theirs.ID, Source: later.Source, ContentHash: later.ContentHash, RefreshedOn: later.RefreshedOn}
		if planText(t, again) != planText(t, want) || !reflect.DeepEqual(again.Origin, want.Origin) {
			t.Errorf("case %s: the second refresh changed more than the origin\n got: %s\nwant: %s", c.name, planText(t, again), planText(t, want))
		}
		if !reportIsEmpty(contract.Report{
			NewQuestions: report.NewQuestions, RemovedQuestions: report.RemovedQuestions, AnswersTaken: report.AnswersTaken,
			AnswersToReview: report.AnswersToReview, NewSections: report.NewSections, RemovedSections: report.RemovedSections,
		}) {
			t.Errorf("case %s: the second refresh reports changes: %s", c.name, reportText(report))
		}
	}
}

func TestResultHoldsNoEntryOfNoKindAndNoEmptySectionMap(t *testing.T) {
	refresher := newRefresher()
	for _, c := range loadCases(t) {
		plan, _, _ := run(t, refresher, c)
		for name, value := range plan.Answers.Plan {
			if value.Kind == model.ValueKindNone {
				t.Errorf("case %s: plan answer %s has no kind", c.name, name)
			}
		}
		for id, values := range plan.Answers.Sections {
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
