package editor_test

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
	contract "onboard-kit/contracts/plan/editor"
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

type expectedCleared struct {
	Answers []string `json:"answers"`
	Holders []string `json:"holders"`
}

type fixtureCase struct {
	Name            string            `json:"name"`
	Operation       string            `json:"operation"`
	Before          string            `json:"before"`
	Question        string            `json:"question"`
	Value           json.RawMessage   `json:"value"`
	Section         string            `json:"section"`
	Enabled         bool              `json:"enabled"`
	Entry           json.RawMessage   `json:"entry"`
	ID              string            `json:"id"`
	Cleared         *expectedCleared  `json:"cleared"`
	After           string            `json:"after"`
	Problems        []expectedProblem `json:"problems"`
	SecretFragments []string          `json:"secretFragments"`
	Source          string            `json:"source"`
}

const markerCharacter = "¤"

const (
	operationSubmit        = "submit"
	operationSwitch        = "switch"
	operationSetContact    = "set-contact"
	operationSetRole       = "set-role"
	operationSetTeam       = "set-team"
	operationRemoveContact = "remove-contact"
	operationRemoveRole    = "remove-role"
	operationRemoveTeam    = "remove-team"

	codeFromValidator = "*"
)

var allOperations = []string{
	operationSubmit, operationSwitch, operationSetContact, operationSetRole, operationSetTeam,
	operationRemoveContact, operationRemoveRole, operationRemoveTeam,
}

func isSet(c fixtureCase) bool {
	return c.Operation == operationSetContact || c.Operation == operationSetRole || c.Operation == operationSetTeam
}

func isRemove(c fixtureCase) bool {
	return c.Operation == operationRemoveContact || c.Operation == operationRemoveRole || c.Operation == operationRemoveTeam
}

func isDirectoryOperation(c fixtureCase) bool {
	return isSet(c) || isRemove(c)
}

var codesEveryCaseSet = []string{
	"answer.wrong-kind",
	"answer.not-an-option",
	"answer.pattern-mismatch",
	"answer.unknown-question",
	"answer.inactive-question",
	"secret.token-prefix",
	"switch.unknown-section",
	"directory.id-malformed",
	"directory.name-empty",
	"directory.holder-not-a-contact",
	"directory.team-role-unknown",
	"directory.entry-not-found",
}

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

func loadPlanFile(t *testing.T, name string) model.Plan {
	t.Helper()
	plan, err := jsonform.Plan(readFixture(t, name))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return plan
}

func loadValue(t *testing.T, c fixtureCase) model.Value {
	t.Helper()
	if len(c.Value) == 0 {
		return model.Value{}
	}
	value, err := jsonform.Value(c.Value)
	if err != nil {
		t.Fatalf("%s: %v", c.Name, err)
	}
	return value
}

type entry struct {
	contact model.Contact
	role    model.Role
	team    model.Team
}

func entryFrom(c fixtureCase) (entry, error) {
	if len(c.Entry) == 0 {
		return entry{}, nil
	}
	kind := strings.TrimPrefix(c.Operation, "set-")
	wrapped := `{"schemaVersion":1,"directory":{"` + kind + `s":[` + string(c.Entry) + `]}}`
	plan, err := jsonform.Plan([]byte(wrapped))
	if err != nil {
		return entry{}, err
	}
	var held entry
	if len(plan.Directory.Contacts) > 0 {
		held.contact = plan.Directory.Contacts[0]
	}
	if len(plan.Directory.Roles) > 0 {
		held.role = plan.Directory.Roles[0]
	}
	if len(plan.Directory.Teams) > 0 {
		held.team = plan.Directory.Teams[0]
	}
	return held, nil
}

func loadEntry(t *testing.T, c fixtureCase) entry {
	t.Helper()
	held, err := entryFrom(c)
	if err != nil {
		t.Fatalf("%s: %v", c.Name, err)
	}
	return held
}

type outcome struct {
	plan     model.Plan
	cleared  model.Cleared
	problems []model.Problem
}

func perform(editor contract.Editor, c fixtureCase, plan model.Plan, value model.Value, held entry) outcome {
	var result outcome
	switch c.Operation {
	case operationSwitch:
		result.plan, result.problems = editor.SetSectionEnabled(plan, c.Section, c.Enabled)
	case operationSetContact:
		result.plan, result.problems = editor.SetContact(plan, held.contact)
	case operationSetRole:
		result.plan, result.problems = editor.SetRole(plan, held.role)
	case operationSetTeam:
		result.plan, result.problems = editor.SetTeam(plan, held.team)
	case operationRemoveContact:
		result.plan, result.cleared, result.problems = editor.RemoveContact(plan, c.ID)
	case operationRemoveRole:
		result.plan, result.cleared, result.problems = editor.RemoveRole(plan, c.ID)
	case operationRemoveTeam:
		result.plan, result.cleared, result.problems = editor.RemoveTeam(plan, c.ID)
	default:
		result.plan, result.problems = editor.Submit(plan, c.Question, value)
	}
	return result
}

