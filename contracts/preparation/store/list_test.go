package store_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	contract "onboard-kit/contracts/preparation/store"
)

type entry struct {
	Path    string  `json:"path"`
	Kind    string  `json:"kind"`
	Fixture string  `json:"fixture"`
	Text    *string `json:"text"`
	Target  string  `json:"target"`
}

type listCase struct {
	Name        string        `json:"name"`
	Source      string        `json:"source"`
	Plan        string        `json:"plan"`
	Entries     []entry       `json:"entries"`
	Outside     []entry       `json:"outside"`
	Code        contract.Code `json:"code"`
	Accounts    []string      `json:"accounts"`
	Destination string        `json:"destination"`
}

func readListCases(t *testing.T) []listCase {
	t.Helper()
	var cases []listCase
	if err := json.Unmarshal(fixture(t, "cases.json"), &cases); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, c := range cases {
		if c.Name == "" || c.Source == "" || seen[c.Name] || c.Code == "" {
			t.Fatalf("case needs a unique name, a source and a code: %#v", c.Name)
		}
		seen[c.Name] = true
	}
	return cases
}

func create(t *testing.T, base, root string, e entry) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(e.Path))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	switch e.Kind {
	case "dir":
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatal(err)
		}
	case "symlink":
		symlink(t, filepath.Join(base, filepath.FromSlash(e.Target)), path)
	case "fifo":
		if err := syscall.Mkfifo(path, 0o600); err != nil {
			t.Fatal(err)
		}
	default:
		value := []byte(nil)
		if e.Fixture != "" {
			value = fixture(t, e.Fixture)
		} else if e.Text != nil {
			value = []byte(*e.Text)
		}
		writeFile(t, path, value)
	}
}

func TestListCases(t *testing.T) {
	t.Parallel()
	for _, c := range readListCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			box := newSandbox(t, c.Plan, "")
			for _, e := range c.Outside {
				create(t, box.Base, filepath.Join(box.Base, "outside"), e)
			}
			if len(c.Outside) > 0 {
				if err := os.MkdirAll(filepath.Join(box.Base, "outside"), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			for _, e := range c.Entries {
				create(t, box.Base, box.Root, e)
			}
			held := snapshot(t, box.Base)
			var result contract.ListResult
			within(t, 10*time.Second, "list", func() { result = newStore().List(contract.ListRequest{Root: box.Root, Plan: box.Plan}) })
			wantCode(t, result.Code, c.Code)
			var accounts []string
			for _, person := range result.People {
				accounts = append(accounts, person.Account)
				if person.Problem != "" {
					t.Errorf("%s: problem %q for a person whose file is readable", person.Account, person.Problem)
				}
				template := c.Destination
				if template == "" {
					template = "records/{account}/preparation.obk.yaml"
				}
				if want := strings.ReplaceAll(template, "{account}", person.Account); person.Destination != want {
					t.Errorf("destination %q; want %q", person.Destination, want)
				}
			}
			if len(accounts) != len(c.Accounts) || (len(accounts) > 0 && !reflect.DeepEqual(accounts, c.Accounts)) {
				t.Fatalf("accounts %q; want %q", accounts, c.Accounts)
			}
			if after := snapshot(t, box.Base); !sameMap(after, held) {
				t.Fatal("list changed the filesystem")
			}
		})
	}
}

func TestListRefusals(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	missing := filepath.Join(t.TempDir(), "absent")
	pipe := filepath.Join(box.Base, "root-pipe")
	if err := syscall.Mkfifo(pipe, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name string
		req  contract.ListRequest
		want contract.Code
	}{
		{name: "root-missing", req: contract.ListRequest{Root: missing, Plan: box.Plan}, want: contract.CodePath},
		{name: "root-empty", req: contract.ListRequest{Root: "", Plan: box.Plan}, want: contract.CodePath},
		{name: "root-is-a-file", req: contract.ListRequest{Root: box.Plan, Plan: box.Plan}, want: contract.CodePath},
		{name: "root-is-a-pipe", req: contract.ListRequest{Root: pipe, Plan: box.Plan}, want: contract.CodePath},
		{name: "root-with-traversal", req: contract.ListRequest{Root: box.Root + "/records/..", Plan: box.Plan}, want: contract.CodePath},
		{name: "plan-missing", req: contract.ListRequest{Root: box.Root, Plan: missing}, want: contract.CodeSource},
		{name: "plan-empty", req: contract.ListRequest{Root: box.Root, Plan: ""}, want: contract.CodeSource},
		{name: "plan-with-traversal", req: contract.ListRequest{Root: box.Root, Plan: box.Root + "/records/../organization.yaml"}, want: contract.CodeSource},
	} {
		var result contract.ListResult
		within(t, 10*time.Second, c.name, func() { result = newStore().List(c.req) })
		if result.Code != c.want || len(result.People) != 0 {
			t.Errorf("%s: %#v; want code %q", c.name, result, c.want)
		}
	}
}

