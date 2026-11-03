package root

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionFlagPrintsTheBuildVersionAndExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"--version"}, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d, stderr %q", code, stderr.String())
	}
	if want := "onboard-kit " + version + "\n"; stdout.String() != want {
		t.Fatalf("got %q, want %q", stdout.String(), want)
	}
}

func TestVersionFlagDefaultsToDevelopmentBuild(t *testing.T) {
	if version != "development" {
		t.Fatalf("version is %q, want development until a build sets it", version)
	}
}