func run(t *testing.T, editor contract.Editor, c fixtureCase, plan model.Plan) outcome {
	t.Helper()
	return perform(editor, c, plan, loadValue(t, c), loadEntry(t, c))
}

func describeValue(value model.Value) string {
	switch value.Kind {
	case model.ValueKindNone:
		return "none"
	case model.ValueKindText:
		return "text " + value.Text
	case model.ValueKindFlag:
		if value.Flag {
			return "flag true"
		}
		return "flag false"
	case model.ValueKindList:
		return "list " + strings.Join(value.List, "|")
	case model.ValueKindDeferred:
		return "deferred"
	}
	return "unknown kind"
}

func describeAnswers(plan model.Plan) []string {
	described := []string{}
	for name, value := range plan.Answers.Plan {
		described = append(described, "plan."+name+" = "+describeValue(value))
	}
	for section, answers := range plan.Answers.Sections {
		for name, value := range answers {
			described = append(described, section+"."+name+" = "+describeValue(value))
		}
	}
	slices.Sort(described)
	return described
}

func withoutAnswers(plan model.Plan) model.Plan {
	plan.Answers = model.Answers{}
	teams := slices.Clone(plan.Directory.Teams)
	for i := range teams {
		if len(teams[i].Holders) == 0 {
			teams[i].Holders = nil
		}
	}
	plan.Directory.Teams = teams
	return plan
}

func withoutDirectory(plan model.Plan) model.Plan {
	plan.Directory = model.Directory{}
	return plan
}

func planText(t *testing.T, plan model.Plan) string {
	t.Helper()
	data, err := jsonform.JSON(withoutAnswers(plan))
	if err != nil {
		t.Fatalf("write plan: %v", err)
	}
	return string(data)
}

func assertSamePlan(t *testing.T, label string, got, want model.Plan) {
	t.Helper()
	if gotAnswers, wantAnswers := describeAnswers(got), describeAnswers(want); !slices.Equal(gotAnswers, wantAnswers) {
		t.Errorf("%s: answers\n  %s\nwant\n  %s", label, strings.Join(gotAnswers, "\n  "), strings.Join(wantAnswers, "\n  "))
	}
	if gotText, wantText := planText(t, got), planText(t, want); gotText != wantText {
		t.Errorf("%s: the plan outside its answers differs\ngot\n%s\nwant\n%s", label, gotText, wantText)
	}
}

