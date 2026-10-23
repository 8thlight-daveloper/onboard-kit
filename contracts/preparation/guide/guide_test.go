package guide_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type scenario struct {
	Name     string   `json:"name"`
	Source   string   `json:"source"`
	Before   string   `json:"before"`
	Input    string   `json:"input"`
	After    string   `json:"after"`
	Code     string   `json:"code"`
	Status   string   `json:"status"`
	Exit     int      `json:"exit"`
	Contains []string `json:"contains"`
}

type sandbox struct {
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

func newSandbox(t *testing.T, before string) sandbox {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	result := sandbox{Root: root, Plan: filepath.Join(root, "organization.yaml"), Prep: filepath.Join(root, "records", "newcomer", "preparation.obk.yaml")}
	writeFile(t, result.Plan, fixture(t, "plan.yaml"))
	if before != "" {
		writeFile(t, result.Prep, fixture(t, before))
	}
	return result
}

func (s sandbox) args() []string {
	return []string{"--plan", s.Plan, "--account", "newcomer", "--root", s.Root}
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
		if entry.IsDir() {
			result[relative+"/"] = "directory"
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			result[relative] = "symlink:" + target
			return err
		}
		value, err := os.ReadFile(path)
		result[relative] = string(value)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

var messages = map[string]string{
	"guide.cancelled":       "Cancelled. No files changed.",
	"guide.saved":           "Preparation saved.",
	"guide.unchanged":       "Preparation unchanged. Original bytes preserved.",
	"guide.usage":           "Use prepare guide --plan FILE --account ID [--root DIR].",
	"guide.source":          "Cannot safely load the selected plan or preparation.",
	"guide.plan":            "The selected plan is invalid.",
	"guide.identity":        "Preparation identity or arrangement IDs are invalid.",
	"guide.path":            "Cannot safely select the preparation destination.",
	"guide.input":           "Cannot read a complete supported input line.",
	"guide.output":          "Cannot display the current draft or review. No files changed.",
	"guide.encode":          "Cannot encode the reviewed preparation within the size limit.",
	"guide.stale-review":    "Source files changed. Reload and review before saving.",
	"guide.commit":          "Preparation was not saved. A commit attempt may have left empty preparation directories.",
	"guide.cleanup-refused": "Preparation was not saved. Owned temporary cleanup needs repair; empty preparation directories may remain.",
	"guide.cleanup-saved":   "Preparation saved. Owned temporary cleanup needs repair.",
	"guide.report-saved":    "Preparation saved, but its completion report could not be written.",
}

func assertReport(t *testing.T, value []byte, status, code string) {
	t.Helper()
	var actual map[string]string
	if err := json.Unmarshal(value, &actual); err != nil {
		t.Fatalf("report is not one JSON object: %q: %v", value, err)
	}
	want := map[string]string{"status": status, "code": code, "message": messages[code]}
	if !reflect.DeepEqual(actual, want) || !bytes.HasSuffix(value, []byte("\n")) {
		t.Fatalf("report = %q; want %#v with final LF", value, want)
	}
}

func TestRecipientScenarios(t *testing.T) {
	var cases []scenario
	if err := json.Unmarshal(fixture(t, "cases.json"), &cases); err != nil {
		t.Fatal(err)
	}
	for _, current := range cases {
		t.Run(current.Name, func(t *testing.T) {
			t.Parallel()
			if current.Source == "" {
				t.Fatal("fixture lacks source")
			}
			box := newSandbox(t, current.Before)
			before := snapshot(t, box.Root)
			var stdout, stderr bytes.Buffer
			got := newCommand().Run(box.args(), strings.NewReader(current.Input), &stdout, &stderr)
			if got != current.Exit {
				t.Fatalf("exit = %d; want %d; stderr %q", got, current.Exit, stderr.String())
			}
			assertReport(t, stderr.Bytes(), current.Status, current.Code)
			for _, part := range current.Contains {
				if !strings.Contains(stdout.String(), part) {
					t.Errorf("stdout missing %q: %q", part, stdout.String())
				}
			}
			want := before
			if current.After != "" {
				want["records/"] = "directory"
				want["records/newcomer/"] = "directory"
				want["records/newcomer/preparation.obk.yaml"] = string(fixture(t, current.After))
			}
			if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, want) {
				t.Fatalf("filesystem = %#v; want %#v", actual, want)
			}
		})
	}
}

type monitoredReader struct {
	Value     string
	ReadCount int
	Failure   error
}

func (r *monitoredReader) Read(target []byte) (int, error) {
	if len(target) == 0 {
		return 0, nil
	}
	if r.ReadCount == len(r.Value) {
		if r.Failure != nil {
			return 0, r.Failure
		}
		return 0, io.EOF
	}
	target[0] = r.Value[r.ReadCount]
	r.ReadCount++
	return 1, nil
}

func TestHelpAndArguments(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"--help", "--bad"}, {"-h", "--plan", "missing"}} {
		reader := &monitoredReader{Value: "PRIVATE"}
		var stdout, stderr bytes.Buffer
		if exit := newCommand().Run(args, reader, &stdout, &stderr); exit != 0 || !bytes.Equal(stdout.Bytes(), fixture(t, "help.txt")) || stderr.Len() != 0 || reader.ReadCount != 0 {
			t.Fatalf("help args %q: exit %d output %q/%q reads %d", args, exit, stdout.String(), stderr.String(), reader.ReadCount)
		}
	}
	for _, args := range [][]string{nil, {"--plan"}, {"--account", "newcomer"}, {"--bad"}, {"--plan=one", "--plan", "two", "--account=newcomer"}, {"--plan=one", "--account="}, {"--plan", "--account", "newcomer"}, {"--plan=one", "--account=newcomer", "--help"}, {"--plan=one", "--account=newcomer", "--root=one", "--root=two"}, {"--plan=one", "--account=newcomer", "extra"}} {
		reader := &monitoredReader{Value: "PRIVATE"}
		var stdout, stderr bytes.Buffer
		if exit := newCommand().Run(args, reader, &stdout, &stderr); exit != 2 || reader.ReadCount != 0 || stdout.Len() != 0 {
			t.Fatalf("usage args %q: exit %d stdout %q reads %d", args, exit, stdout.String(), reader.ReadCount)
		}
		assertReport(t, stderr.Bytes(), "refused", "guide.usage")
	}
	box := newSandbox(t, "")
	var stdout, stderr bytes.Buffer
	if exit := newCommand().Run([]string{"--root=" + box.Root, "--account=newcomer", "--plan=" + box.Plan}, strings.NewReader("cancel\n"), &stdout, &stderr); exit != 0 {
		t.Fatalf("equals args exit %d: %q", exit, stderr.String())
	}
	assertReport(t, stderr.Bytes(), "cancelled", "guide.cancelled")
}

