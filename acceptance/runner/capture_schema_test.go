//go:build darwin || linux

package runner_test

import (
	"fmt"
	"strings"
	"testing"

	"onboard-kit/acceptance/runner"
)

func captureSchema(t *testing.T, body string, accepted bool) {
	t.Helper()
	_, err := runner.Parse([]byte("name: schema\nsteps:\n  - run: [onboard-kit, noop]\n" + body))
	if accepted && err != nil {
		t.Fatalf("valid schema refused: %v", err)
	}
	if !accepted && err == nil {
		t.Fatal("invalid extension schema accepted")
	}
}

func TestCaptureSchemaRejectsAmbiguousDeclarations(t *testing.T) {
	for _, body := range []string{
		"capture: null", "capture: []", "capture: held", "capture: {}",
		"capture: {SANDBOX: {unique-directory-entry: entries}}",
		"capture: {lower: {unique-directory-entry: entries}}",
		"capture: {1BAD: {unique-directory-entry: entries}}",
		"capture: {BAD-NAME: {unique-directory-entry: entries}}",
		"capture: {HELD: null}", "capture: {HELD: {}}",
		"capture: {HELD: {unique-directory-entry: ''}}",
		"capture: {HELD: {unique-directory-entry: 123}}",
		"capture: {HELD: {unique-directory-entry: entries, unknown: true}}",
		"capture: {HELD: {unique-directory-entry: entries, yaml-scalar: {file: a, path: [x]}}}",
		"capture: {HELD: {yaml-scalar: {file: a}}}",
		"capture: {HELD: {yaml-scalar: {file: a, path: []}}}",
		"capture: {HELD: {yaml-scalar: {file: a, path: [0]}}}",
		"capture: {HELD: {yaml-scalar: {file: a, path: ['']}}}",
		"capture: {HELD: {yaml-scalar: {file: a, path: [x], extra: true}}}",
		"capture: {HELD: {unique-directory-entry: entries}, HELD: {unique-directory-entry: entries}}",
		"capture: {HELD: {unique-directory-entry: entries, unique-directory-entry: entries}}",
		"capture: &held {HELD: {unique-directory-entry: entries}}",
		"capture: {HELD: {yaml-scalar: {file: a, path: [&held x]}}}",
	} {
		t.Run(body, func(t *testing.T) { captureSchema(t, "    "+body+"\n", false) })
	}
	for _, name := range []string{"A", "A_0", strings.Repeat("A", 64)} {
		captureSchema(t, "    capture: {"+name+": {unique-directory-entry: entries}}\n", true)
	}
	captureSchema(t, "    capture: {"+strings.Repeat("A", 65)+": {unique-directory-entry: entries}}\n", false)
	captureSchema(t, "    capture: {HELD: {unique-directory-entry: entries}}\n  - run: [onboard-kit, noop]\n    capture: {HELD: {unique-directory-entry: other}}\n", false)
}

func TestCaptureCountIsInclusiveScenarioTotal128(t *testing.T) {
	for _, count := range []int{128, 129} {
		var declarations strings.Builder
		for index := 0; index < count; index++ {
			fmt.Fprintf(&declarations, "      V%d: {yaml-scalar: {file: held.yaml, path: [value]}}\n", index)
		}
		body := "    files: {held.yaml: \"value: held\\n\"}\n    capture:\n" + declarations.String()
		captureSchema(t, body, count == 128)
		if count == 128 {
			steps := "  - " + captureCommand("noop") + body + "  - " + captureCommand("echo", "${V127}") + "    expect: {exit: 0, stdout: {json: [held]}}\n"
			capturePass(t, captureRun(t, steps, nil), "noop", "echo")
		}
	}
	var first, second strings.Builder
	for index := 0; index < 65; index++ {
		fmt.Fprintf(&first, "      A%d: {unique-directory-entry: entries}\n", index)
		fmt.Fprintf(&second, "      B%d: {unique-directory-entry: entries}\n", index)
	}
	captureSchema(t, "    capture:\n"+first.String()+"  - run: [onboard-kit, noop]\n    capture:\n"+second.String(), false)
}

func TestScalarPredicateSchemaHasExactlyOneStrictChoice(t *testing.T) {
	for _, value := range []string{
		"{path: [x], equals: ''}", "{path: [x], equals: text}",
		"{path: [x], absent: true}", "{path: [x], utc-timestamp: true}",
	} {
		captureSchema(t, "    expect: {files: {a.yaml: {yaml-scalars: ["+value+"]}}}\n", true)
	}
	for _, value := range []string{
		"{path: [x], equals: 1}", "{path: [x], equals: null}", "{path: [x], equals: true}",
		"{path: [x], equals: []}", "{path: [x], absent: false}", "{path: [x], absent: 'true'}",
		"{path: [x], absent: null}", "{path: [x], equals: x, absent: true}",
		"{path: [x], equals: x, utc-timestamp: true}", "{path: [x], absent: true, utc-timestamp: true}",
		"{path: [x], equals: x, equals: x}", "{path: [x], equals: x, unknown: y}",
		"{path: [x]}", "{path: [], equals: x}", "{path: [0], equals: x}",
		"{path: [&held x], equals: x}",
	} {
		t.Run(value, func(t *testing.T) {
			captureSchema(t, "    expect: {files: {a.yaml: {yaml-scalars: ["+value+"]}}}\n", false)
		})
	}
}
