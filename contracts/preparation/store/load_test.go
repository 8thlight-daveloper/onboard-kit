package store_test

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	contract "onboard-kit/contracts/preparation/store"
)

func TestLoadReportsHeldState(t *testing.T) {
	t.Parallel()
	for _, held := range []struct {
		name, before string
		exists       bool
		missingDirs  bool
		prepare      func(t *testing.T, box sandbox)
	}{
		{name: "absent-without-directories", exists: false, missingDirs: true},
		{name: "absent-with-directories", exists: false, missingDirs: false, prepare: func(t *testing.T, box sandbox) {
			if err := os.MkdirAll(filepath.Dir(box.Prep), 0o700); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "absent-with-parent-only", exists: false, missingDirs: true, prepare: func(t *testing.T, box sandbox) {
			if err := os.MkdirAll(filepath.Join(box.Root, "records"), 0o700); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "existing-identity", before: "identity.yaml", exists: true},
	} {
		t.Run(held.name, func(t *testing.T) {
			t.Parallel()
			box := newSandbox(t, "plan.yaml", held.before)
			if held.prepare != nil {
				held.prepare(t, box)
			}
			loaded := box.load(t)
			if loaded.Account() != account || loaded.Destination() != destination {
				t.Fatalf("identity %q %q", loaded.Account(), loaded.Destination())
			}
			if loaded.Exists() != held.exists || loaded.DirectoriesMissing() != held.missingDirs {
				t.Fatalf("exists %v missing %v; want %v %v", loaded.Exists(), loaded.DirectoriesMissing(), held.exists, held.missingDirs)
			}
			if held.exists && !bytes.Equal(loaded.Bytes(), fixture(t, held.before)) {
				t.Fatalf("held bytes %q", loaded.Bytes())
			}
			if !held.exists && len(loaded.Bytes()) != 0 {
				t.Fatalf("absent preparation has bytes %q", loaded.Bytes())
			}
			preparation := loaded.Preparation()
			if preparation.SchemaVersion != 1 || preparation.Kind != "preparation" || preparation.Account != account || preparation.PlanID != "planone" {
				t.Fatalf("held identity %#v", preparation)
			}
			if preparation.Recipient.Name != "" || len(preparation.Arrangements) != 0 {
				t.Fatalf("identity-only preparation carries data %#v", preparation)
			}
			plan := loaded.Plan()
			if plan.ID != "planone" || len(plan.Directory.Contacts) != 3 || plan.Directory.Contacts[0].ID != "owner" || plan.Directory.Teams[0].ID != "team" {
				t.Fatalf("plan %#v", plan)
			}
		})
	}
}

func TestLoadReportsHeldContent(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "arrangements-held.yaml")
	preparation := box.load(t).Preparation()
	if preparation.Recipient.Name != "Held newcomer" || preparation.Recipient.Manager != "stale-manager" {
		t.Fatalf("recipient %#v", preparation.Recipient)
	}
	if len(preparation.Arrangements) != 2 || preparation.Arrangements[0].ID != "z-retained" || preparation.Arrangements[1].ID != "a-edit" {
		t.Fatalf("arrangements %#v", preparation.Arrangements)
	}
	if got := preparation.Arrangements[1].Contacts; !reflect.DeepEqual(got, []string{"helper", "manager", "helper"}) {
		t.Fatalf("helpers %#v", got)
	}
	if preparation.UnknownFields["x-retained"] != "PRIVATE-unknown-value" {
		t.Fatalf("unknown fields %#v", preparation.UnknownFields)
	}
	box = newSandbox(t, "plan.yaml", "literal.yaml")
	if got := box.load(t).Preparation().Recipient.Name; got != "Ünal\x1b‮‍\u0085  \nline\t\"\\" {
		t.Fatalf("literal name %q", got)
	}
}

func TestLoadChangesNothing(t *testing.T) {
	t.Parallel()
	for _, before := range []string{"", "identity.yaml", "noop-empty.yaml"} {
		box := newSandbox(t, "plan.yaml", before)
		held := snapshot(t, box.Root)
		stat, err := os.Stat(box.Root)
		if err != nil {
			t.Fatal(err)
		}
		loaded := box.load(t)
		_ = loaded
		if after := snapshot(t, box.Root); !sameMap(after, held) {
			t.Fatalf("load of %q changed the filesystem: %#v", before, after)
		}
		again, err := os.Stat(box.Root)
		if err != nil || !again.ModTime().Equal(stat.ModTime()) {
			t.Fatalf("load of %q changed the root directory", before)
		}
	}
}

func TestLoadedValuesAreIndependentCopies(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "arrangements-held.yaml")
	loaded := box.load(t)
	data := loaded.Bytes()
	data[0] = 'X'
	preparation := loaded.Preparation()
	preparation.Arrangements[1].Contacts[0] = "mutated"
	preparation.Arrangements = preparation.Arrangements[:1]
	preparation.Recipient.Name = "mutated"
	plan := loaded.Plan()
	plan.Directory.Contacts[0].ID = "mutated"
	plan.ID = "mutated"
	if !bytes.Equal(loaded.Bytes(), fixture(t, "arrangements-held.yaml")) {
		t.Fatal("mutating returned bytes changed the held bytes")
	}
	again := loaded.Preparation()
	if again.Recipient.Name != "Held newcomer" || len(again.Arrangements) != 2 || again.Arrangements[1].Contacts[0] != "helper" {
		t.Fatalf("mutating a returned preparation changed the held one: %#v", again)
	}
	if loaded.Plan().ID != "planone" || loaded.Plan().Directory.Contacts[0].ID != "owner" {
		t.Fatal("mutating a returned plan changed the held one")
	}
	result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "arrangements-held.yaml")})
	wantCode(t, result.Code, contract.CodeUnchanged)
}

