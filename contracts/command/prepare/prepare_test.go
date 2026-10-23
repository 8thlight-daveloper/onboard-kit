package prepare_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"

	contract "onboard-kit/contracts/command/prepare"
)

type workspace struct {
	root string
	plan string
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func put(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func setup(t *testing.T, preparation bool) workspace {
	t.Helper()
	w := workspace{root: t.TempDir(), plan: filepath.Join(t.TempDir(), "plan.yaml")}
	put(t, w.plan, fixture(t, "plan.yaml"))
	if preparation {
		put(t, w.preparation(), fixture(t, "preparation.yaml"))
	}
	return w
}

func (w workspace) preparation() string {
	return filepath.Join(w.root, "records", "newcomer-one", "preparation.obk.yaml")
}

func (w workspace) record() string {
	return filepath.Join(w.root, "records", "newcomer-one", "record.obk.yaml")
}

func (w workspace) args(operation string, more ...string) []string {
	args := []string{operation, "--plan", w.plan, "--account", "newcomer-one", "--root", w.root}
	return append(args, more...)
}

func run(t *testing.T, command contract.Command, args []string, want int) string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if got := command.Run(args, &stdout, &stderr); got != want {
		t.Fatalf("exit: got %d, want %d; stdout=%q stderr=%q", got, want, stdout.String(), stderr.String())
	}
	if want == 0 {
		if stderr.Len() != 0 {
			t.Fatalf("successful stderr: %q", stderr.String())
		}
	} else {
		if stdout.Len() != 0 || strings.TrimSpace(stderr.String()) == "" {
			t.Fatalf("failure streams: stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
		if !strings.Contains(strings.ToLower(stderr.String()), "prepar") && !strings.Contains(strings.ToLower(stderr.String()), args[0]) {
			t.Fatalf("diagnostic lacks operation context: %q", stderr.String())
		}
		for _, forbidden := range []string{"PRIVATE-SENTINEL", "Private answer sentinel", "Private unknown sentinel"} {
			if strings.Contains(stderr.String(), forbidden) {
				t.Fatalf("diagnostic leaked personal text: %q", stderr.String())
			}
		}
		for _, argument := range args {
			if filepath.IsAbs(argument) && strings.Contains(stderr.String(), argument) {
				t.Fatalf("diagnostic leaked path: %q", stderr.String())
			}
		}
	}
	return stdout.String()
}

func writeSuccess(t *testing.T, command contract.Command, args []string) {
	t.Helper()
	if got := run(t, command, args, 0); got != "records/newcomer-one/preparation.obk.yaml\n" {
		t.Fatalf("write output: %q", got)
	}
}

func object(t *testing.T, text string) map[string]any {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		t.Fatalf("invalid show JSON: %v: %q", err, text)
	}
	if !strings.HasSuffix(text, "\n") {
		t.Fatal("show lacks final newline")
	}
	return value
}

func equalJSON(t *testing.T, got string, wanted []byte) {
	t.Helper()
	actual := object(t, got)
	expected := object(t, string(wanted))
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("show mismatch\ngot: %s\nwant: %s", got, wanted)
	}
}

func read(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func unchanged(t *testing.T, path string, before []byte) {
	t.Helper()
	if after := read(t, path); !bytes.Equal(after, before) {
		t.Fatalf("changed %s\nbefore: %q\nafter: %q", path, before, after)
	}
}

func noArtifact(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("unexpected artifact %s: %v", path, err)
	}
}

func TestInitAndShow(t *testing.T) {
	t.Parallel()
	w := setup(t, false)
	command := newCommand()
	plan := read(t, w.plan)
	writeSuccess(t, command, w.args("init"))
	want := "schemaVersion: 1\nkind: preparation\naccount: newcomer-one\nplanId: plan-one\n"
	if actual := string(read(t, w.preparation())); actual != want {
		t.Fatalf("init bytes: %q, want %q", actual, want)
	}
	for _, entry := range []struct {
		path string
		mode fs.FileMode
	}{
		{w.preparation(), 0o600},
		{filepath.Dir(w.preparation()), 0o700},
		{filepath.Join(w.root, "records"), 0o700},
	} {
		info, err := os.Stat(entry.path)
		if err != nil || info.Mode().Perm() != entry.mode {
			t.Fatalf("mode for %s: %v, %v", entry.path, info, err)
		}
	}
	before := read(t, w.preparation())
	first := run(t, command, w.args("show"), 0)
	equalJSON(t, first, fixture(t, "empty-show.json"))
	if second := run(t, command, w.args("show"), 0); second != first {
		t.Fatal("show is not deterministic")
	}
	run(t, command, w.args("init"), 1)
	unchanged(t, w.preparation(), before)
	unchanged(t, w.plan, plan)
	noArtifact(t, w.record())
	entries, err := os.ReadDir(filepath.Dir(w.preparation()))
	if err != nil || len(entries) != 1 {
		t.Fatalf("publication leftovers: %v, %v", entries, err)
	}
}

func TestShowExactProgressProjection(t *testing.T) {
	t.Parallel()
	w := setup(t, true)
	put(t, w.record(), fixture(t, "record.yaml"))
	preparation, record, plan := read(t, w.preparation()), read(t, w.record()), read(t, w.plan)
	equalJSON(t, run(t, newCommand(), w.args("show"), 0), fixture(t, "full-show.json"))
	unchanged(t, w.preparation(), preparation)
	unchanged(t, w.record(), record)
	unchanged(t, w.plan, plan)
}

func TestRecipientOverlayAndExplicitClearing(t *testing.T) {
	t.Parallel()
	w := setup(t, true)
	command := newCommand()
	writeSuccess(t, command, w.args("recipient", "--name=Changed", "--company-email", "", "--planned-start-date=", "--team", "team-one"))
	view := object(t, run(t, command, w.args("show"), 0))
	want := map[string]any{
		"name": "Changed", "companyEmail": "", "plannedStartDate": "", "team": "team-one", "manager": "lead-one",
		"onboardingOwner": "owner-one", "buddy": "buddy-one", "workLocation": "remote", "timeZone": "Etc/UTC",
	}
	if !reflect.DeepEqual(view["recipient"], want) {
		t.Fatalf("recipient: %#v", view["recipient"])
	}
	data := string(read(t, w.preparation()))
	for _, retained := range []string{"x-retained:", "empty: []", "purpose: Preserve this work", "note: Example support context"} {
		if !strings.Contains(data, retained) {
			t.Fatalf("lost unrelated data %q", retained)
		}
	}
	writeSuccess(t, command, w.args("recipient", "--name", "Last", "--manager=", "--onboarding-owner=", "--buddy=", "--work-location=", "--time-zone="))
	view = object(t, run(t, command, w.args("show"), 0))
	want["name"] = "Last"
	for _, field := range []string{"manager", "onboardingOwner", "buddy", "workLocation", "timeZone"} {
		want[field] = ""
	}
	if !reflect.DeepEqual(view["recipient"], want) {
		t.Fatalf("clear recipient: %#v", view["recipient"])
	}
}

func TestArrangementOverlayAppendAndRemoval(t *testing.T) {
	t.Parallel()
	w := setup(t, true)
	command := newCommand()
	writeSuccess(t, command, w.args("arrangement", "--id", "first-pairing", "--contact", "lead-one", "--contact", "buddy-one", "--contact", "lead-one", "--note="))
	view := object(t, run(t, command, w.args("show"), 0))
	arrangements := view["arrangements"].([]any)
	first := arrangements[0].(map[string]any)
	want := map[string]any{"id": "first-pairing", "section": "ticket-pairing", "kind": "pairing", "owner": "owner-one", "contacts": []any{"lead-one", "buddy-one", "lead-one"}, "reference": "Example ticket", "targetDate": "2026-11-05", "status": "arranged", "note": ""}
	if !reflect.DeepEqual(first, want) {
		t.Fatalf("overlay: %#v", first)
	}
	writeSuccess(t, command, w.args("arrangement", "--id", "second", "--section", "review", "--kind", "support", "--status", "pending"))
	writeSuccess(t, command, w.args("arrangement", "--id", "first-pairing", "--status", "pending", "--clear-contacts=true", "--owner=", "--reference=", "--target-date="))
	writeSuccess(t, command, w.args("arrangement", "--id", "first-pairing", "--clear-contacts"))
	view = object(t, run(t, command, w.args("show"), 0))
	arrangements = view["arrangements"].([]any)
	want["contacts"], want["status"], want["owner"], want["reference"], want["targetDate"] = []any{}, "pending", "", "", ""
	if len(arrangements) != 2 || !reflect.DeepEqual(arrangements[0], want) || arrangements[1].(map[string]any)["id"] != "second" {
		t.Fatalf("append/clear/order: %#v", arrangements)
	}
	writeSuccess(t, command, w.args("remove", "--id", "first-pairing"))
	writeSuccess(t, command, w.args("remove", "--id", "second"))
	view = object(t, run(t, command, w.args("show"), 0))
	if !reflect.DeepEqual(view["arrangements"], []any{}) {
		t.Fatalf("final removal: %#v", view["arrangements"])
	}
}

func TestAllRecipientFlagsAreWired(t *testing.T) {
	t.Parallel()
	w := setup(t, false)
	command := newCommand()
	writeSuccess(t, command, w.args("init"))
	writeSuccess(t, command, w.args("recipient", "--name", "Example newcomer", "--company-email", "example@company.example", "--planned-start-date", "2026-11-02", "--team", "team-one", "--manager", "lead-one", "--onboarding-owner", "owner-one", "--buddy", "buddy-one", "--work-location", "remote", "--time-zone", "Etc/UTC"))
	view := object(t, run(t, command, w.args("show"), 0))
	wanted := object(t, string(fixture(t, "full-show.json")))
	if !reflect.DeepEqual(view["recipient"], wanted["recipient"]) {
		t.Fatalf("recipient flags: %#v", view["recipient"])
	}
}

func TestUsageHasNoEffects(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{}, {"unknown"}, {"init", "--bogus", "PRIVATE-SENTINEL"}, {"init", "extra"}, {"show", "--name", "PRIVATE-SENTINEL"},
		{"init", "--id", "a"}, {"recipient"}, {"recipient", "--id", "a"}, {"recipient", "--name"},
		{"arrangement", "--id", "a"}, {"arrangement", "--status", "pending"},
		{"arrangement", "--id", "a", "--contact", "", "--clear-contacts"},
		{"remove"}, {"remove", "--id", "a", "--note", "PRIVATE-SENTINEL"},
		{"recipient", "--name", "First", "--name", "Last"},
		{"arrangement", "--id", "a", "--clear-contacts", "--clear-contacts"},
		{"show", "--account", "newcomer-one"},
		{"arrangement", "--id", "a", "--clear-contacts=false"},
		{"arrangement", "--id", "a", "--clear-contacts=no"},
	}
	for _, more := range cases {
		t.Run(strings.Join(more, " "), func(t *testing.T) {
			t.Parallel()
			w := setup(t, false)
			args := append(append([]string{}, more...), "--plan", w.plan, "--account", "newcomer-one", "--root", w.root)
			run(t, newCommand(), args, 2)
			noArtifact(t, filepath.Join(w.root, "records"))
		})
	}
	for _, args := range [][]string{{"init"}, {"init", "--plan", ""}, {"init", "--plan", "unused", "--account", ""}} {
		run(t, newCommand(), args, 2)
	}
}

