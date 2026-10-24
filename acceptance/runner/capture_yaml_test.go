//go:build darwin || linux

package runner_test

import (
	"fmt"
	"strings"
	"testing"
)

func captureYAMLStep(data string) (string, map[string]string) {
	return "  - " + captureCommand("noop") + `    files: {held.yaml: {from: payload.yaml}}
    expect: {exit: 0}
    capture: {HELD: {yaml-scalar: {file: held.yaml, path: [value]}}}
`, map[string]string{"payload.yaml": data}
}

func captureValueCheck(steps, expected string) string {
	return steps + "  - " + captureCommand("stdin") + "    stdin: '${HELD}'\n    expect: {exit: 0, stdout: {equals: " + fmt.Sprintf("%q", expected) + "}}\n"
}

func TestCaptureUsesWholeStrictYAMLProfileAndOnlyStringValues(t *testing.T) {
	for _, data := range []string{
		"", "# only comment\n", "value: [broken\n", "value: held\n---\nother: text\n",
		"value: held\n---\n", "value: held\nprivate: \"" + string([]byte{0xff}) + "\"\n",
		"value: held\nvalue: held\n", "value: held\nprivate: {x: one, x: two}\n",
		"value: held\nprivate: {\"x\": one, \"\\u0078\": two}\n",
		"value: held\nprivate: &unused x\n", "value: &used held\n",
		"value: held\nprivate: &x text\ncopy: *x\n", "value: held\nprivate: {1: text}\n",
		"value: null\n", "value: true\n", "value: 123\n", "value: 2026-10-24\n",
		"value: {}\n", "value: []\n", "other: held\n",
	} {
		t.Run(fmt.Sprintf("data-%x", data), func(t *testing.T) {
			steps, files := captureYAMLStep(data)
			captureFail(t, captureRun(t, steps, files), 1, "capture", "noop")
		})
	}
	for _, item := range []struct{ data, expected string }{
		{"value: ''\n", ""}, {"value: !!str 123\n", "123"}, {"---\nvalue: \"line\\nsecond\"\n...\n", "line\nsecond"},
	} {
		steps, files := captureYAMLStep(item.data)
		capturePass(t, captureRun(t, captureValueCheck(steps, item.expected), files), "noop", "stdin")
	}
}

func TestCaptureValueAndRawFileInclusiveBounds(t *testing.T) {
	for _, size := range []int{4096, 4097} {
		steps, files := captureYAMLStep("value: '" + strings.Repeat("v", size) + "'\n")
		if size == 4096 {
			capturePass(t, captureRun(t, captureValueCheck(steps, strings.Repeat("v", size)), files), "noop", "stdin")
		} else {
			captureFail(t, captureRun(t, steps, files), 1, "capture limit", "noop")
		}
	}
	steps, files := captureYAMLStep("value: \"nul\\u0000here\"\n")
	captureFail(t, captureRun(t, steps, files), 1, "capture limit", "noop")
	for _, size := range []int{1048576, 1048577} {
		prefix := "value: held\n#"
		steps, files := captureYAMLStep(prefix + strings.Repeat("p", size-len(prefix)))
		if size == 1048576 {
			capturePass(t, captureRun(t, captureValueCheck(steps, "held"), files), "noop", "stdin")
		} else {
			captureFail(t, captureRun(t, steps, files), 1, "capture", "noop")
		}
	}
}

func TestScalarEqualityAbsenceAndSequenceSelectionAreNotCoercion(t *testing.T) {
	for _, item := range []struct {
		data, path, predicate string
		pass                  bool
	}{
		{"value: ''\n", "[value]", "equals: ''", true},
		{"value: ' held '\n", "[value]", "equals: ' held '", true},
		{"value: ' held '\n", "[value]", "equals: held", false},
		{"value: 123\n", "[value]", "equals: '123'", false},
		{"value: null\n", "[value]", "absent: true", false},
		{"value: ''\n", "[value]", "absent: true", false},
		{"value: {}\n", "[value]", "absent: true", false},
		{"value: []\n", "[value]", "absent: true", false},
		{"other: held\n", "[value, deeper]", "absent: true", true},
		{"value: scalar\n", "[value, deeper]", "absent: true", false},
		{"value: [first, second]\n", "[value, '1']", "equals: second", true},
		{"value: [first, second]\n", "[value, '0']", "equals: first", true},
		{"value: [first, second]\n", "[value, '2']", "absent: true", true},
		{"value: [first]\n", "[value, '01']", "absent: true", false},
		{"value: [first]\n", "[value, '-1']", "absent: true", false},
		{"value: [first]\n", "[value, '18446744073709551616']", "absent: true", false},
		{"value: {'00': literal}\n", "[value, '00']", "equals: literal", true},
		{"value: [{on: '2026-10-24T12:34:56Z'}]\n", "[value, '0', on]", "utc-timestamp: true", true},
		{"other: held\nprivate: {x: one, x: two}\n", "[value]", "absent: true", false},
	} {
		t.Run(item.data+item.path+item.predicate, func(t *testing.T) {
			steps := "  - " + captureCommand("noop") + "    files: {held.yaml: {from: payload.yaml}}\n    expect:\n      exit: 0\n      files:\n        held.yaml: {yaml-scalars: [{path: " + item.path + ", " + item.predicate + "}]}\n"
			got := captureRun(t, steps, map[string]string{"payload.yaml": item.data})
			if item.pass {
				capturePass(t, got, "noop")
			} else {
				captureFail(t, got, 1, "file yaml-scalars", "noop")
			}
		})
	}
}

func TestCaptureAndEqualityUseTheSameSequenceSelector(t *testing.T) {
	steps := "  - " + captureCommand("noop") + `    files: {held.yaml: "items: [{value: first}, {value: second}]\n"}
    capture: {HELD: {yaml-scalar: {file: held.yaml, path: [items, '1', value]}}}
  - ` + captureCommand("echo", "${HELD}") + `    expect:
      exit: 0
      stdout: {json: [second]}
      files:
        held.yaml: {yaml-scalars: [{path: [items, '1', value], equals: '${HELD}'}]}
`
	capturePass(t, captureRun(t, steps, nil), "noop", "echo")
}
