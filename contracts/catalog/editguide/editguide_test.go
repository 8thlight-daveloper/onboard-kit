package editguide_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

type countedInput struct {
	source   []byte
	position int
	hookAt   int
	hook     func()
}

func (r *countedInput) Read(p []byte) (int, error) {
	if r.hook != nil && r.position == r.hookAt {
		hook := r.hook
		r.hook = nil
		hook()
	}
	if len(p) == 0 {
		return 0, nil
	}
	if r.position == len(r.source) {
		return 0, io.EOF
	}
	p[0] = r.source[r.position]
	r.position++
	return 1, nil
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func setup(t *testing.T, parentMode os.FileMode) (string, string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chown(root, -1, os.Getegid()); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, parentMode); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"original.yaml", "record.yaml", "preparation.yaml"} {
		if err := os.WriteFile(filepath.Join(root, name), fixture(t, name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root, filepath.Join(root, "original.yaml")
}

func personalPreserved(t *testing.T, root string) {
	t.Helper()
	for _, name := range []string{"record.yaml", "preparation.yaml"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil || !bytes.Equal(data, fixture(t, name)) {
			t.Errorf("personal artifact changed: %s", name)
		}
	}
}

func report(t *testing.T, data []byte, status, code, message string) {
	t.Helper()
	var actual map[string]string
	if json.Unmarshal(data, &actual) != nil || !reflect.DeepEqual(actual, map[string]string{"status": status, "code": code, "message": message}) || !bytes.HasSuffix(data, []byte("\n")) {
		t.Errorf("exact private report differs: %q", data)
	}
}

func unchanged(t *testing.T, target string) {
	t.Helper()
	data, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(data, fixture(t, "original.yaml")) {
		t.Error("selected source bytes changed")
	}
}

func document(label string, data []byte) string {
	var out strings.Builder
	out.WriteString(label + ":\n")
	for _, line := range strings.Split(string(data), "\n") {
		out.WriteString("  | " + strconv.Quote(line) + "\n")
	}
	return out.String()
}

func TestInitialHelpAndCLIRefuseWithoutReading(t *testing.T) {
	for _, args := range [][]string{{"--help", "--unknown"}, {"-h", "--plan=/missing"}} {
		input := &countedInput{source: []byte("UNREAD")}
		var out, stderr bytes.Buffer
		if newCommand().Run(args, input, &out, &stderr) != 0 || !bytes.Equal(out.Bytes(), fixture(t, "help.txt")) || stderr.Len() != 0 || input.position != 0 {
			t.Error("help crossed an effect/input boundary")
		}
	}
	for _, args := range [][]string{nil, {"--plan="}, {"--plan"}, {"--plan=a", "--plan=b"}, {"--out=a"}, {"--plan=a", "--help"}, {"--plan", "--path"}} {
		input := &countedInput{source: []byte("UNREAD")}
		var out, stderr bytes.Buffer
		if newCommand().Run(args, input, &out, &stderr) != 2 || input.position != 0 || out.Len() != 0 {
			t.Error("invalid CLI crossed admission")
		}
		report(t, stderr.Bytes(), "refused", "edit.usage", "Use catalog edit --plan FILE.")
	}
}

func TestCancelPartialConsentAndUnchangedSavePreserveExactFiles(t *testing.T) {
	for _, c := range []struct {
		name, input, status, code, message string
		consumed                           int
	}{
		{"cancel", "cancel\nTAIL", "cancelled", "edit.cancelled", "Cancelled. The selected plan was not changed.", 7},
		{"partial", "review\nsave", "cancelled", "edit.cancelled", "Cancelled. The selected plan was not changed.", 11},
		{"unchanged", "review\nsave\nTAIL", "unchanged", "edit.unchanged", "The reviewed proposal leaves the selected plan unchanged.", 12},
		{"back", "review\nback\nsave\ncancel\n", "cancelled", "edit.cancelled", "Cancelled. The selected plan was not changed.", 24},
	} {
		t.Run(c.name, func(t *testing.T) {
			root, target := setup(t, 0755)
			before, err := os.Stat(target)
			if err != nil {
				t.Fatal(err)
			}
			input := &countedInput{source: []byte(c.input)}
			var out, stderr bytes.Buffer
			if newCommand().Run([]string{"--plan", target}, input, &out, &stderr) != 0 || input.position != c.consumed {
				t.Errorf("cancel/no-op/consent boundary differs: %d", input.position)
			}
			report(t, stderr.Bytes(), c.status, c.code, c.message)
			unchanged(t, target)
			personalPreserved(t, root)
			after, err := os.Stat(target)
			if err != nil || !os.SameFile(before, after) {
				t.Error("cancel or unchanged save replaced original inode")
			}
			startup := strings.ReplaceAll(string(fixture(t, "startup.txt")), "@PLAN@", strconv.Quote(target))
			if !strings.HasPrefix(out.String(), startup) {
				t.Error("exact startup missing")
			}
		})
	}
}

func TestReviewedRetirementReplacesOnlyPlanIn0700And0755Parents(t *testing.T) {
	for _, mode := range []os.FileMode{0700, 0755} {
		root, target := setup(t, mode)
		parentBefore, err := os.Stat(root)
		if err != nil {
			t.Fatal(err)
		}
		input := &countedInput{source: fixture(t, "retire.txt")}
		var out, stderr bytes.Buffer
		if newCommand().Run([]string{"--plan=" + target}, input, &out, &stderr) != 0 {
			t.Fatalf("reviewed retirement failed: %q", stderr.String())
		}
		report(t, stderr.Bytes(), "written", "edit.written", "The reviewed plan edits were saved.")
		data, err := os.ReadFile(target)
		if err != nil || !bytes.Equal(data, fixture(t, "retired.yaml")) {
			t.Errorf("independent exact saved retirement differs: %q", data)
		}
		if input.position != len("retire\nresource-old\nreview\nsave\n") {
			t.Error("save read queued tail")
		}
		for _, part := range []string{"Original plan bytes:\n", "Proposed plan bytes:\n", "Save action: replace.\n", "Preparation guidance for newly added instances only: 0\n", "Review complete. Choose save, back, or cancel:\n"} {
			if !strings.Contains(out.String(), part) {
				t.Errorf("complete review missing: %q", part)
			}
		}
		for _, name := range []string{"original.yaml", "retired.yaml"} {
			for _, line := range strings.Split(string(fixture(t, name)), "\n") {
				if !strings.Contains(out.String(), "  | "+strconv.Quote(line)+"\n") {
					t.Errorf("complete literal document line omitted: %q", line)
				}
			}
		}
		exactReview := strings.NewReplacer("@PLAN@", strconv.Quote(target), "@ORIGINAL@", document("Original plan bytes", fixture(t, "original.yaml")), "@PROPOSED@", document("Proposed plan bytes", fixture(t, "retired.yaml"))).Replace(string(fixture(t, "retire-review.txt")))
		if !strings.Contains(out.String(), exactReview) {
			t.Error("exact ordered complete original/proposed retirement review differs")
		}
		info, err := os.Stat(target)
		if err != nil {
			t.Fatal(err)
		}
		native, ok := info.Sys().(*syscall.Stat_t)
		if !ok || info.Mode().Perm() != 0600 || native.Uid != uint32(os.Geteuid()) || native.Gid != uint32(os.Getegid()) || native.Nlink != 1 {
			t.Error("replacement lost supported native profile")
		}
		parentAfter, err := os.Stat(root)
		if err != nil || !os.SameFile(parentBefore, parentAfter) || parentAfter.Mode() != parentBefore.Mode() {
			t.Error("parent ownership/mode context replaced or chmodded")
		}
		entries, err := os.ReadDir(root)
		if err != nil || len(entries) != 3 {
			t.Errorf("unexpected temporary/sidecar artifact: %v %v", entries, err)
		}
		personalPreserved(t, root)
	}
}

func TestFreshSourceBytesAndEqualByteReplacementInodesRefuseStaleSave(t *testing.T) {
	for _, replacement := range []bool{false, true} {
		root, target := setup(t, 0755)
		input := &countedInput{source: []byte("retire\nresource-old\nreview\nsave\n"), hookAt: len("retire\nresource-old\nreview\n")}
		ownerBytes := []byte("INDEPENDENT OWNER CHANGE\n")
		input.hook = func() {
			if replacement {
				ownerBytes = fixture(t, "original.yaml")
				alternate := filepath.Join(root, "owner-replacement.yaml")
				if err := os.WriteFile(alternate, ownerBytes, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(alternate, target); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.WriteFile(target, ownerBytes, 0600); err != nil {
					t.Fatal(err)
				}
			}
		}
		var out, stderr bytes.Buffer
		if newCommand().Run([]string{"--plan=" + target}, input, &out, &stderr) != 1 {
			t.Error("stale source save accepted")
		}
		report(t, stderr.Bytes(), "refused", "edit.stale", "The selected plan context changed; reload and review again.")
		data, err := os.ReadFile(target)
		if err != nil || !bytes.Equal(data, ownerBytes) {
			t.Error("observed independent owner change overwritten")
		}
		personalPreserved(t, root)
	}
}

type displayFailure struct {
	data   bytes.Buffer
	mode   int
	failed bool
}

func (w *displayFailure) Write(p []byte) (int, error) {
	if w.failed {
		return 0, errors.New("fixture display failure")
	}
	joined := append(append([]byte(nil), w.data.Bytes()...), p...)
	if bytes.Contains(joined, []byte("Review complete.")) {
		w.failed = true
		if w.mode == 0 {
			return 0, errors.New("fixture display failure")
		}
		if w.mode == 1 {
			count := len(p) - 1
			if count < 0 {
				count = 0
			}
			w.data.Write(p[:count])
			return count, nil
		}
		w.data.Write(p)
		return len(p), errors.New("fixture display failure")
	}
	return w.data.Write(p)
}

func TestFailedReviewOutputDoesNotConsumeQueuedSaveOrChangeFiles(t *testing.T) {
	for mode := 0; mode < 3; mode++ {
		root, target := setup(t, 0755)
		input := &countedInput{source: []byte("retire\nresource-old\nreview\nsave\nTAIL")}
		out := &displayFailure{mode: mode}
		var stderr bytes.Buffer
		if newCommand().Run([]string{"--plan=" + target}, input, out, &stderr) != 1 || input.position != len("retire\nresource-old\nreview\n") || !out.failed {
			t.Error("unaccepted review consumed queued save or fixture fault was not reached")
		}
		report(t, stderr.Bytes(), "refused", "edit.output", "Required review output failed; no plan replacement was accepted.")
		unchanged(t, target)
		personalPreserved(t, root)
	}
}

func TestInvalidBatchRemainsRepairableByDropButNeverAppliesPrefix(t *testing.T) {
	root, target := setup(t, 0755)
	input := strings.NewReader("retire\nresource-old\nsubmit\nresource-old\nresource_name\ntext\nPRIVATE_REJECTED_VALUE\nreview\ndrop\n2\nreview\nsave\n")
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--plan=" + target}, input, &out, &stderr) != 0 {
		t.Fatalf("drop did not repair staged invalid operation: %q", stderr.String())
	}
	if !strings.Contains(out.String(), "Review unavailable.\n") || !strings.Contains(out.String(), "Code: \"answer.inactive-question\"") || strings.Contains(out.String(), "PRIVATE_REJECTED_VALUE") {
		t.Error("whole-batch refusal/private projection missing")
	}
	data, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(data, fixture(t, "retired.yaml")) {
		t.Error("dropped rejected submission leaked into saved plan")
	}
	personalPreserved(t, root)
}