func TestDataErrorsPreservePreparation(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{"recipient", "--team", "PRIVATE-SENTINEL"}, {"recipient", "--planned-start-date", "2027-02-29"},
		{"recipient", "--manager", "PRIVATE-SENTINEL"}, {"recipient", "--work-location", "PRIVATE-SENTINEL"},
		{"arrangement", "--id", "new", "--section", "ticket-pairing", "--kind", "pairing"},
		{"arrangement", "--id", "new", "--section", "ticket-pairing", "--status", "pending"},
		{"arrangement", "--id", "first-pairing", "--contact", "PRIVATE-SENTINEL"},
		{"arrangement", "--id", "first-pairing", "--status", "verified"},
		{"remove", "--id", "absent"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()
			w := setup(t, true)
			before := read(t, w.preparation())
			run(t, newCommand(), w.args(args[0], args[1:]...), 1)
			unchanged(t, w.preparation(), before)
		})
	}
}

func TestRecordIsReadOnlyAndIrrelevantToEdits(t *testing.T) {
	t.Parallel()
	for _, record := range []string{string(fixture(t, "record.yaml")), "PRIVATE-SENTINEL: [\n", strings.ReplaceAll(string(fixture(t, "record.yaml")), "account: newcomer-one", "account: another-person"), strings.ReplaceAll(string(fixture(t, "record.yaml")), "id: plan-one", "id: another-plan")} {
		t.Run(record, func(t *testing.T) {
			t.Parallel()
			w := setup(t, true)
			put(t, w.record(), []byte(record))
			before := read(t, w.record())
			writeSuccess(t, newCommand(), w.args("recipient", "--name", "Changed"))
			unchanged(t, w.record(), before)
			if record != string(fixture(t, "record.yaml")) {
				run(t, newCommand(), w.args("show"), 1)
			}
		})
	}
}

