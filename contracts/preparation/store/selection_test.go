package store_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	contract "onboard-kit/contracts/preparation/store"
)

func TestRelativeRootAndPlanAreResolvedAgainstTheWorkingDirectory(t *testing.T) {
	for _, c := range []struct {
		name, directory, root, plan string
	}{
		{name: "root-and-plan-below-the-directory", directory: "", root: "root", plan: "root/organization.yaml"},
		{name: "dot-prefixed", directory: "", root: "./root", plan: "./root/organization.yaml"},
		{name: "directory-is-the-root", directory: "root", root: ".", plan: "organization.yaml"},
		{name: "plan-outside-the-root", directory: "", root: "root", plan: "organization.yaml"},
	} {
		t.Run(c.name, func(t *testing.T) {
			box := newSandbox(t, "plan.yaml", "identity.yaml")
			if c.name == "plan-outside-the-root" {
				if err := os.Rename(box.Plan, filepath.Join(box.Base, "organization.yaml")); err != nil {
					t.Fatal(err)
				}
			}
			named := fixture(t, "named.yaml")
			t.Chdir(filepath.Join(box.Base, c.directory))
			request := contract.LoadRequest{Root: c.root, Plan: c.plan, Account: account}
			listed := newStore().List(contract.ListRequest{Root: c.root, Plan: c.plan})
			wantCode(t, listed.Code, contract.CodeListed)
			if len(listed.People) != 1 || listed.People[0].Account != account || listed.People[0].Destination != destination {
				t.Fatalf("people %#v", listed.People)
			}
			loaded := newStore().Load(request)
			wantCode(t, loaded.Code, contract.CodeLoaded)
			if !loaded.Loaded.Exists() || loaded.Loaded.Destination() != destination || loaded.Loaded.Plan().ID != "planone" {
				t.Fatalf("loaded %q exists %v", loaded.Loaded.Destination(), loaded.Loaded.Exists())
			}
			result := newStore().Save(contract.SaveRequest{Loaded: loaded.Loaded, Reviewed: named})
			wantCode(t, result.Code, contract.CodeSaved)
			if got := readBytes(t, box.Prep); string(got) != string(named) {
				t.Fatalf("saved %q", got)
			}
		})
	}
}

func TestRelativePathsKeepTheSafetyRules(t *testing.T) {
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	symlink(t, box.Root, filepath.Join(box.Base, "linked-root"))
	symlink(t, box.Plan, filepath.Join(box.Root, "linked-plan.yaml"))
	t.Chdir(box.Base)
	for _, c := range []struct {
		name, root, plan string
		want             contract.Code
	}{
		{name: "root-is-a-symlink", root: "linked-root", plan: "root/organization.yaml", want: contract.CodePath},
		{name: "root-with-traversal", root: "root/../root", plan: "root/organization.yaml", want: contract.CodePath},
		{name: "root-above-the-directory", root: "../" + filepath.Base(box.Base) + "/root", plan: "root/organization.yaml", want: contract.CodePath},
		{name: "root-missing", root: "absent", plan: "root/organization.yaml", want: contract.CodePath},
		{name: "plan-is-a-symlink", root: "root", plan: "root/linked-plan.yaml", want: contract.CodeSource},
		{name: "plan-parent-is-a-symlink", root: "root", plan: "linked-root/organization.yaml", want: contract.CodeSource},
		{name: "plan-with-traversal", root: "root", plan: "root/../root/organization.yaml", want: contract.CodeSource},
		{name: "plan-missing", root: "root", plan: "root/absent.yaml", want: contract.CodeSource},
	} {
		t.Run(c.name, func(t *testing.T) {
			wantCode(t, newStore().Load(contract.LoadRequest{Root: c.root, Plan: c.plan, Account: account}).Code, c.want)
			wantCode(t, newStore().List(contract.ListRequest{Root: c.root, Plan: c.plan}).Code, c.want)
		})
	}
}

func TestRelativeLoadedStateIsStaleWhenTheWorkingDirectoryMoves(t *testing.T) {
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	other := filepath.Join(box.Base, "elsewhere")
	if err := os.MkdirAll(other, 0o700); err != nil {
		t.Fatal(err)
	}
	named := fixture(t, "named.yaml")
	t.Chdir(box.Base)
	loaded := newStore().Load(contract.LoadRequest{Root: "root", Plan: "root/organization.yaml", Account: account})
	wantCode(t, loaded.Code, contract.CodeLoaded)
	t.Chdir(other)
	held := snapshot(t, box.Base)
	result := newStore().Save(contract.SaveRequest{Loaded: loaded.Loaded, Reviewed: named, CreateDirectories: true})
	wantCode(t, result.Code, contract.CodeStale)
	if result.Committed || result.CreatedDirectories {
		t.Fatalf("refusal reports effects %#v", result)
	}
	if after := snapshot(t, box.Base); !sameMap(after, held) {
		t.Fatal("stale save changed the filesystem")
	}
}

