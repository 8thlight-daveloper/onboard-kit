package store_test

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	contract "onboard-kit/contracts/preparation/store"
)

type foreignLoaded struct{ contract.Loaded }

func TestSavePublishesExactlyTheReviewedBytes(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, plan, before, reviewed string
		consent                      bool
		created                      bool
	}{
		{name: "new-identity-with-directories", plan: "plan.yaml", reviewed: "identity.yaml", consent: true, created: true},
		{name: "new-literal-controls", plan: "plan.yaml", reviewed: "literal.yaml", consent: true, created: true},
		{name: "new-in-existing-directory", plan: "plan.yaml", reviewed: "named.yaml", consent: false},
		{name: "replace-recipient", plan: "plan.yaml", before: "identity.yaml", reviewed: "named.yaml"},
		{name: "replace-keeps-unknown-and-arrangements", plan: "plan.yaml", before: "arrangements-held.yaml", reviewed: "arrangements-held-edited.yaml"},
		{name: "replace-with-noncanonical-bytes-verbatim", plan: "plan.yaml", before: "identity.yaml", reviewed: "replaced-reordered.yaml"},
		{name: "replace-unencodable-original", plan: "plan.yaml", before: "noop-empty.yaml", reviewed: "named.yaml"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			box := newSandbox(t, c.plan, c.before)
			if c.before == "" && !c.consent {
				if err := os.MkdirAll(filepath.Dir(box.Prep), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			loaded := box.load(t)
			planBefore := readBytes(t, box.Plan)
			reviewed := fixture(t, c.reviewed)
			held := append([]byte(nil), reviewed...)
			result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: reviewed, CreateDirectories: c.consent})
			wantCode(t, result.Code, contract.CodeSaved)
			if !result.Committed || result.CreatedDirectories != c.created {
				t.Fatalf("committed %v created %v", result.Committed, result.CreatedDirectories)
			}
			if !bytes.Equal(reviewed, held) {
				t.Fatal("save changed the reviewed bytes it was given")
			}
			if got := readBytes(t, box.Prep); !bytes.Equal(got, held) {
				t.Fatalf("saved %q; want %q", got, held)
			}
			if got := readBytes(t, box.Plan); !bytes.Equal(got, planBefore) {
				t.Fatal("save changed the plan")
			}
			if extras := extraEntries(t, filepath.Dir(box.Prep), filepath.Base(box.Prep)); len(extras) != 0 {
				t.Fatalf("temporary leftovers %v", extras)
			}
			info, err := os.Stat(box.Prep)
			if err != nil || info.Mode().Perm() != 0o600 {
				t.Fatalf("saved file mode %v %v", info, err)
			}
			if c.created {
				for _, path := range []string{filepath.Join(box.Root, "records"), filepath.Dir(box.Prep)} {
					info, err := os.Stat(path)
					if err != nil || info.Mode().Perm() != 0o700 {
						t.Fatalf("created directory %s mode %v %v", path, info, err)
					}
				}
			}
			fresh := box.load(t)
			if !fresh.Exists() || !bytes.Equal(fresh.Bytes(), held) {
				t.Fatal("a fresh load does not return the saved bytes")
			}
		})
	}
}

func TestSaveLeavesRecordAndOtherPeopleAlone(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "")
	record := filepath.Join(box.Root, "records", "newcomer", "record.obk.yaml")
	other := filepath.Join(box.Root, "records", "other", "preparation.obk.yaml")
	writeFile(t, record, []byte("PRIVATE malformed record [\n"))
	writeFile(t, other, fixture(t, "other-account.yaml"))
	loaded := box.load(t)
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml")})
	wantCode(t, result.Code, contract.CodeSaved)
	if got := readBytes(t, record); string(got) != "PRIVATE malformed record [\n" {
		t.Fatal("record changed")
	}
	if got := readBytes(t, other); !bytes.Equal(got, fixture(t, "other-account.yaml")) {
		t.Fatal("another person's preparation changed")
	}
	want := map[string]bool{"records/": true, "records/newcomer/": true, "records/other/": true, "organization.yaml": true, destination: true, "records/newcomer/record.obk.yaml": true, "records/other/preparation.obk.yaml": true}
	for path := range snapshot(t, box.Root) {
		if !want[path] {
			t.Fatalf("unexpected path %s", path)
		}
	}
}

