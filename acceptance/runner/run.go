package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const defaultTimeLimit = 30 * time.Second

type Options struct {
	BinaryPath     string
	RepositoryRoot string
	TimeLimit      time.Duration
	TemporaryRoot  string
}

func (options Options) timeLimit() time.Duration {
	if options.TimeLimit <= 0 {
		return defaultTimeLimit
	}
	return options.TimeLimit
}

func Run(scenario Scenario, options Options) []Failure {
	folder, err := os.MkdirTemp(options.TemporaryRoot, "scenario-*")
	if err != nil {
		return []Failure{{Expectation: "setup", Message: err.Error()}}
	}
	defer os.RemoveAll(folder)
	physical, err := filepath.EvalSymlinks(folder)
	if err != nil {
		return []Failure{{Expectation: "setup", Message: err.Error()}}
	}
	folder = physical
	for index, step := range scenario.Steps {
		failures := runStep(step, folder, options)
		if len(failures) > 0 {
			for position := range failures {
				failures[position].Step = index + 1
			}
			return failures
		}
	}
	return nil
}

func runStep(step Step, folder string, options Options) []Failure {
	if err := placeFiles(step.Files, folder, options.RepositoryRoot); err != nil {
		return []Failure{{Expectation: "setup", Message: err.Error()}}
	}
	before, err := snapshot(step.Expect.Unchanged, folder)
	if err != nil {
		return []Failure{{Expectation: "setup", Message: err.Error()}}
	}
	var result outcome
	if step.IsStdio() {
		result = runStdio(step, folder, options)
	} else {
		result = runCommand(step, folder, options)
	}
	if result.ProcessFailed != nil {
		return []Failure{{Expectation: "start", Message: result.ProcessFailed.Error()}}
	}
	if result.TimedOut {
		return []Failure{{Expectation: "time limit", Message: fmt.Sprintf("the run did not finish within %s and was killed\nstdout: %q\nstderr: %q", options.timeLimit(), result.Stdout, result.Stderr)}}
	}
	failures := append([]Failure{}, result.TalkFailures...)
	return append(failures, check(step.Expect, result, folder, options.RepositoryRoot, before)...)
}

func resolve(root, relative string) (string, error) {
	if !filepath.IsLocal(relative) {
		return "", fmt.Errorf("path %q must stay inside %s", relative, root)
	}
	return filepath.Join(root, relative), nil
}

func placeFiles(files map[string]FileSource, folder, repositoryRoot string) error {
	for path, source := range files {
		target, err := resolve(folder, path)
		if err != nil {
			return err
		}
		content := []byte(source.Text)
		if source.From != "" {
			origin, err := resolve(repositoryRoot, source.From)
			if err != nil {
				return err
			}
			if content, err = os.ReadFile(origin); err != nil {
				return fmt.Errorf("file %s: %w", path, err)
			}
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

type fileState struct {
	Exists  bool
	Content []byte
}

func snapshot(paths []string, folder string) (map[string]fileState, error) {
	states := make(map[string]fileState, len(paths))
	for _, path := range paths {
		state, err := readState(path, folder)
		if err != nil {
			return nil, err
		}
		states[path] = state
	}
	return states, nil
}

func readState(path, folder string) (fileState, error) {
	target, err := resolve(folder, path)
	if err != nil {
		return fileState{}, err
	}
	content, err := os.ReadFile(target)
	if os.IsNotExist(err) {
		return fileState{}, nil
	}
	if err != nil {
		return fileState{}, err
	}
	return fileState{Exists: true, Content: content}, nil
}
