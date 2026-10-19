package runner_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"onboard-kit/acceptance/runner"
)

var fakeBinaryPath string

func TestMain(m *testing.M) {
	directory, err := os.MkdirTemp("", "fake-binary-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fakeBinaryPath = filepath.Join(directory, "fake")
	build := exec.Command("go", "build", "-o", fakeBinaryPath, "./testdata/fake")
	if output, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "building the fake binary failed: %v\n%s", err, output)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(directory)
	os.Exit(code)
}

func runScenario(t *testing.T, text string) []runner.Failure {
	t.Helper()
	return runScenarioWith(t, text, runner.Options{TimeLimit: 20 * time.Second}, nil)
}

func runScenarioWith(t *testing.T, text string, options runner.Options, repositoryFiles map[string]string) []runner.Failure {
	t.Helper()
	scenario, err := runner.Parse([]byte(text))
	if err != nil {
		t.Fatalf("the scenario should parse: %v", err)
	}
	options.BinaryPath = fakeBinaryPath
	options.RepositoryRoot = t.TempDir()
	for path, content := range repositoryFiles {
		target := filepath.Join(options.RepositoryRoot, path)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return runner.Run(scenario, options)
}

func scenarioWithStep(step string) string {
	return "name: sample\nsteps:\n" + step
}

func requirePass(t *testing.T, failures []runner.Failure) {
	t.Helper()
	for _, failure := range failures {
		t.Errorf("unexpected failure: %s", failure)
	}
}

func requireFailure(t *testing.T, failures []runner.Failure, step int, expectation string) runner.Failure {
	t.Helper()
	if len(failures) == 0 {
		t.Fatalf("expected a failure of %q at step %d, got none", expectation, step)
	}
	for _, failure := range failures {
		if failure.Step == step && failure.Expectation == expectation {
			return failure
		}
	}
	t.Fatalf("expected a failure of %q at step %d, got %v", expectation, step, failures)
	return runner.Failure{}
}

func TestPassingScenario(t *testing.T) {
	t.Parallel()
	failures := runScenarioWith(t, `
name: passing
description: Several steps share one folder.
source: somewhere
steps:
  - run: [onboard-kit, write, made.txt, hello]
    files:
      input.txt: text
      copied.txt: {from: golden/copy.txt}
    expect:
      exit: 0
      files:
        made.txt: {same-as: golden/made.txt}
        input.txt: {same-as: golden/input.txt}
        copied.txt: {same-as: golden/copy.txt}
  - run: [onboard-kit, copy, made.txt, second.txt]
    expect:
      exit: 0
      files:
        second.txt: {same-as: golden/made.txt}
      unchanged: [made.txt]
`, runner.Options{TimeLimit: 20 * time.Second}, map[string]string{
		"golden/made.txt":  "hello",
		"golden/input.txt": "text",
		"golden/copy.txt":  "copy me",
	})
	requirePass(t, failures)
}

func TestFailureNamesTheStepAndExpectation(t *testing.T) {
	t.Parallel()
	failures := runScenario(t, `
name: sample
steps:
  - run: [onboard-kit, say, fine, "", "0"]
    expect: {exit: 0, stdout: {equals: fine}}
  - run: [onboard-kit, say, "", oops, "3"]
    expect: {exit: 0}
  - run: [onboard-kit, say, never, "", "0"]
    expect: {exit: 1}
`)
	failure := requireFailure(t, failures, 2, "exit")
	if !strings.Contains(failure.Message, "expected 0, got 3") {
		t.Errorf("the message should give both codes, got %q", failure.Message)
	}
	if len(failures) != 1 {
		t.Errorf("a later step should not run after a failure, got %v", failures)
	}
	if !strings.HasPrefix(failure.String(), "step 2, exit: ") {
		t.Errorf("the text should start with the step and expectation, got %q", failure.String())
	}
}

func TestExpectationKinds(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		step        string
		expectation string
	}{
		{"exit matches", "run: [onboard-kit, say, a, b, '4']\n    expect: {exit: 4}", ""},
		{"exit differs", "run: [onboard-kit, say, a, b, '4']\n    expect: {exit: 0}", "exit"},
		{"exit zero is checked", "run: [onboard-kit, say, a, b, '0']\n    expect: {exit: 1}", "exit"},
		{"stdout contains matches", "run: [onboard-kit, say, hello world, '', '0']\n    expect: {stdout: {contains: lo wo}}", ""},
		{"stdout contains differs", "run: [onboard-kit, say, hello, '', '0']\n    expect: {stdout: {contains: bye}}", "stdout contains"},
		{"stdout equals matches", "run: [onboard-kit, say, hello, '', '0']\n    expect: {stdout: {equals: hello}}", ""},
		{"stdout equals differs", "run: [onboard-kit, say, hello, '', '0']\n    expect: {stdout: {equals: hell}}", "stdout equals"},
		{"stdout json matches", `run: [onboard-kit, say, '{"a":1,"b":2}', '', '0']` + "\n    expect: {stdout: {json: {a: 1}}}", ""},
		{"stdout json differs", `run: [onboard-kit, say, '{"a":1}', '', '0']` + "\n    expect: {stdout: {json: {a: 2}}}", "stdout json"},
		{"stdout json on text", "run: [onboard-kit, say, plain, '', '0']\n    expect: {stdout: {json: {a: 1}}}", "stdout json"},
		{"stdout left out", "run: [onboard-kit, say, anything, anything, '0']\n    expect: {exit: 0}", ""},
		{"stderr contains matches", "run: [onboard-kit, say, '', boom, '1']\n    expect: {stderr: {contains: boo}}", ""},
		{"stderr contains differs", "run: [onboard-kit, say, '', boom, '1']\n    expect: {stderr: {contains: calm}}", "stderr contains"},
		{"stderr equals differs", "run: [onboard-kit, say, '', boom, '1']\n    expect: {stderr: {equals: boo}}", "stderr equals"},
		{"stderr json matches", `run: [onboard-kit, say, '', '{"error":"x","extra":1}', '1']` + "\n    expect: {stderr: {json: {error: x}}}", ""},
		{"stderr json differs", `run: [onboard-kit, say, '', '{"error":"x"}', '1']` + "\n    expect: {stderr: {json: {error: y}}}", "stderr json"},
		{"file json matches", `run: [onboard-kit, write, out.json, '{"a":[1,2],"b":null}']` + "\n    expect: {files: {out.json: {json: {a: [1, 2]}}}}", ""},
		{"file json differs", `run: [onboard-kit, write, out.json, '{"a":[1,2]}']` + "\n    expect: {files: {out.json: {json: {a: [2, 1]}}}}", "file out.json json"},
		{"file json on text", "run: [onboard-kit, write, out.json, nope]\n    expect: {files: {out.json: {json: {a: 1}}}}", "file out.json json"},
		{"file same-as differs", "run: [onboard-kit, write, out.txt, changed]\n    expect: {files: {out.txt: {same-as: golden.txt}}}", "file out.txt same-as"},
		{"file same-as missing file", "run: [onboard-kit, say, '', '', '0']\n    expect: {files: {out.txt: {same-as: golden.txt}}}", "file out.txt"},
		{"file same-as missing golden", "run: [onboard-kit, write, out.txt, golden]\n    expect: {files: {out.txt: {same-as: absent-golden.txt}}}", "file out.txt same-as"},
		{"file absent matches", "run: [onboard-kit, say, '', '', '0']\n    expect: {files: {out.txt: {absent: true}}}", ""},
		{"file absent differs", "run: [onboard-kit, write, out.txt, x]\n    expect: {files: {out.txt: {absent: true}}}", "file out.txt absent"},
		{"stdin reaches the process", "run: [onboard-kit, stdin]\n    stdin: typed\n    expect: {stdout: {equals: typed}}", ""},
		{"placeholder replaced only first", "run: [onboard-kit, arguments, onboard-kit]\n    expect: {stdout: {json: [onboard-kit]}}", ""},
		{"unchanged holds", "run: [onboard-kit, say, '', '', '0']\n    files: {keep.txt: same}\n    expect: {unchanged: [keep.txt]}", ""},
		{"unchanged broken by rewrite", "run: [onboard-kit, write, keep.txt, different]\n    files: {keep.txt: same}\n    expect: {unchanged: [keep.txt]}", "unchanged keep.txt"},
		{"unchanged broken by creation", "run: [onboard-kit, write, new.txt, x]\n    expect: {unchanged: [new.txt]}", "unchanged new.txt"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			failures := runScenarioWith(t, scenarioWithStep("  - "+testCase.step+"\n"), runner.Options{TimeLimit: 20 * time.Second}, map[string]string{"golden.txt": "golden"})
			if testCase.expectation == "" {
				requirePass(t, failures)
				return
			}
			requireFailure(t, failures, 1, testCase.expectation)
		})
	}
}