func TestIdentityGuards(t *testing.T) {
	t.Parallel()
	for _, change := range [][2]string{{"account: newcomer-one", "account: another-person"}, {"planId: plan-one", "planId: another-plan"}, {"kind: preparation", "kind: record"}, {"schemaVersion: 1", "schemaVersion: 2"}} {
		for _, operation := range []string{"show", "recipient", "arrangement", "remove"} {
			t.Run(operation+change[1], func(t *testing.T) {
				t.Parallel()
				w := setup(t, true)
				before := bytes.Replace(fixture(t, "preparation.yaml"), []byte(change[0]), []byte(change[1]), 1)
				put(t, w.preparation(), before)
				more := map[string][]string{"recipient": {"--name", "Changed"}, "arrangement": {"--id", "first-pairing", "--note", "Changed"}, "remove": {"--id", "first-pairing"}}[operation]
				run(t, newCommand(), w.args(operation, more...), 1)
				unchanged(t, w.preparation(), before)
			})
		}
	}
}

func TestScopedRepairAndDuplicateVisibility(t *testing.T) {
	t.Parallel()
	w := setup(t, true)
	data := "schemaVersion: 1\nkind: preparation\naccount: newcomer-one\nplanId: plan-one\narrangements:\n  - id: first\n    section: ticket-pairing\n    kind: support\n    status: pending\n    owner: stale-first\n  - id: second\n    section: review\n    kind: support\n    status: pending\n    owner: stale-second\n"
	put(t, w.preparation(), []byte(data))
	command := newCommand()
	writeSuccess(t, command, w.args("arrangement", "--id", "first", "--owner", "owner-one"))
	view := object(t, run(t, command, w.args("show"), 0))
	problems := view["preparationProblems"].([]any)
	if len(problems) != 1 {
		t.Fatalf("repair problems: %#v", problems)
	}
	problem := problems[0].(map[string]any)
	if len(problem) != 4 || problem["code"] != "preparation.unknown-contact" || problem["path"] != "arrangements[1].owner" || problem["severity"] != "error" || strings.TrimSpace(problem["message"].(string)) == "" {
		t.Fatalf("repair diagnostic: %#v", problem)
	}
	if view["arrangements"].([]any)[1].(map[string]any)["owner"] != "stale-second" {
		t.Fatal("repair discarded unrelated stale reference")
	}
	duplicate := strings.ReplaceAll(data, "id: second", "id: first")
	put(t, w.preparation(), []byte(duplicate))
	view = object(t, run(t, command, w.args("show"), 0))
	if len(view["arrangements"].([]any)) != 2 {
		t.Fatal("show chose a duplicate winner")
	}
	foundDuplicate := false
	for _, entry := range view["preparationProblems"].([]any) {
		problem := entry.(map[string]any)
		if problem["code"] == "preparation.duplicate-id" && problem["path"] == "arrangements[1].id" {
			foundDuplicate = true
		}
	}
	if !foundDuplicate {
		t.Fatal("show concealed duplicate identity diagnostic")
	}
	for _, args := range [][]string{{"recipient", "--name", "Changed"}, {"arrangement", "--id", "first", "--owner", "owner-one"}, {"remove", "--id", "first"}} {
		run(t, command, w.args(args[0], args[1:]...), 1)
		unchanged(t, w.preparation(), []byte(duplicate))
	}
}