func describeProblems(problems []model.Problem) []string {
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

func describeAgainst(problems []model.Problem, want []expectedProblem) []string {
	shaped := slices.Clone(problems)
	for i := range shaped {
		if i < len(want) && want[i].Code == codeFromValidator {
			shaped[i].Code = codeFromValidator
			shaped[i].Detail = nil
		}
	}
	return describeProblems(shaped)
}

func describeExpected(problems []expectedProblem) []string {
	converted := []model.Problem{}
	for _, problem := range problems {
		converted = append(converted, model.Problem{
			Code: problem.Code, Severity: model.Severity(problem.Severity), Question: problem.Question, Section: problem.Section, Detail: problem.Detail,
		})
	}
	return describeProblems(converted)
}

func hasError(problems []expectedProblem) bool {
	return slices.ContainsFunc(problems, func(problem expectedProblem) bool { return problem.Severity == "error" })
}

func assertProblems(t *testing.T, got []model.Problem, want []expectedProblem, fragments []string) {
	t.Helper()
	if !slices.Equal(describeAgainst(got, want), describeExpected(want)) {
		t.Errorf("problems\n  %s\nwant\n  %s", strings.Join(describeAgainst(got, want), "\n  "), strings.Join(describeExpected(want), "\n  "))
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

func assertCleared(t *testing.T, label string, got model.Cleared, want *expectedCleared) {
	t.Helper()
	if want == nil {
		want = &expectedCleared{}
	}
	if !slices.Equal(got.Answers, want.Answers) && (len(got.Answers) > 0 || len(want.Answers) > 0) {
		t.Errorf("%s: cleared answers %v, want %v", label, got.Answers, want.Answers)
	}
	if !slices.Equal(got.Holders, want.Holders) && (len(got.Holders) > 0 || len(want.Holders) > 0) {
		t.Errorf("%s: cleared holders %v, want %v", label, got.Holders, want.Holders)
	}
}

func isApplied(c fixtureCase) bool {
	return !hasError(c.Problems)
}

func isRefusal(c fixtureCase) bool {
	return hasError(c.Problems)
}

func isUnanswered(value model.Value) bool {
	switch value.Kind {
	case model.ValueKindNone:
		return true
	case model.ValueKindText:
		return value.Text == ""
	case model.ValueKindList:
		return !slices.ContainsFunc(value.List, func(item string) bool { return item != "" })
	}
	return false
}

func TestEveryCaseCitesItsSourceAndTheCasesCoverEveryCode(t *testing.T) {
	names := map[string]bool{}
	codes := map[string]bool{}
	severities := map[string]bool{}
	operations := map[string]bool{}
	details := map[string]bool{}
	var sawClearByNoValue, sawClearByEmptyText, sawClearByList, sawDeferred, sawEnabledTrue, sawEnabledFalse bool
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
		if names[c.Name] {
			t.Errorf("case %s is listed twice", c.Name)
		}
		names[c.Name] = true
		if !slices.Contains(allOperations, c.Operation) {
			t.Errorf("case %s has the operation %q", c.Name, c.Operation)
		}
		if isRemove(c) && c.Cleared == nil {
			t.Errorf("case %s removes an entry and lists no cleared places", c.Name)
		}
		if !isRemove(c) && c.Cleared != nil {
			t.Errorf("case %s lists cleared places and does not remove", c.Name)
		}
		if c.Cleared != nil {
			if !slices.IsSorted(c.Cleared.Answers) || !slices.IsSorted(c.Cleared.Holders) {
				t.Errorf("case %s lists its cleared places out of order", c.Name)
			}
			if isRefusal(c) && (len(c.Cleared.Answers) > 0 || len(c.Cleared.Holders) > 0) {
				t.Errorf("case %s is refused and lists cleared places", c.Name)
			}
		}
		operations[c.Operation] = true
		before := loadPlanFile(t, c.Before)
		after := loadPlanFile(t, c.After)
		if hasError(c.Problems) && c.After != c.Before {
			t.Errorf("case %s expects an error, so its plan after must be its plan before", c.Name)
		}
		if hasError(c.Problems) {
			assertSamePlan(t, c.Name+": before and after files", after, before)
		}
		for _, problem := range c.Problems {
			codes[problem.Code] = true
			severities[problem.Severity] = true
			for key := range problem.Detail {
				details[key] = true
			}
		}
		if isDirectoryOperation(c) {
			continue
		}
		if c.Operation == operationSubmit {
			value := loadValue(t, c)
			sawClearByNoValue = sawClearByNoValue || (len(c.Value) == 0 && isApplied(c))
			sawClearByEmptyText = sawClearByEmptyText || (value.Kind == model.ValueKindText && value.Text == "")
			sawClearByList = sawClearByList || (value.Kind == model.ValueKindList && isUnanswered(value))
			sawDeferred = sawDeferred || value.Kind == model.ValueKindDeferred
		} else {
			sawEnabledTrue = sawEnabledTrue || c.Enabled
			sawEnabledFalse = sawEnabledFalse || !c.Enabled
		}
	}
	if !codes[codeFromValidator] {
		t.Error("no case expects a refusal by the validator's own code")
	}
	for _, code := range codesEveryCaseSet {
		if !codes[code] {
			t.Errorf("no case expects the code %s", code)
		}
	}
	for _, severity := range []string{"error"} {
		if !severities[severity] {
			t.Errorf("no case expects a problem of severity %s", severity)
		}
	}
	for _, key := range []string{"expected", "found"} {
		if !details[key] {
			t.Errorf("no case expects the detail %s", key)
		}
	}
	for _, operation := range allOperations {
		if !operations[operation] {
			t.Errorf("no case uses the operation %s", operation)
		}
	}
	for label, seen := range map[string]bool{
		"a Submit with no value": sawClearByNoValue, "a Submit with empty text": sawClearByEmptyText, "a Submit with an empty list": sawClearByList,
		"a Submit with a Deferred value": sawDeferred, "a switch to true": sawEnabledTrue, "a switch to false": sawEnabledFalse,
	} {
		if !seen {
			t.Errorf("no case has %s", label)
		}
	}
}

func TestEachCaseGivesTheExpectedPlanAndProblems(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			got := run(t, newEditor(), c, loadPlanFile(t, c.Before))
			assertProblems(t, got.problems, c.Problems, c.SecretFragments)
			assertSamePlan(t, "plan", got.plan, loadPlanFile(t, c.After))
			assertCleared(t, "cleared", got.cleared, c.Cleared)
		})
	}
}

func TestAnErrorKeepsThePlanAndAWarningDoesNot(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		before := loadPlanFile(t, c.Before)
		got := run(t, newEditor(), c, before)
		severities := map[model.Severity]bool{}
		for _, problem := range got.problems {
			severities[problem.Severity] = true
		}
		if severities[model.SeverityError] {
			assertSamePlan(t, c.Name+": an error", got.plan, before)
		}
		if !severities[model.SeverityError] {
			assertSamePlan(t, c.Name+": no error", got.plan, loadPlanFile(t, c.After))
		}
	}
}