func TestJSONSubsetRules(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		actual   string
		expected string
		matches  bool
	}{
		{"extra keys are fine", `{"a":1,"b":2}`, `{a: 1}`, true},
		{"missing key", `{"b":2}`, `{a: 1}`, false},
		{"key order does not matter", `{"b":2,"a":1}`, `{a: 1, b: 2}`, true},
		{"empty object matches any object", `{"a":1}`, `{}`, true},
		{"nested objects by subset", `{"a":{"b":1,"c":2},"d":3}`, `{a: {b: 1}}`, true},
		{"nested mismatch", `{"a":{"b":2}}`, `{a: {b: 1}}`, false},
		{"same list in order", `[1,2,3]`, `[1, 2, 3]`, true},
		{"list order matters", `[1,2,3]`, `[3, 2, 1]`, false},
		{"list order matters for objects", `[{"a":1},{"a":2}]`, `[{a: 2}, {a: 1}]`, false},
		{"list needs the same length when longer", `[1,2,3]`, `[1, 2]`, false},
		{"list needs the same length when shorter", `[1]`, `[1, 2]`, false},
		{"empty list matches empty list", `{"problems":[]}`, `{problems: []}`, true},
		{"empty list does not match full list", `{"problems":["x"]}`, `{problems: []}`, false},
		{"objects in lists match by subset", `[{"a":1,"b":2}]`, `[{a: 1}]`, true},
		{"null matches null", `{"a":null}`, `{a: null}`, true},
		{"null does not match a value", `{"a":0}`, `{a: null}`, false},
		{"null does not match a missing key", `{}`, `{a: null}`, false},
		{"a value does not match null", `{"a":null}`, `{a: 1}`, false},
		{"top level null", `null`, `null`, true},
		{"strings match exactly", `{"a":"x"}`, `{a: x}`, true},
		{"a quoted number is text", `{"a":"1"}`, `{a: 1}`, false},
		{"numbers match across forms", `{"a":1.0}`, `{a: 1}`, true},
		{"booleans", `{"a":true}`, `{a: false}`, false},
		{"object against list", `[1]`, `{a: 1}`, false},
		{"list against object", `{"a":1}`, `[1]`, false},
		{"yes stays text", `{"a":"yes"}`, `{a: yes}`, true},
		{"a date stays text", `{"a":"2026-10-15"}`, `{a: 2026-10-15}`, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			step := fmt.Sprintf("  - run: [onboard-kit, say, '%s', '', '0']\n    expect: {stdout: {json: %s}}\n", testCase.actual, testCase.expected)
			failures := runScenario(t, scenarioWithStep(step))
			if testCase.matches {
				requirePass(t, failures)
				return
			}
			requireFailure(t, failures, 1, "stdout json")
		})
	}
}

