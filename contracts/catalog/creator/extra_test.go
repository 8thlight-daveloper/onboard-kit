package creator_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestOpaquePlanIdentityIsPreserved(t *testing.T) {
	root, target := sandbox(t)
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--id= Plan α ", "--out=" + target}, strings.NewReader("review\nsave\n"), &out, &stderr) != 0 {
		t.Fatalf("opaque identity refused: %q", stderr.String())
	}
	body, err := os.ReadFile(target)
	if err != nil || string(body) != "schemaVersion: 1\nid: \" Plan α \"\n" {
		t.Errorf("opaque identity changed: %q err=%v", body, err)
	}
	if !strings.Contains(out.String(), "Plan: \" Plan α \"\n") {
		t.Error("opaque identity was not safely displayed")
	}
	preserved(t, root)
}

func TestExactAdvisoryInstructionsAreDisplayedInActualGuideOrder(t *testing.T) {
	root, target := sandbox(t)
	var out, stderr bytes.Buffer
	if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, bytes.NewReader(fixture(t, "repeated.txt")), &out, &stderr) != 0 {
		t.Fatalf("repeated creation failed: %q", stderr.String())
	}
	previous := -1
	for _, id := range []string{"invitation-prepare", "reading-assign", "resource-access-arrange"} {
		var exact strings.Builder
		exact.WriteString("Advisory instructions:\n")
		for _, line := range strings.Split(string(fixture(t, id+".md")), "\n") {
			exact.WriteString("  | ")
			exact.WriteString(strconv.Quote(line))
			exact.WriteByte('\n')
		}
		start := strings.Index(out.String(), "Guide: "+strconv.Quote(id)+"\n")
		if start < 0 || start <= previous {
			t.Errorf("guide missing/out of order: %s", id)
			continue
		}
		end := strings.Index(out.String()[start:], exact.String())
		if end < 0 {
			t.Errorf("literal full instruction block missing for %s", id)
			continue
		}
		next := strings.Index(out.String()[start+1:], "Guide: ")
		if next >= 0 && end > next+1 {
			t.Errorf("instruction block associated with a different guide: %s", id)
		}
		previous = start
	}
	preserved(t, root)
}

type borrowedReader struct {
	*strings.Reader
	closed bool
}

func (r *borrowedReader) Close() error { r.closed = true; return nil }

type borrowedWriter struct {
	bytes.Buffer
	closed bool
}

func (w *borrowedWriter) Close() error { w.closed = true; return nil }

type reportFailure struct {
	calls    int
	accepted bytes.Buffer
}

func (w *reportFailure) Write(p []byte) (int, error) {
	w.calls++
	if w.calls == 1 {
		return 0, errors.New("owned report failure")
	}
	return w.accepted.Write(p)
}

func TestReportFailureAfterPublicationKeepsReviewedPlanAndBorrowedStreams(t *testing.T) {
	root, target := sandbox(t)
	input := &borrowedReader{Reader: strings.NewReader("review\nsave\nUNREAD")}
	var out borrowedWriter
	var stderr reportFailure
	if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, &out, &stderr) != 1 {
		t.Error("failed final report did not return failure")
	}
	if input.closed || out.closed {
		t.Error("borrowed stream was closed")
	}
	remaining, err := io.ReadAll(input)
	if err != nil || string(remaining) != "UNREAD" {
		t.Errorf("save consumed following bytes: %q %v", remaining, err)
	}
	var actual map[string]string
	if stderr.calls < 2 || json.Unmarshal(stderr.accepted.Bytes(), &actual) != nil || !reflect.DeepEqual(actual, map[string]string{"status": "written", "code": "create.report-written", "message": "The plan was created, but its report could not be written."}) || !bytes.HasSuffix(stderr.accepted.Bytes(), []byte("\n")) {
		t.Errorf("fixed single report fallback missing: calls=%d bytes=%q", stderr.calls, stderr.accepted.Bytes())
	}
	body, err := os.ReadFile(target)
	if err != nil || !bytes.Equal(body, fixture(t, "empty-plan.yaml")) {
		t.Error("published reviewed plan was rolled back or changed after report failure")
	}
	preserved(t, root)
}

