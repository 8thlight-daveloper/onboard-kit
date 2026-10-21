package validator_test

import (
	"bytes"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type expectedProblem struct {
	Code            string            `json:"code"`
	Severity        string            `json:"severity"`
	Question        string            `json:"question"`
	Section         string            `json:"section"`
	Detail          map[string]string `json:"detail"`
	MessageContains string            `json:"messageContains"`
}

type expectedFile struct {
	Problems        []expectedProblem `json:"problems"`
	SecretFragments []string          `json:"secretFragments"`
}

type fixtureCheck struct {
	Name            string            `json:"name"`
	Question        string            `json:"question"`
	Value           json.RawMessage   `json:"value"`
	Problems        []expectedProblem `json:"problems"`
	SecretFragments []string          `json:"secretFragments"`
}

type fixtureCase struct {
	Name     string         `json:"name"`
	Plan     string         `json:"plan"`
	Expected string         `json:"expected"`
	Checks   []fixtureCheck `json:"checks"`
	Source   string         `json:"source"`
}

const markerCharacter = "¤"

var answerCodes = []string{
	"answer.wrong-kind",
	"answer.not-an-option",
	"answer.pattern-mismatch",
	"answer.thin-contact",
	"answer.single-person",
	"answer.unknown-question",
	"answer.inactive-question",
	"secret.token-prefix",
	"secret.url-credential",
	"secret.private-key",
	"directory.duplicate-id",
	"directory.holder-not-a-contact",
	"directory.team-role-unknown",
	"directory.unknown-id",
	"directory.role-vacant",
	"directory.email-malformed",
	"directory.contact-unreachable",
	"plan.duplicate-section-id",
	"plan.duplicate-team-variable",
}

var judgedPlanCodes = []string{"plan.duplicate-section-id", "plan.duplicate-team-variable"}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return bytes.ReplaceAll(data, []byte(markerCharacter), nil)
}

func loadCases(t *testing.T) []fixtureCase {
	t.Helper()
	var cases []fixtureCase
	if err := json.Unmarshal(readFixture(t, "cases.json"), &cases); err != nil {
		t.Fatalf("cases.json: %v", err)
	}
	return cases
}

func loadPlan(t *testing.T, c fixtureCase) model.Plan {
	t.Helper()
	plan, err := jsonform.Plan(readFixture(t, c.Plan))
	if err != nil {
		t.Fatalf("%s: %v", c.Plan, err)
	}
	return plan
}