func TestStdioTalk(t *testing.T) {
	t.Parallel()
	failures := runScenario(t, `
name: sample
steps:
  - run: [onboard-kit, stdio]
    talk:
      - send: {command: questions}
        expect: {ok: true, count: 1, problems: []}
      - send: {command: submit, value: established}
        expect: {request: {value: established}, count: 2}
      - send: {command: noted}
    expect:
      exit: 0
      stdout: {contains: '"count":3'}
`)
	requirePass(t, failures)
}

func TestStdioResponseMismatchNamesTheExchange(t *testing.T) {
	t.Parallel()
	failures := runScenario(t, `
name: sample
steps:
  - run: [onboard-kit, stdio]
    talk:
      - send: {command: questions}
        expect: {ok: true}
      - send: {command: submit}
        expect: {ok: false}
    expect: {exit: 0}
`)
	requireFailure(t, failures, 1, "talk 2")
}

func TestStdioResponseThatIsNotJSON(t *testing.T) {
	t.Parallel()
	failures := runScenario(t, `
name: sample
steps:
  - run: [onboard-kit, stdio]
    talk:
      - send: {command: garble}
        expect: {ok: true}
    expect: {exit: 0}
`)
	requireFailure(t, failures, 1, "talk 1")
}

func TestStdioProcessThatStopsAnswering(t *testing.T) {
	t.Parallel()
	failures := runScenario(t, `
name: sample
steps:
  - run: [onboard-kit, stdio]
    talk:
      - send: {command: quit}
        expect: {ok: true}
    expect: {exit: 0}
`)
	requireFailure(t, failures, 1, "talk 1")
}

