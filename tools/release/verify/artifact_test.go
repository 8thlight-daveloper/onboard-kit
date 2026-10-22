package main

import (
	"debug/macho"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLibrariesMustBeInTheExplicitBaseline(t *testing.T) {
	t.Parallel()
	if err := checkLibraries([]string{"/usr/lib/libSystem.B.dylib", "/usr/lib/libresolv.9.dylib"}); err != nil {
		t.Fatal(err)
	}
	for _, library := range []string{
		"/opt/homebrew/lib/libextra.dylib",
		"/nix/store/example/lib/libextra.dylib",
		"@rpath/libextra.dylib",
		"@executable_path/libextra.dylib",
		"/usr/lib/../../opt/libextra.dylib",
		"/usr/lib/unreviewed.dylib",
		"",
	} {
		if err := checkLibraries([]string{"/usr/lib/libSystem.B.dylib", library}); err == nil || !strings.Contains(err.Error(), "baseline") {
			t.Errorf("library %q must be refused, got %v", library, err)
		}
	}
}

func TestArtifactMustBeAnExecutableArm64MachO(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name    string
		cpu     macho.Cpu
		kind    macho.Type
		mode    os.FileMode
		allowed bool
	}{
		{"arm64-executable", macho.CpuArm64, macho.TypeExec, 0o755, true},
		{"intel-executable", macho.CpuAmd64, macho.TypeExec, 0o755, false},
		{"arm64-library", macho.CpuArm64, macho.TypeDylib, 0o755, false},
		{"not-executable", macho.CpuArm64, macho.TypeExec, 0o644, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "artifact")
			header := make([]byte, 32)
			binary.LittleEndian.PutUint32(header, macho.Magic64)
			binary.LittleEndian.PutUint32(header[4:], uint32(c.cpu))
			binary.LittleEndian.PutUint32(header[12:], uint32(c.kind))
			if err := os.WriteFile(path, header, c.mode); err != nil {
				t.Fatal(err)
			}
			_, err := inspectArtifact(path)
			if (err == nil) != c.allowed {
				t.Fatalf("allowed %t, got %v", c.allowed, err)
			}
		})
	}
}

func TestMissingMalformedAndDirectoryArtifactsFail(t *testing.T) {
	t.Parallel()
	folder := t.TempDir()
	malformed := filepath.Join(folder, "text")
	if err := os.WriteFile(malformed, []byte("not a release binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{folder, filepath.Join(folder, "missing"), malformed} {
		if _, err := inspectArtifact(path); err == nil {
			t.Errorf("artifact %s must be refused", path)
		}
	}
}

func TestHashPinsTheArtifactBytes(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "artifact")
	if err := os.WriteFile(path, []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := fileHash(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("unexpected artifact checksum %s", got)
	}
	if _, err := fileHash(path + "-missing"); err == nil {
		t.Fatal("missing artifact returned a checksum")
	}
	if _, err := fileHash(filepath.Dir(path)); err == nil {
		t.Fatal("directory returned a checksum")
	}
}

func TestAcceptanceEnvironmentHasNoInheritedToolsOrCredentials(t *testing.T) {
	t.Parallel()
	scratch := t.TempDir()
	environment, err := isolatedEnvironment(scratch)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"PATH":     filepath.Join(scratch, "bin"),
		"TMPDIR":   filepath.Join(scratch, "temporary"),
		"OBK_HOME": filepath.Join(scratch, "state"),
		"TERM":     "dumb",
	}
	for _, entry := range environment {
		key, value, _ := strings.Cut(entry, "=")
		if expected, present := want[key]; !present || value != expected {
			t.Fatalf("unexpected environment entry %q", entry)
		}
		delete(want, key)
	}
	if len(want) != 0 {
		t.Fatalf("missing isolated environment entries: %v", want)
	}
	entries, err := os.ReadDir(filepath.Join(scratch, "bin"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("acceptance PATH must be empty: %v, %v", entries, err)
	}
}