func TestExplicitEmptyListDoesNotBecomeMissingOrNone(t *testing.T) {
	root, target := sandbox(t)
	var out, stderr bytes.Buffer
	input := "add\nonboarding-help\nhelp-a\nbind\nhelp-a\ngeneral_contacts\nlist\n0\nreview\nsave\n"
	if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, strings.NewReader(input), &out, &stderr) != 0 {
		t.Fatalf("explicit empty list refused: %q", stderr.String())
	}
	plan := savedPlan(t, target)
	answers, ok := plan["answers"].(map[string]any)
	if !ok {
		t.Fatal("explicit empty list scope was dropped")
	}
	sections, ok := answers["sections"].(map[string]any)
	if !ok {
		t.Fatal("explicit empty list scope was dropped")
	}
	local, ok := sections["help-a"].(map[string]any)
	if !ok {
		t.Fatal("explicit empty list binding was dropped")
	}
	list, ok := local["general_contacts"].([]any)
	if !ok || len(list) != 0 {
		t.Errorf("explicit empty list changed: %#v", local)
	}
	preserved(t, root)
}

func TestRejectedPrivateSourceAndBindingValuesNeverEcho(t *testing.T) {
	root, target := sandbox(t)
	var out, stderr bytes.Buffer
	input := "add\nPRIVATE_REJECTED_SOURCE\nunused\nadd\nrepository-access\nrepository-a\nbind\nrepository-a\nPRIVATE_REJECTED_NAME\ntext\nPRIVATE_REJECTED_VALUE\nreview\nsave\n"
	if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, strings.NewReader(input), &out, &stderr) != 0 {
		t.Fatalf("repairable draft did not save: %q", stderr.String())
	}
	combined := out.String() + stderr.String()
	for _, private := range []string{"PRIVATE_REJECTED_SOURCE", "PRIVATE_REJECTED_NAME", "PRIVATE_REJECTED_VALUE"} {
		if strings.Contains(combined, private) {
			t.Errorf("rejected private input echoed: %s", private)
		}
	}
	for _, repair := range []string{"Repair: \"Choose an exact section-template ID from catalog list.\"", "Repair: \"Bind an exact variable declared in this scope.\""} {
		if !strings.Contains(combined, repair) {
			t.Errorf("missing useful fixed repair: %s", repair)
		}
	}
	plan := savedPlan(t, target)
	if _, exists := plan["answers"]; exists {
		t.Error("rejected binding was silently installed")
	}
	preserved(t, root)
}

func TestFreshRemoveReaddAndNonexampleRepetition(t *testing.T) {
	root, target := sandbox(t)
	var out, stderr bytes.Buffer
	input := "add\nrepository-access\nactivity-a\nremove\nactivity-a\nadd\nworking-hours\nactivity-a\nadd\nworking-hours\nactivity-b\nreview\nsave\n"
	if newCommand().Run([]string{"--id=plan-one", "--out=" + target}, strings.NewReader(input), &out, &stderr) != 0 {
		t.Fatalf("fresh unsaved reselect failed: %q", stderr.String())
	}
	sections, ok := savedPlan(t, target)["sections"].([]any)
	if !ok || len(sections) != 2 {
		t.Fatal("nonexample repeated instances not created")
	}
	for i, raw := range sections {
		section := raw.(map[string]any)
		want := []string{"activity-a", "activity-b"}[i]
		if section["id"] != want || section["source"].(map[string]any)["template"] != "working-hours" {
			t.Errorf("reselected identity/source changed: %#v", section)
		}
	}
	if strings.Contains(out.String(), "Guide: \"resource-access-arrange\"") {
		t.Error("retired unsaved selection kept irrelevant preparation advice")
	}
	preserved(t, root)
}