func TestControlBytesOtherThanNulAreOrdinaryNameBytes(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ name, root, plan string }{
		{name: "start-of-heading-in-the-root", root: "root\x01", plan: "organization.yaml"},
		{name: "unit-separator-in-the-root", root: "root\x1f", plan: "organization.yaml"},
		{name: "start-of-heading-in-the-plan", root: "root", plan: "organization\x01.yaml"},
		{name: "newline-in-both", root: "root\n", plan: "organization\n.yaml"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			box := newSandbox(t, "plan.yaml", "identity.yaml")
			root := filepath.Join(box.Base, c.root)
			if root != box.Root {
				if err := os.Rename(box.Root, root); err != nil {
					t.Fatal(err)
				}
			}
			plan := filepath.Join(root, c.plan)
			if err := os.Rename(filepath.Join(root, "organization.yaml"), plan); err != nil {
				t.Fatal(err)
			}
			listed := newStore().List(contract.ListRequest{Root: root, Plan: plan})
			wantCode(t, listed.Code, contract.CodeListed)
			if len(listed.People) != 1 || listed.People[0].Account != account {
				t.Fatalf("people %#v", listed.People)
			}
			loaded := newStore().Load(contract.LoadRequest{Root: root, Plan: plan, Account: account})
			wantCode(t, loaded.Code, contract.CodeLoaded)
			result := newStore().Save(contract.SaveRequest{Loaded: loaded.Loaded, Reviewed: fixture(t, "named.yaml")})
			wantCode(t, result.Code, contract.CodeSaved)
			if got := readBytes(t, filepath.Join(root, filepath.FromSlash(destination))); string(got) != string(fixture(t, "named.yaml")) {
				t.Fatalf("saved %q", got)
			}
		})
	}
}

func TestReplacedRootIsStaleWhenThePlanLivesOutsideIt(t *testing.T) {
	t.Parallel()
	for _, consent := range []bool{false, true} {
		for _, reviewed := range []string{"named.yaml", "identity.yaml"} {
			t.Run(fmt.Sprintf("%s-consent-%v", reviewed, consent), func(t *testing.T) {
				t.Parallel()
				box := newSandbox(t, "plan.yaml", "")
				plan := filepath.Join(box.Base, "organization.yaml")
				if err := os.Rename(box.Plan, plan); err != nil {
					t.Fatal(err)
				}
				loaded := newStore().Load(contract.LoadRequest{Root: box.Root, Plan: plan, Account: account})
				wantCode(t, loaded.Code, contract.CodeLoaded)
				if err := os.Rename(box.Root, box.Root+"-moved"); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(box.Root, 0o700); err != nil {
					t.Fatal(err)
				}
				held := snapshot(t, box.Base)
				result := newStore().Save(contract.SaveRequest{Loaded: loaded.Loaded, Reviewed: fixture(t, reviewed), CreateDirectories: consent})
				wantCode(t, result.Code, contract.CodeStale)
				if result.Committed || result.CreatedDirectories {
					t.Fatalf("refusal reports effects %#v", result)
				}
				if after := snapshot(t, box.Base); !sameMap(after, held) {
					t.Fatal("stale save changed the filesystem")
				}
			})
		}
	}
}

func TestRootAndPlanBelowDirectoriesWithUnusualNamesAreOrdinary(t *testing.T) {
	for _, name := range []string{"-dash", " space", "+plus", "!bang", "#hash", ",comma", "-", " "} {
		t.Run(name, func(t *testing.T) {
			base, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			root := filepath.Join(base, name, "root")
			plan := filepath.Join(root, "organization.yaml")
			prep := filepath.Join(root, filepath.FromSlash(destination))
			writeFile(t, plan, fixture(t, "plan.yaml"))
			writeFile(t, prep, fixture(t, "identity.yaml"))
			request := contract.LoadRequest{Root: root, Plan: plan, Account: account}
			loaded := newStore().Load(request)
			wantCode(t, loaded.Code, contract.CodeLoaded)
			listed := newStore().List(contract.ListRequest{Root: root, Plan: plan})
			wantCode(t, listed.Code, contract.CodeListed)
			if len(listed.People) != 1 || listed.People[0].Account != account {
				t.Fatalf("listed %#v; want the one person %q", listed.People, account)
			}
			saved := newStore().Save(contract.SaveRequest{Loaded: loaded.Loaded, Reviewed: fixture(t, "named.yaml")})
			wantCode(t, saved.Code, contract.CodeSaved)
			if got := readBytes(t, prep); string(got) != string(fixture(t, "named.yaml")) {
				t.Fatal("saved file does not hold the reviewed bytes")
			}
		})
	}
}

func TestRootThatCannotBeSearchedIsStillARootWhenWrittenWithATrailingDot(t *testing.T) {
	for _, mode := range []os.FileMode{0o600, 0o400} {
		t.Run(mode.String(), func(t *testing.T) {
			base, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			closed := filepath.Join(base, "closed")
			if err := os.Mkdir(closed, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(closed, mode); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(closed, 0o700) })
			if os.Geteuid() == 0 {
				t.Skip("root ignores directory permission bits")
			}
			root := closed + "/."
			plan := filepath.Join(base, "absent.yaml")
			listed := newStore().List(contract.ListRequest{Root: root, Plan: plan})
			wantCode(t, listed.Code, contract.CodeSource)
			loaded := newStore().Load(contract.LoadRequest{Root: root, Plan: plan, Account: account})
			wantCode(t, loaded.Code, contract.CodeSource)
		})
	}
}