func TestSaveRefusesReviewedBytesThatAreNotThisPersons(t *testing.T) {
	t.Parallel()
	oversize := padded(t, fixture(t, "identity.yaml"), byteLimit+1)
	for _, c := range []struct {
		name     string
		reviewed []byte
		want     contract.Code
	}{
		{name: "nil", reviewed: nil, want: contract.CodeDraft},
		{name: "empty", reviewed: []byte{}, want: contract.CodeDraft},
		{name: "not-yaml", reviewed: fixture(t, "malformed.yaml"), want: contract.CodeDraft},
		{name: "invalid-utf8", reviewed: fixture(t, "invalid-utf8.yaml"), want: contract.CodeDraft},
		{name: "wrong-kind", reviewed: fixture(t, "wrong-kind.yaml"), want: contract.CodeDraft},
		{name: "someone-else", reviewed: fixture(t, "other-account.yaml"), want: contract.CodeDraft},
		{name: "another-plan", reviewed: fixture(t, "other-plan.yaml"), want: contract.CodeDraft},
		{name: "lower-sorting-account", reviewed: fixture(t, "lower-account.yaml"), want: contract.CodeDraft},
		{name: "lower-sorting-plan", reviewed: fixture(t, "lower-plan.yaml"), want: contract.CodeDraft},
		{name: "one-byte-over-the-limit", reviewed: oversize, want: contract.CodeTooLarge},
		{name: "garbage-over-the-limit", reviewed: bytes.Repeat([]byte{0xff}, byteLimit+1), want: contract.CodeTooLarge},
	} {
		for _, before := range []string{"", "identity.yaml"} {
			t.Run(c.name+"-before-"+before, func(t *testing.T) {
				t.Parallel()
				box := newSandbox(t, "plan.yaml", before)
				loaded := box.load(t)
				held := snapshot(t, box.Root)
				result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: c.reviewed, CreateDirectories: true})
				wantCode(t, result.Code, c.want)
				if result.Committed || result.CreatedDirectories {
					t.Fatalf("refusal reports effects %#v", result)
				}
				if after := snapshot(t, box.Root); !sameMap(after, held) {
					t.Fatal("refused save changed the filesystem")
				}
			})
		}
	}
}

func TestSaveAcceptsExactlyTheSizeLimit(t *testing.T) {
	t.Parallel()
	for _, before := range []string{"", "identity.yaml"} {
		t.Run("before-"+before, func(t *testing.T) {
			t.Parallel()
			box := newSandbox(t, "plan.yaml", before)
			loaded := box.load(t)
			reviewed := padded(t, fixture(t, "named.yaml"), byteLimit)
			result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: reviewed, CreateDirectories: true})
			wantCode(t, result.Code, contract.CodeSaved)
			got := readBytes(t, box.Prep)
			if len(got) != byteLimit || !bytes.Equal(got, reviewed) {
				t.Fatalf("limit-sized preparation was not saved exactly: %d bytes", len(got))
			}
			again := newStore().Load(box.request())
			wantCode(t, again.Code, contract.CodeLoaded)
			if !bytes.Equal(again.Loaded.Bytes(), reviewed) {
				t.Fatal("limit-sized preparation did not load back exactly")
			}
		})
	}
}

