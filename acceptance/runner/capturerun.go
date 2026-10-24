package runner

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

func extensionFailure(category string) []Failure {
	messages := map[string]string{
		"interpolation": "Scenario interpolation is invalid.",
		"path safety":   "Scenario path is not a safe local path.",
		"capture":       "Capture could not be established.",
		"capture limit": "Capture budget exceeded.",
	}
	message, known := messages[category]
	if !known {
		message = "Scenario expectation failed."
	}
	return []Failure{{Expectation: category, Message: message}}
}

func runExtended(scenario Scenario, options Options) []Failure {
	if validateCaptures(scenario) != nil || validate(scenario) != nil {
		failures := extensionFailure("setup")
		failures[0].Step = 1
		return failures
	}
	folder, err := os.MkdirTemp(options.TemporaryRoot, "scenario-*")
	if err != nil {
		failures := extensionFailure("setup")
		failures[0].Step = 1
		return failures
	}
	defer os.RemoveAll(folder)
	folder, err = filepath.EvalSymlinks(folder)
	if err != nil {
		failures := extensionFailure("setup")
		failures[0].Step = 1
		return failures
	}
	root, err := openSandbox(folder)
	if err != nil {
		failures := extensionFailure("path safety")
		failures[0].Step = 1
		return failures
	}
	root.Close()
	values := map[string]string{"SANDBOX": folder}
	retained := 0
	for index, original := range scenario.Steps {
		step, err := expandStep(original, values)
		var failures []Failure
		if err != nil {
			failures = extensionFailure("interpolation")
		} else {
			failures = runExtendedStep(step, folder, options)
		}
		if len(failures) == 0 {
			var fresh map[string]string
			fresh, failures = captureStep(step, folder, retained)
			if len(failures) == 0 {
				for name, value := range fresh {
					values[name] = value
					retained += len(name) + len(value)
				}
			}
		}
		if len(failures) > 0 {
			for position := range failures {
				failures[position].Step = index + 1
			}
			return failures
		}
	}
	return nil
}

func stepLexicalPaths(step Step) error {
	for path := range step.Files {
		if _, err := localParts(path); err != nil {
			return err
		}
	}
	for path := range step.Expect.Files {
		if _, err := localParts(path); err != nil {
			return err
		}
	}
	for _, path := range step.Expect.Unchanged {
		if _, err := localParts(path); err != nil {
			return err
		}
	}
	for _, capture := range step.Capture {
		if capture.UniqueDirectoryEntry != nil {
			if _, err := localParts(*capture.UniqueDirectoryEntry); err != nil {
				return err
			}
		}
		if capture.YAMLScalar != nil {
			if _, err := localParts(capture.YAMLScalar.File); err != nil {
				return err
			}
		}
	}
	return nil
}

func proveStepPaths(step Step, folder string) error {
	for path := range step.Files {
		if err := proveNative(folder, path, true); err != nil {
			return err
		}
	}
	return proveExpectationPaths(step.Expect, folder)
}

func proveExpectationPaths(expect Expectation, folder string) error {
	for path := range expect.Files {
		if err := proveNative(folder, path, true); err != nil {
			return err
		}
	}
	for _, path := range expect.Unchanged {
		if err := proveNative(folder, path, true); err != nil {
			return err
		}
	}
	return nil
}

func runExtendedStep(step Step, folder string, options Options) []Failure {
	if stepLexicalPaths(step) != nil || proveStepPaths(step, folder) != nil {
		return extensionFailure("path safety")
	}
	if err := placeNativeFiles(step.Files, folder, options.RepositoryRoot); err != nil {
		if errors.Is(err, errUnsafePath) {
			return extensionFailure("path safety")
		}
		return extensionFailure("setup")
	}
	before, err := snapshot(step.Expect.Unchanged, folder)
	if err != nil {
		return extensionFailure("setup")
	}
	var result outcome
	if step.IsStdio() {
		result = runStdio(step, folder, options)
	} else {
		result = runCommand(step, folder, options)
	}
	if result.ProcessFailed != nil {
		return extensionFailure("start")
	}
	if result.TimedOut {
		return extensionFailure("time limit")
	}
	if proveExpectationPaths(step.Expect, folder) != nil {
		return extensionFailure("path safety")
	}
	failures := append([]Failure(nil), result.TalkFailures...)
	failures = append(failures, check(step.Expect, result, folder, options.RepositoryRoot, before)...)
	for index := range failures {
		label := failures[index].Expectation
		switch {
		case strings.HasPrefix(label, "file "):
			if strings.HasSuffix(label, " yaml-scalars") {
				label = "file yaml-scalars"
			} else {
				label = "file"
			}
		case strings.HasPrefix(label, "unchanged "):
			label = "unchanged"
		}
		failures[index].Expectation = label
		failures[index].Message = "Scenario expectation failed."
	}
	return failures
}

func captureStep(step Step, folder string, retained int) (map[string]string, []Failure) {
	names := make([]string, 0, len(step.Capture))
	for name := range step.Capture {
		names = append(names, name)
	}
	sort.Strings(names)
	fresh := make(map[string]string, len(names))
	for _, name := range names {
		capture := step.Capture[name]
		var value string
		var err error
		if capture.UniqueDirectoryEntry != nil {
			value, err = captureDirectory(folder, *capture.UniqueDirectoryEntry)
		} else if capture.YAMLScalar != nil {
			value, err = captureYAML(folder, *capture.YAMLScalar)
		} else {
			err = errors.New("capture kind")
		}
		if err != nil {
			if errors.Is(err, errUnsafePath) {
				return nil, extensionFailure("path safety")
			}
			return nil, extensionFailure("capture")
		}
		if !utf8.ValidString(value) || len(value) > 4096 || strings.IndexByte(value, 0) >= 0 {
			return nil, extensionFailure("capture limit")
		}
		retained += len(name) + len(value)
		if retained > 1048576 {
			return nil, extensionFailure("capture limit")
		}
		fresh[name] = value
	}
	return fresh, nil
}
