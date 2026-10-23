package guide_test

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
)

const interactionBudget = 1_048_576

type budgetReader struct {
	Value            *strings.Reader
	Consumed         int
	AfterBudgetReads int
	FinalDataEOF     bool
}

func (r *budgetReader) Read(value []byte) (int, error) {
	if r.Consumed >= interactionBudget {
		r.AfterBudgetReads++
	}
	count, err := r.Value.Read(value)
	r.Consumed += count
	if count > 0 && r.Value.Len() == 0 && r.FinalDataEOF {
		return count, io.EOF
	}
	return count, err
}

func budgetPadding(t *testing.T, count int, framing string, multibyte bool) string {
	t.Helper()
	var value strings.Builder
	for count > 0 {
		lineBytes := min(count, 65_536+len(framing))
		if lineBytes < len(framing) {
			t.Fatal("fixture cannot frame final padding line")
		}
		payloadBytes := lineBytes - len(framing)
		if multibyte {
			value.WriteString(strings.Repeat("Ü", payloadBytes/2))
			if payloadBytes%2 != 0 {
				value.WriteByte('x')
			}
		} else {
			value.WriteString(strings.Repeat("x", payloadBytes))
		}
		value.WriteString(framing)
		count -= lineBytes
	}
	return value.String()
}

func TestTotalInteractionBudget(t *testing.T) {
	for _, current := range []struct {
		Name         string
		Tail         string
		Suffix       string
		Framing      string
		Unicode      bool
		FinalDataEOF bool
		Exit         int
		Status       string
		Code         string
	}{
		{Name: "save-exact-LF-leaves-trailing-input", Tail: "review\nsave\n", Suffix: "PRIVATE unused trailing input\n", Framing: "\n", Exit: 0, Status: "saved", Code: "guide.saved"},
		{Name: "save-exact-CRLF-charges-both-framing-bytes", Tail: "review\r\nsave\r\n", Framing: "\r\n", Unicode: true, Exit: 0, Status: "saved", Code: "guide.saved"},
		{Name: "save-exact-final-byte-with-EOF", Tail: "review\nsave\n", Framing: "\n", FinalDataEOF: true, Exit: 0, Status: "saved", Code: "guide.saved"},
		{Name: "complete-review-budget-ends-before-queued-save", Tail: "review\n", Suffix: "save\n", Framing: "\n", Exit: 1, Status: "refused", Code: "guide.input"},
		{Name: "complete-line-budget-precedes-unread-EOF", Tail: "review\n", Framing: "\n", Exit: 1, Status: "refused", Code: "guide.input"},
		{Name: "complete-line-budget-precedes-queued-cancel", Tail: "review\n", Suffix: "cancel\n", Framing: "\n", Exit: 1, Status: "refused", Code: "guide.input"},
		{Name: "partial-save-budget-stops-before-next-byte", Tail: "review\nsa", Suffix: "ve\n", Framing: "\n", Exit: 1, Status: "refused", Code: "guide.input"},
		{Name: "partial-save-budget-final-data-EOF-is-not-cancel", Tail: "review\nsa", Framing: "\n", FinalDataEOF: true, Exit: 1, Status: "refused", Code: "guide.input"},
		{Name: "complete-cancel-at-budget-is-cancel", Tail: "cancel\n", Suffix: "save\n", Framing: "\n", Exit: 0, Status: "cancelled", Code: "guide.cancelled"},
	} {
		t.Run(current.Name, func(t *testing.T) {
			box := newSandbox(t, "")
			before := snapshot(t, box.Root)
			prefix := budgetPadding(t, interactionBudget-len(current.Tail), current.Framing, current.Unicode) + current.Tail
			if len(prefix) != interactionBudget {
				t.Fatal("invalid exact-budget fixture")
			}
			reader := &budgetReader{Value: strings.NewReader(prefix + current.Suffix), FinalDataEOF: current.FinalDataEOF}
			var stderr bytes.Buffer
			if exit := newCommand().Run(box.args(), reader, io.Discard, &stderr); exit != current.Exit {
				t.Fatalf("exit %d; want %d; report %q", exit, current.Exit, stderr.String())
			}
			assertReport(t, stderr.Bytes(), current.Status, current.Code)
			if reader.Consumed != interactionBudget {
				t.Fatalf("consumed %d bytes; want exactly %d without suffix prefetch", reader.Consumed, interactionBudget)
			}
			if reader.AfterBudgetReads != 0 {
				t.Fatal("called input reader after budget exhaustion, including an unread EOF check")
			}
			want := before
			if current.Status == "saved" {
				want["records/"] = "directory"
				want["records/newcomer/"] = "directory"
				want["records/newcomer/preparation.obk.yaml"] = string(fixture(t, "identity.yaml"))
			}
			if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, want) {
				t.Fatal("budget handling changed unexpected files")
			}
		})
	}
}

func TestBudgetDoesNotInspectUnusedSuffixAndPreservesBelowLimitEOF(t *testing.T) {
	for _, current := range []struct {
		Input, Status, Code string
		Consumed            int
		Exit                int
		FinalDataEOF        bool
	}{
		{Input: "cancel\n" + strings.Repeat("PRIVATE", 200_000), Status: "cancelled", Code: "guide.cancelled", Consumed: len("cancel\n"), Exit: 0},
		{Input: "review\nsave", Status: "cancelled", Code: "guide.cancelled", Consumed: len("review\nsave"), Exit: 0, FinalDataEOF: true},
		{Input: "review\nsave\n", Status: "saved", Code: "guide.saved", Consumed: len("review\nsave\n"), Exit: 0, FinalDataEOF: true},
	} {
		box := newSandbox(t, "")
		before := snapshot(t, box.Root)
		reader := &budgetReader{Value: strings.NewReader(current.Input), FinalDataEOF: current.FinalDataEOF}
		var stderr bytes.Buffer
		if exit := newCommand().Run(box.args(), reader, io.Discard, &stderr); exit != current.Exit {
			t.Fatalf("below-budget exit %d; report %q", exit, stderr.String())
		}
		assertReport(t, stderr.Bytes(), current.Status, current.Code)
		if reader.Consumed != current.Consumed {
			t.Fatalf("consumed %d bytes; want %d", reader.Consumed, current.Consumed)
		}
		if current.Status == "saved" {
			before["records/"] = "directory"
			before["records/newcomer/"] = "directory"
			before["records/newcomer/preparation.obk.yaml"] = string(fixture(t, "identity.yaml"))
		}
		if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
			t.Fatal("below-budget EOF/cancel changed unexpected files")
		}
	}
}
