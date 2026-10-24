//go:build darwin || linux

package runner_test

import (
	"fmt"
	"strings"
	"testing"
)

func TestCaptureDirectoryYAMLAndPriorValuesReachActualChildAndPaths(t *testing.T) {
	steps := "  - " + captureCommand("seed", "one") + `    files:
      held.yaml: "value: 'two words'\ndirectory: entries\n"
    expect: {exit: 0}
    capture:
      ITEM: {unique-directory-entry: entries}
      HELD: {yaml-scalar: {file: held.yaml, path: [value]}}
      DIRECTORY: {yaml-scalar: {file: held.yaml, path: [directory]}}
  - ` + captureCommand("root", "${SANDBOX}") + `    expect: {exit: 0, stdout: {equals: physical-root}}
    capture: {AGAIN: {unique-directory-entry: '${DIRECTORY}'}}
  - ` + captureCommand("echo", "${AGAIN}", "${HELD}", "$HOME", "$(touch NEVER)", "`literal`") + `    files:
      '${ITEM}.yaml': "value: 'two words'\n"
    expect:
      exit: 0
      stdout: {json: [only, two words, '$HOME', '$(touch NEVER)', '` + "`literal`" + `']}
      unchanged: ['${ITEM}.yaml']
      files:
        '${ITEM}.yaml': {yaml-scalars: [{path: [value], equals: '${HELD}'}]}
        NEVER: {absent: true}
    capture: {LATER: {yaml-scalar: {file: '${ITEM}.yaml', path: [value]}}}
  - ` + captureCommand("stdin") + `    stdin: '${HELD}'
    expect: {exit: 0, stdout: {equals: two words}}
`
	capturePass(t, captureRun(t, steps, nil), "seed", "root", "echo", "stdin")
}

func TestCaptureReplacementIsSinglePassAndFixtureBodiesStayLiteral(t *testing.T) {
	steps := "  - " + captureCommand("noop") + `    files: {held.yaml: "value: '${UNKNOWN}'\n"}
    expect: {exit: 0}
    capture: {TEXT: {yaml-scalar: {file: held.yaml, path: [value]}}}
  - ` + captureCommand("echo", "${TEXT}") + `    files: {literal.txt: '${TEXT}'}
    expect: {exit: 0, stdout: {json: ['${UNKNOWN}']}}
  - ` + captureCommand("read", "literal.txt") + `    expect: {exit: 0, stdout: {equals: '${TEXT}'}}
`
	capturePass(t, captureRun(t, steps, nil), "noop", "echo", "read")
}

func TestInvalidOriginalReferencesFailBeforeAnyChild(t *testing.T) {
	for _, token := range []string{"${MISSING}", "${}", "${lower}", "${UNFINISHED", "${A:-fallback}"} {
		t.Run(token, func(t *testing.T) {
			steps := "  - " + captureCommand("echo", token) + "    expect: {exit: 0}\n"
			captureFail(t, captureRun(t, steps, nil), 1, "interpolation")
		})
	}
	steps := "  - " + captureCommand("echo", "${OWN}") + `    files: {held.yaml: "value: held\n"}
    capture: {OWN: {yaml-scalar: {file: held.yaml, path: [value]}}}
`
	captureFail(t, captureRun(t, steps, nil), 1, "interpolation")
	steps = "  - " + captureCommand("noop") + `    files: {held.yaml: "value: next\n", next.yaml: "value: held\n"}
    capture:
      A: {yaml-scalar: {file: held.yaml, path: [value]}}
      B: {yaml-scalar: {file: '${A}.yaml', path: [value]}}
`
	captureFail(t, captureRun(t, steps, nil), 1, "interpolation")
	captureFail(t, captureRun(t, "  - run: ['${SANDBOX}/executable']\n", nil), 1, "interpolation")
}

func TestRepositoryReferencesAndSelectorKeysAreNotTemplated(t *testing.T) {
	steps := "  - " + captureCommand("noop") + `    files: {held.yaml: "'${KEY}': selected\n"}
    capture: {HELD: {yaml-scalar: {file: held.yaml, path: ['${KEY}']}}}
  - ` + captureCommand("read", "copied.txt") + `    files: {copied.txt: {from: 'literal/${HELD}.txt'}}
    expect: {exit: 0, stdout: {equals: literal-reference}}
`
	capturePass(t, captureRun(t, steps, map[string]string{"literal/${HELD}.txt": "literal-reference"}), "noop", "read")
}

func TestCaptureRunsOnlyAfterEveryOrdinaryExpectationPasses(t *testing.T) {
	steps := "  - " + captureCommand("noop") + `    expect: {exit: 7}
    capture: {HELD: {yaml-scalar: {file: missing.yaml, path: [value]}}}
  - ` + captureCommand("noop") + "    expect: {exit: 0}\n"
	captureFail(t, captureRun(t, steps, nil), 1, "exit", "noop")
}

func TestCaptureInvocationStateDoesNotLeakAcrossRuns(t *testing.T) {
	text := "name: isolated\nsteps:\n  - " + captureCommand("noop") + `    files: {held.yaml: "value: held\n"}
    capture: {HELD: {yaml-scalar: {file: held.yaml, path: [value]}}}
  - ` + captureCommand("echo", "${HELD}") + "    expect: {exit: 0}\n"
	for i := 0; i < 2; i++ {
		capturePass(t, captureRun(t, strings.TrimPrefix(text, "name: isolated\nsteps:\n"), nil), "noop", "echo")
	}
	captureFail(t, captureRun(t, "  - "+captureCommand("echo", "${HELD}"), nil), 1, "interpolation")
}

func TestCapturedPrivateValueNeverAppearsInLaterFailure(t *testing.T) {
	secret := "PRIVATE-FIXTURE-DO-NOT-DUMP"
	steps := "  - " + captureCommand("noop") + fmt.Sprintf("    files: {held.yaml: %q}\n", "value: "+secret+"\n") + `    capture: {HELD: {yaml-scalar: {file: held.yaml, path: [value]}}}
  - ` + captureCommand("echo", "${HELD}") + "    expect: {exit: 0, stdout: {equals: deliberately-wrong}}\n"
	got := captureRun(t, steps, nil)
	captureFail(t, got, 2, "stdout equals", "noop", "echo")
	for _, failure := range got.failures {
		if strings.Contains(failure.String(), secret) {
			t.Fatal("captured private value leaked")
		}
	}
}
