package store_test

import (
	"bytes"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"unsafe"

	contract "onboard-kit/contracts/preparation/store"
)

func setImmutable(path string, on bool) {
	pointer, err := syscall.BytePtrFromString(path)
	if err != nil {
		return
	}
	flags := uintptr(0)
	if on {
		flags = 2
	}
	_, _, _ = syscall.Syscall(syscall.SYS_CHFLAGS, uintptr(unsafe.Pointer(pointer)), flags, 0)
}

func releaseAll(directory string) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return
	}
	for _, entry := range entries {
		setImmutable(filepath.Join(directory, entry.Name()), false)
	}
}

func TestCleanupReportMatchesWhatIsLeftBehind(t *testing.T) {
	for _, c := range []struct {
		name, before, reviewed string
	}{
		{name: "replace-existing", before: "identity.yaml", reviewed: "named.yaml"},
		{name: "create-new", before: "", reviewed: "named.yaml"},
	} {
		t.Run(c.name, func(t *testing.T) {
			outcomes := map[contract.Code]int{}
			cleanupSeen := 0
			for attempt := 0; attempt < attempts; attempt++ {
				box := newSandbox(t, "plan.yaml", c.before)
				directory := filepath.Dir(box.Prep)
				if err := os.MkdirAll(directory, 0o700); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { releaseAll(directory) })
				loaded := box.load(t)
				reviewed := padded(t, fixture(t, c.reviewed), byteLimit)
				watching := watchDirectory(directory, filepath.Base(box.Prep), func(path string) { setImmutable(path, true) })
				result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: reviewed})
				watching.finish()
				releaseAll(directory)
				outcomes[result.Code]++
				extras := extraEntries(t, directory, filepath.Base(box.Prep))
				destinationBytes, readErr := os.ReadFile(box.Prep)
				switch result.Code {
				case contract.CodeCleanupRefused:
					cleanupSeen++
					if len(extras) == 0 || result.Committed {
						t.Fatalf("cleanup-refused with extras %v committed %v", extras, result.Committed)
					}
					if c.before != "" && !bytes.Equal(destinationBytes, fixture(t, c.before)) {
						t.Fatal("refused save changed the existing preparation")
					}
					if c.before == "" && readErr == nil {
						t.Fatal("refused save created the preparation")
					}
				case contract.CodeCleanupSaved:
					cleanupSeen++
					if len(extras) == 0 || !result.Committed || !bytes.Equal(destinationBytes, reviewed) {
						t.Fatalf("cleanup-saved with extras %v committed %v", extras, result.Committed)
					}
				case contract.CodeSaved:
					if len(extras) != 0 || !result.Committed || !bytes.Equal(destinationBytes, reviewed) {
						t.Fatalf("saved while %v remain, committed %v", extras, result.Committed)
					}
				default:
					if len(extras) != 0 {
						t.Fatalf("code %q left %v", result.Code, extras)
					}
					if result.Committed {
						t.Fatalf("code %q reports a commit", result.Code)
					}
					if c.before != "" && !bytes.Equal(destinationBytes, fixture(t, c.before)) {
						t.Fatalf("code %q changed the existing preparation", result.Code)
					}
				}
			}
			if cleanupSeen == 0 {
				t.Fatalf("no attempt reached a cleanup failure: %v", outcomes)
			}
			t.Logf("outcomes %v", outcomes)
		})
	}
}

func TestSaveOverAFileThatCannotBeReplacedIsRefusedWithoutACommit(t *testing.T) {
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	loaded := box.load(t)
	setImmutable(box.Prep, true)
	t.Cleanup(func() { setImmutable(box.Prep, false) })
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "named.yaml")})
	setImmutable(box.Prep, false)
	if result.Committed || result.Code != contract.CodeCommit {
		t.Fatalf("code %q committed %v for a file that cannot be replaced; want %q", result.Code, result.Committed, contract.CodeCommit)
	}
	if got := readBytes(t, box.Prep); !bytes.Equal(got, fixture(t, "identity.yaml")) {
		t.Fatal("the immutable preparation changed")
	}
	if extras := extraEntries(t, filepath.Dir(box.Prep), filepath.Base(box.Prep)); len(extras) != 0 {
		t.Fatalf("code %q left %v", result.Code, extras)
	}
}
