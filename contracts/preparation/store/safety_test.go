package store_test

import (
	"bytes"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	contract "onboard-kit/contracts/preparation/store"
)

type unsafeSelectionCase struct {
	name  string
	build func(t *testing.T, box sandbox, outside string) contract.LoadRequest
	want  contract.Code
}

func unsafeSelectionCases() []unsafeSelectionCase {
	return []unsafeSelectionCase{
		{name: "plan-is-a-symlink", want: contract.CodeSource, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			link := filepath.Join(box.Root, "selected.yaml")
			symlink(t, box.Plan, link)
			request := box.request()
			request.Plan = link
			return request
		}},
		{name: "plan-parent-is-a-symlink", want: contract.CodeSource, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			link := filepath.Join(box.Root, "selected-parent")
			symlink(t, box.Root, link)
			request := box.request()
			request.Plan = filepath.Join(link, "organization.yaml")
			return request
		}},
		{name: "plan-with-traversal-staying-inside-root", want: contract.CodeSource, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			if err := os.MkdirAll(filepath.Join(box.Root, "records"), 0o700); err != nil {
				t.Fatal(err)
			}
			request := box.request()
			request.Plan = box.Root + "/records/../organization.yaml"
			return request
		}},
		{name: "plan-with-trailing-separator", want: contract.CodeSource, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			request := box.request()
			request.Plan = box.Plan + "/"
			return request
		}},
		{name: "plan-with-nul", want: contract.CodeSource, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			request := box.request()
			request.Plan = box.Plan + "\x00"
			return request
		}},
		{name: "plan-is-a-fifo", want: contract.CodeSource, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			path := filepath.Join(box.Root, "fifo.yaml")
			if err := syscall.Mkfifo(path, 0o600); err != nil {
				t.Fatal(err)
			}
			request := box.request()
			request.Plan = path
			return request
		}},
		{name: "root-is-a-symlink", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			link := filepath.Join(box.Base, "selected-root")
			symlink(t, box.Root, link)
			request := box.request()
			request.Root = link
			return request
		}},
		{name: "root-parent-is-a-symlink", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			link := filepath.Join(box.Base, "selected-parent")
			symlink(t, box.Base, link)
			request := box.request()
			request.Root = filepath.Join(link, "root")
			return request
		}},
		{name: "root-with-traversal-staying-inside", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			if err := os.MkdirAll(filepath.Join(box.Root, "records"), 0o700); err != nil {
				t.Fatal(err)
			}
			request := box.request()
			request.Root = box.Root + "/records/.."
			return request
		}},
		{name: "root-with-nul", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			request := box.request()
			request.Root = box.Root + "\x00"
			return request
		}},
		{name: "records-is-a-symlink-to-outside", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			writeFile(t, filepath.Join(outside, "newcomer", "preparation.obk.yaml"), fixture(t, "identity.yaml"))
			symlink(t, outside, filepath.Join(box.Root, "records"))
			return box.request()
		}},
		{name: "person-directory-is-a-symlink-to-outside", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			writeFile(t, filepath.Join(outside, "preparation.obk.yaml"), fixture(t, "identity.yaml"))
			if err := os.MkdirAll(filepath.Join(box.Root, "records"), 0o700); err != nil {
				t.Fatal(err)
			}
			symlink(t, outside, filepath.Join(box.Root, "records", "newcomer"))
			return box.request()
		}},
		{name: "preparation-is-a-symlink-to-a-file", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			writeFile(t, filepath.Join(outside, "held.yaml"), fixture(t, "identity.yaml"))
			if err := os.MkdirAll(filepath.Dir(box.Prep), 0o700); err != nil {
				t.Fatal(err)
			}
			symlink(t, filepath.Join(outside, "held.yaml"), box.Prep)
			return box.request()
		}},
		{name: "preparation-is-a-dangling-symlink", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			if err := os.MkdirAll(filepath.Dir(box.Prep), 0o700); err != nil {
				t.Fatal(err)
			}
			symlink(t, filepath.Join(outside, "nothing-here.yaml"), box.Prep)
			return box.request()
		}},
		{name: "preparation-is-a-directory", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			if err := os.MkdirAll(box.Prep, 0o700); err != nil {
				t.Fatal(err)
			}
			return box.request()
		}},
		{name: "preparation-is-a-fifo", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			if err := os.MkdirAll(filepath.Dir(box.Prep), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := syscall.Mkfifo(box.Prep, 0o600); err != nil {
				t.Fatal(err)
			}
			return box.request()
		}},
		{name: "person-directory-is-a-file", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			writeFile(t, filepath.Join(box.Root, "records", "newcomer"), []byte("not a directory\n"))
			return box.request()
		}},
		{name: "root-is-a-fifo", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			path := filepath.Join(box.Base, "root-fifo")
			if err := syscall.Mkfifo(path, 0o600); err != nil {
				t.Fatal(err)
			}
			request := box.request()
			request.Root = path
			return request
		}},
		{name: "records-is-a-fifo", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			if err := syscall.Mkfifo(filepath.Join(box.Root, "records"), 0o600); err != nil {
				t.Fatal(err)
			}
			return box.request()
		}},
		{name: "person-directory-is-a-fifo", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			if err := os.MkdirAll(filepath.Join(box.Root, "records"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := syscall.Mkfifo(filepath.Join(box.Root, "records", account), 0o600); err != nil {
				t.Fatal(err)
			}
			return box.request()
		}},
		{name: "preparation-is-the-plan-through-a-hard-link", want: contract.CodePath, build: func(t *testing.T, box sandbox, outside string) contract.LoadRequest {
			if err := os.MkdirAll(filepath.Dir(box.Prep), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Link(box.Plan, box.Prep); err != nil {
				t.Fatal(err)
			}
			return box.request()
		}},
	}
}