type loadRefusalCase struct {
	name   string
	plan   string
	before string
	change func(t *testing.T, box sandbox, request *contract.LoadRequest)
	want   contract.Code
}

func loadRefusalCases(missing string) []loadRefusalCase {
	return []loadRefusalCase{
		{name: "root-missing", plan: "plan.yaml", want: contract.CodePath, change: func(t *testing.T, box sandbox, request *contract.LoadRequest) { request.Root = missing }},
		{name: "root-is-a-file", plan: "plan.yaml", want: contract.CodePath, change: func(t *testing.T, box sandbox, request *contract.LoadRequest) { request.Root = box.Plan }},
		{name: "root-empty", plan: "plan.yaml", want: contract.CodePath, change: func(t *testing.T, box sandbox, request *contract.LoadRequest) { request.Root = "" }},
		{name: "plan-missing", plan: "plan.yaml", want: contract.CodeSource, change: func(t *testing.T, box sandbox, request *contract.LoadRequest) { request.Plan = missing }},
		{name: "plan-empty", plan: "plan.yaml", want: contract.CodeSource, change: func(t *testing.T, box sandbox, request *contract.LoadRequest) { request.Plan = "" }},
		{name: "plan-is-a-directory", plan: "plan.yaml", want: contract.CodeSource, change: func(t *testing.T, box sandbox, request *contract.LoadRequest) { request.Plan = box.Root }},
		{name: "plan-not-yaml", plan: "malformed.yaml", want: contract.CodeSource},
		{name: "plan-invalid-utf8", plan: "invalid-utf8.yaml", want: contract.CodeSource},
		{name: "plan-without-id", plan: "plan-no-id.yaml", want: contract.CodePlan},
		{name: "flat-results-path", plan: "plan-flat.yaml", want: contract.CodePath},
		{name: "account-empty", plan: "plan.yaml", want: contract.CodePath, change: func(t *testing.T, box sandbox, request *contract.LoadRequest) { request.Account = "" }},
		{name: "account-with-separator", plan: "plan.yaml", want: contract.CodePath, change: func(t *testing.T, box sandbox, request *contract.LoadRequest) { request.Account = "a/b" }},
		{name: "account-with-traversal", plan: "plan.yaml", want: contract.CodePath, change: func(t *testing.T, box sandbox, request *contract.LoadRequest) { request.Account = ".." }},
		{name: "preparation-not-yaml", plan: "plan.yaml", before: "malformed.yaml", want: contract.CodeSource},
		{name: "preparation-invalid-utf8", plan: "plan.yaml", before: "invalid-utf8.yaml", want: contract.CodeSource},
		{name: "preparation-wrong-kind", plan: "plan.yaml", before: "wrong-kind.yaml", want: contract.CodeSource},
		{name: "preparation-for-someone-else", plan: "plan.yaml", before: "other-account.yaml", want: contract.CodeIdentity},
		{name: "preparation-for-another-plan", plan: "plan.yaml", before: "other-plan.yaml", want: contract.CodeIdentity},
		{name: "preparation-for-a-lower-sorting-account", plan: "plan.yaml", before: "lower-account.yaml", want: contract.CodeIdentity},
		{name: "preparation-for-a-lower-sorting-plan", plan: "plan.yaml", before: "lower-plan.yaml", want: contract.CodeIdentity},
		{name: "plan-over-the-limit", plan: "plan.yaml", want: contract.CodeSource, change: func(t *testing.T, box sandbox, request *contract.LoadRequest) {
			writeFile(t, box.Plan, padded(t, fixture(t, "plan.yaml"), byteLimit+1))
		}},
		{name: "preparation-over-the-limit", plan: "plan.yaml", before: "identity.yaml", want: contract.CodeSource, change: func(t *testing.T, box sandbox, request *contract.LoadRequest) {
			writeFile(t, box.Prep, padded(t, fixture(t, "identity.yaml"), byteLimit+1))
		}},
		{name: "preparation-for-someone-else-with-other-account-requested", plan: "plan.yaml", before: "identity.yaml", want: contract.CodeIdentity, change: func(t *testing.T, box sandbox, request *contract.LoadRequest) {
			writeFile(t, filepath.Join(box.Root, "records", "second", "preparation.obk.yaml"), fixture(t, "identity.yaml"))
			request.Account = "second"
		}},
	}
}