func TestArgumentsAreNotChanged(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		plan := loadPlanFile(t, c.Before)
		snapshot := loadPlanFile(t, c.Before)
		value := loadValue(t, c)
		valueSnapshot := loadValue(t, c)
		held := loadEntry(t, c)
		heldSnapshot := loadEntry(t, c)
		perform(newEditor(), c, plan, value, held)
		if !reflect.DeepEqual(plan, snapshot) {
			t.Fatalf("%s: the plan was changed", c.Name)
		}
		if !reflect.DeepEqual(value, valueSnapshot) {
			t.Fatalf("%s: the value was changed", c.Name)
		}
		if !reflect.DeepEqual(held, heldSnapshot) {
			t.Fatalf("%s: the entry was changed", c.Name)
		}
	}
}

func TestEverySectionOfTheResultKeepsItsParent(t *testing.T) {
	t.Parallel()
	parents := 0
	for _, c := range loadCases(t) {
		before := loadPlanFile(t, c.Before)
		got := run(t, newEditor(), c, before).plan
		if len(got.Sections) != len(before.Sections) {
			t.Fatalf("%s: got %d sections, want %d", c.Name, len(got.Sections), len(before.Sections))
		}
		for i, section := range before.Sections {
			if got.Sections[i].ID != section.ID || got.Sections[i].Parent != section.Parent {
				t.Errorf("%s: section %d is %q with parent %q, want %q with parent %q", c.Name, i, got.Sections[i].ID, got.Sections[i].Parent, section.ID, section.Parent)
			}
			if section.Parent != "" {
				parents++
			}
		}
	}
	if parents == 0 {
		t.Fatal("no fixture section holds a parent")
	}
}

func TestResultsDoNotDependOnEarlierCallsOrOnRepeats(t *testing.T) {
	t.Parallel()
	cases := loadCases(t)
	shared := newEditor()
	first := map[string]outcome{}
	for _, c := range cases {
		first[c.Name] = run(t, newEditor(), c, loadPlanFile(t, c.Before))
	}
	for pass := 0; pass < 2; pass++ {
		for i := range cases {
			index := i
			if pass == 1 {
				index = len(cases) - 1 - i
			}
			c := cases[index]
			for repeat := 0; repeat < 5; repeat++ {
				if got := run(t, shared, c, loadPlanFile(t, c.Before)); !reflect.DeepEqual(got, first[c.Name]) {
					t.Fatalf("%s on pass %d, repeat %d: the result changed", c.Name, pass, repeat)
				}
			}
		}
	}
}

