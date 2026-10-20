package resolver_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type fixtureProblem struct {
	Code     string `json:"code"`
	Question string `json:"question"`
}

type fixtureCase struct {
	Name     string           `json:"name"`
	Section  string           `json:"section"`
	Answers  string           `json:"answers"`
	Resolved string           `json:"resolved"`
	Problems []fixtureProblem `json:"problems"`
	Source   string           `json:"source"`
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}

func loadCases(t *testing.T) []fixtureCase {
	t.Helper()
	var cases []fixtureCase
	if err := json.Unmarshal(readFixture(t, "cases.json"), &cases); err != nil {
		t.Fatalf("cases.json: %v", err)
	}
	return cases
}

func loadSection(t *testing.T, name string) model.Section {
	t.Helper()
	section, err := jsonform.Section(readFixture(t, name))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return section
}

func loadAnswers(t *testing.T, name string) model.Answers {
	t.Helper()
	answers, err := jsonform.Answers(readFixture(t, name))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return answers
}

func sectionText(t *testing.T, section model.Section) string {
	t.Helper()
	data, err := jsonform.SectionJSON(section)
	if err != nil {
		t.Fatalf("write section: %v", err)
	}
	return string(bytes.TrimSpace(data))
}

func TestEveryCaseCitesItsSource(t *testing.T) {
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
	}
}

func TestFillsTheSectionFromItsOwnAnswers(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			section := loadSection(t, c.Section)
			answers := loadAnswers(t, c.Answers)
			want := sectionText(t, loadSection(t, c.Resolved))
			got, _ := newResolver().Resolve(section, answers)
			if text := sectionText(t, got); text != want {
				t.Errorf("section\ngot:\n%s\nwant:\n%s", text, want)
			}
		})
	}
}

func TestReportsTheVariablesItCouldNotFill(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			section := loadSection(t, c.Section)
			_, problems := newResolver().Resolve(section, loadAnswers(t, c.Answers))
			if len(problems) != len(c.Problems) {
				t.Fatalf("got %d problems, want %d: %+v", len(problems), len(c.Problems), problems)
			}
			for i, want := range c.Problems {
				got := problems[i]
				if got.Code != want.Code || got.Question != want.Question {
					t.Errorf("problem %d: got %s at %s, want %s at %s", i, got.Code, got.Question, want.Code, want.Question)
				}
				if got.Severity != model.SeverityWarning {
					t.Errorf("problem %d: got severity %q, want warning", i, got.Severity)
				}
				if got.Section != section.ID {
					t.Errorf("problem %d: got section %q, want %q", i, got.Section, section.ID)
				}
			}
		})
	}
}

func TestResultsDoNotDependOnEarlierCalls(t *testing.T) {
	t.Parallel()
	type call struct {
		section model.Section
		answers model.Answers
		text    string
		codes   []string
	}
	var calls []call
	for _, c := range loadCases(t) {
		var codes []string
		for _, problem := range c.Problems {
			codes = append(codes, problem.Code+" "+problem.Question)
		}
		calls = append(calls, call{loadSection(t, c.Section), loadAnswers(t, c.Answers), sectionText(t, loadSection(t, c.Resolved)), codes})
	}
	resolver := newResolver()
	for pass := 0; pass < 2; pass++ {
		for i := range calls {
			index := i
			if pass == 1 {
				index = len(calls) - 1 - i
			}
			got, problems := resolver.Resolve(calls[index].section, calls[index].answers)
			if text := sectionText(t, got); text != calls[index].text {
				t.Fatalf("call %d on pass %d: section differs\ngot:\n%s\nwant:\n%s", index, pass, text, calls[index].text)
			}
			var codes []string
			for _, problem := range problems {
				codes = append(codes, problem.Code+" "+problem.Question)
			}
			if !reflect.DeepEqual(codes, calls[index].codes) {
				t.Fatalf("call %d on pass %d: problems %v, want %v", index, pass, codes, calls[index].codes)
			}
		}
	}
}

func TestInputsAreNotChanged(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		section := loadSection(t, c.Section)
		answers := loadAnswers(t, c.Answers)
		sectionSnapshot := loadSection(t, c.Section)
		answersSnapshot := loadAnswers(t, c.Answers)
		newResolver().Resolve(section, answers)
		if !reflect.DeepEqual(section, sectionSnapshot) {
			t.Fatalf("%s: the section was changed", c.Name)
		}
		if !reflect.DeepEqual(answers, answersSnapshot) {
			t.Fatalf("%s: the answers were changed", c.Name)
		}
	}
}

func TestResultKeepsEveryVariableExample(t *testing.T) {
	t.Parallel()
	examples := 0
	for _, c := range loadCases(t) {
		section := loadSection(t, c.Section)
		got, _ := newResolver().Resolve(section, loadAnswers(t, c.Answers))
		if len(got.Variables) != len(section.Variables) {
			t.Fatalf("%s: got %d variables, want %d", c.Name, len(got.Variables), len(section.Variables))
		}
		for i, variable := range section.Variables {
			if got.Variables[i].Example != variable.Example {
				t.Errorf("%s: variable %s: got example %q, want %q", c.Name, variable.Name, got.Variables[i].Example, variable.Example)
			}
			if variable.Example != "" {
				examples++
			}
		}
	}
	if examples == 0 {
		t.Fatal("no fixture variable holds an example")
	}
}

