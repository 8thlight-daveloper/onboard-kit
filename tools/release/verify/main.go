package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "release check:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	options := flag.NewFlagSet("release-check", flag.ContinueOnError)
	options.SetOutput(stderr)
	binary := options.String("binary", "", "verify an existing artifact instead of building one")
	if err := options.Parse(args); err != nil {
		return err
	}
	if options.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", options.Args())
	}
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		return fmt.Errorf("run on macOS arm64 (darwin/arm64), not %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		return fmt.Errorf("run from the repository root: %w", err)
	}
	scenarios, err := filepath.Glob(filepath.Join(root, "acceptance", "scenarios", "*.yaml"))
	if err != nil {
		return err
	}
	if len(scenarios) == 0 {
		return fmt.Errorf("no acceptance scenarios found")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	buildEnvironment := append(os.Environ(), "CGO_ENABLED=0", "GOOS=darwin", "GOARCH=arm64", "GOMAXPROCS=2")
	if *binary == "" {
		*binary = filepath.Join(root, "output", "release", "onboard-kit")
		if err := os.MkdirAll(filepath.Dir(*binary), 0o755); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "Building darwin/arm64 artifact with CGO disabled.")
		if err := command(ctx, root, buildEnvironment, stdout, stderr, "go", "build", "-p=2", "-trimpath", "-o", *binary, "./binary"); err != nil {
			return err
		}
	}
	artifact, err := filepath.Abs(*binary)
	if err != nil {
		return err
	}
	before, err := fileHash(artifact)
	if err != nil {
		return err
	}
	libraries, err := inspectArtifact(artifact)
	if err != nil {
		return err
	}
	for _, library := range libraries {
		fmt.Fprintln(stdout, "System library:", library)
	}
	if err := command(ctx, root, os.Environ(), stdout, stderr, "/usr/bin/codesign", "--verify", "--strict", artifact); err != nil {
		return err
	}
	scratchRoot := filepath.Join(root, ".scratch")
	if err := os.MkdirAll(scratchRoot, 0o755); err != nil {
		return err
	}
	scratch, err := os.MkdirTemp(scratchRoot, "release-check-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)
	harness := filepath.Join(scratch, "acceptance.test")
	if err := command(ctx, root, buildEnvironment, stdout, stderr, "go", "test", "-p=2", "-c", "-o", harness, "./acceptance"); err != nil {
		return err
	}
	environment, err := isolatedEnvironment(scratch)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Testing %d scenarios against %s with an empty PATH directory and no inherited environment.\n", len(scenarios), artifact)
	if err := command(ctx, filepath.Join(root, "acceptance"), environment, stdout, stderr, harness,
		"-test.run=^TestScenarios$", "-test.count=1", "-test.parallel=2", "-test.timeout=5m", "-test.v", "-binary="+artifact); err != nil {
		return err
	}
	after, err := fileHash(artifact)
	if err != nil {
		return err
	}
	if before != after {
		return fmt.Errorf("artifact changed during acceptance: before %s, after %s", before, after)
	}
	fmt.Fprintf(stdout, "Verified artifact: %s\nSHA-256: %s\n", artifact, after)
	return nil
}

func isolatedEnvironment(scratch string) ([]string, error) {
	for _, folder := range []string{"bin", "temporary", "state"} {
		if err := os.Mkdir(filepath.Join(scratch, folder), 0o700); err != nil {
			return nil, err
		}
	}
	return []string{
		"PATH=" + filepath.Join(scratch, "bin"),
		"TMPDIR=" + filepath.Join(scratch, "temporary"),
		"OBK_HOME=" + filepath.Join(scratch, "state"),
		"TERM=dumb",
	}, nil
}

func command(ctx context.Context, directory string, environment []string, stdout, stderr io.Writer, name string, args ...string) error {
	process := exec.CommandContext(ctx, name, args...)
	process.Dir = directory
	process.Env = environment
	process.Stdout = stdout
	process.Stderr = stderr
	process.WaitDelay = time.Second
	if err := process.Run(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}
