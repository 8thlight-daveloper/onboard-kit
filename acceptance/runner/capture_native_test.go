//go:build darwin || linux

package runner_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestUniqueDirectoryCaptureCountsAllDirectEntriesAndRejectsAliases(t *testing.T) {
	for _, kind := range []string{"empty", "one", "one-deep", "two", "hidden", "file", "link", "invalid-name"} {
		t.Run(kind, func(t *testing.T) {
			if kind == "invalid-name" {
				if err := os.Mkdir(filepath.Join(t.TempDir(), string([]byte{0xff})), 0700); err != nil {
					t.Skipf("actual host cannot create an invalid-UTF8 native basename: %v", err)
				}
			}
			steps := "  - " + captureCommand("seed", kind) + "    expect: {exit: 0}\n    capture: {ITEM: {unique-directory-entry: entries}}\n"
			got := captureRun(t, steps, nil)
			switch kind {
			case "one", "one-deep":
				capturePass(t, got, "seed")
			case "link":
				captureFail(t, got, 1, "path safety", "seed")
			case "invalid-name":
				captureFail(t, got, 1, "capture limit", "seed")
			default:
				captureFail(t, got, 1, "capture", "seed")
			}
		})
	}
}

func TestExpandedPathsAreValidatedBeforeSetupAndCannotCollide(t *testing.T) {
	for _, value := range []string{"../escape", "/absolute", "a/../escape"} {
		steps := "  - " + captureCommand("noop") + fmt.Sprintf("    files: {held.yaml: %q}\n", "value: '"+value+"'\n") + `    capture: {HELD: {yaml-scalar: {file: held.yaml, path: [value]}}}
  - ` + captureCommand("noop") + "    files: {'${HELD}': forbidden}\n"
		captureFail(t, captureRun(t, steps, nil), 2, "path safety", "noop")
	}
	steps := "  - " + captureCommand("noop") + `    files: {held.yaml: "value: same.txt\n"}
    capture: {HELD: {yaml-scalar: {file: held.yaml, path: [value]}}}
  - ` + captureCommand("noop") + "    files: {'${HELD}': first, same.txt: second}\n"
	captureFail(t, captureRun(t, steps, nil), 2, "interpolation", "noop")
	steps = "  - " + captureCommand("noop") + `    files: {held.yaml: "value: same.txt\n"}
    capture: {HELD: {yaml-scalar: {file: held.yaml, path: [value]}}}
  - ` + captureCommand("noop") + "    expect: {files: {'${HELD}': {absent: true}, same.txt: {absent: true}}}\n"
	captureFail(t, captureRun(t, steps, nil), 2, "interpolation", "noop")
}

func TestCaptureRejectsActualLeafSymlinkEvenWithinOwnedSandbox(t *testing.T) {
	steps := "  - " + captureCommand("symlink", "alias.yaml", "real.yaml") + `    files: {real.yaml: "value: held\n"}
    expect: {exit: 0}
    capture: {HELD: {yaml-scalar: {file: alias.yaml, path: [value]}}}
`
	captureFail(t, captureRun(t, steps, nil), 1, "path safety", "symlink")
}

func TestActualSymlinkAncestorCannotExposeOrOverwriteOutsideFixture(t *testing.T) {
	outside, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(outside, "held.yaml")
	original := "value: OUTSIDE-PRIVATE-FIXTURE\n"
	if err := os.WriteFile(target, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"capture", "setup", "expectation", "unchanged"} {
		t.Run(action, func(t *testing.T) {
			steps := "  - " + captureCommand("symlink", "escape", outside) + "    expect: {exit: 0}\n"
			switch action {
			case "capture":
				steps += "    capture: {HELD: {yaml-scalar: {file: escape/held.yaml, path: [value]}}}\n"
			case "setup":
				steps += "  - " + captureCommand("noop") + "    files: {escape/held.yaml: CHANGED}\n    capture: {HELD: {yaml-scalar: {file: held.yaml, path: [value]}}}\n"
			case "expectation":
				steps += "  - " + captureCommand("noop") + "    expect: {files: {escape/held.yaml: {yaml-scalars: [{path: [value], equals: no}]}}}\n"
			case "unchanged":
				steps += "  - " + captureCommand("noop") + "    expect: {unchanged: [escape/held.yaml]}\n    capture: {HELD: {unique-directory-entry: missing}}\n"
			}
			step := 2
			if action == "capture" {
				step = 1
			}
			got := captureRun(t, steps, nil)
			captureFail(t, got, step, "path safety", "symlink")
			body, err := os.ReadFile(target)
			if err != nil || string(body) != original {
				t.Fatal("actual outside bytes changed")
			}
			after, err := os.Stat(target)
			if err != nil || !os.SameFile(before, after) {
				t.Fatal("actual outside native identity changed")
			}
			for _, failure := range got.failures {
				if failure.Message != "Scenario path is not a safe local path." {
					t.Fatal("outside capture was exposed in diagnostic")
				}
			}
		})
	}
}
