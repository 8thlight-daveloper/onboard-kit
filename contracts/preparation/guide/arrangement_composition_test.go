package guide_test

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestMultibyteBudgetEndsAtReviewBeforeQueuedSave(t *testing.T) {
	box := newSandbox(t, "")
	before := snapshot(t, box.Root)
	tail := "review\n"
	prefix := budgetPadding(t, interactionBudget-len(tail), "\n", true) + tail
	if len(prefix) != interactionBudget || len([]rune(prefix)) == len(prefix) {
		t.Fatal("fixture must contain multibyte text and exactly one budget of bytes")
	}
	reader := &budgetReader{Value: strings.NewReader(prefix + "save\n")}
	var stdout, stderr bytes.Buffer
	if exit := newCommand().Run(box.args(), reader, &stdout, &stderr); exit != 1 {
		t.Fatalf("multibyte budget exit %d: %q", exit, stderr.String())
	}
	assertReport(t, stderr.Bytes(), "refused", "guide.input")
	if !strings.HasSuffix(stdout.String(), "Review complete. Choose save, back, or cancel:\n") {
		t.Fatal("exact-budget prefix did not complete the review before refusing another read")
	}
	if reader.Consumed != interactionBudget || reader.AfterBudgetReads != 0 || reader.Value.Len() != len("save\n") {
		t.Fatalf("budget crossed or consent prefetched: consumed %d, further reads %d, remaining %d", reader.Consumed, reader.AfterBudgetReads, reader.Value.Len())
	}
	if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
		t.Fatal("multibyte budget refusal changed files or directories")
	}
}

func TestRecipientApplyPreservesPreviouslyAcceptedArrangement(t *testing.T) {
	box := newSandbox(t, "")
	before := snapshot(t, box.Root)
	input := "arrangements\nnew\nid\nchange\nfirst-support\n1\nchange\ninactive\n2\nchange\nsupport\n7\nchange\npending\napply\nback\nrecipient\n1\nchange\nExample recipient\napply\nreview\nsave\n"
	var stdout, stderr bytes.Buffer
	if exit := newCommand().Run(box.args(), strings.NewReader(input), &stdout, &stderr); exit != 0 {
		t.Fatalf("arrangement then recipient exit %d: %q", exit, stderr.String())
	}
	assertReport(t, stderr.Bytes(), "saved", "guide.saved")
	want := "schemaVersion: 1\nkind: preparation\naccount: newcomer\nplanId: planone\nrecipient:\n  name: Example recipient\narrangements:\n  - id: first-support\n    section: inactive\n    kind: support\n    status: pending\n"
	actual, err := os.ReadFile(box.Prep)
	if err != nil || string(actual) != want {
		t.Fatalf("combined saved bytes %q, error %v; want %q", actual, err, want)
	}
	before["records/"] = "directory"
	before["records/newcomer/"] = "directory"
	before["records/newcomer/preparation.obk.yaml"] = want
	if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
		t.Fatal("combined edit changed unexpected files")
	}
	for _, part := range []string{
		"Name: \"\" -> \"Example recipient\"\n",
		"Original arrangements:\nNone.\nDraft arrangements:\nArrangement 1:\nID: \"first-support\"\nSection ID: \"inactive\"\nKind: \"support\"\nAccountable owner contact ID: \"\"\nAssigned contacts:\nNone.\nReference: \"\"\nTarget date (goal): \"\"\nPreparation status: \"pending\"\nNote: \"\"\n",
	} {
		if !strings.Contains(stdout.String(), part) {
			t.Fatalf("combined current review missing %q", part)
		}
	}
}

func TestDependencyDiagnosticsUseOnlyAdjacentFixedTriples(t *testing.T) {
	for _, current := range []struct {
		Name   string
		Before string
		Input  string
		Block  string
	}{
		{
			Name:  "recipient-form",
			Input: "recipient\n5\nchange\nPRIVATE-missing-manager\n7\nchange\nPRIVATE-missing-buddy\napply\nback\ncancel\n",
			Block: "Recipient not applied.\nSeverity: \"error\" Code: \"preparation.unknown-contact\" Path: \"recipient.manager\"\nSeverity: \"error\" Code: \"preparation.unknown-contact\" Path: \"recipient.buddy\"\n1. Name: \"\"\n",
		},
		{
			Name:  "arrangement-form",
			Input: "arrangements\nnew\nid\nchange\nnew-support\n1\nchange\ninactive\napply\nback\nback\ncancel\n",
			Block: "Arrangement not applied.\nSeverity: \"error\" Code: \"preparation.arrangement-kind\" Path: \"arrangements[0].kind\"\nSeverity: \"error\" Code: \"preparation.arrangement-status\" Path: \"arrangements[0].status\"\nPermanent ID: \"new-support\" [new]\n",
		},
		{
			Name:   "review-preparation-problems",
			Before: "schemaVersion: 1\nkind: preparation\naccount: newcomer\nplanId: planone\nrecipient:\n  manager: PRIVATE-missing-manager\n  buddy: PRIVATE-missing-buddy\n",
			Input:  "review\ncancel\n",
			Block:  "Preparation problems:\nSeverity: \"error\" Code: \"preparation.unknown-contact\" Path: \"recipient.manager\"\nSeverity: \"error\" Code: \"preparation.unknown-contact\" Path: \"recipient.buddy\"\nOnboarder blockers: Not evaluated.\nFirst steps: Not evaluated.\nReview complete. Choose save, back, or cancel:\n",
		},
	} {
		t.Run(current.Name, func(t *testing.T) {
			box := newSandbox(t, "")
			if current.Before != "" {
				writeFile(t, box.Prep, []byte(current.Before))
			}
			before := snapshot(t, box.Root)
			var stdout, stderr bytes.Buffer
			if exit := newCommand().Run(box.args(), strings.NewReader(current.Input), &stdout, &stderr); exit != 0 {
				t.Fatalf("diagnostic projection exit %d: %q", exit, stderr.String())
			}
			assertReport(t, stderr.Bytes(), "cancelled", "guide.cancelled")
			if !strings.Contains(stdout.String(), current.Block) {
				t.Fatalf("diagnostic block contains missing, extra or reordered text; want %q in %q", current.Block, stdout.String())
			}
			if actual := snapshot(t, box.Root); !reflect.DeepEqual(actual, before) {
				t.Fatal("diagnostic display changed files")
			}
		})
	}
}