func TestExplicitEmptyIDCanRepairMalformedItem(t *testing.T) {
	t.Parallel()
	w := setup(t, true)
	data := "schemaVersion: 1\nkind: preparation\naccount: newcomer-one\nplanId: plan-one\narrangements:\n  - id: \"\"\n    owner: stale\n"
	put(t, w.preparation(), []byte(data))
	writeSuccess(t, newCommand(), w.args("remove", "--id", ""))
	view := object(t, run(t, newCommand(), w.args("show"), 0))
	if !reflect.DeepEqual(view["arrangements"], []any{}) {
		t.Fatalf("empty removal: %#v", view)
	}
	put(t, w.preparation(), []byte(data+"  - id: \"\"\n"))
	before := read(t, w.preparation())
	run(t, newCommand(), w.args("remove", "--id="), 1)
	unchanged(t, w.preparation(), before)
}

func TestPlanValidationAndLocationBeforeEffects(t *testing.T) {
	t.Parallel()
	base := string(fixture(t, "plan.yaml"))
	cases := []string{
		strings.ReplaceAll(base, "id: plan-one\n", ""),
		"schemaVersion: 1\nid: plan-one\nPRIVATE-SENTINEL: [\n",
		base + "resultsPath: records/{account}.yaml\n",
		base + "resultsPath: ../{account}/record.yaml\n",
		base + "resultsPath: records/{account}/PREPARATION.OBK.YAML\n",
		strings.ReplaceAll(base, "name: organization", "name: organization\n    type: choice\n    options: [one, two]\n  - name: organization"),
		strings.ReplaceAll(base, "name: Example team", "name: Example team\n      contact: PRIVATE-SENTINEL"),
	}
	for _, plan := range cases {
		t.Run(plan, func(t *testing.T) {
			t.Parallel()
			w := setup(t, false)
			put(t, w.plan, []byte(plan))
			run(t, newCommand(), w.args("init"), 1)
			noArtifact(t, filepath.Join(w.root, "records"))
			unchanged(t, w.plan, []byte(plan))
		})
	}
}

