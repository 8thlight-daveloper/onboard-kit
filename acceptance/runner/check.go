package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

func check(expectation Expectation, result outcome, folder, repositoryRoot string, before map[string]fileState) []Failure {
	var failures []Failure
	if expectation.Exit != nil && *expectation.Exit != result.ExitCode {
		failures = append(failures, Failure{Expectation: "exit", Message: fmt.Sprintf("expected %d, got %d\nstdout: %q\nstderr: %q", *expectation.Exit, result.ExitCode, result.Stdout, result.Stderr)})
	}
	failures = append(failures, checkOutput("stdout", expectation.Stdout, result.Stdout)...)
	failures = append(failures, checkOutput("stderr", expectation.Stderr, result.Stderr)...)
	for _, path := range sortedKeys(expectation.Files) {
		failures = append(failures, checkFile(path, expectation.Files[path], folder, repositoryRoot)...)
	}
	for _, path := range expectation.Unchanged {
		failures = append(failures, checkUnchanged(path, folder, before[path])...)
	}
	return failures
}

func sortedKeys(files map[string]FileExpectation) []string {
	keys := make([]string, 0, len(files))
	for key := range files {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func checkOutput(name string, expectation *OutputExpectation, actual string) []Failure {
	if expectation == nil {
		return nil
	}
	var failures []Failure
	if expectation.JSON != nil {
		if message := jsonMismatch(expectation.JSON.Value, []byte(actual)); message != "" {
			failures = append(failures, Failure{Expectation: name + " json", Message: message})
		}
	}
	if expectation.Contains != nil && !strings.Contains(actual, *expectation.Contains) {
		failures = append(failures, Failure{Expectation: name + " contains", Message: fmt.Sprintf("%q not found in %q", *expectation.Contains, actual)})
	}
	if expectation.Equals != nil && *expectation.Equals != actual {
		failures = append(failures, Failure{Expectation: name + " equals", Message: fmt.Sprintf("expected %q, got %q", *expectation.Equals, actual)})
	}
	return failures
}

func jsonMismatch(expected any, content []byte) string {
	var actual any
	if err := json.Unmarshal(content, &actual); err != nil {
		return fmt.Sprintf("not JSON: %v in %q", err, content)
	}
	return subsetMismatch(expected, actual, "$")
}

func checkFile(path string, expectation FileExpectation, folder, repositoryRoot string) []Failure {
	label := "file " + path
	state, err := readState(path, folder)
	if err != nil {
		return []Failure{{Expectation: label, Message: err.Error()}}
	}
	if expectation.Absent {
		if state.Exists {
			return []Failure{{Expectation: label + " absent", Message: "the file exists"}}
		}
		return nil
	}
	if !state.Exists {
		return []Failure{{Expectation: label, Message: "the file does not exist"}}
	}
	var failures []Failure
	if expectation.SameAs != nil {
		failures = append(failures, checkSameAs(label, *expectation.SameAs, state.Content, repositoryRoot)...)
	}
	if expectation.JSON != nil {
		if message := jsonMismatch(expectation.JSON.Value, state.Content); message != "" {
			failures = append(failures, Failure{Expectation: label + " json", Message: message})
		}
	}
	return failures
}

func checkSameAs(label, reference string, content []byte, repositoryRoot string) []Failure {
	origin, err := resolve(repositoryRoot, reference)
	if err != nil {
		return []Failure{{Expectation: label + " same-as", Message: err.Error()}}
	}
	wanted, err := os.ReadFile(origin)
	if err != nil {
		return []Failure{{Expectation: label + " same-as", Message: err.Error()}}
	}
	if !bytes.Equal(wanted, content) {
		return []Failure{{Expectation: label + " same-as", Message: fmt.Sprintf("differs from %s\nexpected: %q\nactual:   %q", reference, wanted, content)}}
	}
	return nil
}

func checkUnchanged(path, folder string, before fileState) []Failure {
	after, err := readState(path, folder)
	if err != nil {
		return []Failure{{Expectation: "unchanged " + path, Message: err.Error()}}
	}
	if after.Exists != before.Exists || !bytes.Equal(after.Content, before.Content) {
		return []Failure{{Expectation: "unchanged " + path, Message: fmt.Sprintf("changed from %q to %q", before.Content, after.Content)}}
	}
	return nil
}