func TestOneEditorServesCallsAtTheSameTime(t *testing.T) {
	t.Parallel()
	cases := loadCases(t)
	shared := newEditor()
	first := map[string]outcome{}
	planData := map[string][]byte{}
	for _, c := range cases {
		first[c.Name] = run(t, newEditor(), c, loadPlanFile(t, c.Before))
		planData[c.Name] = readFixture(t, c.Before)
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
				value := model.Value{}
				if len(c.Value) > 0 {
					value, err = jsonform.Value(c.Value)
					if err != nil {
						failures <- err.Error()
						return
					}
				}
				held, err := entryFrom(c)
				if err != nil {
					failures <- err.Error()
					return
				}
				got := perform(shared, c, plan, value, held)
				if !reflect.DeepEqual(got, first[c.Name]) {
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

func changeEveryAnswerAndSection(plan *model.Plan) {
	for name := range plan.Answers.Plan {
		plan.Answers.Plan[name] = model.Value{Kind: model.ValueKindText, Text: "changed"}
	}
	if plan.Answers.Plan != nil {
		plan.Answers.Plan["added"] = model.Value{Kind: model.ValueKindText, Text: "added"}
	}
	for section, answers := range plan.Answers.Sections {
		for name, value := range answers {
			for i := range value.List {
				value.List[i] = "changed"
			}
			answers[name] = model.Value{Kind: model.ValueKindText, Text: "changed"}
		}
		answers["added"] = model.Value{Kind: model.ValueKindText, Text: "added"}
		plan.Answers.Sections[section] = answers
	}
	for i := range plan.Sections {
		if plan.Sections[i].Enabled != nil {
			*plan.Sections[i].Enabled = !*plan.Sections[i].Enabled
		}
		plan.Sections[i].Enabled = new(bool)
	}
	slices.Reverse(plan.Sections)
}

func changeDirectory(plan *model.Plan) {
	directory := plan.Directory
	for i := range directory.Contacts {
		directory.Contacts[i].Name = "changed"
		directory.Contacts[i].Email = "changed"
		directory.Contacts[i].Slack = "changed"
		directory.Contacts[i].SlackID = "changed"
	}
	for i := range directory.Roles {
		directory.Roles[i].Name = "changed"
		for j := range directory.Roles[i].Holders {
			directory.Roles[i].Holders[j] = "changed"
		}
	}
	for i := range directory.Teams {
		directory.Teams[i].Name = "changed"
		for role, holders := range directory.Teams[i].Holders {
			for j := range holders {
				holders[j] = "changed"
			}
			directory.Teams[i].Holders[role] = append(holders[:len(holders):len(holders)], "added")
		}
		if directory.Teams[i].Holders != nil {
			directory.Teams[i].Holders["added"] = []string{"added"}
		}
	}
	slices.Reverse(directory.Contacts)
	slices.Reverse(directory.Roles)
	slices.Reverse(directory.Teams)
}

func changeEveryStoredList(plan *model.Plan) {
	for _, value := range plan.Answers.Plan {
		for i := range value.List {
			value.List[i] = "changed"
		}
	}
	for _, answers := range plan.Answers.Sections {
		for _, value := range answers {
			for i := range value.List {
				value.List[i] = "changed"
			}
		}
	}
}

func TestResultSharesNothingWithTheArgumentThatAnEditCouldChange(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		argument := loadPlanFile(t, c.Before)
		snapshot := loadPlanFile(t, c.Before)
		editor := newEditor()
		got := run(t, editor, c, argument)
		baseline := run(t, newEditor(), c, loadPlanFile(t, c.Before))

		changeEveryAnswerAndSection(&argument)
		changeEveryStoredList(&argument)
		if isDirectoryOperation(c) {
			changeDirectory(&argument)
		}
		if !reflect.DeepEqual(got, baseline) {
			t.Fatalf("%s: changing the argument after the call changed the result", c.Name)
		}

		argument = loadPlanFile(t, c.Before)
		got = run(t, editor, c, argument)
		changeEveryAnswerAndSection(&got.plan)
		changeEveryStoredList(&got.plan)
		if isDirectoryOperation(c) {
			changeDirectory(&got.plan)
		}
		if !reflect.DeepEqual(argument, snapshot) {
			t.Fatalf("%s: changing the result changed the argument", c.Name)
		}
		if again := run(t, editor, c, argument); !reflect.DeepEqual(again, baseline) {
			t.Fatalf("%s: changing the result changed the next call", c.Name)
		}
	}
}

func TestStoredListIsNotTheListThatWasSubmitted(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		value := loadValue(t, c)
		if c.Operation != operationSubmit || value.Kind != model.ValueKindList || isUnanswered(value) || !isApplied(c) {
			continue
		}
		got, _ := newEditor().Submit(loadPlanFile(t, c.Before), c.Question, value)
		stored := describeAnswers(got)
		submitted := slices.Clone(value.List)
		for i := range value.List {
			value.List[i] = "changed"
		}
		if !slices.Equal(describeAnswers(got), stored) {
			t.Errorf("%s: changing the submitted list changed the plan", c.Name)
		}
		fresh := loadValue(t, c)
		again, _ := newEditor().Submit(loadPlanFile(t, c.Before), c.Question, fresh)
		changeEveryStoredList(&again)
		if !slices.Equal(fresh.List, submitted) {
			t.Errorf("%s: changing the stored list changed the submitted one", c.Name)
		}
	}
}

func TestSubmitChangesOnlyTheOneAnswer(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		if c.Operation != operationSubmit || !isApplied(c) {
			continue
		}
		before := loadPlanFile(t, c.Before)
		got, _ := newEditor().Submit(before, c.Question, loadValue(t, c))
		scope, name, _ := strings.Cut(c.Question, ".")
		key := scope + "." + name
		wanted := map[string]string{}
		for _, line := range describeAnswers(before) {
			left, right, _ := strings.Cut(line, " = ")
			wanted[left] = right
		}
		if value := loadValue(t, c); isUnanswered(value) {
			delete(wanted, key)
		} else {
			wanted[key] = describeValue(value)
		}
		expected := []string{}
		for left, right := range wanted {
			expected = append(expected, left+" = "+right)
		}
		slices.Sort(expected)
		if gotAnswers := describeAnswers(got); !slices.Equal(gotAnswers, expected) {
			t.Errorf("%s: answers\n  %s\nwant\n  %s", c.Name, strings.Join(gotAnswers, "\n  "), strings.Join(expected, "\n  "))
		}
		if planText(t, got) != planText(t, before) {
			t.Errorf("%s: the plan changed outside its answers", c.Name)
		}
	}
}

func switchedOnlyAt(plan model.Plan, section string) model.Plan {
	sections := slices.Clone(plan.Sections)
	for i := range sections {
		if sections[i].ID == section {
			sections[i].Enabled = nil
		}
	}
	plan.Sections = sections
	return plan
}

