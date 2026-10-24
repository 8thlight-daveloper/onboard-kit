package guide_test

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

const boundaryPreparation = "# Preserve this noncanonical source and its missing final newline.\nschemaVersion: 1\nkind: preparation\naccount: newcomer\nplanId: planone\narrangements: [{id: held, section: inactive, kind: support, contacts: [helper, manager], status: pending, note: Original note}]"

func boundarySandbox(t *testing.T) sandbox {
	t.Helper()
	box := newSandbox(t, "")
	writeFile(t, box.Prep, []byte(boundaryPreparation))
	return box
}

func TestValidLocalArrangementChangesStayDiscardable(t *testing.T) {
	for _, current := range []struct {
		Name  string
		Local string
	}{
		{"nested-apply-then-enclosing-back", "4\nclear\napply\nback\n"},
		{"nested-back-after-valid-scalar-change", "8\nchange\nDiscarded note\n4\nclear\nback\nback\n"},
		{"enclosing-back-after-valid-scalar-change", "8\nchange\nDiscarded note\nback\n"},
		{"nested-helper-replacement-then-enclosing-back", "4\nclear\nadd\nmanager\napply\nback\n"},
	} {
		t.Run(current.Name, func(t *testing.T) {
			box := boundarySandbox(t)
			before := snapshot(t, box.Root)
			original, err := os.Stat(box.Prep)
			if err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			input := "arrangements\nedit\n1\n" + current.Local + "back\nreview\nsave\n"
			if exit := newCommand().Run(box.args(), strings.NewReader(input), &stdout, &stderr); exit != 0 {
				t.Fatalf("discardable valid form exit %d: %q", exit, stderr.String())
			}
			assertReport(t, stderr.Bytes(), "unchanged", "guide.unchanged")
			if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
				t.Fatal("local form return changed preparation or another path")
			}
			after, err := os.Stat(box.Prep)
			if err != nil || !os.SameFile(original, after) {
				t.Fatalf("discarded form replaced preparation inode: %v", err)
			}
			want := "Draft arrangements:\nArrangement 1:\nID: \"held\"\nSection ID: \"inactive\"\nKind: \"support\"\nAccountable owner contact ID: \"\"\nAssigned contacts:\n1. \"helper\"\n2. \"manager\"\nReference: \"\"\nTarget date (goal): \"\"\nPreparation status: \"pending\"\nNote: \"Original note\"\n"
			if !strings.Contains(stdout.String(), want) {
				t.Fatal("review did not retain the whole original arrangement")
			}
			if strings.Contains(stdout.String(), "Arrangement not applied.") {
				t.Fatal("valid pending proposal was refused, masking an early apply")
			}
		})
	}
}

type ordinalBoundaryReader struct {
	Value    *strings.Reader
	Output   *bytes.Buffer
	Observed string
	Reached  bool
	Consumed int
}

func (r *ordinalBoundaryReader) Read(value []byte) (int, error) {
	if len(value) == 0 {
		return 0, nil
	}
	if r.Value.Len() == 0 {
		if !r.Reached {
			r.Reached = true
			r.Observed = r.Output.String()
		}
		return 0, io.EOF
	}
	count, err := r.Value.Read(value[:1])
	r.Consumed += count
	return count, err
}

func TestNoncanonicalOrdinalsRetryBeforeAnyFurtherInput(t *testing.T) {
	for _, menu := range []struct {
		Name   string
		Prefix string
		Prompt string
	}{
		{"arrangement-edit", "arrangements\nedit\n", "Choose arrangement number or back:\n"},
		{"arrangement-remove", "arrangements\nremove\n", "Choose arrangement number or back:\n"},
		{"helper-remove", "arrangements\nedit\n1\n4\nremove\n", "Choose helper number or back:\n"},
	} {
		for _, bad := range []struct{ Name, Value string }{
			{"empty", ""},
			{"zero", "0"},
			{"leading-zero", "01"},
			{"plus-sign", "+1"},
			{"minus-sign", "-1"},
			{"leading-space", " 1"},
			{"trailing-space", "1 "},
			{"leading-tab", "\t1"},
			{"trailing-tab", "1\t"},
			{"leading-nonbreaking-space", "\u00a01"},
			{"trailing-nonbreaking-space", "1\u00a0"},
			{"suffix", "1x"},
			{"decimal", "1.0"},
			{"exponent", "1e0"},
			{"overflow", "18446744073709551617"},
			{"arabic-indic", "١"},
			{"fullwidth", "１"},
			{"out-of-range", "3"},
			{"permanent-id-is-not-ordinal", "held"},
		} {
			t.Run(menu.Name+"/"+bad.Name, func(t *testing.T) {
				box := boundarySandbox(t)
				before := snapshot(t, box.Root)
				var stdout, stderr bytes.Buffer
				input := menu.Prefix + bad.Value + "\n"
				reader := &ordinalBoundaryReader{Value: strings.NewReader(input), Output: &stdout}
				if exit := newCommand().Run(box.args(), reader, &stdout, &stderr); exit != 0 {
					t.Fatalf("invalid ordinal exit %d: %q", exit, stderr.String())
				}
				assertReport(t, stderr.Bytes(), "cancelled", "guide.cancelled")
				want := "Choose one of the displayed actions.\n" + menu.Prompt
				if !reader.Reached || reader.Consumed != len(input) || !strings.HasSuffix(reader.Observed, want) {
					t.Fatalf("next input read did not follow exact selection retry: reached %t, consumed %d/%d, output %q", reader.Reached, reader.Consumed, len(input), reader.Observed)
				}
				if strings.Count(reader.Observed, menu.Prompt) != 2 {
					t.Fatalf("selection prompt count is not initial plus one retry: %q", reader.Observed)
				}
				if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
					t.Fatal("invalid ordinal changed files")
				}
			})
		}
	}
}