func TestResultSharesNothingWithTheInput(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		section := loadSection(t, c.Section)
		snapshot := loadSection(t, c.Section)
		got, _ := newResolver().Resolve(section, loadAnswers(t, c.Answers))
		for i, variable := range section.Variables {
			if got.Variables[i].Example != variable.Example {
				t.Fatalf("%s: variable %s: got example %q, want %q", c.Name, variable.Name, got.Variables[i].Example, variable.Example)
			}
		}
		overwrite(reflect.ValueOf(&got).Elem())
		if !reflect.DeepEqual(section, snapshot) {
			t.Fatalf("%s: changing the result changed the input section", c.Name)
		}
	}
}

func TestConditionTestsAreNotSharedWithTheInput(t *testing.T) {
	t.Parallel()
	section := loadSection(t, "settled-conditions.section.json")
	snapshot := loadSection(t, "settled-conditions.section.json")
	got, _ := newResolver().Resolve(section, loadAnswers(t, "settled-conditions.answers.json"))

	var inputTests, resultTests []*bool
	collect := func(tests *[]*bool, condition *model.Condition) {
		if condition == nil {
			t.Fatal("a condition is missing")
		}
		for _, clause := range condition.All {
			if clause.Answer.IsSettled == nil {
				t.Fatal("an isSettled test is missing")
			}
			*tests = append(*tests, clause.Answer.IsSettled)
		}
	}
	collect(&inputTests, section.IncludeIf)
	collect(&resultTests, got.IncludeIf)
	for i := range section.Variables {
		collect(&inputTests, section.Variables[i].AskIf)
		collect(&resultTests, got.Variables[i].AskIf)
	}
	if len(inputTests) != 4 || len(resultTests) != 4 {
		t.Fatalf("got %d input tests and %d result tests, want 4 each", len(inputTests), len(resultTests))
	}
	for i := range resultTests {
		if *resultTests[i] != *inputTests[i] {
			t.Errorf("isSettled test %d: got %v, want %v", i, *resultTests[i], *inputTests[i])
		}
		*resultTests[i] = !*resultTests[i]
	}
	if !reflect.DeepEqual(section, snapshot) {
		t.Fatal("changing an isSettled test in the result changed the input section")
	}
}

func TestSwitchAndSectionClausesAreNotSharedWithTheInput(t *testing.T) {
	t.Parallel()
	section := loadSection(t, "switch-clauses.section.json")
	snapshot := loadSection(t, "switch-clauses.section.json")
	got, _ := newResolver().Resolve(section, loadAnswers(t, "switch-clauses.answers.json"))

	if section.Enabled == nil || *section.Enabled {
		t.Fatal("the fixture section must hold enabled false")
	}
	if got.Enabled == nil {
		t.Fatal("the switch was dropped")
	}
	if *got.Enabled {
		t.Error("the switch was turned on")
	}
	if got.Enabled == section.Enabled {
		t.Error("the switch is the same pointer as the input's")
	}
	*got.Enabled = true

	var inputClauses, resultClauses []model.Clause
	inputClauses = append(inputClauses, section.IncludeIf.All...)
	resultClauses = append(resultClauses, got.IncludeIf.All...)
	for i := range section.Variables {
		if section.Variables[i].AskIf != nil {
			inputClauses = append(inputClauses, section.Variables[i].AskIf.All...)
			resultClauses = append(resultClauses, got.Variables[i].AskIf.All...)
		}
	}
	if len(inputClauses) != 5 || len(resultClauses) != 5 {
		t.Fatalf("got %d input clauses and %d result clauses, want 5 each", len(inputClauses), len(resultClauses))
	}
	for i := range resultClauses {
		if !reflect.DeepEqual(resultClauses[i], inputClauses[i]) {
			t.Errorf("clause %d: got %+v, want %+v", i, resultClauses[i], inputClauses[i])
		}
	}
	for i := range resultClauses {
		resultClauses[i].Enabled = !resultClauses[i].Enabled
		resultClauses[i].Section = "overwritten"
	}
	if !reflect.DeepEqual(section, snapshot) {
		t.Fatal("changing the switch or a section clause in the result changed the input section")
	}
}

func overwrite(v reflect.Value) {
	switch v.Kind() {
	case reflect.String:
		if v.CanSet() {
			v.SetString("overwritten")
		}
	case reflect.Bool:
		if v.CanSet() {
			v.SetBool(!v.Bool())
		}
	case reflect.Pointer:
		if !v.IsNil() {
			overwrite(v.Elem())
		}
	case reflect.Interface:
		if !v.IsNil() {
			overwriteReference(v.Elem())
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			overwrite(v.Field(i))
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			overwrite(v.Index(i))
		}
	case reflect.Map:
		overwriteReference(v)
	}
}

func overwriteReference(v reflect.Value) {
	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			entry := v.MapIndex(key)
			if entry.Kind() == reflect.Interface && !entry.IsNil() {
				if entry.Elem().Kind() == reflect.String {
					v.SetMapIndex(key, reflect.ValueOf("overwritten"))
				} else {
					overwriteReference(entry.Elem())
				}
			}
			if entry.Kind() == reflect.String {
				v.SetMapIndex(key, reflect.ValueOf("overwritten"))
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if item.Kind() == reflect.Interface && !item.IsNil() {
				overwriteReference(item.Elem())
			}
			if item.Kind() == reflect.String {
				item.SetString("overwritten")
			}
			if item.Kind() == reflect.Interface && item.Elem().Kind() == reflect.String {
				item.Set(reflect.ValueOf("overwritten"))
			}
		}
	}
}

func TestZeroSectionAndAnswersGiveTheZeroSection(t *testing.T) {
	t.Parallel()
	got, problems := newResolver().Resolve(model.Section{}, model.Answers{})
	if text, want := sectionText(t, got), sectionText(t, model.Section{}); text != want {
		t.Errorf("got %s, want %s", text, want)
	}
	if len(problems) != 0 {
		t.Errorf("got problems %+v, want none", problems)
	}
}
