package store_test

import (
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	contract "onboard-kit/contracts/preparation/store"
)

type inodeIdentity struct {
	device int32
	inode  uint64
}

func identityOf(t *testing.T, path string) inodeIdentity {
	t.Helper()
	var stat syscall.Stat_t
	if err := syscall.Stat(path, &stat); err != nil {
		t.Fatal(err)
	}
	return inodeIdentity{device: stat.Dev, inode: stat.Ino}
}

func openDescriptors(t *testing.T, identities map[string]inodeIdentity) []string {
	t.Helper()
	directory, err := os.Open("/dev/fd")
	if err != nil {
		t.Fatal("cannot observe descriptor names")
	}
	names, readErr := directory.Readdirnames(-1)
	closeErr := directory.Close()
	if readErr != nil || closeErr != nil {
		t.Fatal("cannot finish descriptor-name observation")
	}
	var held []string
	for _, name := range names {
		descriptor, err := strconv.Atoi(name)
		if err != nil {
			continue
		}
		var stat syscall.Stat_t
		if err := syscall.Fstat(descriptor, &stat); err != nil {
			continue
		}
		for label, identity := range identities {
			if identity.device == stat.Dev && identity.inode == stat.Ino {
				held = append(held, label)
			}
		}
	}
	return held
}

func TestNoDescriptorOutlivesACall(t *testing.T) {
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	writeFile(t, filepath.Join(box.Root, "records", "second", "preparation.obk.yaml"), fixture(t, "identity.yaml"))
	identities := map[string]inodeIdentity{
		"root":    identityOf(t, box.Root),
		"plan":    identityOf(t, box.Plan),
		"records": identityOf(t, filepath.Join(box.Root, "records")),
		"person":  identityOf(t, filepath.Dir(box.Prep)),
		"second":  identityOf(t, filepath.Join(box.Root, "records", "second")),
		"prep":    identityOf(t, box.Prep),
	}
	if held := openDescriptors(t, identities); len(held) != 0 {
		t.Fatalf("fixture already holds descriptors %v", held)
	}
	check := func(step string) {
		t.Helper()
		if held := openDescriptors(t, identities); len(held) != 0 {
			t.Fatalf("after %s the store still holds descriptors on %v", step, held)
		}
	}
	listed := newStore().List(contract.ListRequest{Root: box.Root, Plan: box.Plan})
	wantCode(t, listed.Code, contract.CodeListed)
	check("list")
	loaded := box.load(t)
	check("load")
	second := newStore().Load(contract.LoadRequest{Root: box.Root, Plan: box.Plan, Account: "second"})
	wantCode(t, second.Code, contract.CodeIdentity)
	check("refused load for another identity")
	wantCode(t, newStore().Load(contract.LoadRequest{Root: box.Root, Plan: box.Plan, Account: "missing"}).Code, contract.CodeLoaded)
	check("load of an absent person")
	wantCode(t, newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml")}).Code, contract.CodeUnchanged)
	check("unchanged save")
	wantCode(t, newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "other-plan.yaml")}).Code, contract.CodeDraft)
	check("refused draft")
	again := box.load(t)
	identities["prep"] = identityOf(t, box.Prep)
	wantCode(t, newStore().Save(contract.SaveRequest{Loaded: again, Reviewed: fixture(t, "named.yaml")}).Code, contract.CodeSaved)
	check("saved")
	wantCode(t, newStore().Save(contract.SaveRequest{Loaded: again, Reviewed: fixture(t, "named.yaml")}).Code, contract.CodeStale)
	check("stale save")
	writeFile(t, box.Plan, []byte(externalMessage))
	wantCode(t, newStore().Load(box.request()).Code, contract.CodeSource)
	check("refused load")
}

func descriptorSet(t *testing.T) map[int]bool {
	t.Helper()
	directory, err := os.Open("/dev/fd")
	if err != nil {
		t.Fatal("cannot observe descriptor names")
	}
	listing := int(directory.Fd())
	names, readErr := directory.Readdirnames(-1)
	closeErr := directory.Close()
	if readErr != nil || closeErr != nil {
		t.Fatal("cannot finish descriptor-name observation")
	}
	result := map[int]bool{}
	for _, name := range names {
		descriptor, err := strconv.Atoi(name)
		if err != nil || descriptor == listing {
			continue
		}
		var stat syscall.Stat_t
		if syscall.Fstat(descriptor, &stat) != nil {
			continue
		}
		result[descriptor] = true
	}
	return result
}

func requireSameDescriptors(t *testing.T, step string, before map[int]bool) {
	t.Helper()
	after := descriptorSet(t)
	for descriptor := range after {
		if !before[descriptor] {
			t.Errorf("after %s descriptor %d is still open", step, descriptor)
		}
	}
	for descriptor := range before {
		if !after[descriptor] {
			t.Errorf("after %s descriptor %d was closed that the caller owned", step, descriptor)
		}
	}
}

func returnsDescriptors(t *testing.T, step string, action func()) {
	t.Helper()
	before := descriptorSet(t)
	within(t, 10*time.Second, step, action)
	requireSameDescriptors(t, step, before)
}

