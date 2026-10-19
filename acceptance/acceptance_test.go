package acceptance

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"onboard-kit/acceptance/runner"
)

func TestScenarios(t *testing.T) {
	repositoryRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	sources, err := filepath.Glob(filepath.Join(repositoryRoot, "binary", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) == 0 {
		t.Skip("no Go files in ./binary yet, so there is no binary to run the scenarios against")
	}
	binaryPath := filepath.Join(t.TempDir(), "onboard-kit")
	build := exec.Command("go", "build", "-o", binaryPath, "./binary")
	build.Dir = repositoryRoot
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building the binary failed: %v\n%s", err, output)
	}
	scenarioFiles, err := filepath.Glob(filepath.Join(repositoryRoot, "acceptance", "scenarios", "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, scenarioFile := range scenarioFiles {
		name := strings.TrimSuffix(filepath.Base(scenarioFile), ".yaml")
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			scenario, err := runner.Load(scenarioFile)
			if err != nil {
				t.Fatal(err)
			}
			failures := runner.Run(scenario, runner.Options{BinaryPath: binaryPath, RepositoryRoot: repositoryRoot})
			for _, failure := range failures {
				t.Error(failure.String())
			}
		})
	}
}
