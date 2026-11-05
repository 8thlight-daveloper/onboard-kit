package registry

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var monday = time.Date(2026, 11, 2, 9, 30, 0, 0, time.UTC)

func paths(entries []Entry) []string {
	var found []string
	for _, entry := range entries {
		found = append(found, entry.Path)
	}
	return found
}

func TestAMissingListHasNoPlansAndNoProblem(t *testing.T) {
	entries, err := Load(filepath.Join(t.TempDir(), "not-made-yet"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("got %v, %v", entries, err)
	}
}

func TestRecordingCreatesThePrivateHomeAndWritesTheDocumentedFormat(t *testing.T) {
	home := filepath.Join(t.TempDir(), "state", "obk")
	if err := Record(home, "/work/acme/acme.obk.yaml", monday.Add(1500*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(Path(home))
	if err != nil {
		t.Fatal(err)
	}
	if want := "onboard-kit plans 1\n2026-11-02T09:30:01Z\t/work/acme/acme.obk.yaml\n"; string(data) != want {
		t.Fatalf("got %q, want %q", data, want)
	}
	info, err := os.Stat(home)
	if err != nil || info.Mode().Perm() != 0o700 {
		t.Fatalf("home mode %v, %v", info.Mode().Perm(), err)
	}
}

func TestTheLastOpenedPlanComesFirstAndAppearsOnce(t *testing.T) {
	home := t.TempDir()
	for index, path := range []string{"/a.obk.yaml", "/b.obk.yaml", "/a.obk.yaml", "/c.obk.yaml"} {
		if err := Record(home, path, monday.Add(time.Duration(index)*time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := Load(home)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(paths(entries), " "); got != "/c.obk.yaml /a.obk.yaml /b.obk.yaml" {
		t.Fatalf("got %s", got)
	}
	if !entries[1].Opened.Equal(monday.Add(2 * time.Hour)) {
		t.Fatalf("the second open of a was not kept: %v", entries[1].Opened)
	}
}

func TestForgettingRemovesOnlyTheListLineAndNeverTheFile(t *testing.T) {
	home := t.TempDir()
	plan := filepath.Join(t.TempDir(), "acme.obk.yaml")
	if err := os.WriteFile(plan, []byte("schemaVersion: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{plan, "/other.obk.yaml"} {
		if err := Record(home, path, monday); err != nil {
			t.Fatal(err)
		}
	}
	if err := Forget(home, plan); err != nil {
		t.Fatal(err)
	}
	entries, _ := Load(home)
	if got := paths(entries); len(got) != 1 || got[0] != "/other.obk.yaml" {
		t.Fatalf("got %v", got)
	}
	if _, err := os.Stat(plan); err != nil {
		t.Fatalf("the plan file was touched: %v", err)
	}
	if err := Forget(home, "/never-listed.obk.yaml"); err != nil {
		t.Fatalf("forgetting an unlisted plan: %v", err)
	}
}

func TestTheListKeepsTheNewestHundredPlans(t *testing.T) {
	home := t.TempDir()
	for index := range MaximumEntries + 5 {
		if err := Record(home, fmt.Sprintf("/plans/%03d.obk.yaml", index), monday.Add(time.Duration(index)*time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := Load(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != MaximumEntries || entries[0].Path != "/plans/104.obk.yaml" || entries[len(entries)-1].Path != "/plans/005.obk.yaml" {
		t.Fatalf("got %d, first %s, last %s", len(entries), entries[0].Path, entries[len(entries)-1].Path)
	}
}

func TestAPathThatIsNotFullPlainTextIsNotRecorded(t *testing.T) {
	home := t.TempDir()
	for _, path := range []string{"", "relative.obk.yaml", "/with\nnewline.obk.yaml", "/with\ttab.obk.yaml", "/" + strings.Repeat("x", MaximumPath), "/bad\xff.obk.yaml"} {
		if err := Record(home, path, monday); !errors.Is(err, ErrUnfitPath) {
			t.Errorf("%q: got %v", path, err)
		}
	}
	if _, err := os.Stat(Path(home)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("a list was written: %v", err)
	}
}

func TestADamagedListIsReportedAndLeftAsItIs(t *testing.T) {
	cases := map[string]string{
		"not a list\n": "its first line is not onboard-kit plans 1",
		"onboard-kit plans 1\nyesterday\t/a.obk.yaml\n":                             "line 2 is not a time, a tab, and a full path",
		"onboard-kit plans 1\n2026-11-02T09:30:00Z\n":                               "line 2 is not a time, a tab, and a full path",
		"onboard-kit plans 1\n2026-11-02T09:30:00Z\ta.obk":                          "line 2 is not a time, a tab, and a full path",
		"onboard-kit plans 1\n2026-11-02T09:30:00Z\t/a\n2026-11-02T09:31:00Z\t/a\n": "line 3 names a plan a second time",
		"onboard-kit plans 1\n\xff\n":                                               "it is not UTF-8 text",
	}
	for content, reason := range cases {
		home := t.TempDir()
		if err := os.WriteFile(Path(home), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := Load(home)
		var damage *Damage
		if !errors.As(err, &damage) || damage.Reason != reason || damage.File != Path(home) {
			t.Errorf("%q: got %v, want %q", content, err, reason)
		}
		if err := Record(home, "/new.obk.yaml", monday); !errors.As(err, &damage) {
			t.Errorf("%q: recording over a damaged list gave %v", content, err)
		}
		if err := Forget(home, "/a"); !errors.As(err, &damage) {
			t.Errorf("%q: forgetting in a damaged list gave %v", content, err)
		}
		if data, _ := os.ReadFile(Path(home)); string(data) != content {
			t.Errorf("%q: the damaged list was changed to %q", content, data)
		}
	}
}

func TestAListLargerThanTheBoundIsDamagedAndTooManyPlansAreToo(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(Path(home), []byte(FirstLine+"\n"+strings.Repeat(" ", MaximumBytes)), 0o600); err != nil {
		t.Fatal(err)
	}
	var damage *Damage
	if _, err := Load(home); !errors.As(err, &damage) || damage.Reason != "it is larger than 256 KiB" {
		t.Fatalf("got %v", err)
	}
	var many strings.Builder
	many.WriteString(FirstLine + "\n")
	for index := range MaximumEntries + 1 {
		fmt.Fprintf(&many, "2026-11-02T09:30:00Z\t/plans/%d.obk.yaml\n", index)
	}
	if _, reason := Parse([]byte(many.String())); reason != "it holds more than 100 plans" {
		t.Fatalf("got %q", reason)
	}
}

func TestBlankLinesAndAMissingLastNewlineAreRead(t *testing.T) {
	entries, reason := Parse([]byte("onboard-kit plans 1\n\n2026-11-02T09:30:00Z\t/a.obk.yaml\n2026-11-03T09:30:00+02:00\t/b.obk.yaml"))
	if reason != "" || strings.Join(paths(entries), " ") != "/b.obk.yaml /a.obk.yaml" {
		t.Fatalf("got %v, %q", paths(entries), reason)
	}
	if entries[0].Opened.Location() != time.UTC {
		t.Fatalf("times are not held in UTC: %v", entries[0].Opened)
	}
}

func TestReplacingLeavesNoTemporaryFileBehind(t *testing.T) {
	home := t.TempDir()
	if err := Record(home, "/a.obk.yaml", monday); err != nil {
		t.Fatal(err)
	}
	names, err := os.ReadDir(home)
	if err != nil || len(names) != 1 || names[0].Name() != FileName {
		t.Fatalf("got %v, %v", names, err)
	}
}
