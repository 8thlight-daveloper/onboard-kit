package guide_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func arrangementCases(t *testing.T) []scenario {
	t.Helper()
	var cases []scenario
	if err := json.Unmarshal(fixture(t, "arrangement-cases.json"), &cases); err != nil {
		t.Fatal(err)
	}
	return cases
}

func TestArrangementScenarios(t *testing.T) {
	for _, current := range arrangementCases(t) {
		t.Run(current.Name, func(t *testing.T) {
			t.Parallel()
			if current.Source == "" {
				t.Fatal("arrangement fixture lacks source")
			}
			box := newSandbox(t, current.Before)
			writeFile(t, box.Plan, fixture(t, "arrangement-plan.yaml"))
			before := snapshot(t, box.Root)
			var original os.FileInfo
			if current.Before != "" {
				var err error
				original, err = os.Stat(box.Prep)
				if err != nil {
					t.Fatal(err)
				}
			}
			var stdout, stderr bytes.Buffer
			if exit := newCommand().Run(box.args(), strings.NewReader(current.Input), &stdout, &stderr); exit != current.Exit {
				t.Fatalf("exit %d; want %d; report %q", exit, current.Exit, stderr.String())
			}
			assertReport(t, stderr.Bytes(), current.Status, current.Code)
			for _, part := range current.Contains {
				if !strings.Contains(stdout.String(), part) {
					t.Errorf("missing arrangement presentation %q", part)
				}
			}
			if strings.Contains(stdout.String(), "PRIVATE-unknown-value") || strings.Contains(stderr.String(), "PRIVATE-unknown-value") {
				t.Fatal("unknown preparation value displayed")
			}
			want := before
			if current.After != "" {
				want["records/"] = "directory"
				want["records/newcomer/"] = "directory"
				want["records/newcomer/preparation.obk.yaml"] = string(fixture(t, current.After))
			}
			if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, want) {
				t.Fatalf("arrangement save/preservation mismatch: actual %#v; expected %#v", actual, want)
			}
			if current.Status == "unchanged" {
				after, err := os.Stat(box.Prep)
				if err != nil || !os.SameFile(original, after) {
					t.Fatal("arrangement no-op replaced original inode")
				}
			}
		})
	}
}

func TestCompleteOriginalAndDraftArrangementReview(t *testing.T) {
	t.Parallel()
	var input string
	for _, current := range arrangementCases(t) {
		if current.Name == "edit-whole-arrangement-clears-in-place" {
			input = current.Input
		}
	}
	if input == "" {
		t.Fatal("missing independent review input")
	}
	box := newSandbox(t, "arrangements-before.yaml")
	writeFile(t, box.Plan, fixture(t, "arrangement-plan.yaml"))
	record := filepath.Join(box.Root, "records", "newcomer", "record.obk.yaml")
	writeFile(t, record, []byte("PRIVATE malformed record [\n"))
	before := snapshot(t, box.Root)
	var stdout, stderr bytes.Buffer
	if exit := newCommand().Run(box.args(), strings.NewReader(input), &stdout, &stderr); exit != 0 {
		t.Fatalf("complete review exit %d: %q", exit, stderr.String())
	}
	assertReport(t, stderr.Bytes(), "saved", "guide.saved")
	if !strings.Contains(stdout.String(), string(fixture(t, "arrangement-review.txt"))) {
		t.Fatal("review omitted or changed complete original/draft known-field blocks")
	}
	if strings.Contains(stdout.String(), "PRIVATE") || strings.Contains(stderr.String(), "PRIVATE") {
		t.Fatal("record or unknown preparation value displayed")
	}
	before["records/newcomer/preparation.obk.yaml"] = string(fixture(t, "arrangements-edited.yaml"))
	if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
		t.Fatal("arrangement save modified plan, record or unrelated paths")
	}
}

func TestArrangementScalarDisplayRemainsLiteral(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "")
	data := "schemaVersion: 1\nkind: preparation\naccount: newcomer\nplanId: planone\narrangements:\n  - id: literal-note\n    section: inactive\n    kind: access\n    status: pending\n    note: \"save\\x1b\\u202e\\u200d\\nnext\"\n"
	writeFile(t, box.Prep, []byte(data))
	before := snapshot(t, box.Root)
	var stdout, stderr bytes.Buffer
	if exit := newCommand().Run(box.args(), strings.NewReader("arrangements\nedit\n1\n8\nkeep\nback\nback\nreview\ncancel\n"), &stdout, &stderr); exit != 0 {
		t.Fatalf("literal arrangement review exit %d", exit)
	}
	assertReport(t, stderr.Bytes(), "cancelled", "guide.cancelled")
	if !strings.Contains(stdout.String(), `Note: "save\u001b\u202e\u200d\nnext"`) {
		t.Fatal("unsafe held arrangement scalar not quoted literally")
	}
	for _, unsafe := range []string{"\x1b", "\u202e", "\u200d"} {
		if strings.Contains(stdout.String(), unsafe) {
			t.Fatal("arrangement display emitted terminal control")
		}
	}
	if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
		t.Fatal("literal arrangement display changed held bytes")
	}
}

type failedArrangementReview struct {
	Value  bytes.Buffer
	Failed bool
}

func (w *failedArrangementReview) Write(value []byte) (int, error) {
	shouldFail := !w.Failed && strings.Contains(w.Value.String()+string(value), "Draft arrangements:")
	count, err := w.Value.Write(value)
	if shouldFail {
		w.Failed = true
		return count, errors.New("PRIVATE arrangement display failure")
	}
	return count, err
}

func TestArrangementReviewMustFinishBeforeReadingSave(t *testing.T) {
	t.Parallel()
	box := newSandbox(t, "arrangements-before.yaml")
	writeFile(t, box.Plan, fixture(t, "arrangement-plan.yaml"))
	before := snapshot(t, box.Root)
	prefix := "arrangements\nremove\n2\nremove\nback\nreview\n"
	reader := &chunkReader{Reader: strings.NewReader(prefix + "save\n")}
	var stdout failedArrangementReview
	var stderr bytes.Buffer
	if exit := newCommand().Run(box.args(), reader, &stdout, &stderr); exit != 1 {
		t.Fatalf("arrangement review display failure exit %d", exit)
	}
	assertReport(t, stderr.Bytes(), "refused", "guide.output")
	if !stdout.Failed {
		t.Fatal("fixture never reached the failed arrangement review block")
	}
	if reader.ReadCount != len(prefix) {
		t.Fatal("save input read before complete arrangement review succeeded")
	}
	if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
		t.Fatal("failed arrangement review published staged removal")
	}
}
