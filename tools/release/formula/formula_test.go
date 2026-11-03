package main

import (
	"strings"
	"testing"
)

const checksums = "aaa  onboard-kit_v0.1.0_darwin_arm64.tar.gz\nbbb  onboard-kit_v0.1.0_linux_arm64.tar.gz\nccc  onboard-kit_v0.1.0_linux_amd64.tar.gz\n"

func TestFormulaNamesEachArchiveWithItsChecksum(t *testing.T) {
	got, err := formula("v0.1.0", "example-org/onboard-kit", checksums)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`url "https://github.com/example-org/onboard-kit/releases/download/v0.1.0/onboard-kit_v0.1.0_darwin_arm64.tar.gz"`,
		`sha256 "aaa"`,
		`url "https://github.com/example-org/onboard-kit/releases/download/v0.1.0/onboard-kit_v0.1.0_linux_arm64.tar.gz"`,
		`sha256 "bbb"`,
		`url "https://github.com/example-org/onboard-kit/releases/download/v0.1.0/onboard-kit_v0.1.0_linux_amd64.tar.gz"`,
		`sha256 "ccc"`,
		`homepage "https://github.com/example-org/onboard-kit"`,
		`license "MIT"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("formula should hold %q, got:\n%s", want, got)
		}
	}
}

func TestFormulaVersionDropsTheTagPrefix(t *testing.T) {
	got, err := formula("v0.1.0", "example-org/onboard-kit", checksums)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `version "0.1.0"`) {
		t.Errorf("formula should hold version 0.1.0, got:\n%s", got)
	}
}

func TestFormulaRefusesAMissingChecksum(t *testing.T) {
	short := strings.Split(checksums, "\n")[0] + "\n"
	if _, err := formula("v0.1.0", "example-org/onboard-kit", short); err == nil {
		t.Fatal("expected an error")
	}
}

func TestFormulaIgnoresChecksumsOfAnotherVersion(t *testing.T) {
	if _, err := formula("v0.2.0", "example-org/onboard-kit", checksums); err == nil {
		t.Fatal("expected an error")
	}
}