func TestInputBoundsAndFailures(t *testing.T) {
	t.Parallel()
	for _, current := range []struct {
		Name, Input string
		Failure     error
	}{
		{Name: "invalid-utf8", Input: "\xff\n"},
		{Name: "one-over-limit", Input: strings.Repeat("x", 65_537) + "\n"},
		{Name: "multibyte-over-byte-limit", Input: strings.Repeat("Ü", 32_769) + "\n"},
		{Name: "read-error", Failure: errors.New("PRIVATE input error")},
	} {
		t.Run(current.Name, func(t *testing.T) {
			box := newSandbox(t, "")
			before := snapshot(t, box.Root)
			reader := &monitoredReader{Value: current.Input, Failure: current.Failure}
			var stdout, stderr bytes.Buffer
			if exit := newCommand().Run(box.args(), reader, &stdout, &stderr); exit != 1 {
				t.Fatalf("exit %d", exit)
			}
			assertReport(t, stderr.Bytes(), "refused", "guide.input")
			if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
				t.Fatal("input failure changed filesystem")
			}
		})
	}
	for _, framing := range []string{"\n", "\r\n"} {
		for _, value := range []string{strings.Repeat("x", 65_536), strings.Repeat("Ü", 32_768)} {
			box := newSandbox(t, "")
			var stdout, stderr bytes.Buffer
			input := value + framing + "cancel" + framing
			if exit := newCommand().Run(box.args(), strings.NewReader(input), &stdout, &stderr); exit != 0 {
				t.Fatalf("exact boundary %q exit %d report %q", framing, exit, stderr.String())
			}
			assertReport(t, stderr.Bytes(), "cancelled", "guide.cancelled")
		}
	}
}

type reviewWriter struct {
	Value  bytes.Buffer
	Action func()
	Mode   string
	Done   bool
}

func (w *reviewWriter) Write(value []byte) (int, error) {
	joined := w.Value.String() + string(value)
	if !w.Done && strings.Contains(joined, "Review — not saved.") {
		w.Done = true
		if w.Action != nil {
			w.Action()
		}
		switch w.Mode {
		case "short":
			return 0, nil
		case "error":
			return 0, errors.New("PRIVATE output error")
		case "full-error":
			_, _ = w.Value.Write(value)
			return len(value), errors.New("PRIVATE output error")
		}
	}
	return w.Value.Write(value)
}