func TestSetSectionEnabledChangesOnlyTheSwitchOfThatSection(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		if c.Operation != operationSwitch || !isApplied(c) {
			continue
		}
		before := loadPlanFile(t, c.Before)
		got, problems := newEditor().SetSectionEnabled(before, c.Section, c.Enabled)
		if len(problems) != 0 {
			t.Errorf("%s: got problems %v", c.Name, describeProblems(problems))
		}
		if !slices.Equal(describeAnswers(got), describeAnswers(before)) {
			t.Errorf("%s: the answers changed", c.Name)
		}
		if planText(t, switchedOnlyAt(got, c.Section)) != planText(t, switchedOnlyAt(before, c.Section)) {
			t.Errorf("%s: the plan changed outside the section's switch", c.Name)
		}
		for _, section := range got.Sections {
			if section.ID != c.Section {
				continue
			}
			if c.Enabled && section.Enabled != nil {
				t.Errorf("%s: a section turned on keeps the field", c.Name)
			}
			if !c.Enabled && (section.Enabled == nil || *section.Enabled) {
				t.Errorf("%s: a section turned off must hold enabled false", c.Name)
			}
		}
	}
}

func TestRepeatingAnEditChangesNothingMore(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		if !isApplied(c) {
			continue
		}
		editor := newEditor()
		first := run(t, editor, c, loadPlanFile(t, c.Before))
		second := run(t, editor, c, first.plan)
		assertSamePlan(t, c.Name+": repeated", second.plan, first.plan)
		if isRemove(c) {
			assertProblems(t, second.problems, []expectedProblem{{Code: "directory.entry-not-found", Severity: "error", Detail: map[string]string{"kind": strings.TrimPrefix(c.Operation, "remove-"), "id": c.ID}}}, nil)
			assertCleared(t, c.Name+": repeated", second.cleared, nil)
			continue
		}
		if !slices.Equal(describeProblems(second.problems), describeProblems(first.problems)) {
			t.Errorf("%s: the repeat gives other problems", c.Name)
		}
	}
}

func TestClearingTakesBackWhatAnAnswerAdded(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		value := loadValue(t, c)
		if c.Operation != operationSubmit || !isApplied(c) || isUnanswered(value) {
			continue
		}
		before := loadPlanFile(t, c.Before)
		scope, name, _ := strings.Cut(c.Question, ".")
		if _, stored := before.Answers.Plan[name]; scope == "plan" && stored {
			continue
		}
		if _, stored := before.Answers.Sections[scope][name]; scope != "plan" && stored {
			continue
		}
		editor := newEditor()
		answered, _ := editor.Submit(before, c.Question, value)
		cleared, problems := editor.Submit(answered, c.Question, model.Value{})
		if len(problems) != 0 {
			t.Errorf("%s: clearing gave problems %v", c.Name, describeProblems(problems))
		}
		assertSamePlan(t, c.Name+": cleared", cleared, before)
	}
}

func TestSwitchingOffAndOnAgainGivesBackThePlan(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		if c.Operation != operationSwitch || c.Enabled || !isApplied(c) {
			continue
		}
		before := loadPlanFile(t, c.Before)
		stateBefore := before.Sections
		var wasOnWithNoField bool
		for _, section := range stateBefore {
			if section.ID == c.Section && section.Enabled == nil {
				wasOnWithNoField = true
			}
		}
		if !wasOnWithNoField {
			continue
		}
		editor := newEditor()
		off, _ := editor.SetSectionEnabled(before, c.Section, false)
		on, _ := editor.SetSectionEnabled(off, c.Section, true)
		assertSamePlan(t, c.Name+": off and on", on, before)
	}
}

func TestResultDoesNotDependOnTheOrderOfTheSections(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		want := run(t, newEditor(), c, loadPlanFile(t, c.Before))
		count := len(loadPlanFile(t, c.Before).Sections)
		variants := []func(*model.Plan){}
		for turn := 1; turn < count; turn++ {
			variants = append(variants, func(p *model.Plan) {
				p.Sections = append(slices.Clone(p.Sections[turn:]), p.Sections[:turn]...)
			})
		}
		variants = append(variants, func(p *model.Plan) { slices.Reverse(p.Sections) })
		for _, variant := range variants {
			changed := loadPlanFile(t, c.Before)
			variant(&changed)
			got := run(t, newEditor(), c, changed)
			if !slices.Equal(describeProblems(got.problems), describeProblems(want.problems)) {
				t.Errorf("%s: sections reordered, the problems changed", c.Name)
			}
			if !slices.Equal(describeAnswers(got.plan), describeAnswers(want.plan)) {
				t.Errorf("%s: sections reordered, the answers changed", c.Name)
			}
			assertCleared(t, c.Name+": sections reordered", got.cleared, &expectedCleared{Answers: want.cleared.Answers, Holders: want.cleared.Holders})
			gotOrder, argumentOrder := []string{}, []string{}
			for _, section := range got.plan.Sections {
				gotOrder = append(gotOrder, section.ID)
			}
			for _, section := range changed.Sections {
				argumentOrder = append(argumentOrder, section.ID)
			}
			if !slices.Equal(gotOrder, argumentOrder) {
				t.Errorf("%s: the sections of the result are %v, the argument had %v", c.Name, gotOrder, argumentOrder)
			}
		}
	}
}

