package store_test

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	contract "onboard-kit/contracts/preparation/store"
)

const (
	sentinel         = "FOREIGN FILE the store did not create\n"
	occupiedNames    = 5
	discoveryTimeout = 30 * time.Second
)

func occupy(t *testing.T, directory string, names []string) {
	t.Helper()
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(sentinel), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func discoverNextTemporaryName(t *testing.T, before string, occupied []string) string {
	t.Helper()
	for attempt := 0; attempt < attempts; attempt++ {
		box := newSandbox(t, "plan.yaml", before)
		directory := filepath.Dir(box.Prep)
		occupy(t, directory, occupied)
		loaded := box.load(t)
		stop := make(chan struct{})
		done := make(chan struct{})
		var once sync.Once
		name := ""
		go func() {
			defer close(done)
			for {
				select {
				case <-stop:
					return
				default:
				}
				entries, err := os.ReadDir(directory)
				if err != nil {
					continue
				}
				for _, entry := range entries {
					if entry.Name() != filepath.Base(box.Prep) && !slices.Contains(occupied, entry.Name()) {
						once.Do(func() { name = entry.Name() })
						return
					}
				}
			}
		}()
		var result contract.SaveResult
		within(t, discoveryTimeout, "save while discovering the next temporary name", func() {
			result = newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: padded(t, fixture(t, "named.yaml"), byteLimit), CreateDirectories: before == ""})
		})
		close(stop)
		<-done
		wantCode(t, result.Code, contract.CodeSaved)
		if name != "" {
			return name
		}
	}
	t.Fatalf("no attempt observed a new entry beside the destination in %d saves with %d names occupied", attempts, len(occupied))
	return ""
}

func TestSaveNeverTouchesAFileItDidNotCreateAndStillSaves(t *testing.T) {
	for _, c := range []struct{ name, before string }{
		{name: "replace", before: "identity.yaml"},
		{name: "create-new", before: ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			var occupied []string
			for len(occupied) < occupiedNames {
				occupied = append(occupied, discoverNextTemporaryName(t, c.before, occupied))
				if slices.Contains(occupied[:len(occupied)-1], occupied[len(occupied)-1]) {
					t.Fatalf("name %q was offered twice", occupied[len(occupied)-1])
				}
			}
			box := newSandbox(t, "plan.yaml", c.before)
			directory := filepath.Dir(box.Prep)
			occupy(t, directory, occupied)
			before := map[string]os.FileInfo{}
			for _, name := range occupied {
				info, err := os.Stat(filepath.Join(directory, name))
				if err != nil {
					t.Fatal(err)
				}
				before[name] = info
			}
			loaded := box.load(t)
			reviewed := fixture(t, "named.yaml")
			var result contract.SaveResult
			within(t, 10*time.Second, "save beside files it did not create", func() {
				result = newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: reviewed})
			})
			for _, name := range occupied {
				got, err := os.ReadFile(filepath.Join(directory, name))
				if err != nil || string(got) != sentinel {
					t.Fatalf("file %q the call did not create is now %q, error %v; code %q", name, got, err, result.Code)
				}
				if info, err := os.Stat(filepath.Join(directory, name)); err != nil || !os.SameFile(before[name], info) {
					t.Fatalf("file %q the call did not create was replaced: %v", name, err)
				}
			}
			wantCode(t, result.Code, contract.CodeSaved)
			if !result.Committed {
				t.Fatal("saved but not committed")
			}
			if got := readBytes(t, box.Prep); !bytes.Equal(got, reviewed) {
				t.Fatalf("saved %q; want %q", got, reviewed)
			}
			extras := extraEntries(t, directory, filepath.Base(box.Prep))
			slices.Sort(extras)
			want := slices.Clone(occupied)
			slices.Sort(want)
			if !slices.Equal(extras, want) {
				t.Fatalf("entries beside the destination %v; want only the files the call did not create, %v", extras, want)
			}
		})
	}
}