func TestCustomPathsAndTwoPeople(t *testing.T) {
	t.Parallel()
	w := setup(t, false)
	put(t, w.plan, append(fixture(t, "plan.yaml"), []byte("resultsPath: Cohort/{account}/progress.yaml\n")...))
	command := newCommand()
	for _, account := range []string{"newcomer-one", "newcomer-two"} {
		args := []string{"init", "--plan", w.plan, "--account", account, "--root", w.root}
		if got := run(t, command, args, 0); got != "Cohort/"+account+"/preparation.obk.yaml\n" {
			t.Fatalf("custom path: %q", got)
		}
	}
	second := filepath.Join(w.root, "Cohort", "newcomer-two", "preparation.obk.yaml")
	before := read(t, second)
	run(t, command, w.args("recipient", "--buddy", "buddy-one"), 0)
	unchanged(t, second, before)
	view := object(t, run(t, command, w.args("show"), 0))
	if view["recordPath"] != "Cohort/newcomer-one/progress.yaml" || view["preparationPath"] != "Cohort/newcomer-one/preparation.obk.yaml" {
		t.Fatalf("custom show paths: %#v", view)
	}
	noArtifact(t, filepath.Join(w.root, "records"))
}

func TestSymlinkBoundaries(t *testing.T) {
	t.Parallel()
	for _, target := range []string{"root", "root-trailing-slash", "plan", "records", "person", "preparation", "record"} {
		t.Run(target, func(t *testing.T) {
			t.Parallel()
			w := setup(t, false)
			outside := t.TempDir()
			sentinel := filepath.Join(outside, "sentinel")
			put(t, sentinel, []byte("PRIVATE-SENTINEL"))
			link := ""
			destination := outside
			switch target {
			case "root", "root-trailing-slash":
				link = filepath.Join(t.TempDir(), "root-link")
				w.root = link
				if target == "root-trailing-slash" {
					w.root += string(os.PathSeparator)
				}
			case "plan":
				link = filepath.Join(t.TempDir(), "plan-link")
				destination = w.plan
				w.plan = link
			case "records":
				link = filepath.Join(w.root, "records")
			case "person":
				if err := os.Mkdir(filepath.Join(w.root, "records"), 0o700); err != nil {
					t.Fatal(err)
				}
				link = filepath.Dir(w.preparation())
			case "preparation", "record":
				if err := os.MkdirAll(filepath.Dir(w.preparation()), 0o700); err != nil {
					t.Fatal(err)
				}
				link = w.preparation()
				if target == "record" {
					link = w.record()
					put(t, w.preparation(), fixture(t, "preparation.yaml"))
				}
				destination = sentinel
			}
			if err := os.Symlink(destination, link); err != nil {
				t.Fatal(err)
			}
			operation := "init"
			if target == "record" {
				operation = "show"
			}
			run(t, newCommand(), w.args(operation), 1)
			unchanged(t, sentinel, []byte("PRIVATE-SENTINEL"))
			if target == "record" {
				writeSuccess(t, newCommand(), w.args("recipient", "--name", "Changed"))
				info, err := os.Lstat(link)
				if err != nil || info.Mode()&os.ModeSymlink == 0 {
					t.Fatalf("edit touched record symlink: %v, %v", info, err)
				}
			}
			entries, err := os.ReadDir(outside)
			if err != nil || len(entries) != 1 {
				t.Fatalf("escaped root: %v, %v", entries, err)
			}
		})
	}
}

