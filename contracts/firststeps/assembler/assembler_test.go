package assembler_test

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"

	contract "onboard-kit/contracts/firststeps/assembler"
	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type expectedProblem struct {
	Code     string
	Severity model.Severity
	Message  string
	Question string
	Section  string
	Detail   map[string]string
	Path     string
}

type fixture struct {
	Name        string
	Source      string
	Context     contract.Context
	Plan        json.RawMessage
	Preparation model.Preparation
	Golden      string
	Fragment    string
	Contains    []string
	Absent      []string
	Problems    []expectedProblem
}

func fixtures(t *testing.T) []fixture {
	t.Helper()
	var all []fixture
	seen := make(map[string]bool)
	for _, path := range []string{"testdata/cases.json", "testdata/fragments.json"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var group []fixture
		decoder := json.NewDecoder(strings.NewReader(string(data)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&group); err != nil {
			t.Fatal(err)
		}
		for _, item := range group {
			if item.Name == "" || item.Source == "" || seen[item.Name] || (item.Golden == "" && item.Fragment == "" && len(item.Contains) == 0) {
				t.Fatal("fixture missing provenance, oracle, or distinct name")
			}
			seen[item.Name] = true
		}
		all = append(all, group...)
	}
	if len(all) != 15 {
		t.Fatal("expected fifteen literal first-steps cases")
	}
	return all
}

func snapshot(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func check(t *testing.T, item fixture, shared contract.Assembler, reverse bool) {
	t.Helper()
	plan, err := jsonform.Plan(item.Plan)
	if err != nil {
		t.Fatal(err)
	}
	if reverse {
		for i, j := 0, len(plan.Sections)-1; i < j; i, j = i+1, j-1 {
			plan.Sections[i], plan.Sections[j] = plan.Sections[j], plan.Sections[i]
		}
		for i, j := 0, len(plan.Directory.Contacts)-1; i < j; i, j = i+1, j-1 {
			plan.Directory.Contacts[i], plan.Directory.Contacts[j] = plan.Directory.Contacts[j], plan.Directory.Contacts[i]
		}
	}
	before := snapshot(t, []any{item.Context, plan, item.Preparation})
	document, problems := shared.Assemble(item.Context, plan, item.Preparation)
	if snapshot(t, []any{item.Context, plan, item.Preparation}) != before {
		t.Fatal("assembly changed an argument or backing storage")
	}
	if item.Golden != "" {
		want, err := os.ReadFile("testdata/" + item.Golden)
		if err != nil {
			t.Fatal(err)
		}
		if document != string(want) {
			t.Fatalf("%s: literal guide mismatch\ngot %q\nwant %q", item.Name, document, want)
		}
	} else if item.Fragment != "" && !strings.Contains(document, "## Do these first\n\n"+item.Fragment+"\n## Connect to the shared repository\n") {
		t.Fatalf("%s: exact literal Bootstrap fragment missing or rewritten", item.Name)
	}
	for _, text := range item.Contains {
		if !strings.Contains(document, text) {
			t.Fatalf("%s: literal display missing: %q", item.Name, text)
		}
	}
	for _, text := range item.Absent {
		if strings.Contains(document, text) {
			t.Fatalf("%s: unsafe display present: %q", item.Name, text)
		}
	}
	if len(item.Contains) > 0 && strings.Count("\n"+document, "\n## ") != 12 {
		t.Fatal("wrapper value injected an extra heading")
	}
	if len(problems) != len(item.Problems) {
		t.Fatalf("%s: problem count got %d want %d", item.Name, len(problems), len(item.Problems))
	}
	for index, want := range item.Problems {
		got := problems[index]
		if got.Code != want.Code || got.Severity != want.Severity || got.Question != want.Question || got.Section != want.Section {
			t.Fatalf("%s: problem %d mismatch: %#v", item.Name, index, got)
		}
		if want.Message != "" && got.Message != want.Message {
			t.Fatalf("%s: problem message differs", item.Name)
		}
		if want.Path != "" {
			if len(got.Detail) != 1 || got.Detail["path"] != want.Path || strings.TrimSpace(got.Message) == "" {
				t.Fatalf("%s: dependency diagnostic shape or useful repair missing", item.Name)
			}
		} else if len(got.Detail) != len(want.Detail) || !mapsEqual(got.Detail, want.Detail) {
			t.Fatalf("%s: detail keys/values differ", item.Name)
		}
	}
	if strings.Contains(document, "PRIVATE") || strings.Contains(snapshot(t, problems), "PRIVATE") {
		t.Fatal("omitted or unknown personal/reference data leaked")
	}
	if document != "" && (!strings.HasSuffix(document, "\n") || strings.HasSuffix(document, "\n\n") || strings.Contains(document, "\r")) {
		t.Fatal("guide has invalid final newline or carriage return")
	}
	firstDocument := document
	firstProblems := snapshot(t, problems)
	for index := range problems {
		problems[index].Code = "changed-result"
		for key := range problems[index].Detail {
			problems[index].Detail[key] = "changed-result"
		}
	}
	againDocument, againProblems := shared.Assemble(item.Context, plan, item.Preparation)
	if firstDocument != againDocument || firstProblems != snapshot(t, againProblems) || before != snapshot(t, []any{item.Context, plan, item.Preparation}) {
		t.Fatal("state or returned diagnostic maps alias arguments or another result")
	}
}

func mapsEqual(left, right map[string]string) bool {
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func TestLiteralGuides(t *testing.T) {
	for _, item := range fixtures(t) {
		t.Run(item.Name, func(t *testing.T) {
			t.Parallel()
			check(t, item, newAssembler(), false)
		})
	}
}

func TestSharedInstanceAndPermutations(t *testing.T) {
	shared := newAssembler()
	cases := fixtures(t)
	for round := 0; round < 2; round++ {
		for index := range cases {
			item := cases[index]
			if round == 1 {
				item = cases[len(cases)-1-index]
			}
			check(t, item, shared, round == 1)
		}
	}
}

func TestConcurrentSharedInstance(t *testing.T) {
	shared := newAssembler()
	var wait sync.WaitGroup
	for _, item := range fixtures(t) {
		wait.Add(1)
		go func() {
			defer wait.Done()
			check(t, item, shared, false)
		}()
	}
	wait.Wait()
}