func loadExpected(t *testing.T, c fixtureCase) expectedFile {
	t.Helper()
	var expected expectedFile
	decoder := json.NewDecoder(bytes.NewReader(readFixture(t, c.Expected)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&expected); err != nil {
		t.Fatalf("%s: %v", c.Expected, err)
	}
	return expected
}

func loadValue(t *testing.T, check fixtureCheck) model.Value {
	t.Helper()
	if len(check.Value) == 0 {
		return model.Value{}
	}
	value, err := jsonform.Value(check.Value)
	if err != nil {
		t.Fatalf("%s: %v", check.Name, err)
	}
	return value
}

func casesWithExpectedProblems(cases []fixtureCase) []fixtureCase {
	return slices.DeleteFunc(slices.Clone(cases), func(c fixtureCase) bool { return c.Expected == "" })
}

func judgedProblems(problems []model.Problem) []model.Problem {
	kept := []model.Problem{}
	for _, problem := range problems {
		if strings.HasPrefix(problem.Code, "answer.") || strings.HasPrefix(problem.Code, "secret.") ||
			strings.HasPrefix(problem.Code, "directory.") || slices.Contains(judgedPlanCodes, problem.Code) {
			kept = append(kept, problem)
		}
	}
	return kept
}

func describe(problems []model.Problem) []string {
	described := []string{}
	for _, problem := range problems {
		detail := []string{}
		for _, key := range slices.Sorted(maps.Keys(problem.Detail)) {
			detail = append(detail, key+"="+problem.Detail[key])
		}
		described = append(described, strings.Join([]string{
			problem.Code, string(problem.Severity), "question=" + problem.Question, "section=" + problem.Section, "detail[" + strings.Join(detail, ",") + "]",
		}, " "))
	}
	return described
}

func describeExpected(problems []expectedProblem) []string {
	converted := []model.Problem{}
	for _, problem := range problems {
		converted = append(converted, model.Problem{
			Code: problem.Code, Severity: model.Severity(problem.Severity), Question: problem.Question, Section: problem.Section, Detail: problem.Detail,
		})
	}
	return describe(converted)
}

func sameProblemsIgnoringMessages(left, right []model.Problem) bool {
	return slices.Equal(describe(left), describe(right))
}

func assertProblems(t *testing.T, got []model.Problem, want []expectedProblem, fragments []string) {
	t.Helper()
	got = judgedProblems(got)
	if !slices.Equal(describe(got), describeExpected(want)) {
		t.Errorf("got\n  %s\nwant\n  %s", strings.Join(describe(got), "\n  "), strings.Join(describeExpected(want), "\n  "))
		return
	}
	for i, problem := range got {
		if strings.TrimSpace(problem.Message) == "" {
			t.Errorf("problem %d, %s, has no message", i, problem.Code)
		}
		if want[i].MessageContains != "" && !strings.Contains(problem.Message, want[i].MessageContains) {
			t.Errorf("problem %d, %s: message %q does not hold %q", i, problem.Code, problem.Message, want[i].MessageContains)
		}
		for _, fragment := range fragments {
			shown := []string{problem.Code, problem.Message, problem.Question, problem.Section}
			for key, value := range problem.Detail {
				shown = append(shown, key, value)
			}
			for _, text := range shown {
				if strings.Contains(text, fragment) {
					t.Errorf("problem %d, %s, repeats part of the secret in %q", i, problem.Code, text)
				}
			}
		}
	}
}

func TestEveryCaseCitesItsSourceAndTheCasesCoverEveryCode(t *testing.T) {
	names := map[string]bool{}
	codes := map[string]bool{}
	severities := map[string]bool{}
	details := map[string]bool{}
	note := func(problems []expectedProblem) {
		for _, problem := range problems {
			codes[problem.Code] = true
			severities[problem.Severity] = true
			for key := range problem.Detail {
				details[key] = true
			}
		}
	}
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
		if names[c.Name] {
			t.Errorf("case %s is listed twice", c.Name)
		}
		names[c.Name] = true
		loadPlan(t, c)
		if c.Expected != "" {
			note(loadExpected(t, c).Problems)
		}
		for _, check := range c.Checks {
			note(check.Problems)
		}
		if c.Expected == "" && len(c.Checks) == 0 {
			t.Errorf("case %s checks nothing", c.Name)
		}
	}
	for _, code := range answerCodes {
		if !codes[code] {
			t.Errorf("no case expects the code %s", code)
		}
	}
	for _, severity := range []string{"error", "warning"} {
		if !severities[severity] {
			t.Errorf("no case expects a problem of severity %s", severity)
		}
	}
	for _, key := range []string{"item", "expected", "found", "kind", "id", "contact", "role", "team", "holder", "variables"} {
		if !details[key] {
			t.Errorf("no case expects the detail %s", key)
		}
	}
}

func TestValidateReportsTheStoredAnswersThatAreNotAcceptable(t *testing.T) {
	for _, c := range casesWithExpectedProblems(loadCases(t)) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			expected := loadExpected(t, c)
			assertProblems(t, newValidator().Validate(loadPlan(t, c)), expected.Problems, expected.SecretFragments)
		})
	}
}

func TestCheckValueReportsTheProblemsOfOneValue(t *testing.T) {
	for _, c := range loadCases(t) {
		for _, check := range c.Checks {
			t.Run(c.Name+"/"+check.Name, func(t *testing.T) {
				t.Parallel()
				got := newValidator().CheckValue(loadPlan(t, c), check.Question, loadValue(t, check))
				assertProblems(t, got, check.Problems, check.SecretFragments)
			})
		}
	}
}

func storedAnswers(plan model.Plan) map[string]model.Value {
	stored := map[string]model.Value{}
	for name, value := range plan.Answers.Plan {
		stored["plan."+name] = value
	}
	for section, answers := range plan.Answers.Sections {
		for name, value := range answers {
			stored[section+"."+name] = value
		}
	}
	return stored
}

