package acceptance

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"onboard-kit/acceptance/runner"
)

var existingBinary = flag.String("binary", "", "run scenarios against this existing binary without building it")

func TestScenarios(t *testing.T) {
	repositoryRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	binaryPath := scenarioBinary(t, repositoryRoot)
	scenarioFiles, err := filepath.Glob(filepath.Join(repositoryRoot, "acceptance", "scenarios", "*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(scenarioFiles) == 0 {
		t.Fatal("no acceptance scenarios found")
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

func scenarioBinary(t *testing.T, repositoryRoot string) string {
	t.Helper()
	if *existingBinary != "" {
		path, err := filepath.Abs(*existingBinary)
		if err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("existing binary: %v", err)
		}
		if !info.Mode().IsRegular() {
			t.Fatal("existing binary must be a regular file")
		}
		return path
	}
	sources, err := filepath.Glob(filepath.Join(repositoryRoot, "binary", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) == 0 {
		t.Skip("no Go files in ./binary yet, so there is no binary to run the scenarios against")
	}
	path := filepath.Join(t.TempDir(), "onboard-kit")
	build := exec.Command("go", "build", "-o", path, "./binary")
	build.Dir = repositoryRoot
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building the binary failed: %v\n%s", err, output)
	}
	return path
}
