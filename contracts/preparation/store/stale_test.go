package store_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	contract "onboard-kit/contracts/preparation/store"
)

func replaceWithNewInode(t *testing.T, path string, value []byte) {
	t.Helper()
	temporary := path + ".swap"
	writeFile(t, temporary, value)
	if err := os.Rename(temporary, path); err != nil {
		t.Fatal(err)
	}
}

type staleCase struct {
	name   string
	before string
	change func(t *testing.T, box sandbox)
}

func staleCases() []staleCase {
	return []staleCase{
		{name: "plan-bytes-changed", before: "identity.yaml", change: func(t *testing.T, box sandbox) {
			writeFile(t, box.Plan, []byte(externalMessage))
		}},
		{name: "plan-same-length-edit", before: "identity.yaml", change: func(t *testing.T, box sandbox) {
			value := fixture(t, "plan.yaml")
			value[len(value)-3] ^= 1
			writeFile(t, box.Plan, value)
		}},
		{name: "plan-rewritten-with-same-bytes-in-new-file", before: "identity.yaml", change: func(t *testing.T, box sandbox) {
			replaceWithNewInode(t, box.Plan, fixture(t, "plan.yaml"))
		}},
		{name: "plan-mode-changed", before: "identity.yaml", change: func(t *testing.T, box sandbox) {
			if err := os.Chmod(box.Plan, 0o640); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "plan-removed", before: "identity.yaml", change: func(t *testing.T, box sandbox) {
			if err := os.Remove(box.Plan); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "preparation-bytes-changed", before: "identity.yaml", change: func(t *testing.T, box sandbox) {
			writeFile(t, box.Prep, []byte(externalMessage))
		}},
		{name: "preparation-same-length-edit", before: "identity.yaml", change: func(t *testing.T, box sandbox) {
			value := fixture(t, "identity.yaml")
			value[len(value)-3] ^= 1
			writeFile(t, box.Prep, value)
		}},
		{name: "preparation-rewritten-with-same-bytes-in-new-file", before: "identity.yaml", change: func(t *testing.T, box sandbox) {
			replaceWithNewInode(t, box.Prep, fixture(t, "identity.yaml"))
		}},
		{name: "preparation-mode-changed", before: "identity.yaml", change: func(t *testing.T, box sandbox) {
			if err := os.Chmod(box.Prep, 0o640); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "preparation-removed", before: "identity.yaml", change: func(t *testing.T, box sandbox) {
			if err := os.Remove(box.Prep); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "preparation-appeared", change: func(t *testing.T, box sandbox) {
			writeFile(t, box.Prep, []byte(externalMessage))
		}},
		{name: "preparation-appeared-with-identical-bytes", change: func(t *testing.T, box sandbox) {
			writeFile(t, box.Prep, fixture(t, "identity.yaml"))
		}},
		{name: "root-replaced-by-a-copy", before: "identity.yaml", change: func(t *testing.T, box sandbox) {
			moved := box.Root + "-moved"
			if err := os.Rename(box.Root, moved); err != nil {
				t.Fatal(err)
			}
			writeFile(t, box.Plan, fixture(t, "plan.yaml"))
			writeFile(t, box.Prep, fixture(t, "identity.yaml"))
		}},
	}
}

func TestStaleLoadedStateIsRefusedBeforeAnyWrite(t *testing.T) {
	t.Parallel()
	for _, c := range staleCases() {
		for _, reviewed := range []string{"named.yaml", "identity.yaml"} {
			t.Run(c.name+"-saving-"+reviewed, func(t *testing.T) {
				t.Parallel()
				box := newSandbox(t, "plan.yaml", c.before)
				loaded := box.load(t)
				c.change(t, box)
				held := snapshot(t, box.Base)
				result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, reviewed), CreateDirectories: true})
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

func TestStaleCheckHappensBeforeDirectoriesAreCreated(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "")
	loaded := box.load(t)
	writeFile(t, box.Plan, []byte(externalMessage))
	held := snapshot(t, box.Base)
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml"), CreateDirectories: true})
	wantCode(t, result.Code, contract.CodeStale)
	if result.CreatedDirectories {
		t.Fatal("reports created directories")
	}
	if after := snapshot(t, box.Base); !sameMap(after, held) {
		t.Fatalf("stale save created paths: %#v", after)
	}
}

func TestStaleCheckComesBeforeTheConsentCheck(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "")
	loaded := box.load(t)
	writeFile(t, box.Plan, []byte(externalMessage))
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml")})
	wantCode(t, result.Code, contract.CodeStale)
}

func TestExternalEditorBetweenLoadAndSaveIsNotOverwritten(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	loaded := box.load(t)
	writeFile(t, box.Prep, []byte(externalMessage))
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "named.yaml")})
	wantCode(t, result.Code, contract.CodeStale)
	if got := readBytes(t, box.Prep); string(got) != externalMessage {
		t.Fatalf("external bytes overwritten: %q", got)
	}
	fresh := newStore().Load(box.request())
	wantCode(t, fresh.Code, contract.CodeSource)
}

func TestStaleCheckRunsAgainAfterTheTemporaryIsWritten(t *testing.T) {
	for _, c := range []struct {
		name   string
		before string
		change func(t *testing.T, box sandbox)
	}{
		{name: "plan", before: "identity.yaml", change: func(t *testing.T, box sandbox) {
			_ = os.WriteFile(box.Plan, []byte(externalMessage), 0o600)
		}},
		{name: "existing-preparation", before: "identity.yaml", change: func(t *testing.T, box sandbox) {
			_ = os.WriteFile(box.Prep, []byte(externalMessage), 0o600)
		}},
		{name: "appearing-preparation", change: func(t *testing.T, box sandbox) {
			_ = os.WriteFile(box.Prep, []byte(externalMessage), 0o600)
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			stale := 0
			for attempt := 0; attempt < attempts; attempt++ {
				box := newSandbox(t, "plan.yaml", c.before)
				if err := os.MkdirAll(filepath.Dir(box.Prep), 0o700); err != nil {
					t.Fatal(err)
				}
				loaded := box.load(t)
				reviewed := padded(t, fixture(t, "named.yaml"), byteLimit)
				watching := watchDirectory(filepath.Dir(box.Prep), filepath.Base(box.Prep), func(string) { c.change(t, box) })
				result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: reviewed})
				watching.finish()
				switch result.Code {
				case contract.CodeStale:
					stale++
					if result.Committed {
						t.Fatal("stale save reports a commit")
					}
					if extras := extraEntries(t, filepath.Dir(box.Prep), filepath.Base(box.Prep)); len(extras) != 0 {
						t.Fatalf("stale save left %v", extras)
					}
					switch c.name {
					case "plan":
						if !bytes.Equal(readBytes(t, box.Prep), fixture(t, "identity.yaml")) {
							t.Fatal("stale save published")
						}
					default:
						if got := readBytes(t, box.Prep); string(got) != externalMessage {
							t.Fatalf("external bytes were overwritten: %.40q", got)
						}
					}
				case contract.CodeSaved:
				default:
					t.Fatalf("code %q", result.Code)
				}
			}
			if stale == 0 {
				t.Fatal("a change made while the temporary was written was never refused")
			}
			t.Logf("refused %d of %d attempts", stale, attempts)
		})
	}
}
