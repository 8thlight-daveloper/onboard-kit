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

func TestScenarioWorkingDirectoryUsesPhysicalTemporaryRoot(t *testing.T) {
	t.Parallel()
	for _, failLastStep := range []bool{false, true} {
		name := "successful scenario"
		if failLastStep {
			name = "failed scenario"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			base, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			physical := filepath.Join(base, "owned")
			if err := os.Mkdir(physical, 0o700); err != nil {
				t.Fatal(err)
			}
			alias := filepath.Join(base, "logical-only-alias")
			if err := os.Symlink(physical, alias); err != nil {
				t.Fatal(err)
			}
			sentinel := filepath.Join(physical, "parent-owned.txt")
			if err := os.WriteFile(sentinel, []byte("PARENT OWNED\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			parentBefore, err := os.Stat(physical)
			if err != nil {
				t.Fatal(err)
			}
			aliasBefore, err := os.Lstat(alias)
			if err != nil {
				t.Fatal(err)
			}
			throughAlias, err := os.Stat(alias)
			if err != nil || !os.SameFile(parentBefore, throughAlias) {
				t.Fatalf("the owned alias must reach the physical parent: %v", err)
			}
			workingDirectory, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			ambientTemp, ambientTempSet := os.LookupEnv("TMPDIR")
			ambientPWD, ambientPWDSet := os.LookupEnv("PWD")

			control := exec.Command(fakeBinaryPath, "folder")
			control.Dir = alias
			control.Env = append(os.Environ(), "PWD="+alias)
			controlOutput, err := control.Output()
			if err != nil {
				t.Fatalf("the independently owned logical-PWD control must run: %v", err)
			}
			if string(controlOutput) != alias {
				t.Fatalf("the folder helper must expose a valid logical PWD, got %q", controlOutput)
			}

			step := fmt.Sprintf("  - run: [onboard-kit, folder]\n    expect:\n      exit: 0\n      stdout: {contains: %q}\n", physical+string(filepath.Separator)+"scenario-")
			step += "  - run: [onboard-kit, write, owned-child.txt, child]\n    expect: {exit: 0}\n"
			if failLastStep {
				step += "  - run: [onboard-kit, say, '', '', '1']\n    expect: {exit: 0}\n"
			}
			failures := runScenarioWith(t, scenarioWithStep(step), runner.Options{TimeLimit: 20 * time.Second, TemporaryRoot: alias}, nil)
			if failLastStep {
				requireFailure(t, failures, 3, "exit")
				if len(failures) != 1 {
					t.Errorf("only the authored final failure is expected: %v", failures)
				}
			} else {
				requirePass(t, failures)
			}

			parentAfter, err := os.Stat(physical)
			if err != nil {
				t.Fatal(err)
			}
			if !os.SameFile(parentBefore, parentAfter) || parentBefore.Mode() != parentAfter.Mode() {
				t.Error("the requested physical parent must retain its identity and mode")
			}
			aliasAfter, err := os.Lstat(alias)
			if err != nil {
				t.Fatal(err)
			}
			if aliasAfter.Mode()&os.ModeSymlink == 0 || !os.SameFile(aliasBefore, aliasAfter) {
				t.Error("the owned alias must remain the same symlink")
			}
			target, err := os.Readlink(alias)
			if err != nil || target != physical {
				t.Fatalf("the owned alias target must remain unchanged: %v", err)
			}
			entries, err := os.ReadDir(physical)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 || entries[0].Name() != "parent-owned.txt" {
				t.Errorf("the runner must remove only its child sandbox, got %v", entries)
			}
			content, err := os.ReadFile(sentinel)
			if err != nil || string(content) != "PARENT OWNED\n" {
				t.Fatalf("the parent's existing file must remain unchanged: %v", err)
			}
			currentDirectory, err := os.Getwd()
			if err != nil || currentDirectory != workingDirectory {
				t.Fatalf("the runner must not change the parent process working directory: %v", err)
			}
			currentTemp, currentTempSet := os.LookupEnv("TMPDIR")
			currentPWD, currentPWDSet := os.LookupEnv("PWD")
			if currentTemp != ambientTemp || currentTempSet != ambientTempSet || currentPWD != ambientPWD || currentPWDSet != ambientPWDSet {
				t.Error("the runner must preserve ambient TMPDIR and PWD")
			}
			if strings.Contains(physical, "logical-only-alias") {
				t.Fatal("the fixture must distinguish its physical and logical spellings")
			}
		})
	}
}
