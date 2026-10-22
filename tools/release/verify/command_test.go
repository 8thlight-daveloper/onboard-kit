package main

import (
	"bytes"
	"context"
	"flag"
	"os"
	"strings"
	"testing"
	"time"
)

var environmentProbe = flag.Bool("release-environment-probe", false, "check the environment in a test subprocess")

func TestCommandEnvironmentProbe(t *testing.T) {
	if !*environmentProbe {
		return
	}
	if os.Getenv("RELEASE_CHECK_PROBE") != "1" || len(os.Environ()) != 2 {
		t.Fatal("the subprocess did not receive exactly its isolated environment")
	}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if key != "PATH" && key != "RELEASE_CHECK_PROBE" {
			t.Fatal("the subprocess inherited an unexpected environment variable")
		}
	}
}

func TestCommandPassesOnlyTheProvidedEnvironment(t *testing.T) {
	t.Setenv("RELEASE_CHECK_PARENT_SENTINEL", "must-not-reach-the-child")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	folder := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var output bytes.Buffer
	err = command(ctx, folder, []string{"PATH=" + folder, "RELEASE_CHECK_PROBE=1"}, &output, &output,
		executable, "-test.run=^TestCommandEnvironmentProbe$", "-release-environment-probe")
	if err != nil {
		t.Fatalf("isolated subprocess: %v\n%s", err, output.String())
	}
}