func TestReviewOutputFailurePrecedesConsent(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"short", "error", "full-error"} {
		box := newSandbox(t, "")
		before := snapshot(t, box.Root)
		reader := &monitoredReader{Value: "review\nsave\n"}
		writer := &reviewWriter{Mode: mode}
		var stderr bytes.Buffer
		if exit := newCommand().Run(box.args(), reader, writer, &stderr); exit != 1 {
			t.Fatalf("%s exit %d", mode, exit)
		}
		assertReport(t, stderr.Bytes(), "refused", "guide.output")
		if reader.ReadCount != len("review\n") {
			t.Fatalf("%s consumed consent: %d bytes", mode, reader.ReadCount)
		}
		if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
			t.Fatalf("%s changed filesystem", mode)
		}
	}
}

func TestStaleReviewAndAppearingDestination(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"plan", "existing-preparation", "appeared"} {
		beforeName := ""
		if mode == "existing-preparation" {
			beforeName = "identity.yaml"
		}
		box := newSandbox(t, beforeName)
		writer := &reviewWriter{Action: func() {
			path := box.Plan
			if mode != "plan" {
				path = box.Prep
			}
			writeFile(t, path, []byte("EXTERNAL concurrent bytes\n"))
		}}
		var stderr bytes.Buffer
		if exit := newCommand().Run(box.args(), strings.NewReader("review\nsave\n"), writer, &stderr); exit != 1 {
			t.Fatalf("%s exit %d", mode, exit)
		}
		assertReport(t, stderr.Bytes(), "refused", "guide.stale-review")
		path := box.Plan
		if mode != "plan" {
			path = box.Prep
		}
		if got, err := os.ReadFile(path); err != nil || string(got) != "EXTERNAL concurrent bytes\n" {
			t.Fatalf("%s overwrote external bytes: %q %v", mode, got, err)
		}
	}
}

func TestLiteralDisplayAndNoopInode(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "literal.yaml")
	before, err := os.Stat(box.Prep)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if exit := newCommand().Run(box.args(), strings.NewReader("review\nsave\n"), &stdout, &stderr); exit != 0 {
		t.Fatalf("exit %d: %q", exit, stderr.String())
	}
	assertReport(t, stderr.Bytes(), "unchanged", "guide.unchanged")
	want := `Name: "Ünal\u001b\u202e\u200d\u0085\u2028\u2029\nline\t\"\\" -> "Ünal\u001b\u202e\u200d\u0085\u2028\u2029\nline\t\"\\"`
	if !strings.Contains(stdout.String(), want) {
		t.Fatalf("literal display missing %q in %q", want, stdout.String())
	}
	for _, forbidden := range []string{"\x1b", "\u202e", "\u200d", "\u0085", "\u2028", "\u2029"} {
		if strings.Contains(stdout.String(), forbidden) {
			t.Errorf("unsafe literal terminal control %q", forbidden)
		}
	}
	after, err := os.Stat(box.Prep)
	if err != nil || !os.SameFile(before, after) {
		t.Fatalf("no-op replaced inode: %v", err)
	}
	if got, err := os.ReadFile(box.Prep); err != nil || !bytes.Equal(got, fixture(t, "literal.yaml")) {
		t.Fatalf("no-op changed bytes: %q %v", got, err)
	}
}

func TestIdentityAndAmbiguityRefusal(t *testing.T) {
	t.Parallel()
	for _, current := range []struct{ Name, Data, Code string }{
		{"version", "schemaVersion: 2\nkind: preparation\naccount: newcomer\nplanId: planone\n", "guide.source"},
		{"kind", "schemaVersion: 1\nkind: record\naccount: newcomer\nplanId: planone\n", "guide.source"},
		{"account", "schemaVersion: 1\nkind: preparation\naccount: PRIVATE-other\nplanId: planone\n", "guide.identity"},
		{"plan", "schemaVersion: 1\nkind: preparation\naccount: newcomer\nplanId: PRIVATE-other\n", "guide.identity"},
		{"duplicate", "schemaVersion: 1\nkind: preparation\naccount: newcomer\nplanId: planone\narrangements:\n  - id: same\n  - id: same\n", "guide.identity"},
		{"duplicate-empty", "schemaVersion: 1\nkind: preparation\naccount: newcomer\nplanId: planone\narrangements: [{}, {}]\n", "guide.identity"},
		{"malformed", "PRIVATE-broken: [\n", "guide.source"},
	} {
		t.Run(current.Name, func(t *testing.T) {
			box := newSandbox(t, "")
			writeFile(t, box.Prep, []byte(current.Data))
			before := snapshot(t, box.Root)
			reader := &monitoredReader{Value: "review\nsave\n"}
			var stdout, stderr bytes.Buffer
			if exit := newCommand().Run(box.args(), reader, &stdout, &stderr); exit != 1 {
				t.Fatalf("exit %d", exit)
			}
			assertReport(t, stderr.Bytes(), "refused", current.Code)
			if reader.ReadCount != 0 {
				t.Fatal("read interaction after startup refusal")
			}
			if strings.Contains(stderr.String(), "PRIVATE") {
				t.Fatal("diagnostic leaked source")
			}
			if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
				t.Fatal("startup refusal changed files")
			}
		})
	}
}

