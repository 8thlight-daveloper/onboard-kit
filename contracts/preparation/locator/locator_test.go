package locator_test

import (
	"encoding/json"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	contract "onboard-kit/contracts/preparation/locator"
	"onboard-kit/internal/model"
)

type expectedProblem struct {
	Code string `json:"code"`
	Path string `json:"path"`
}

type fixture struct {
	Name        string            `json:"name"`
	Source      string            `json:"source"`
	ResultsPath string            `json:"resultsPath"`
	Account     string            `json:"account"`
	Expected    contract.Paths    `json:"expected"`
	Problems    []expectedProblem `json:"problems"`
}

func readFixtures(t *testing.T) []fixture {
	t.Helper()
	data, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []fixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Fatal("empty fixture manifest")
	}
	seen := make(map[string]bool, len(fixtures))
	for _, f := range fixtures {
		if f.Name == "" || seen[f.Name] || f.Source == "" {
			t.Fatalf("fixture needs unique name and source: %#v", f)
		}
		seen[f.Name] = true
		if len(f.Problems) == 0 && (f.Expected.Record == "" || f.Expected.Preparation == "") {
			t.Fatalf("fixture %s has no observable expectation", f.Name)
		}
		if len(f.Problems) > 0 && f.Expected != (contract.Paths{}) {
			t.Fatalf("fixture %s returns paths after rejection", f.Name)
		}
	}
	return fixtures
}

func TestFixtures(t *testing.T) {
	t.Parallel()
	for _, f := range readFixtures(t) {
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			before := f
			locator := newLocator()
			got, problems := locator.Locate(f.ResultsPath, f.Account)
			if got != f.Expected {
				t.Errorf("paths %#v; want %#v", got, f.Expected)
			}
			checkProblems(t, problems, f.Problems)
			again, repeated := locator.Locate(f.ResultsPath, f.Account)
			if got != again || !reflect.DeepEqual(problems, repeated) {
				t.Fatal("not deterministic")
			}
			if !reflect.DeepEqual(f, before) {
				t.Fatal("input changed")
			}
		})
	}
}

func checkProblems(t *testing.T, got []model.Problem, expected []expectedProblem) {
	t.Helper()
	if len(got) != len(expected) {
		t.Fatalf("problems %#v; want %#v", got, expected)
	}
	for i, problem := range got {
		want := expected[i]
		if problem.Code != want.Code || problem.Severity != model.SeverityError || problem.Question != "" || problem.Section != "" || !reflect.DeepEqual(problem.Detail, map[string]string{"path": want.Path}) {
			t.Errorf("problem %d %#v; want code %s at %s", i, problem, want.Code, want.Path)
		}
		if strings.TrimSpace(problem.Message) == "" {
			t.Errorf("problem %d has no explanation", i)
		}
		for _, marker := range []string{"PrivatePersonIdentity", "PrivatePathIdentity"} {
			if strings.Contains(problem.Message, marker) {
				t.Errorf("diagnostic exposes supplied data: %q", problem.Message)
			}
		}
	}
}

func TestDiagnosticsDoNotShareState(t *testing.T) {
	t.Parallel()
	locator := newLocator()
	_, first := locator.Locate("../PrivatePathIdentity/{account}/record.obk.yaml", "newcomer-one")
	want := []expectedProblem{{Code: "preparation.results-path", Path: "resultsPath"}}
	checkProblems(t, first, want)
	first[0].Detail["path"] = "changed by caller"
	_, second := locator.Locate("../PrivatePathIdentity/{account}/record.obk.yaml", "newcomer-one")
	checkProblems(t, second, want)
	if first[0].Detail["path"] != "changed by caller" {
		t.Fatal("later call changed a previous result")
	}
}

func TestCallsDoNotKeepAccountOrPath(t *testing.T) {
	t.Parallel()
	locator := newLocator()
	first, problems := locator.Locate("", "newcomer-one")
	checkProblems(t, problems, nil)
	wantFirst := contract.Paths{Record: "records/newcomer-one/record.obk.yaml", Preparation: "records/newcomer-one/preparation.obk.yaml"}
	if first != wantFirst {
		t.Fatalf("first paths %#v; want %#v", first, wantFirst)
	}
	second, problems := locator.Locate("people/{account}/custom.yaml", "newcomer-two")
	checkProblems(t, problems, nil)
	wantSecond := contract.Paths{Record: "people/newcomer-two/custom.yaml", Preparation: "people/newcomer-two/preparation.obk.yaml"}
	if second != wantSecond || first != wantFirst {
		t.Fatalf("calls retained state: first %#v, second %#v", first, second)
	}
	_, problems = locator.Locate("", "")
	checkProblems(t, problems, []expectedProblem{{Code: "preparation.account", Path: "account"}})
	again, problems := locator.Locate("", "newcomer-one")
	checkProblems(t, problems, nil)
	if again != wantFirst {
		t.Fatalf("failure changed later resolution: %#v", again)
	}
}

func TestInvalidUTF8(t *testing.T) {
	t.Parallel()
	fixtures := []fixture{
		{Name: "invalid-path-byte", Source: "contract.md#path-syntax", ResultsPath: "records/{account}/rec\xfford", Account: "newcomer-one", Problems: []expectedProblem{{Code: "preparation.results-path", Path: "resultsPath"}}},
		{Name: "invalid-account-byte", Source: "contract.md#resolution-phases", Account: "person\xff", Problems: []expectedProblem{{Code: "preparation.account", Path: "account"}}},
	}
	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			paths, problems := newLocator().Locate(f.ResultsPath, f.Account)
			if paths != (contract.Paths{}) {
				t.Errorf("returned paths after invalid UTF-8: %#v", paths)
			}
			checkProblems(t, problems, f.Problems)
		})
	}
}

