package assembler_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"onboard-kit/contracts/harness/jsonform"
	"onboard-kit/internal/model"
)

type expectedProblem struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Question string `json:"question"`
	Section  string `json:"section"`
}

type fixtureCase struct {
	Name     string            `json:"name"`
	Plan     string            `json:"plan"`
	Document string            `json:"document"`
	Problems []expectedProblem `json:"problems"`
	Source   string            `json:"source"`
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

func loadPlan(t *testing.T, c fixtureCase) model.Plan {
	t.Helper()
	plan, err := jsonform.Plan(readFixture(t, c.Plan))
	if err != nil {
		t.Fatalf("%s: %v", c.Plan, err)
	}
	return plan
}

func problemsOf(problems []model.Problem) []expectedProblem {
	got := []expectedProblem{}
	for _, problem := range problems {
		got = append(got, expectedProblem{
			Code:     problem.Code,
			Severity: string(problem.Severity),
			Question: problem.Question,
			Section:  problem.Section,
		})
	}
	return got
}

func expectedProblems(c fixtureCase) []expectedProblem {
	if c.Problems == nil {
		return []expectedProblem{}
	}
	return c.Problems
}

func gapTableRows(document string) ([]string, bool) {
	lines := strings.Split(document, "\n")
	start := slices.Index(lines, "## Gaps to fill")
	if start < 0 {
		return nil, false
	}
	rows := []string{}
	for _, line := range lines[start+1:] {
		if line == "" || strings.HasPrefix(line, "| Gap |") || strings.HasPrefix(line, "| --- |") {
			continue
		}
		if !strings.HasPrefix(line, "|") {
			break
		}
		rows = append(rows, line)
	}
	return rows, true
}

func TestEveryCaseCitesItsSourceAndHasItsFiles(t *testing.T) {
	for _, c := range loadCases(t) {
		if strings.TrimSpace(c.Source) == "" {
			t.Errorf("case %s has no source", c.Name)
		}
		if c.Plan == "" || c.Document == "" {
			t.Errorf("case %s needs a plan and a document", c.Name)
			continue
		}
		loadPlan(t, c)
		readFixture(t, c.Document)
	}
}

func TestAssemblesTheDocumentAndReportsTheGaps(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			document, problems := newAssembler().Assemble(loadPlan(t, c))
			if want := string(readFixture(t, c.Document)); document != want {
				t.Errorf("document differs\n--- got ---\n%s\n--- want ---\n%s", document, want)
			}
			if got, want := problemsOf(problems), expectedProblems(c); !reflect.DeepEqual(got, want) {
				t.Errorf("problems differ\ngot  %+v\nwant %+v", got, want)
			}
		})
	}
}

func TestGapTableHasOneRowForEachProblem(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			document, problems := newAssembler().Assemble(loadPlan(t, c))
			rows, found := gapTableRows(document)
			if !found {
				t.Fatalf("the document has no gap table\n%s", document)
			}
			if len(rows) != len(problems) {
				t.Errorf("the table has %d rows and there are %d problems\n%s", len(rows), len(problems), document)
			}
		})
	}
}

func TestEndsTheDocumentWithOneLineBreak(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			document, _ := newAssembler().Assemble(loadPlan(t, c))
			if !strings.HasSuffix(document, "\n") || strings.HasSuffix(document, "\n\n") {
				t.Errorf("the document must end with exactly one line break, it ends with %q", document[max(0, len(document)-10):])
			}
			if strings.Contains(document, "\r") {
				t.Errorf("the document holds a carriage return")
			}
		})
	}
}

func TestLeavesThePlanAsItWas(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			plan := loadPlan(t, c)
			newAssembler().Assemble(plan)
			if !reflect.DeepEqual(plan, loadPlan(t, c)) {
				t.Errorf("the plan changed")
			}
		})
	}
}

func TestChangingTheProblemsChangesNothingInThePlan(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			plan := loadPlan(t, c)
			assembler := newAssembler()
			_, problems := assembler.Assemble(plan)
			for i := range problems {
				problems[i].Code = "changed"
				problems[i].Detail = map[string]string{"changed": "changed"}
			}
			if !reflect.DeepEqual(plan, loadPlan(t, c)) {
				t.Errorf("the plan changed")
			}
			again, _ := assembler.Assemble(plan)
			if want := string(readFixture(t, c.Document)); again != want {
				t.Errorf("the next call gave another document")
			}
		})
	}
}

func TestKeepsNoMemoryBetweenCalls(t *testing.T) {
	cases := loadCases(t)
	assembler := newAssembler()
	check := func(c fixtureCase) {
		t.Helper()
		document, problems := assembler.Assemble(loadPlan(t, c))
		if want := string(readFixture(t, c.Document)); document != want {
			t.Errorf("%s: the document differs on a shared assembler\n--- got ---\n%s\n--- want ---\n%s", c.Name, document, want)
		}
		if got, want := problemsOf(problems), expectedProblems(c); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: problems differ on a shared assembler\ngot  %+v\nwant %+v", c.Name, got, want)
		}
	}
	for _, c := range cases {
		check(c)
	}
	for i := len(cases) - 1; i >= 0; i-- {
		check(cases[i])
	}
}

func TestGivesTheSameDocumentWhateverTheOrderOfSections(t *testing.T) {
	for _, c := range loadCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			want := string(readFixture(t, c.Document))
			plan := loadPlan(t, c)
			count := len(plan.Sections)
			for shift := 1; shift < count; shift++ {
				rotated := plan
				rotated.Sections = append(append([]model.Section{}, plan.Sections[shift:]...), plan.Sections[:shift]...)
				if document, _ := newAssembler().Assemble(rotated); document != want {
					t.Errorf("rotating the sections by %d changed the document\n%s", shift, document)
				}
			}
			reversed := plan
			reversed.Sections = nil
			for i := count - 1; i >= 0; i-- {
				reversed.Sections = append(reversed.Sections, plan.Sections[i])
			}
			if document, _ := newAssembler().Assemble(reversed); document != want {
				t.Errorf("reversing the sections changed the document\n%s", document)
			}
		})
	}
}

func TestAssemblesThePlanWithNothingInIt(t *testing.T) {
	document, problems := newAssembler().Assemble(model.Plan{})
	if len(problems) != 0 {
		t.Errorf("got problems %+v", problems)
	}
	if !strings.HasSuffix(document, "\n") {
		t.Errorf("the document must end with a line break")
	}
}