func TestStdioExitIsChecked(t *testing.T) {
	t.Parallel()
	failures := runScenario(t, `
name: sample
steps:
  - run: [onboard-kit, stdio]
    talk:
      - send: {command: questions}
    expect: {exit: 2}
`)
	requireFailure(t, failures, 1, "exit")
}

func TestStdioClosesInputAfterTheList(t *testing.T) {
	t.Parallel()
	failures := runScenario(t, `
name: sample
steps:
  - run: [onboard-kit, stdio]
    talk: []
    expect: {exit: 0}
`)
	requirePass(t, failures)
}

func TestTimeLimitKillsARunThatDoesNotFinish(t *testing.T) {
	t.Parallel()
	started := time.Now()
	failures := runScenarioWith(t, scenarioWithStep("  - run: [onboard-kit, sleep]\n    expect: {exit: 0}\n"), runner.Options{TimeLimit: 300 * time.Millisecond}, nil)
	requireFailure(t, failures, 1, "time limit")
	if elapsed := time.Since(started); elapsed > 10*time.Second {
		t.Errorf("the run should be killed near the limit, took %s", elapsed)
	}
}

func TestTimeLimitKillsAStdioRunThatDoesNotAnswer(t *testing.T) {
	t.Parallel()
	started := time.Now()
	failures := runScenarioWith(t, `
name: sample
steps:
  - run: [onboard-kit, stdio]
    talk:
      - send: {command: hang}
        expect: {ok: true}
`, runner.Options{TimeLimit: 300 * time.Millisecond}, nil)
	requireFailure(t, failures, 1, "time limit")
	if elapsed := time.Since(started); elapsed > 10*time.Second {
		t.Errorf("the run should be killed near the limit, took %s", elapsed)
	}
}