func TestUnchangedDraftKeepsOriginalBytesAndInode(t *testing.T) {
	t.Parallel()
	old := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	for _, c := range []struct{ name, held, reviewed string }{
		{name: "same-bytes", held: "named.yaml", reviewed: "named.yaml"},
		{name: "literal-controls", held: "literal.yaml", reviewed: "literal.yaml"},
		{name: "unencodable-empty-arrangement", held: "noop-empty.yaml", reviewed: "noop-empty.yaml"},
		{name: "same-meaning-other-bytes", held: "reordered.yaml", reviewed: "identity.yaml"},
		{name: "arrangements-and-unknown-data", held: "arrangements-held.yaml", reviewed: "arrangements-held.yaml"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			box := newSandbox(t, "plan.yaml", c.held)
			if err := os.Chtimes(box.Prep, old, old); err != nil {
				t.Fatal(err)
			}
			before, err := os.Stat(box.Prep)
			if err != nil {
				t.Fatal(err)
			}
			held := snapshot(t, box.Root)
			loaded := box.load(t)
			result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, c.reviewed), CreateDirectories: true})
			wantCode(t, result.Code, contract.CodeUnchanged)
			if result.Committed || result.CreatedDirectories {
				t.Fatalf("no-op reports effects %#v", result)
			}
			after, err := os.Stat(box.Prep)
			if err != nil || !os.SameFile(before, after) || !after.ModTime().Equal(old) || after.Mode() != before.Mode() {
				t.Fatalf("no-op touched the file: %v", err)
			}
			if now := snapshot(t, box.Root); !sameMap(now, held) {
				t.Fatal("no-op changed the filesystem")
			}
		})
	}
}

func TestUnchangedDraftNeedsNoWriteAccess(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	if err := os.Chmod(box.Prep, 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Dir(box.Prep), 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Dir(box.Prep), 0o700) })
	loaded := box.load(t)
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml")})
	wantCode(t, result.Code, contract.CodeUnchanged)
}

func TestAbsentIdentityOnlyDraftIsCreatedNotSkipped(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "")
	loaded := box.load(t)
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml"), CreateDirectories: true})
	wantCode(t, result.Code, contract.CodeSaved)
	if !bytes.Equal(readBytes(t, box.Prep), fixture(t, "identity.yaml")) {
		t.Fatal("identity-only preparation not created")
	}
}

func TestSaveNeedsConsentBeforeCreatingDirectories(t *testing.T) {
	t.Parallel()
	for _, partial := range []bool{false, true} {
		box := newSandbox(t, "plan.yaml", "")
		if partial {
			if err := os.MkdirAll(filepath.Join(box.Root, "records"), 0o700); err != nil {
				t.Fatal(err)
			}
		}
		loaded := box.load(t)
		if !loaded.DirectoriesMissing() {
			t.Fatal("directories should be missing")
		}
		held := snapshot(t, box.Root)
		result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml")})
		wantCode(t, result.Code, contract.CodeDirectories)
		if result.Committed || result.CreatedDirectories {
			t.Fatalf("refusal reports effects %#v", result)
		}
		if after := snapshot(t, box.Root); !sameMap(after, held) {
			t.Fatal("unconsented save changed the filesystem")
		}
		again := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml"), CreateDirectories: true})
		wantCode(t, again.Code, contract.CodeSaved)
		if !again.CreatedDirectories {
			t.Fatal("consented save does not report created directories")
		}
	}
}

func TestConsentWithNothingToCreateCreatesNothing(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	loaded := box.load(t)
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "named.yaml"), CreateDirectories: true})
	wantCode(t, result.Code, contract.CodeSaved)
	if result.CreatedDirectories {
		t.Fatal("reports directories created when none were missing")
	}
}

func TestSavedLoadedStateCannotSaveAgain(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	loaded := box.load(t)
	wantCode(t, newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "named.yaml")}).Code, contract.CodeSaved)
	held := snapshot(t, box.Root)
	for _, reviewed := range []string{"replaced.yaml", "named.yaml", "identity.yaml"} {
		result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, reviewed)})
		wantCode(t, result.Code, contract.CodeStale)
		if result.Committed {
			t.Fatal("second save reports a commit")
		}
	}
	if after := snapshot(t, box.Root); !sameMap(after, held) {
		t.Fatal("second save changed the filesystem")
	}
}

