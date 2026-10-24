package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func workspaceRepository(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"tracked", "configured-temp"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module release-workspace-fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tracked", "marker"), []byte("Keep this independent repository marker.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func requireWorkspace(t *testing.T, repository, workspace string) {
	t.Helper()
	if !filepath.IsAbs(workspace) || filepath.Clean(workspace) != workspace {
		t.Fatalf("workspace is not an absolute canonical path: %q", workspace)
	}
	physical, err := filepath.EvalSymlinks(workspace)
	if err != nil {
		t.Fatalf("workspace physical lookup: %v", err)
	}
	if physical != workspace {
		t.Fatalf("workspace is not physical: %q resolves to %q", workspace, physical)
	}
	relative, err := filepath.Rel(repository, workspace)
	if err != nil {
		t.Fatalf("workspace relation: %v", err)
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		t.Fatalf("workspace remains inside the repository: %q", workspace)
	}
	info, err := os.Lstat(workspace)
	if err != nil {
		t.Fatalf("workspace inspection: %v", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o700 {
		t.Fatalf("workspace must be a nonsymlink0700 directory: %v", info.Mode())
	}
	entries, err := os.ReadDir(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("new workspace is not empty: %d entries", len(entries))
	}
	t.Cleanup(func() {
		if err := os.Remove(workspace); err != nil && !os.IsNotExist(err) {
			t.Errorf("remove empty owned workspace: %v", err)
		}
	})
}

func repositorySnapshot(t *testing.T, repository string) map[string]struct {
	mode fs.FileMode
	body string
} {
	t.Helper()
	result := make(map[string]struct {
		mode fs.FileMode
		body string
	})
	err := filepath.WalkDir(repository, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name, err := filepath.Rel(repository, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		body := ""
		if info.Mode().IsRegular() {
			bytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			body = string(bytes)
		}
		result[name] = struct {
			mode fs.FileMode
			body string
		}{mode: info.Mode(), body: body}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestVerificationWorkspaceOutsideRepositoryAndPhysicalPrivate(t *testing.T) {
	repository := workspaceRepository(t)
	t.Setenv("TMPDIR", filepath.Join(repository, "configured-temp"))
	t.Setenv("TMP", filepath.Join(repository, "configured-temp"))
	t.Setenv("TEMP", filepath.Join(repository, "configured-temp"))
	before := repositorySnapshot(t, repository)
	workspace, err := verificationWorkspace(repository)
	if err != nil {
		t.Fatalf("create verification workspace: %v", err)
	}
	requireWorkspace(t, repository, workspace)
	if !reflect.DeepEqual(repositorySnapshot(t, repository), before) {
		t.Fatal("workspace creation modified repository names, modes or file bytes")
	}
}

func TestVerificationWorkspacesAreDistinctAndLeaveRepositoryUntouched(t *testing.T) {
	repository := workspaceRepository(t)
	before := repositorySnapshot(t, repository)
	first, err := verificationWorkspace(repository)
	if err != nil {
		t.Fatal(err)
	}
	requireWorkspace(t, repository, first)
	marker, err := os.CreateTemp(first, "owned-marker-")
	if err != nil {
		t.Fatal(err)
	}
	markerPath := marker.Name()
	t.Cleanup(func() {
		if err := os.Remove(markerPath); err != nil && !os.IsNotExist(err) {
			t.Errorf("remove owned marker: %v", err)
		}
	})
	if _, err := marker.WriteString("First invocation retains its own file.\n"); err != nil {
		_ = marker.Close()
		t.Fatal(err)
	}
	if err := marker.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := verificationWorkspace(repository)
	if err != nil {
		t.Fatal(err)
	}
	requireWorkspace(t, repository, second)
	if first == second {
		t.Fatal("two invocations reused one workspace pathname")
	}
	firstInfo, err := os.Stat(first)
	if err != nil {
		t.Fatal(err)
	}
	secondInfo, err := os.Stat(second)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(firstInfo, secondInfo) {
		t.Fatal("two workspace names identify the same native directory")
	}
	body, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "First invocation retains its own file.\n" {
		t.Fatal("second invocation changed the first workspace file")
	}
	if !reflect.DeepEqual(repositorySnapshot(t, repository), before) {
		t.Fatal("workspace creation changed the repository fixture")
	}
}