func TestUnsafeSelectionsAreRefusedWithoutFollowing(t *testing.T) {
	t.Parallel()
	for _, c := range unsafeSelectionCases() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			box := newSandbox(t, "plan.yaml", "")
			outside := filepath.Join(box.Base, "outside")
			if err := os.MkdirAll(outside, 0o700); err != nil {
				t.Fatal(err)
			}
			request := c.build(t, box, outside)
			held := snapshot(t, box.Base)
			var result contract.LoadResult
			within(t, 10*time.Second, "load", func() { result = newStore().Load(request) })
			wantCode(t, result.Code, c.want)
			if result.Loaded != nil {
				t.Fatal("refused load returned a loaded state")
			}
			if after := snapshot(t, box.Base); !sameMap(after, held) {
				t.Fatal("refused load changed the filesystem")
			}
		})
	}
}

func symlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}

func TestSaveNeverFollowsASymlinkIntoAnotherDirectory(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name   string
		before string
		swap   func(t *testing.T, box sandbox, outside string)
	}{
		{name: "records-replaced-by-symlink-before-directories-are-created", swap: func(t *testing.T, box sandbox, outside string) {
			symlink(t, outside, filepath.Join(box.Root, "records"))
		}},
		{name: "person-directory-replaced-by-symlink", before: "identity.yaml", swap: func(t *testing.T, box sandbox, outside string) {
			writeFile(t, filepath.Join(outside, "preparation.obk.yaml"), fixture(t, "identity.yaml"))
			dir := filepath.Dir(box.Prep)
			if err := os.Rename(dir, dir+"-moved"); err != nil {
				t.Fatal(err)
			}
			symlink(t, outside, dir)
		}},
		{name: "preparation-replaced-by-symlink", before: "identity.yaml", swap: func(t *testing.T, box sandbox, outside string) {
			writeFile(t, filepath.Join(outside, "held.yaml"), fixture(t, "identity.yaml"))
			if err := os.Remove(box.Prep); err != nil {
				t.Fatal(err)
			}
			symlink(t, filepath.Join(outside, "held.yaml"), box.Prep)
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			box := newSandbox(t, "plan.yaml", c.before)
			outside := filepath.Join(box.Base, "outside")
			if err := os.MkdirAll(outside, 0o700); err != nil {
				t.Fatal(err)
			}
			loaded := box.load(t)
			c.swap(t, box, outside)
			held := snapshot(t, box.Base)
			result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "named.yaml"), CreateDirectories: true})
			wantCode(t, result.Code, contract.CodeStale)
			if result.Committed || result.CreatedDirectories {
				t.Fatalf("refusal reports effects %#v", result)
			}
			if after := snapshot(t, box.Base); !sameMap(after, held) {
				t.Fatal("save wrote through a swapped path")
			}
		})
	}
}

func TestPublishedFileIsNotReadableByOthers(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "")
	loaded := box.load(t)
	wantCode(t, newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "named.yaml"), CreateDirectories: true}).Code, contract.CodeSaved)
	for _, path := range []string{filepath.Join(box.Root, "records"), filepath.Dir(box.Prep), box.Prep} {
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm()&0o077 != 0 {
			t.Fatalf("%s is not private: %v %v", path, info, err)
		}
	}
	if !bytes.Equal(readBytes(t, box.Prep), fixture(t, "named.yaml")) {
		t.Fatal("bytes differ")
	}
}