func warmUpStore(t *testing.T) {
	t.Helper()
	box := newSandbox(t, "plan.yaml", "identity.yaml")
	loaded := box.load(t)
	newStore().List(contract.ListRequest{Root: box.Root, Plan: box.Plan})
	newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "identity.yaml")})
}

func TestEveryLoadAndListOutcomeReturnsDescriptors(t *testing.T) {
	warmUpStore(t)
	check := func(name string, request contract.LoadRequest) {
		returnsDescriptors(t, name+" load", func() { newStore().Load(request) })
		returnsDescriptors(t, name+" list", func() { newStore().List(contract.ListRequest{Root: request.Root, Plan: request.Plan}) })
	}
	for _, c := range unsafeSelectionCases() {
		box := newSandbox(t, "plan.yaml", "")
		outside := filepath.Join(box.Base, "outside")
		if err := os.MkdirAll(outside, 0o700); err != nil {
			t.Fatal(err)
		}
		check("unsafe selection "+c.name, c.build(t, box, outside))
	}
	missing := filepath.Join(t.TempDir(), "absent")
	for _, c := range loadRefusalCases(missing) {
		box := newSandbox(t, c.plan, c.before)
		request := box.request()
		if c.change != nil {
			c.change(t, box, &request)
		}
		check("refusal "+c.name, request)
	}
	for _, before := range []string{"", "identity.yaml", "arrangements-held.yaml", "noop-empty.yaml"} {
		box := newSandbox(t, "plan.yaml", before)
		check("held "+before, box.request())
		if err := os.MkdirAll(filepath.Dir(box.Prep), 0o700); err != nil {
			t.Fatal(err)
		}
		check("held with directories "+before, box.request())
	}
	for _, c := range readListCases(t) {
		box := newSandbox(t, c.Plan, "")
		for _, e := range c.Outside {
			create(t, box.Base, filepath.Join(box.Base, "outside"), e)
		}
		for _, e := range c.Entries {
			create(t, box.Base, box.Root, e)
		}
		check("list case "+c.Name, box.request())
	}
}

func TestEverySaveOutcomeReturnsDescriptors(t *testing.T) {
	warmUpStore(t)
	save := func(name string, loaded contract.Loaded, reviewed []byte, consent bool) {
		returnsDescriptors(t, name, func() {
			newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: reviewed, CreateDirectories: consent})
		})
	}
	for _, c := range staleCases() {
		for _, reviewed := range []string{"named.yaml", "identity.yaml"} {
			box := newSandbox(t, "plan.yaml", c.before)
			loaded := box.load(t)
			c.change(t, box)
			save("stale save after "+c.name+" of "+reviewed, loaded, fixture(t, reviewed), true)
		}
	}
	for _, before := range []string{"", "identity.yaml"} {
		for name, reviewed := range map[string][]byte{
			"nil":                 nil,
			"not yaml":            fixture(t, "malformed.yaml"),
			"another account":     fixture(t, "other-account.yaml"),
			"another plan":        fixture(t, "other-plan.yaml"),
			"lower account":       fixture(t, "lower-account.yaml"),
			"lower plan":          fixture(t, "lower-plan.yaml"),
			"over the limit":      padded(t, fixture(t, "identity.yaml"), byteLimit+1),
			"unchanged":           fixture(t, "identity.yaml"),
			"reordered unchanged": fixture(t, "reordered.yaml"),
		} {
			box := newSandbox(t, "plan.yaml", before)
			save("refused or unchanged save of "+name+" before "+before, box.load(t), reviewed, true)
		}
	}
	save("nil loaded state", nil, fixture(t, "identity.yaml"), true)
	box := newSandbox(t, "plan.yaml", "")
	save("save without consent", box.load(t), fixture(t, "named.yaml"), false)
	save("save creating directories", box.load(t), fixture(t, "named.yaml"), true)
	box = newSandbox(t, "plan.yaml", "named.yaml")
	save("save replacing a file", box.load(t), fixture(t, "identity.yaml"), false)
	box = newSandbox(t, "plan.yaml", "named.yaml")
	loaded := box.load(t)
	if err := os.Chmod(box.Prep, 0o400); err != nil {
		t.Fatal(err)
	}
	save("save after a mode change", loaded, fixture(t, "identity.yaml"), false)
}

func TestRefusedGroupMismatchSaveReturnsDescriptors(t *testing.T) {
	warmUpStore(t)
	for _, c := range []struct {
		name   string
		target func(box sandbox) string
	}{
		{name: "preparation-group-differs-from-its-directory", target: func(box sandbox) string { return box.Prep }},
		{name: "directory-group-differs-from-the-preparation", target: func(box sandbox) string { return filepath.Dir(box.Prep) }},
	} {
		box := newSandbox(t, "plan.yaml", "identity.yaml")
		other := otherGroup(t, groupOf(t, box.Prep))
		if err := os.Chown(c.target(box), -1, other); err != nil {
			t.Skipf("cannot change the group to %d: %v", other, err)
		}
		loaded := box.load(t)
		returnsDescriptors(t, "group mismatch save, "+c.name, func() {
			result := newStore().Save(contract.SaveRequest{Loaded: loaded, Reviewed: fixture(t, "named.yaml")})
			wantCode(t, result.Code, contract.CodeCommit)
		})
	}
}