func TestNonregularArtifacts(t *testing.T) {
	t.Parallel()
	for _, target := range []string{"plan", "preparation", "record"} {
		for _, kind := range []string{"directory", "fifo"} {
			t.Run(target+kind, func(t *testing.T) {
				t.Parallel()
				w := setup(t, target == "record")
				path := w.preparation()
				if target == "record" {
					path = w.record()
				} else if target == "plan" {
					path = filepath.Join(t.TempDir(), "special-plan")
					w.plan = path
				}
				if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					t.Fatal(err)
				}
				if kind == "directory" {
					if err := os.Mkdir(path, 0o700); err != nil {
						t.Fatal(err)
					}
				} else if err := syscall.Mkfifo(path, 0o600); err != nil {
					t.Fatal(err)
				}
				operation := "init"
				if target != "plan" {
					operation = "show"
				}
				run(t, newCommand(), w.args(operation), 1)
			})
		}
	}
	for _, root := range []string{filepath.Join(t.TempDir(), "missing"), filepath.Join(t.TempDir(), "regular")} {
		w := setup(t, false)
		if strings.HasSuffix(root, "regular") {
			put(t, root, []byte("PRIVATE-SENTINEL"))
		}
		w.root = root
		run(t, newCommand(), w.args("init"), 1)
	}
}

func TestHardLinkedPreparationIsReplacedWithoutTruncation(t *testing.T) {
	t.Parallel()
	w := setup(t, true)
	other := filepath.Join(t.TempDir(), "other-preparation.yaml")
	if err := os.Link(w.preparation(), other); err != nil {
		t.Fatal(err)
	}
	before := read(t, other)
	writeSuccess(t, newCommand(), w.args("recipient", "--name", "Changed"))
	unchanged(t, other, before)
	if bytes.Equal(read(t, w.preparation()), before) {
		t.Fatal("selected pathname was not changed")
	}
	entries, err := os.ReadDir(filepath.Dir(w.preparation()))
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary files remain: %v, %v", entries, err)
	}
}

func TestInputByteBudgets(t *testing.T) {
	t.Parallel()
	for _, target := range []string{"plan", "preparation", "record"} {
		for _, extra := range []int{0, 1} {
			t.Run(target+string(rune('0'+extra)), func(t *testing.T) {
				t.Parallel()
				w := setup(t, target != "plan")
				path, base := w.plan, fixture(t, "plan.yaml")
				if target == "preparation" {
					path, base = w.preparation(), fixture(t, "preparation.yaml")
				} else if target == "record" {
					path, base = w.record(), fixture(t, "record.yaml")
				}
				data := append(append([]byte{}, base...), '#')
				data = append(data, bytes.Repeat([]byte("x"), (1<<20)+extra-len(data))...)
				put(t, path, data)
				operation := "show"
				if target == "plan" {
					operation = "init"
				}
				want := 0
				if extra != 0 {
					want = 1
				}
				output := run(t, newCommand(), w.args(operation), want)
				if want == 0 {
					if operation == "init" {
						if output != "records/newcomer-one/preparation.obk.yaml\n" {
							t.Fatalf("budget-boundary init output: %q", output)
						}
						identity := "schemaVersion: 1\nkind: preparation\naccount: newcomer-one\nplanId: plan-one\n"
						if actual := string(read(t, w.preparation())); actual != identity {
							t.Fatalf("budget-boundary init bytes: %q", actual)
						}
					} else {
						view := object(t, output)
						if view["account"] != "newcomer-one" || view["planId"] != "plan-one" {
							t.Fatalf("budget-boundary show identity: %#v", view)
						}
						record, valid := view["record"].(map[string]any)
						if !valid || record["present"] != (target == "record") {
							t.Fatalf("budget-boundary record presence: %#v", view["record"])
						}
					}
				}
				unchanged(t, path, data)
			})
		}
	}
}