func TestSafePathsAndHardlinks(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"plan-symlink", "root-symlink", "parent-symlink", "preparation-symlink", "traversal", "alias"} {
		box := newSandbox(t, "identity.yaml")
		args := box.args()
		switch mode {
		case "plan-symlink":
			path := filepath.Join(box.Root, "selected.yaml")
			if err := os.Symlink(box.Plan, path); err != nil {
				t.Fatal(err)
			}
			args[1] = path
		case "root-symlink":
			path := filepath.Join(box.Root, "selected-root")
			if err := os.Symlink(box.Root, path); err != nil {
				t.Fatal(err)
			}
			args[5] = path
		case "parent-symlink":
			path := filepath.Join(box.Root, "selected-parent")
			if err := os.Symlink(box.Root, path); err != nil {
				t.Fatal(err)
			}
			args[1] = filepath.Join(path, "organization.yaml")
		case "preparation-symlink":
			if err := os.Remove(box.Prep); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(box.Plan, box.Prep); err != nil {
				t.Fatal(err)
			}
		case "traversal":
			args[1] = box.Root + "/records/../organization.yaml"
		case "alias":
			writeFile(t, box.Plan, []byte("schemaVersion: 1\nkind: preparation\nid: planone\naccount: newcomer\nplanId: planone\n"))
			if err := os.Remove(box.Prep); err != nil {
				t.Fatal(err)
			}
			if err := os.Link(box.Plan, box.Prep); err != nil {
				t.Fatal(err)
			}
		}
		before := snapshot(t, box.Root)
		var stdout, stderr bytes.Buffer
		if exit := newCommand().Run(args, strings.NewReader("review\nsave\n"), &stdout, &stderr); exit != 1 {
			t.Fatalf("%s exit %d: %q", mode, exit, stderr.String())
		}
		var report map[string]string
		if err := json.Unmarshal(stderr.Bytes(), &report); err != nil || report["status"] != "refused" {
			t.Fatalf("%s report %q", mode, stderr.String())
		}
		if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
			t.Fatalf("%s changed files", mode)
		}
	}
	box := newSandbox(t, "named.yaml")
	link := filepath.Join(box.Root, "held-preparation.yaml")
	if err := os.Link(box.Prep, link); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if exit := newCommand().Run(box.args(), strings.NewReader("recipient\n1\nclear\napply\nreview\nsave\n"), &stdout, &stderr); exit != 0 {
		t.Fatalf("replace exit %d %q", exit, stderr.String())
	}
	assertReport(t, stderr.Bytes(), "saved", "guide.saved")
	if got, err := os.ReadFile(link); err != nil || !bytes.Equal(got, fixture(t, "named.yaml")) {
		t.Fatalf("hardlink inode truncated: %q %v", got, err)
	}
	if got, err := os.ReadFile(box.Prep); err != nil || !bytes.Equal(got, fixture(t, "identity.yaml")) {
		t.Fatalf("selected path not replaced: %q %v", got, err)
	}
}

type failingWriter struct {
	Full  bool
	Value bytes.Buffer
}

func (w *failingWriter) Write(value []byte) (int, error) {
	if w.Full {
		_, _ = w.Value.Write(value)
		return len(value), errors.New("PRIVATE stderr error")
	}
	return 0, errors.New("PRIVATE stderr error")
}