func TestResultProblemsAreNotTheNextCalls(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		editor := newEditor()
		first := run(t, editor, c, loadPlanFile(t, c.Before))
		want := describeProblems(first.problems)
		slices.Reverse(first.problems)
		for i := range first.problems {
			if first.problems[i].Detail != nil {
				first.problems[i].Detail["changed"] = "changed"
			}
			first.problems[i].Message = "changed"
		}
		if len(first.problems) > 0 {
			first.problems[0] = model.Problem{Code: "changed"}
		}
		for i := range first.cleared.Answers {
			first.cleared.Answers[i] = "changed"
		}
		for i := range first.cleared.Holders {
			first.cleared.Holders[i] = "changed"
		}
		next := run(t, editor, c, loadPlanFile(t, c.Before))
		if got := describeProblems(next.problems); !slices.Equal(got, want) {
			t.Fatalf("%s: changing the problems changed the next call", c.Name)
		}
		assertCleared(t, c.Name+": after the result was changed", next.cleared, c.Cleared)
	}
}

func TestPlanWithNothingHasNoQuestionAndNoSection(t *testing.T) {
	t.Parallel()
	editor := newEditor()
	got, problems := editor.Submit(model.Plan{}, "plan.anything", model.Value{Kind: model.ValueKindText, Text: "x"})
	assertProblems(t, problems, []expectedProblem{{Code: "answer.unknown-question", Severity: "error", Question: "plan.anything"}}, nil)
	assertSamePlan(t, "submit", got, model.Plan{})
	got, problems = editor.SetSectionEnabled(model.Plan{}, "anything", false)
	assertProblems(t, problems, []expectedProblem{{Code: "switch.unknown-section", Severity: "error", Section: "anything"}}, nil)
	assertSamePlan(t, "switch", got, model.Plan{})
}

func answerLines(plan model.Plan) map[string]string {
	lines := map[string]string{}
	for _, line := range describeAnswers(plan) {
		key, value, _ := strings.Cut(line, " = ")
		lines[key] = value
	}
	return lines
}

func TestDirectoryEditsChangeOnlyTheDirectoryAndTheAnswersTheyClear(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		if !isDirectoryOperation(c) || !isApplied(c) {
			continue
		}
		before := loadPlanFile(t, c.Before)
		got := run(t, newEditor(), c, before)
		if len(got.problems) != 0 {
			t.Errorf("%s: got problems %v", c.Name, describeProblems(got.problems))
		}
		if planText(t, withoutDirectory(got.plan)) != planText(t, withoutDirectory(before)) {
			t.Errorf("%s: the plan changed outside the directory and its answers", c.Name)
		}
		changed := []string{}
		was, now := answerLines(before), answerLines(got.plan)
		for key, value := range was {
			if now[key] != value {
				changed = append(changed, key)
			}
		}
		for key := range now {
			if _, held := was[key]; !held {
				t.Errorf("%s: the answer %s was added", c.Name, key)
			}
		}
		slices.Sort(changed)
		if len(changed) != len(got.cleared.Answers) || (len(changed) > 0 && !slices.Equal(changed, got.cleared.Answers)) {
			t.Errorf("%s: the answers that changed are %v, and the cleared answers are %v", c.Name, changed, got.cleared.Answers)
		}
		if isSet(c) && (len(changed) > 0 || len(got.cleared.Holders) > 0) {
			t.Errorf("%s: a set changed an answer or reports cleared places", c.Name)
		}
	}
}

func TestClearedAnswersLeaveNoEmptySectionMapAndNoEmptyEntry(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		if !isDirectoryOperation(c) {
			continue
		}
		got := run(t, newEditor(), c, loadPlanFile(t, c.Before)).plan
		for section, answers := range got.Answers.Sections {
			if len(answers) == 0 {
				t.Errorf("%s: the section %s holds an empty map of answers", c.Name, section)
			}
			for name, value := range answers {
				if value.Kind == model.ValueKindNone {
					t.Errorf("%s: %s.%s holds a value of no kind", c.Name, section, name)
				}
			}
		}
		for name, value := range got.Answers.Plan {
			if value.Kind == model.ValueKindNone {
				t.Errorf("%s: plan.%s holds a value of no kind", c.Name, name)
			}
		}
	}
}