func problemsOfQuestion(problems []model.Problem, question string) []model.Problem {
	found := []model.Problem{}
	for _, problem := range problems {
		if problem.Question == question {
			found = append(found, problem)
		}
	}
	return found
}

func asStored(problems []model.Problem) []model.Problem {
	converted := slices.Clone(problems)
	for i := range converted {
		if converted[i].Code == "directory.unknown-id" {
			converted[i].Severity = model.SeverityWarning
		}
	}
	return converted
}

func TestCheckValueAgreesWithValidateAboutEachStoredAnswer(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		plan := loadPlan(t, c)
		validator := newValidator()
		reported := judgedProblems(validator.Validate(plan))
		for question, value := range storedAnswers(plan) {
			checked := judgedProblems(validator.CheckValue(plan, question, value))
			inReport := problemsOfQuestion(reported, question)
			if len(checked) == 1 && (checked[0].Code == "answer.inactive-question" || checked[0].Code == "answer.unknown-question") {
				for _, problem := range inReport {
					if !strings.HasPrefix(problem.Code, "secret.") && problem.Code != "answer.unknown-question" {
						t.Errorf("%s, %s: not an active question, but Validate reports %s", c.Name, question, problem.Code)
					}
				}
				continue
			}
			for _, problem := range checked {
				if problem.Code == "directory.unknown-id" && problem.Severity != model.SeverityError {
					t.Errorf("%s, %s: CheckValue gives %s with severity %s, want error", c.Name, question, problem.Code, problem.Severity)
				}
			}
			for _, problem := range inReport {
				if problem.Code == "directory.unknown-id" && problem.Severity != model.SeverityWarning {
					t.Errorf("%s, %s: Validate gives %s with severity %s, want warning", c.Name, question, problem.Code, problem.Severity)
				}
			}
			if !sameProblemsIgnoringMessages(asStored(checked), inReport) {
				t.Errorf("%s, %s: CheckValue gives\n  %s\nValidate gives\n  %s", c.Name, question,
					strings.Join(describe(checked), "\n  "), strings.Join(describe(inReport), "\n  "))
			}
		}
	}
}

func TestPlanAndValueAreNotChanged(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		plan := loadPlan(t, c)
		snapshot := loadPlan(t, c)
		validator := newValidator()
		validator.Validate(plan)
		for _, check := range c.Checks {
			value := loadValue(t, check)
			valueSnapshot := loadValue(t, check)
			validator.CheckValue(plan, check.Question, value)
			if !reflect.DeepEqual(value, valueSnapshot) {
				t.Fatalf("%s, %s: the value was changed", c.Name, check.Name)
			}
		}
		if !reflect.DeepEqual(plan, snapshot) {
			t.Fatalf("%s: the plan was changed", c.Name)
		}
	}
}

type results struct {
	validated []model.Problem
	checked   [][]model.Problem
}

func resultsOf(t *testing.T, validator interface {
	Validate(model.Plan) []model.Problem
	CheckValue(model.Plan, string, model.Value) []model.Problem
}, c fixtureCase) results {
	plan := loadPlan(t, c)
	r := results{validated: validator.Validate(plan)}
	for _, check := range c.Checks {
		r.checked = append(r.checked, validator.CheckValue(plan, check.Question, loadValue(t, check)))
	}
	return r
}

func TestResultsDoNotDependOnEarlierCallsOrOnRepeats(t *testing.T) {
	t.Parallel()
	cases := loadCases(t)
	shared := newValidator()
	first := map[string]results{}
	for _, c := range cases {
		first[c.Name] = resultsOf(t, newValidator(), c)
	}
	for pass := 0; pass < 2; pass++ {
		for i := range cases {
			index := i
			if pass == 1 {
				index = len(cases) - 1 - i
			}
			c := cases[index]
			for repeat := 0; repeat < 10; repeat++ {
				if got := resultsOf(t, shared, c); !reflect.DeepEqual(got, first[c.Name]) {
					t.Fatalf("%s on pass %d, repeat %d: the result changed", c.Name, pass, repeat)
				}
			}
		}
	}
}

