package store_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	contract "onboard-kit/contracts/preparation/store"
)

func TestDarwinRefusedDirectoryCreationUnderInheritableACLLeavesNothingGranted(t *testing.T) {
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" {
		t.Skip("native filesec fixture specifies little-endian Darwin arm64 and amd64")
	}
	box := newSandbox(t, "plan.yaml", "")
	value, err := exec.Command("/bin/chmod", "+a", "everyone allow read,file_inherit,directory_inherit", box.Root).CombinedOutput()
	if err != nil {
		t.Fatalf("cannot add inheritable ACL to owned root: output bytes %d", len(value))
	}
	loaded := box.load(t)
	held := snapshot(t, box.Root)
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml"), CreateDirectories: true})
	if changed := !sameMap(snapshot(t, box.Root), held); result.CreatedDirectories != changed {
		t.Fatalf("code %q reports CreatedDirectories %v but the disk changed %v", result.Code, result.CreatedDirectories, changed)
	}
	if result.Committed != (result.Code == contract.CodeSaved) {
		t.Fatalf("code %q reports Committed %v", result.Code, result.Committed)
	}
	err = filepath.WalkDir(box.Root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || path == box.Root || path == box.Plan {
			return walkErr
		}
		if hasNativeReadGrant(t, readNativeFilesec(t, path), 0) {
			t.Errorf("code %q left %s holding an inherited read grant", result.Code, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRefusedDirectoryCreationInAnUnwritableRootReportsNothingCreated(t *testing.T) {
	for _, mode := range []os.FileMode{0o500, 0o550} {
		box := newSandbox(t, "plan.yaml", "")
		loaded := box.load(t)
		held := snapshot(t, box.Root)
		if err := os.Chmod(box.Root, mode); err != nil {
			t.Fatal(err)
		}
		result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml"), CreateDirectories: true})
		if err := os.Chmod(box.Root, 0o700); err != nil {
			t.Fatal(err)
		}
		wantCode(t, result.Code, contract.CodeCommit)
		if result.Committed || result.CreatedDirectories {
			t.Fatalf("mode %o: refusal reports effects %#v", mode, result)
		}
		if after := snapshot(t, box.Root); !sameMap(after, held) {
			t.Fatalf("mode %o: refused save changed the filesystem", mode)
		}
	}
}

func TestRefusedDirectoryCreationBelowAnUnwritableRecordsDirectoryReportsNothingCreated(t *testing.T) {
	box := newSandbox(t, "plan.yaml", "")
	records := filepath.Join(box.Root, "records")
	if err := os.MkdirAll(records, 0o700); err != nil {
		t.Fatal(err)
	}
	loaded := box.load(t)
	held := snapshot(t, box.Root)
	if err := os.Chmod(records, 0o500); err != nil {
		t.Fatal(err)
	}
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml"), CreateDirectories: true})
	if err := os.Chmod(records, 0o700); err != nil {
		t.Fatal(err)
	}
	wantCode(t, result.Code, contract.CodeCommit)
	if result.Committed || result.CreatedDirectories {
		t.Fatalf("refusal reports effects %#v", result)
	}
	if after := snapshot(t, box.Root); !sameMap(after, held) {
		t.Fatal("refused save changed the filesystem")
	}
}