func TestOversizedEncodedEditPreservesBytes(t *testing.T) {
	t.Parallel()
	w := setup(t, true)
	before := read(t, w.preparation())
	run(t, newCommand(), w.args("recipient", "--name", strings.Repeat("x", 1<<20)), 1)
	unchanged(t, w.preparation(), before)
}

func TestConcurrentInitPublishesOnce(t *testing.T) {
	t.Parallel()
	w := setup(t, false)
	start := make(chan struct{})
	results := make(chan int, 2)
	for range 2 {
		go func() {
			<-start
			results <- newCommand().Run(w.args("init"), io.Discard, io.Discard)
		}()
	}
	close(start)
	first, second := <-results, <-results
	if !((first == 0 && second == 1) || (first == 1 && second == 0)) {
		t.Fatalf("concurrent create: %d, %d", first, second)
	}
	equalJSON(t, run(t, newCommand(), w.args("show"), 0), fixture(t, "empty-show.json"))
	entries, err := os.ReadDir(filepath.Dir(w.preparation()))
	if err != nil || len(entries) != 1 {
		t.Fatalf("concurrent publication leftovers: %v, %v", entries, err)
	}
}

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) {
	return 0, errors.New("PRIVATE-SENTINEL")
}

func TestOutputFailureRetainsPublishedEdit(t *testing.T) {
	t.Parallel()
	w := setup(t, true)
	var stderr bytes.Buffer
	if got := newCommand().Run(w.args("recipient", "--name", "Changed"), failedWriter{}, &stderr); got != 1 {
		t.Fatalf("output failure exit: %d", got)
	}
	message := strings.ToLower(stderr.String())
	if !strings.Contains(message, "saved") || !strings.Contains(message, "display") || strings.Contains(message, "private-sentinel") {
		t.Fatalf("post-publication diagnostic: %q", message)
	}
	view := object(t, run(t, newCommand(), w.args("show"), 0))
	if view["recipient"].(map[string]any)["name"] != "Changed" {
		t.Fatal("stdout error rolled back a published edit")
	}
	before := read(t, w.preparation())
	stderr.Reset()
	if got := newCommand().Run(w.args("show"), failedWriter{}, &stderr); got != 1 || stderr.Len() == 0 {
		t.Fatalf("show output error: %d, %q", got, stderr.String())
	}
	unchanged(t, w.preparation(), before)
}

func TestFlagLookingTextRequiresEquals(t *testing.T) {
	t.Parallel()
	w := setup(t, true)
	command := newCommand()
	before := read(t, w.preparation())
	run(t, command, w.args("recipient", "--name", "--literal"), 2)
	unchanged(t, w.preparation(), before)
	writeSuccess(t, command, w.args("recipient", "--name=--literal"))
	view := object(t, run(t, command, w.args("show"), 0))
	if view["recipient"].(map[string]any)["name"] != "--literal" {
		t.Fatal("equals spelling lost the literal flag-looking value")
	}
}

func TestHelpNeedsNoFilesystem(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"--help"}, {"init", "--help"}, {"show", "--help"}, {"recipient", "--help"}, {"arrangement", "--help"}, {"remove", "--help"}, {"--help", "--unknown", "PRIVATE-SENTINEL"}, {"show", "--help", "--account", "PRIVATE-SENTINEL"}} {
		output := run(t, newCommand(), args, 0)
		for _, word := range []string{"--plan", "--account"} {
			if !strings.Contains(output, word) {
				t.Fatalf("help omits %q: %q", word, output)
			}
		}
	}
}