func TestClearedPlacesAreSortedAndListedOnce(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		if !isRemove(c) {
			continue
		}
		got := run(t, newEditor(), c, loadPlanFile(t, c.Before)).cleared
		for label, places := range map[string][]string{"answers": got.Answers, "holders": got.Holders} {
			if !slices.IsSorted(places) || len(slices.Compact(slices.Clone(places))) != len(places) {
				t.Errorf("%s: cleared %s %v are not sorted and distinct", c.Name, label, places)
			}
		}
	}
}

func TestRemovingTakesBackAnEntryThatASetAdded(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		if !isSet(c) || !isApplied(c) {
			continue
		}
		before := loadPlanFile(t, c.Before)
		held := loadEntry(t, c)
		kind, id := strings.TrimPrefix(c.Operation, "set-"), ""
		existing := map[string]bool{}
		for _, contact := range before.Directory.Contacts {
			existing["contact:"+contact.ID] = true
		}
		for _, role := range before.Directory.Roles {
			existing["role:"+role.ID] = true
		}
		for _, team := range before.Directory.Teams {
			existing["team:"+team.ID] = true
		}
		id = map[string]string{"contact": held.contact.ID, "role": held.role.ID, "team": held.team.ID}[kind]
		if existing[kind+":"+id] {
			continue
		}
		editor := newEditor()
		added := run(t, editor, c, before).plan
		removal := fixtureCase{Operation: "remove-" + kind, ID: id}
		removed := perform(editor, removal, added, model.Value{}, entry{})
		if len(removed.problems) != 0 {
			t.Errorf("%s: removing the new entry gave %v", c.Name, describeProblems(removed.problems))
		}
		assertCleared(t, c.Name+": removing the new entry", removed.cleared, nil)
		assertSamePlan(t, c.Name+": added and removed", removed.plan, before)
	}
}

func TestStoredEntryIsNotTheEntryThatWasSubmitted(t *testing.T) {
	t.Parallel()
	for _, c := range loadCases(t) {
		if !isSet(c) {
			continue
		}
		editor := newEditor()
		held := loadEntry(t, c)
		got := perform(editor, c, loadPlanFile(t, c.Before), model.Value{}, held)
		stored := planText(t, got.plan)
		for i := range held.role.Holders {
			held.role.Holders[i] = "changed"
		}
		for role, holders := range held.team.Holders {
			for i := range holders {
				holders[i] = "changed"
			}
			held.team.Holders[role] = nil
		}
		if held.team.Holders != nil {
			held.team.Holders["added"] = []string{"added"}
		}
		if planText(t, got.plan) != stored {
			t.Errorf("%s: changing the submitted entry changed the plan", c.Name)
		}
		fresh := loadEntry(t, c)
		again := perform(editor, c, loadPlanFile(t, c.Before), model.Value{}, fresh)
		changeDirectory(&again.plan)
		if !reflect.DeepEqual(fresh, loadEntry(t, c)) {
			t.Errorf("%s: changing the stored entry changed the submitted one", c.Name)
		}
	}
}

func TestDirectoryOperationsOnAPlanWithNothing(t *testing.T) {
	t.Parallel()
	editor := newEditor()
	contact := model.Contact{ID: "sam-okafor", Name: "Sam Okafor"}
	got, problems := editor.SetContact(model.Plan{}, contact)
	assertProblems(t, problems, nil, nil)
	if !reflect.DeepEqual(got.Directory.Contacts, []model.Contact{contact}) || len(got.Directory.Roles) != 0 || len(got.Directory.Teams) != 0 {
		t.Errorf("set contact on nothing gives %+v", got.Directory)
	}
	for _, removal := range []struct {
		kind   string
		remove func() (model.Plan, model.Cleared, []model.Problem)
	}{
		{"contact", func() (model.Plan, model.Cleared, []model.Problem) {
			return editor.RemoveContact(model.Plan{}, "sam-okafor")
		}},
		{"role", func() (model.Plan, model.Cleared, []model.Problem) {
			return editor.RemoveRole(model.Plan{}, "sam-okafor")
		}},
		{"team", func() (model.Plan, model.Cleared, []model.Problem) {
			return editor.RemoveTeam(model.Plan{}, "sam-okafor")
		}},
	} {
		changed, cleared, problems := removal.remove()
		assertProblems(t, problems, []expectedProblem{{Code: "directory.entry-not-found", Severity: "error", Detail: map[string]string{"kind": removal.kind, "id": "sam-okafor"}}}, nil)
		assertSamePlan(t, "remove "+removal.kind, changed, model.Plan{})
		assertCleared(t, "remove "+removal.kind, cleared, nil)
	}
}