func TestListShowsThePersonLoadOpens(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	result := newStore().List(contract.ListRequest{Root: box.Root, Plan: box.Plan})
	wantCode(t, result.Code, contract.CodeListed)
	if len(result.People) != 1 || result.People[0].Account != account || result.People[0].Destination != destination || result.People[0].Problem != "" {
		t.Fatalf("people %#v", result.People)
	}
	loaded := newStore().Load(contract.LoadRequest{Root: box.Root, Plan: box.Plan, Account: result.People[0].Account})
	wantCode(t, loaded.Code, contract.CodeLoaded)
	if loaded.Loaded.Destination() != result.People[0].Destination {
		t.Fatal("list and load disagree on the destination")
	}
}

func TestListShowsAPersonWhosePreparationCannotBeOpenedWithTheLoadProblem(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("a file with mode 000 stays readable to root")
	}
	box := newSandbox(t, "plan.yaml", "")
	for _, name := range []string{"alpha", "beta", "gamma"} {
		writeFile(t, filepath.Join(box.Root, "records", name, "preparation.obk.yaml"), fixture(t, "identity.yaml"))
	}
	closed := filepath.Join(box.Root, "records", "beta", "preparation.obk.yaml")
	if err := os.Chmod(closed, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(closed, 0o600) })
	held := snapshotModes(t, box.Root)
	result := newStore().List(contract.ListRequest{Root: box.Root, Plan: box.Plan})
	wantCode(t, result.Code, contract.CodeListed)
	if len(result.People) != 3 {
		t.Fatalf("people %#v; want alpha, beta and gamma", result.People)
	}
	for i, name := range []string{"alpha", "beta", "gamma"} {
		person := result.People[i]
		if person.Account != name || person.Destination != "records/"+name+"/preparation.obk.yaml" {
			t.Fatalf("person %d is %#v; want %q", i, person, name)
		}
	}
	loaded := newStore().Load(contract.LoadRequest{Root: box.Root, Plan: box.Plan, Account: "beta"})
	wantCode(t, loaded.Code, contract.CodeSource)
	if loaded.Loaded != nil {
		t.Fatal("load of a preparation that cannot be opened returned a state")
	}
	if result.People[1].Problem != contract.CodeSource {
		t.Fatalf("problem %q; want %q", result.People[1].Problem, contract.CodeSource)
	}
	if result.People[0].Problem != "" || result.People[2].Problem != "" {
		t.Fatalf("problems %q and %q for readable preparations", result.People[0].Problem, result.People[2].Problem)
	}
	if after := snapshotModes(t, box.Root); !reflect.DeepEqual(after, held) {
		t.Fatal("list changed the filesystem")
	}
}

func snapshotModes(t *testing.T, root string) map[string]os.FileMode {
	t.Helper()
	result := map[string]os.FileMode{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		result[path] = info.Mode()
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestListSkipsANonRegularEntryAtThePreparationPathWhateverItsMode(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores mode 0")
	}
	for _, kind := range []string{"directory", "pipe"} {
		t.Run(kind, func(t *testing.T) {
			box := newSandbox(t, "plan.yaml", "")
			for _, name := range []string{"alpha", "beta"} {
				writeFile(t, filepath.Join(box.Root, "records", name, "preparation.obk.yaml"), fixture(t, "identity.yaml"))
			}
			path := filepath.Join(box.Root, "records", "alpha", "preparation.obk.yaml")
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "directory":
				if err := os.Mkdir(path, 0); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(path, 0o700) })
			default:
				if err := syscall.Mkfifo(path, 0); err != nil {
					t.Fatal(err)
				}
			}
			var result contract.ListResult
			within(t, 10*time.Second, "list", func() { result = newStore().List(contract.ListRequest{Root: box.Root, Plan: box.Plan}) })
			wantCode(t, result.Code, contract.CodeListed)
			if len(result.People) != 1 || result.People[0].Account != "beta" || result.People[0].Problem != "" {
				t.Fatalf("people %#v; want only beta", result.People)
			}
		})
	}
}