func TestFolderIsRemovedAfterwards(t *testing.T) {
	t.Parallel()
	temporaryRoot := t.TempDir()
	failures := runScenarioWith(t, scenarioWithStep("  - run: [onboard-kit, write, left.txt, here]\n  - run: [onboard-kit, say, '', '', '1']\n    expect: {exit: 0}\n"), runner.Options{TimeLimit: 20 * time.Second, TemporaryRoot: temporaryRoot}, nil)
	requireFailure(t, failures, 2, "exit")
	leftovers, err := os.ReadDir(temporaryRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(leftovers) != 0 {
		t.Errorf("the temporary folder should be removed, found %d entries", len(leftovers))
	}
}

func TestRunsHappenInsideTheTemporaryFolder(t *testing.T) {
	t.Parallel()
	temporaryRoot := t.TempDir()
	failures := runScenarioWith(t, scenarioWithStep("  - run: [onboard-kit, folder]\n    expect: {stdout: {contains: scenario-}}\n"), runner.Options{TimeLimit: 20 * time.Second, TemporaryRoot: temporaryRoot}, nil)
	requirePass(t, failures)
}

func TestEachScenarioGetsItsOwnFolder(t *testing.T) {
	t.Parallel()
	first := runScenario(t, "name: one\nsteps:\n  - run: [onboard-kit, write, shared.txt, one]\n")
	requirePass(t, first)
	second := runScenario(t, "name: two\nsteps:\n  - run: [onboard-kit, say, '', '', '0']\n    expect: {files: {shared.txt: {absent: true}}}\n")
	requirePass(t, second)
}

func TestSetupFailures(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"missing from file":   "  - run: [onboard-kit, say, '', '', '0']\n    files: {a.txt: {from: nowhere.txt}}\n",
		"from escapes repo":   "  - run: [onboard-kit, say, '', '', '0']\n    files: {a.txt: {from: ../outside.txt}}\n",
		"expected file path":  "  - run: [onboard-kit, say, '', '', '0']\n    expect: {files: {../outside.txt: {absent: true}}}\n",
		"unchanged path":      "  - run: [onboard-kit, say, '', '', '0']\n    expect: {unchanged: [../outside.txt]}\n",
		"program not found":   "  - run: [/nonexistent/program]\n",
		"stdio not startable": "  - run: [/nonexistent/program]\n    talk: []\n",
	}
	for name, step := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			failures := runScenario(t, scenarioWithStep(step))
			if len(failures) == 0 {
				t.Fatal("expected a failure")
			}
			if failures[0].Step != 1 {
				t.Errorf("the failure should name step 1, got %v", failures[0])
			}
		})
	}
}

func TestScenarioThatCannotParse(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"empty":                "",
		"not yaml":             "name: [unclosed",
		"no name":              "steps:\n  - run: [onboard-kit]\n",
		"no steps":             "name: sample\n",
		"step without run":     "name: sample\nsteps:\n  - expect: {exit: 0}\n",
		"unknown scenario key": "name: sample\nsteps:\n  - run: [onboard-kit]\nextra: 1\n",
		"unknown step key":     "name: sample\nsteps:\n  - run: [onboard-kit]\n    surprise: 1\n",
		"unknown expect key":   "name: sample\nsteps:\n  - run: [onboard-kit]\n    expect: {exits: 0}\n",
		"unknown output key":   "name: sample\nsteps:\n  - run: [onboard-kit]\n    expect: {stdout: {matches: x}}\n",
		"unknown file key":     "name: sample\nsteps:\n  - run: [onboard-kit]\n    expect: {files: {a: {like: b}}}\n",
		"exit is text":         "name: sample\nsteps:\n  - run: [onboard-kit]\n    expect: {exit: zero}\n",
		"talk with stdin":      "name: sample\nsteps:\n  - run: [onboard-kit, stdio]\n    stdin: x\n    talk: []\n",
		"send is missing":      "name: sample\nsteps:\n  - run: [onboard-kit, stdio]\n    talk:\n      - expect: {ok: true}\n",
		"file path escapes":    "name: sample\nsteps:\n  - run: [onboard-kit]\n    files: {../a: text}\n",
		"file source unknown":  "name: sample\nsteps:\n  - run: [onboard-kit]\n    files: {a: {copy: b}}\n",
	}
	for name, text := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := runner.Parse([]byte(text)); err == nil {
				t.Error("expected a parse error")
			}
		})
	}
}

func TestLoadChecksTheNameAgainstTheFileName(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	good := filepath.Join(directory, "sample.yaml")
	bad := filepath.Join(directory, "other.yaml")
	text := []byte("name: sample\nsteps:\n  - run: [onboard-kit]\n")
	for _, path := range []string{good, bad} {
		if err := os.WriteFile(path, text, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := runner.Load(good); err != nil {
		t.Errorf("a matching name should load: %v", err)
	}
	if _, err := runner.Load(bad); err == nil {
		t.Error("a name that differs from the file name should not load")
	}
	if _, err := runner.Load(filepath.Join(directory, "missing.yaml")); err == nil {
		t.Error("a missing file should not load")
	}
}
