package guide_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestIncompleteEOFStillChecksUTF8AndLogicalPayloadSize(t *testing.T) {
	for _, current := range []struct {
		Name            string
		Input           string
		Exit            int
		Status          string
		Code            string
		CompleteFraming bool
	}{
		{Name: "valid-incomplete-action", Input: "review", Exit: 0, Status: "cancelled", Code: "guide.cancelled"},
		{Name: "valid-incomplete-exact-multibyte-boundary", Input: strings.Repeat("Ü", 32_768), Exit: 0, Status: "cancelled", Code: "guide.cancelled"},
		{Name: "invalid-partial-UTF8", Input: "PRIVATE\xff", Exit: 1, Status: "refused", Code: "guide.input"},
		{Name: "incomplete-partial-UTF8-code-point", Input: "PRIVATE\xc3", Exit: 1, Status: "refused", Code: "guide.input"},
		{Name: "oversized-incomplete-logical-payload", Input: strings.Repeat("x", 65_537), Exit: 1, Status: "refused", Code: "guide.input"},
		{Name: "bare-CR-within-payload-bound", Input: strings.Repeat("x", 65_535) + "\r", Exit: 0, Status: "cancelled", Code: "guide.cancelled"},
		{Name: "bare-CR-exceeds-payload-bound", Input: strings.Repeat("x", 65_536) + "\r", Exit: 1, Status: "refused", Code: "guide.input"},
		{Name: "complete-CRLF-exact-payload-bound", Input: strings.Repeat("x", 65_536) + "\r\ncancel\n", Exit: 0, Status: "cancelled", Code: "guide.cancelled", CompleteFraming: true},
	} {
		for _, dataEOF := range []bool{false, true} {
			name := current.Name + "/separate-EOF"
			if dataEOF {
				name = current.Name + "/final-data-and-EOF"
			}
			t.Run(name, func(t *testing.T) {
				box := newSandbox(t, "")
				before := snapshot(t, box.Root)
				reader := &budgetReader{Value: strings.NewReader(current.Input), FinalDataEOF: dataEOF}
				var stdout, stderr bytes.Buffer
				if exit := newCommand().Run(box.args(), reader, &stdout, &stderr); exit != current.Exit {
					t.Fatalf("incomplete/frame input exit %d; want %d; report %q", exit, current.Exit, stderr.String())
				}
				assertReport(t, stderr.Bytes(), current.Status, current.Code)
				if strings.Contains(stdout.String(), "PRIVATE") {
					t.Fatal("rejected incomplete personal input was echoed")
				}
				if current.CompleteFraming && reader.Consumed != len(current.Input) {
					t.Fatal("complete boundary CRLF did not return to read the explicit cancel action")
				}
				if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
					t.Fatal("incomplete/frame input changed original filesystem")
				}
			})
		}
	}
}