func TestSaveRefusesLoadedStateItDidNotIssue(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	held := snapshot(t, box.Root)
	issued := box.load(t)
	typedNil := reflect.Zero(reflect.TypeOf(issued)).Interface().(contract.Loaded)
	for name, loaded := range map[string]contract.Loaded{"nil": nil, "foreign": foreignLoaded{Loaded: issued}, "typed-nil-of-the-issued-type": typedNil} {
		var result contract.SaveResult
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Errorf("%s: save panicked: %v", name, recovered)
				}
			}()
			result = newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "named.yaml"), CreateDirectories: true})
		}()
		if result.Code != contract.CodeLoadedState || result.Committed {
			t.Fatalf("%s: %#v", name, result)
		}
	}
	if after := snapshot(t, box.Root); !sameMap(after, held) {
		t.Fatal("refused save changed the filesystem")
	}
}

func TestLoadedStateWorksAcrossStoreValues(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	loaded := box.load(t)
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "named.yaml")})
	wantCode(t, result.Code, contract.CodeSaved)
}

func TestSavedFileKeepsOnlyTheSelectedPathname(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "named.yaml")
	link := filepath.Join(box.Root, "held-preparation.yaml")
	if err := os.Link(box.Prep, link); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(box.Prep)
	if err != nil {
		t.Fatal(err)
	}
	loaded := box.load(t)
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml")})
	wantCode(t, result.Code, contract.CodeSaved)
	if got := readBytes(t, link); !bytes.Equal(got, fixture(t, "named.yaml")) {
		t.Fatalf("hard link inode was truncated: %q", got)
	}
	if got := readBytes(t, box.Prep); !bytes.Equal(got, fixture(t, "identity.yaml")) {
		t.Fatalf("selected pathname not replaced: %q", got)
	}
	after, err := os.Stat(box.Prep)
	if err != nil || os.SameFile(before, after) {
		t.Fatal("replacement must publish a new inode at the selected pathname")
	}
}

func TestSavedModesNeverBroaden(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name  string
		mode  os.FileMode
		saved os.FileMode
	}{
		{name: "group-and-world-readable", mode: 0o644, saved: 0o600},
		{name: "group-readable", mode: 0o640, saved: 0o600},
		{name: "owner-read-only", mode: 0o400, saved: 0o400},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			box := newSandbox(t, "plan.yaml", "identity.yaml")
			if err := os.Chmod(box.Prep, c.mode); err != nil {
				t.Fatal(err)
			}
			loaded := box.load(t)
			held := snapshot(t, box.Root)
			result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "named.yaml")})
			info, err := os.Stat(box.Prep)
			if err != nil {
				t.Fatal(err)
			}
			if result.Code == contract.CodeSaved {
				if info.Mode().Perm() != c.saved {
					t.Fatalf("saved mode %v; want %v", info.Mode().Perm(), c.saved)
				}
				if got := readBytes(t, box.Prep); !bytes.Equal(got, fixture(t, "named.yaml")) {
					t.Fatalf("saved %q", got)
				}
				return
			}
			if c.mode != 0o400 {
				t.Fatalf("code %q; only a read-only file may be refused", result.Code)
			}
			wantCode(t, result.Code, contract.CodeCommit)
			if info.Mode().Perm() != c.mode || result.Committed {
				t.Fatalf("refusal changed mode to %v", info.Mode().Perm())
			}
			if after := snapshot(t, box.Root); !sameMap(after, held) {
				t.Fatal("refused save changed the filesystem")
			}
		})
	}
}

func TestSaveDoesNotChangeTheReviewedBytesOrLoadedState(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	loaded := box.load(t)
	preparation := loaded.Preparation()
	data := loaded.Bytes()
	wantCode(t, newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "named.yaml")}).Code, contract.CodeSaved)
	if !reflect.DeepEqual(preparation, loaded.Preparation()) || !bytes.Equal(data, loaded.Bytes()) || !loaded.Exists() {
		t.Fatal("save changed what the loaded state reports")
	}
	if !strings.HasPrefix(string(readBytes(t, box.Prep)), "schemaVersion: 1") {
		t.Fatal("saved file is not the reviewed document")
	}
}
