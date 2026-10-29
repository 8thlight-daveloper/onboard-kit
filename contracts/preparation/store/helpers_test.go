package store_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	contract "onboard-kit/contracts/preparation/store"
)

const (
	account         = "newcomer"
	destination     = "records/newcomer/preparation.obk.yaml"
	byteLimit       = 1048576
	externalMessage = "EXTERNAL concurrent bytes\n"
	attempts        = 20
)

type sandbox struct {
	Base string
	Root string
	Plan string
	Prep string
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	value, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func writeFile(t *testing.T, path string, value []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, value, 0o600); err != nil {
		t.Fatal(err)
	}
}

func padded(t *testing.T, base []byte, size int) []byte {
	t.Helper()
	fill := size - len(base) - 2
	if fill < 0 {
		t.Fatalf("base of %d bytes does not fit %d", len(base), size)
	}
	result := append([]byte{}, base...)
	result = append(result, '#')
	result = append(result, []byte(strings.Repeat("x", fill))...)
	result = append(result, '\n')
	if len(result) != size {
		t.Fatalf("padding produced %d bytes; want %d", len(result), size)
	}
	return result
}

func newSandbox(t *testing.T, plan, before string) sandbox {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(base, "root")
	result := sandbox{Base: base, Root: root, Plan: filepath.Join(root, "organization.yaml"), Prep: filepath.Join(root, filepath.FromSlash(destination))}
	writeFile(t, result.Plan, fixture(t, plan))
	if before != "" {
		writeFile(t, result.Prep, fixture(t, before))
	}
	return result
}

func (s sandbox) load(t *testing.T) contract.Loaded {
	t.Helper()
	result := newStore().Load(contract.LoadRequest{Root: s.Root, Plan: s.Plan, Account: account})
	if result.Code != contract.CodeLoaded || result.Loaded == nil {
		t.Fatalf("load code %q; want %q", result.Code, contract.CodeLoaded)
	}
	return result.Loaded
}

func (s sandbox) request() contract.LoadRequest {
	return contract.LoadRequest{Root: s.Root, Plan: s.Plan, Account: account}
}

func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		switch {
		case entry.IsDir():
			result[relative+"/"] = "directory"
		case entry.Type()&os.ModeSymlink != 0:
			target, err := os.Readlink(path)
			result[relative] = "symlink:" + target
			return err
		case entry.Type()&os.ModeNamedPipe != 0:
			result[relative] = "fifo"
		default:
			value, err := os.ReadFile(path)
			result[relative] = string(value)
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func sameMap(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if other, ok := right[key]; !ok || other != value {
			return false
		}
	}
	return true
}

func wantCode(t *testing.T, got, want contract.Code) {
	t.Helper()
	if got != want {
		t.Fatalf("code %q; want %q", got, want)
	}
}

func readBytes(t *testing.T, path string) []byte {
	t.Helper()
	value, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func entryNames(t *testing.T, directory string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func extraEntries(t *testing.T, directory, kept string) []string {
	t.Helper()
	var extras []string
	for _, name := range entryNames(t, directory) {
		if name != kept {
			extras = append(extras, name)
		}
	}
	return extras
}

func within(t *testing.T, limit time.Duration, name string, action func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		action()
	}()
	select {
	case <-done:
	case <-time.After(limit):
		t.Fatalf("%s did not return within %s", name, limit)
	}
}

type watcher struct {
	stop chan struct{}
	done chan struct{}
}

func watchDirectory(directory, kept string, action func(path string)) *watcher {
	result := &watcher{stop: make(chan struct{}), done: make(chan struct{})}
	go func() {
		defer close(result.done)
		for {
			select {
			case <-result.stop:
				return
			default:
			}
			entries, err := os.ReadDir(directory)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if entry.Name() != kept {
					action(filepath.Join(directory, entry.Name()))
					return
				}
			}
		}
	}()
	return result
}

func (w *watcher) finish() {
	close(w.stop)
	<-w.done
}