func TestPostcommitReportFailureRetainsSavedBytes(t *testing.T) {
	t.Parallel()
	for _, full := range []bool{false, true} {
		box := newSandbox(t, "")
		var stdout bytes.Buffer
		stderr := &failingWriter{Full: full}
		if exit := newCommand().Run(box.args(), strings.NewReader("review\nsave\n"), &stdout, stderr); exit != 1 {
			t.Fatalf("full %v exit %d", full, exit)
		}
		if got, err := os.ReadFile(box.Prep); err != nil || !bytes.Equal(got, fixture(t, "identity.yaml")) {
			t.Fatalf("committed bytes lost: %q %v", got, err)
		}
		if strings.Contains(stderr.Value.String(), "No files changed") || strings.Contains(stderr.Value.String(), "PRIVATE") {
			t.Fatalf("false or private report: %q", stderr.Value.String())
		}
		if full && (!strings.Contains(stderr.Value.String(), "guide.saved") || !strings.Contains(stderr.Value.String(), "guide.report-saved")) {
			t.Fatalf("missing truthful best effort reports: %q", stderr.Value.String())
		}
	}
}

func TestRecordIsNotReadAndPrivateDirectories(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "")
	record := filepath.Join(box.Root, "records", "newcomer", "record.obk.yaml")
	writeFile(t, record, []byte("PRIVATE malformed record [\n"))
	var stdout, stderr bytes.Buffer
	if exit := newCommand().Run(box.args(), strings.NewReader("review\nsave\n"), &stdout, &stderr); exit != 0 {
		t.Fatalf("record interfered: exit %d %q", exit, stderr.String())
	}
	assertReport(t, stderr.Bytes(), "saved", "guide.saved")
	if got, err := os.ReadFile(record); err != nil || string(got) != "PRIVATE malformed record [\n" {
		t.Fatal("record changed")
	}
	if strings.Contains(stdout.String(), "PRIVATE") || strings.Contains(stderr.String(), "PRIVATE") {
		t.Fatal("record displayed")
	}
	info, err := os.Stat(box.Prep)
	if err != nil || info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("preparation not private: %v %v", info, err)
	}
	other := newSandbox(t, "")
	stdout.Reset()
	stderr.Reset()
	if exit := newCommand().Run(other.args(), strings.NewReader("review\nsave\n"), &stdout, &stderr); exit != 0 {
		t.Fatalf("new nested dirs exit %d", exit)
	}
	for _, path := range []string{filepath.Join(other.Root, "records"), filepath.Dir(other.Prep)} {
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm()&0o077 != 0 {
			t.Fatalf("directory not private: %v %v", info, err)
		}
	}
}

func TestDirectoryChoicesAreExplicit(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "")
	var stdout, stderr bytes.Buffer
	input := "recipient\n4\nchange\nteam\n5\nchange\nmanager\n6\nchange\nowner\n7\nchange\nhelper\napply\nreview\ncancel\n"
	if exit := newCommand().Run(box.args(), strings.NewReader(input), &stdout, &stderr); exit != 0 {
		t.Fatalf("exit %d %q", exit, stderr.String())
	}
	assertReport(t, stderr.Bytes(), "cancelled", "guide.cancelled")
	for _, part := range []string{
		"Available teams:\nID: \"team\" Name: \"Example team\"\n",
		"Available contacts:\nID: \"owner\" Name: \"Example owner\"\nID: \"helper\" Name: \"Example helper\"\nID: \"manager\" Name: \"Example manager\"\n",
		"Team ID: \"\" -> \"team\"", "Manager contact ID: \"\" -> \"manager\"",
		"Onboarding owner contact ID: \"\" -> \"owner\"", "Buddy contact ID: \"\" -> \"helper\"",
	} {
		if !strings.Contains(stdout.String(), part) {
			t.Errorf("missing exact choices/change %q in %q", part, stdout.String())
		}
	}
	if _, err := os.Stat(filepath.Join(box.Root, "records")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("cancel created preparation directories: %v", err)
	}
}

func TestSourceSizeBoundary(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"plan", "preparation"} {
		for _, excess := range []int{0, 1} {
			box := newSandbox(t, "identity.yaml")
			path := box.Plan
			name := "plan.yaml"
			if source == "preparation" {
				path = box.Prep
				name = "identity.yaml"
			}
			base := fixture(t, name)
			data := append(append([]byte{}, base...), []byte("#"+strings.Repeat("x", 1_048_576+excess-len(base)-1))...)
			writeFile(t, path, data)
			before := snapshot(t, box.Root)
			var stdout, stderr bytes.Buffer
			exit := newCommand().Run(box.args(), strings.NewReader("review\nsave\n"), &stdout, &stderr)
			if excess == 0 {
				if exit != 0 {
					t.Fatalf("exact %s size exit %d %q", source, exit, stderr.String())
				}
				assertReport(t, stderr.Bytes(), "unchanged", "guide.unchanged")
			} else {
				if exit != 1 {
					t.Fatalf("oversize %s exit %d", source, exit)
				}
				assertReport(t, stderr.Bytes(), "refused", "guide.source")
			}
			if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
				t.Fatalf("%s boundary changed filesystem", source)
			}
		}
	}
}

