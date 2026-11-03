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
	"strings"
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
	target := runtime.GOOS + "/" + runtime.GOARCH
	if _, supported := linuxMachines[runtime.GOARCH]; target != "darwin/arm64" && !(runtime.GOOS == "linux" && supported) {
		return fmt.Errorf("run on darwin/arm64, linux/amd64 or linux/arm64, not %s", target)
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
	buildEnvironment := append(os.Environ(), "CGO_ENABLED=0", "GOOS="+runtime.GOOS, "GOARCH="+runtime.GOARCH, "GOMAXPROCS=2")
	if *binary == "" {
		*binary = filepath.Join(root, "output", "release", "onboard-kit")
		if err := os.MkdirAll(filepath.Dir(*binary), 0o755); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Building %s artifact with CGO disabled.\n", target)
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
	if runtime.GOOS == "darwin" {
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
	} else if err := inspectStaticArtifact(artifact, linuxMachines[runtime.GOARCH]); err != nil {
		return err
	}
	scratch, err := verificationWorkspace(root)
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

func verificationWorkspace(root string) (string, error) {
	folder, err := os.MkdirTemp("/tmp", "onboard-kit-release-")
	if err != nil {
		return "", err
	}
	success := false
	defer func() {
		if !success {
			_ = os.Remove(folder)
		}
	}()
	workspace, err := os.Open(folder)
	if err != nil {
		return "", err
	}
	if err := workspace.Chown(-1, os.Getegid()); err != nil {
		_ = workspace.Close()
		return "", err
	}
	if err := workspace.Close(); err != nil {
		return "", err
	}
	physical, err := filepath.EvalSymlinks(folder)
	if err != nil {
		return "", err
	}
	repository, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(repository, physical)
	if err != nil {
		return "", err
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("verification workspace must be outside the repository")
	}
	info, err := os.Lstat(physical)
	if err != nil {
		return "", err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o700 {
		return "", fmt.Errorf("verification workspace must be a private native directory")
	}
	success = true
	return physical, nil
}

func isolatedEnvironment(scratch string) ([]string, error) {
	for _, folder := range []string{"bin", "temporary", "state", "home", "cache"} {
		if err := os.Mkdir(filepath.Join(scratch, folder), 0o700); err != nil {
			return nil, err
		}
	}
	return []string{
		"PATH=" + filepath.Join(scratch, "bin"),
		"TMPDIR=" + filepath.Join(scratch, "temporary"),
		"OBK_HOME=" + filepath.Join(scratch, "state"),
		"HOME=" + filepath.Join(scratch, "home"),
		"XDG_CACHE_HOME=" + filepath.Join(scratch, "cache"),
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