func TestExpandedPathBudget(t *testing.T) {
	t.Parallel()
	tooLong := []expectedProblem{{Code: "preparation.path-too-long", Path: "resultsPath"}}
	fixtures := []fixture{
		{
			Name: "record-exact-byte-budget", Source: "docs/decisions/0210-person-specific-preparation-paths.md; contract.md#resolution-phases",
			ResultsPath: "{account}/" + strings.Repeat("r", 4094), Account: "a",
			Expected: contract.Paths{Record: "a/" + strings.Repeat("r", 4094), Preparation: "a/preparation.obk.yaml"},
		},
		{
			Name: "record-one-byte-over", Source: "docs/decisions/0210-person-specific-preparation-paths.md; contract.md#resolution-phases",
			ResultsPath: "{account}/" + strings.Repeat("r", 4095), Account: "a", Problems: tooLong,
		},
		{
			Name: "preparation-exact-byte-budget", Source: "docs/decisions/0210-person-specific-preparation-paths.md; contract.md#resolution-phases",
			ResultsPath: strings.Repeat("d", 4073) + "/{account}/r", Account: "a",
			Expected: contract.Paths{Record: strings.Repeat("d", 4073) + "/a/r", Preparation: strings.Repeat("d", 4073) + "/a/preparation.obk.yaml"},
		},
		{
			Name: "preparation-only-one-byte-over", Source: "docs/decisions/0210-person-specific-preparation-paths.md; contract.md#resolution-phases",
			ResultsPath: strings.Repeat("d", 4074) + "/{account}/r", Account: "a", Problems: tooLong,
		},
		{
			Name: "convention-exact-byte-budget", Source: "docs/decisions/0210-person-specific-preparation-paths.md; contract.md#resolution-phases",
			Account:  strings.Repeat("a", 4067),
			Expected: contract.Paths{Record: "records/" + strings.Repeat("a", 4067) + "/record.obk.yaml", Preparation: "records/" + strings.Repeat("a", 4067) + "/preparation.obk.yaml"},
		},
		{
			Name: "convention-preparation-over-budget", Source: "docs/decisions/0210-person-specific-preparation-paths.md; contract.md#resolution-phases",
			Account: strings.Repeat("a", 4068), Problems: tooLong,
		},
		{
			Name: "unicode-exact-byte-budget", Source: "docs/decisions/0210-person-specific-preparation-paths.md; contract.md#resolution-phases",
			ResultsPath: "{account}/" + strings.Repeat("é", 2047), Account: "a",
			Expected: contract.Paths{Record: "a/" + strings.Repeat("é", 2047), Preparation: "a/preparation.obk.yaml"},
		},
		{
			Name: "unicode-bytes-not-runes", Source: "docs/decisions/0210-person-specific-preparation-paths.md; contract.md#resolution-phases",
			ResultsPath: "{account}/" + strings.Repeat("é", 2048), Account: "a", Problems: tooLong,
		},
		{
			Name: "bounded-input-rejects-amplification", Source: "docs/decisions/0210-person-specific-preparation-paths.md; contract.md#resolution-phases",
			ResultsPath: strings.Repeat("{account}", 4096) + "/record", Account: strings.Repeat("a", 8192), Problems: tooLong,
		},
		{
			Name: "syntax-before-budget", Source: "contract.md#resolution-phases",
			ResultsPath: "../{account}/record", Account: strings.Repeat("a", 4096), Problems: []expectedProblem{{Code: "preparation.results-path", Path: "resultsPath"}},
		},
		{
			Name: "directory-before-budget", Source: "contract.md#resolution-phases",
			ResultsPath: "records/{account}.yaml", Account: strings.Repeat("a", 4096), Problems: []expectedProblem{{Code: "preparation.account-directory", Path: "resultsPath"}},
		},
		{
			Name: "budget-before-collision", Source: "contract.md#resolution-phases",
			ResultsPath: "{account}/preparation.obk.yaml", Account: strings.Repeat("a", 4096), Problems: tooLong,
		},
	}
	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			got, problems := newLocator().Locate(f.ResultsPath, f.Account)
			if got != f.Expected {
				t.Errorf("path lengths record=%d preparation=%d; want record=%d preparation=%d", len(got.Record), len(got.Preparation), len(f.Expected.Record), len(f.Expected.Preparation))
			}
			checkProblems(t, problems, f.Problems)
		})
	}
}

func TestRejectsAmplificationBeforeAllocation(t *testing.T) {
	resultsPath := strings.Repeat("{account}", 4096) + "/record"
	account := strings.Repeat("a", 8192)
	locator := newLocator()
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	paths, problems := locator.Locate(resultsPath, account)
	runtime.ReadMemStats(&after)
	if paths != (contract.Paths{}) {
		t.Error("returned paths after exceeding the expansion budget")
	}
	checkProblems(t, problems, []expectedProblem{{Code: "preparation.path-too-long", Path: "resultsPath"}})
	if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 8*1024*1024 {
		t.Errorf("rejection allocated %d bytes; expansion must be rejected before constructing the paths", allocated)
	}
}