func TestActualPlanValidationAndBlockerReview(t *testing.T) {
	t.Parallel()
	for _, duplicate := range []bool{false, true} {
		box := newSandbox(t, "")
		plan := string(fixture(t, "plan.yaml")) + "variables:\n  - name: organization\n    required: true\n    audience: onboarder\n    default: Example default\n"
		if duplicate {
			plan += "  - name: organization\n"
		}
		writeFile(t, box.Plan, []byte(plan))
		before := snapshot(t, box.Root)
		var stdout, stderr bytes.Buffer
		exit := newCommand().Run(box.args(), strings.NewReader("review\ncancel\n"), &stdout, &stderr)
		if duplicate {
			if exit != 1 {
				t.Fatalf("ambiguous plan exit %d", exit)
			}
			assertReport(t, stderr.Bytes(), "refused", "guide.plan")
		} else {
			if exit != 0 {
				t.Fatalf("blocker review exit %d %q", exit, stderr.String())
			}
			assertReport(t, stderr.Bytes(), "cancelled", "guide.cancelled")
			for _, part := range []string{"Onboarder blockers:\nQuestion: \"plan.organization\"", "First-steps document:\n| \"# First steps\"", "| \"Organization: Organization not specified.\"", "| \"General help contacts not specified. Ask the onboarding owner to provide a help route that works before company access.\""} {
				if !strings.Contains(stdout.String(), part) {
					t.Errorf("missing actual blocker/guide %q in %q", part, stdout.String())
				}
			}
			if strings.Contains(stdout.String(), "Example default") {
				t.Fatal("unaccepted organization default displayed as guide content")
			}
		}
		if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
			t.Fatal("review/refusal changed filesystem")
		}
	}
}

func TestUnencodableSizeExpansionDoesNotPublish(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "")
	data := string(fixture(t, "identity.yaml")) + "x-expansion: \"" + strings.Repeat("\u2028", 180_000) + "\"\n"
	if len(data) > 1_048_576 {
		t.Fatal("fixture source exceeds input bound")
	}
	writeFile(t, box.Prep, []byte(data))
	before := snapshot(t, box.Root)
	var stdout, stderr bytes.Buffer
	input := "recipient\n1\nchange\nExample newcomer\napply\nreview\nsave\n"
	if exit := newCommand().Run(box.args(), strings.NewReader(input), &stdout, &stderr); exit != 1 {
		t.Fatalf("expanded encoding exit %d %q", exit, stderr.String())
	}
	assertReport(t, stderr.Bytes(), "refused", "guide.encode")
	if strings.Contains(stdout.String(), "Review complete. Choose save") {
		t.Fatal("offered save for oversized encoding")
	}
	if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
		t.Fatal("oversize encoding changed filesystem")
	}
}

func TestInvalidSelectionBeforeInteraction(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"blank-plan-id", "unsafe-results-path", "invalid-account"} {
		box := newSandbox(t, "")
		args := box.args()
		code := "guide.plan"
		switch mode {
		case "blank-plan-id":
			writeFile(t, box.Plan, []byte("schemaVersion: 1\n"))
		case "unsafe-results-path":
			writeFile(t, box.Plan, append(fixture(t, "plan.yaml"), []byte("resultsPath: ../records/{account}/record.obk.yaml\n")...))
			code = "guide.path"
		case "invalid-account":
			args[3] = "../PRIVATE"
			code = "guide.path"
		}
		before := snapshot(t, box.Root)
		reader := &monitoredReader{Value: "review\nsave\n"}
		var stdout, stderr bytes.Buffer
		if exit := newCommand().Run(args, reader, &stdout, &stderr); exit != 1 {
			t.Fatalf("%s exit %d %q", mode, exit, stderr.String())
		}
		assertReport(t, stderr.Bytes(), "refused", code)
		if reader.ReadCount != 0 {
			t.Fatalf("%s consumed interaction", mode)
		}
		if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
			t.Fatalf("%s changed files", mode)
		}
	}
}
