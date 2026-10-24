package runner_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"onboard-kit/acceptance/runner"
)

func captureFixtureDispatch() bool {
	position := -1
	for index, argument := range os.Args {
		if argument == "--capture-fixture" {
			position = index
			break
		}
	}
	if position < 0 {
		return false
	}
	arguments := os.Args[position+1:]
	if len(arguments) < 2 {
		os.Exit(90)
	}
	proof, err := os.OpenFile(arguments[0], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		os.Exit(91)
	}
	if _, err := fmt.Fprintln(proof, arguments[1]); err != nil {
		os.Exit(92)
	}
	if err := proof.Close(); err != nil {
		os.Exit(93)
	}
	mode, args := arguments[1], arguments[2:]
	checked := func(err error) {
		if err != nil {
			os.Exit(94)
		}
	}
	switch mode {
	case "noop":
	case "echo":
		body, err := json.Marshal(args)
		checked(err)
		_, err = os.Stdout.Write(body)
		checked(err)
	case "stdin":
		_, err := io.Copy(os.Stdout, os.Stdin)
		checked(err)
	case "root":
		cwd, err := os.Getwd()
		checked(err)
		if len(args) != 1 || args[0] != cwd {
			os.Exit(95)
		}
		fmt.Print("physical-root")
	case "read":
		body, err := os.ReadFile(args[0])
		checked(err)
		_, err = os.Stdout.Write(body)
		checked(err)
	case "symlink":
		checked(os.Symlink(args[1], args[0]))
	case "seed":
		checked(os.Mkdir("entries", 0700))
		switch args[0] {
		case "empty":
		case "one":
			checked(os.Mkdir("entries/only", 0700))
		case "one-deep":
			checked(os.Mkdir("entries/only", 0700))
			checked(os.WriteFile("entries/only/private", []byte{0xff}, 0600))
		case "two":
			checked(os.Mkdir("entries/one", 0700))
			checked(os.Mkdir("entries/two", 0700))
		case "hidden":
			checked(os.Mkdir("entries/only", 0700))
			checked(os.Mkdir("entries/.hidden", 0700))
		case "file":
			checked(os.WriteFile("entries/only", []byte("not a directory"), 0600))
		case "link":
			checked(os.Mkdir("real", 0700))
			checked(os.Symlink("../real", "entries/only"))
		case "invalid-name":
			checked(os.Mkdir("entries/"+string([]byte{0xff}), 0700))
		default:
			os.Exit(96)
		}
	default:
		os.Exit(97)
	}
	os.Exit(0)
	return true
}

func TestCaptureFixtureProcess(t *testing.T) {
	if captureFixtureDispatch() {
		t.Fatal("fixture dispatcher returned after native child exit")
	}
}

type captureResult struct {
	failures []runner.Failure
	calls    []string
}

func captureCommand(mode string, args ...string) string {
	values := []string{"onboard-kit", "-test.run=^TestCaptureFixtureProcess$", "--", "--capture-fixture"}
	parts := make([]string, 0, len(values)+len(args)+2)
	for _, value := range values {
		parts = append(parts, strconv.Quote(value))
	}
	parts = append(parts, "CAPTURE_PROOF_PATH", strconv.Quote(mode))
	for _, arg := range args {
		parts = append(parts, strconv.Quote(arg))
	}
	return "run: [" + strings.Join(parts, ", ") + "]\n"
}

func captureRun(t *testing.T, steps string, repositoryFiles map[string]string) captureResult {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	proof := filepath.Join(base, "proof")
	repository := filepath.Join(base, "repository")
	if err := os.Mkdir(repository, 0700); err != nil {
		t.Fatal(err)
	}
	for path, text := range repositoryFiles {
		target := filepath.Join(repository, path)
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	text := "name: capture-probe\nsteps:\n" + strings.ReplaceAll(steps, "CAPTURE_PROOF_PATH", strconv.Quote(proof))
	scenario, err := runner.Parse([]byte(text))
	if err != nil {
		t.Fatalf("independently valid extension schema must parse: %v", err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	failures := runner.Run(scenario, runner.Options{BinaryPath: executable, RepositoryRoot: repository, TemporaryRoot: base, TimeLimit: 10 * time.Second})
	fresh, err := runner.Parse([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(scenario, fresh) {
		t.Fatal("Run mutated the supplied parsed scenario")
	}
	body, err := os.ReadFile(proof)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	var calls []string
	if len(body) != 0 {
		calls = strings.Split(strings.TrimSuffix(string(body), "\n"), "\n")
	}
	return captureResult{failures, calls}
}

func capturePass(t *testing.T, got captureResult, calls ...string) {
	t.Helper()
	if len(got.failures) != 0 {
		t.Fatalf("expected actual passing runner execution, got %v", got.failures)
	}
	if strings.Join(got.calls, "\n") != strings.Join(calls, "\n") {
		t.Fatalf("actual native child proof: got %q, want %q", got.calls, calls)
	}
}

func captureFail(t *testing.T, got captureResult, step int, category string, calls ...string) {
	t.Helper()
	want := map[string]string{"interpolation": "Scenario interpolation is invalid.", "path safety": "Scenario path is not a safe local path.", "capture": "Capture could not be established.", "capture limit": "Capture budget exceeded."}[category]
	if want == "" {
		want = "Scenario expectation failed."
	}
	if len(got.failures) != 1 || got.failures[0].Step != step || got.failures[0].Expectation != category || got.failures[0].Message != want {
		t.Fatalf("exact static failure missing: %#v", got.failures)
	}
	if strings.Join(got.calls, "\n") != strings.Join(calls, "\n") {
		t.Fatalf("wrong actual child execution prefix: %q; want %q", got.calls, calls)
	}
}