func TestOneValidatorServesCallsAtTheSameTime(t *testing.T) {
	t.Parallel()
	cases := loadCases(t)
	shared := newValidator()
	first := map[string]results{}
	for _, c := range cases {
		first[c.Name] = resultsOf(t, newValidator(), c)
	}
	planData := map[string][]byte{}
	for _, c := range cases {
		planData[c.Name] = readFixture(t, c.Plan)
	}
	var group sync.WaitGroup
	failures := make(chan string, len(cases)*8)
	for worker := 0; worker < 8; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for _, c := range cases {
				plan, err := jsonform.Plan(planData[c.Name])
				if err != nil {
					failures <- err.Error()
					return
				}
				if got := shared.Validate(plan); !reflect.DeepEqual(got, first[c.Name].validated) {
					failures <- c.Name + ": the result changed"
				}
			}
		}()
	}
	group.Wait()
	close(failures)
	for failure := range failures {
		t.Error(failure)
	}
}

func TestResultDoesNotDependOnTheOrderOfTheSections(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		plan := loadPlan(t, c)
		want := resultsOf(t, newValidator(), c)
		count := len(plan.Sections)
		variants := []func(*model.Plan){}
		for turn := 1; turn < count; turn++ {
			variants = append(variants, func(p *model.Plan) {
				p.Sections = append(slices.Clone(p.Sections[turn:]), p.Sections[:turn]...)
			})
		}
		variants = append(variants, func(p *model.Plan) { slices.Reverse(p.Sections) })
		for _, variant := range variants {
			changed := loadPlan(t, c)
			variant(&changed)
			if got := newValidator().Validate(changed); !reflect.DeepEqual(got, want.validated) {
				t.Errorf("%s: sections reordered, got\n  %s\nwant\n  %s", c.Name, strings.Join(describe(got), "\n  "), strings.Join(describe(want.validated), "\n  "))
			}
			for i, check := range c.Checks {
				if got := newValidator().CheckValue(changed, check.Question, loadValue(t, check)); !reflect.DeepEqual(got, want.checked[i]) {
					t.Errorf("%s, %s: sections reordered, the result changed", c.Name, check.Name)
				}
			}
		}
	}
}

func TestResultDoesNotDependOnTheOrderOfTheDirectory(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		want := resultsOf(t, newValidator(), c)
		changed := loadPlan(t, c)
		slices.Reverse(changed.Directory.Contacts)
		slices.Reverse(changed.Directory.Roles)
		slices.Reverse(changed.Directory.Teams)
		for _, role := range changed.Directory.Roles {
			slices.Reverse(role.Holders)
		}
		for _, team := range changed.Directory.Teams {
			for _, holders := range team.Holders {
				slices.Reverse(holders)
			}
		}
		if got := newValidator().Validate(changed); !reflect.DeepEqual(got, want.validated) {
			t.Errorf("%s: directory reversed, got\n  %s\nwant\n  %s", c.Name, strings.Join(describe(got), "\n  "), strings.Join(describe(want.validated), "\n  "))
		}
		for i, check := range c.Checks {
			if got := newValidator().CheckValue(changed, check.Question, loadValue(t, check)); !reflect.DeepEqual(got, want.checked[i]) {
				t.Errorf("%s, %s: directory reversed, the result changed", c.Name, check.Name)
			}
		}
	}
}

func TestResultListsAreNotTheNextCalls(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		plan := loadPlan(t, c)
		validator := newValidator()
		problems := validator.Validate(plan)
		want := describe(problems)
		slices.Reverse(problems)
		for i := range problems {
			if problems[i].Detail != nil {
				problems[i].Detail["changed"] = "changed"
			}
			problems[i].Message = "changed"
		}
		if len(problems) > 0 {
			problems[0] = model.Problem{Code: "changed"}
		}
		if got := describe(validator.Validate(plan)); !slices.Equal(got, want) {
			t.Fatalf("%s: changing the result changed the next call", c.Name)
		}
	}
}

func TestPlanWithNothingHasNoProblemsAndNoQuestion(t *testing.T) {
	t.Parallel()
	validator := newValidator()
	if got := validator.Validate(model.Plan{}); len(got) != 0 {
		t.Errorf("got problems %v, want none", describe(got))
	}
	got := validator.CheckValue(model.Plan{}, "plan.anything", model.Value{Kind: model.ValueKindText, Text: "x"})
	assertProblems(t, got, []expectedProblem{{Code: "answer.unknown-question", Severity: "error", Question: "plan.anything"}}, nil)
}