func TestLoadRefusals(t *testing.T) {
	t.Parallel()
	for _, c := range loadRefusalCases(filepath.Join(t.TempDir(), "absent")) {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			box := newSandbox(t, c.plan, c.before)
			request := box.request()
			if c.change != nil {
				c.change(t, box, &request)
			}
			held := snapshot(t, box.Root)
			result := newStore().Load(request)
			wantCode(t, result.Code, c.want)
			if result.Loaded != nil {
				t.Fatal("refused load returned a loaded state")
			}
			if after := snapshot(t, box.Root); !sameMap(after, held) {
				t.Fatal("refused load changed the filesystem")
			}
		})
	}
}

func TestSourceSizeBoundary(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"plan", "preparation"} {
		for _, excess := range []int{0, 1} {
			box := newSandbox(t, "plan.yaml", "identity.yaml")
			path, name := box.Plan, "plan.yaml"
			if source == "preparation" {
				path, name = box.Prep, "identity.yaml"
			}
			writeFile(t, path, padded(t, fixture(t, name), byteLimit+excess))
			held := snapshot(t, box.Root)
			result := newStore().Load(box.request())
			if excess == 0 {
				wantCode(t, result.Code, contract.CodeLoaded)
			} else {
				wantCode(t, result.Code, contract.CodeSource)
			}
			if after := snapshot(t, box.Root); !sameMap(after, held) {
				t.Fatalf("%s boundary changed the filesystem", source)
			}
		}
	}
}

func TestLoadReturnsEveryByteOfAFileAtTheSizeLimit(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	held := padded(t, fixture(t, "named.yaml"), byteLimit)
	writeFile(t, box.Prep, held)
	loaded := box.load(t)
	got := loaded.Bytes()
	if len(got) != byteLimit || !bytes.Equal(got, held) {
		t.Fatalf("loaded %d bytes; want all %d", len(got), byteLimit)
	}
	if loaded.Preparation().Recipient.Name != "Example newcomer" {
		t.Fatalf("recipient %#v", loaded.Preparation().Recipient)
	}
	wantCode(t, newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: held}).Code, contract.CodeUnchanged)
}

func TestLoadIsDeterministic(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "arrangements-held.yaml")
	first, second := box.load(t), box.load(t)
	if !reflect.DeepEqual(first.Preparation(), second.Preparation()) || !reflect.DeepEqual(first.Plan(), second.Plan()) || !bytes.Equal(first.Bytes(), second.Bytes()) || first.Destination() != second.Destination() {
		t.Fatal("two loads of the same files differ")
	}
}

func TestLoadDoesNotReadTheRecord(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	record := filepath.Join(box.Root, "records", "newcomer", "record.obk.yaml")
	writeFile(t, record, []byte("PRIVATE malformed record [\n"))
	if err := os.Chmod(record, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(record, 0o600) })
	_ = box.load(t)
	wantCode(t, newStore().Save(contract.SaveRequest{Loaded: box.load(t), Reviewed: fixture(t, "identity.yaml")}).Code, contract.CodeUnchanged)
}
