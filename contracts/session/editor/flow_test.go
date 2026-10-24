//go:build darwin || linux

package editor_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestActualAnswerAndFollowPlanAreRecordOnlyAtomicWrites(t *testing.T) {
	s := session(t, planBody, recordBody)
	if err := os.Chmod(s.record, 0644); err != nil {
		t.Fatal(err)
	}
	anchor := filepath.Join(s.base, "old-record-anchor")
	if err := os.Link(s.record, anchor); err != nil {
		t.Fatal(err)
	}
	before, held := snapshot(t, s.base), nativeObjectAt(t, s.record)
	clock := &clockFact{value: editEvent}
	prefix := confirm + "answer personal.choice text \"org-a\"\nsave\n"
	got := run(t, s, clock, prefix+"UNREAD_TAIL\n")
	report(t, got, "edit.saved", 0, true)
	if clock.calls != 0 || got.consumed != len(prefix) || !strings.Contains(got.out, "Stored override after save: true\n") {
		t.Fatal("answer used event, lost explicit equal ownership, or read ahead")
	}
	savedRecord(t, s, held, ownedRecord)
	companions(t, s, before)
	if !reflect.DeepEqual(nativeObjectAt(t, anchor), held) {
		t.Fatal("old hardlink anchor was overwritten in place")
	}
	put(t, s.plan, strings.Replace(planBody, "choice: org-a", "choice: org-b", 1))
	before, held = snapshot(t, s.base), nativeObjectAt(t, s.record)
	got = run(t, s, clock, confirm+"follow-plan personal.choice\nsave\n")
	report(t, got, "edit.saved", 0, true)
	if clock.calls != 0 || !strings.Contains(got.out, "After answer: text \"org-b\"\n") || !strings.Contains(got.out, "Stored override after save: false\n") {
		t.Fatal("follow-plan did not reveal current shared answer or treated historical registration hash as live seal")
	}
	savedRecord(t, s, held, followerRecord)
	companions(t, s, before)
}

func TestActualOutcomeUsesOneInjectedEventAndLiteralCurrentHash(t *testing.T) {
	s := session(t, outcomePlan, recordBody)
	before, held := snapshot(t, s.base), nativeObjectAt(t, s.record)
	clock := &clockFact{value: editEvent}
	got := run(t, s, clock, confirm+"outcome access verified \"\" \"\"\nsave\n")
	report(t, got, "edit.saved", 0, true)
	if clock.calls != 1 || !strings.Contains(got.out, "New outcome time: sampled after save\n") || strings.Contains(got.out, "2000-01-01T00:00:00Z") {
		t.Fatal("outcome event is dishonest or resampled")
	}
	savedRecord(t, s, held, verifiedRecord)
	companions(t, s, before)
}

func TestFailedAndDeferredAreRealExplicitRecordOutcomes(t *testing.T) {
	for _, kind := range []string{"failed", "deferred"} {
		t.Run(kind, func(t *testing.T) {
			s := session(t, outcomePlan, recordBody)
			before, held := snapshot(t, s.base), nativeObjectAt(t, s.record)
			clock := &clockFact{value: editEvent}
			got := run(t, s, clock, confirm+"outcome access "+kind+" \"\" \"\"\nsave\n")
			report(t, got, "edit.saved", 0, true)
			if clock.calls != 1 {
				t.Fatal("explicit outcome was not stamped exactly once")
			}
			savedRecord(t, s, held, strings.Replace(verifiedRecord, "outcome: verified", "outcome: "+kind, 1))
			companions(t, s, before)
		})
	}
}

func TestCurrentPrerequisiteGateRefusesVerifiedBeforeEventAndSaveInput(t *testing.T) {
	plan := "schemaVersion: 1\nid: plan-one\nsections:\n  - id: access\n    title: Access\n    summary: Request access\n    requires:\n      - a\n  - id: a\n    title: Prerequisite\n    summary: Complete prerequisite\n"
	s := session(t, plan, recordBody)
	before := snapshot(t, s.base)
	clock := &clockFact{value: editEvent}
	prefix := confirm + "outcome access verified \"\" \"\"\n"
	got := run(t, s, clock, prefix+"save\n")
	report(t, got, "edit.review", 1, false)
	if got.consumed != len(prefix) || clock.calls != 0 || !strings.Contains(got.out, "outcome.prerequisite-unverified") {
		t.Fatal("verified bypassed actual current prerequisite gate or consumed save/time")
	}
	unchanged(t, s, before)
}

func TestNoOpAndAllConsentStopsRetainExactNativeObjects(t *testing.T) {
	for _, c := range []struct {
		name, input, code string
		exit              int
	}{
		{"identity-cancel", "cancel\n", "edit.cancelled", 0},
		{"identity-partial", "confirm \"newcomer\"", "edit.cancelled", 0},
		{"action-exit", confirm + "exit\n", "edit.cancelled", 0},
		{"save-cancel", confirm + "answer personal.choice text \"org-a\"\ncancel\n", "edit.cancelled", 0},
		{"save-partial", confirm + "answer personal.choice text \"org-a\"\nsave", "edit.cancelled", 0},
		{"no-op", confirm + "follow-plan personal.choice\nsave\n", "edit.unchanged", 0},
		{"invalid-json-surrogate", confirm + "answer personal.choice text \"\\ud800\"\n", "edit.input", 1},
		{"blank-not-default", confirm + "\n", "edit.input", 1},
		{"empty-text-not-clear", confirm + "answer personal.choice text \"\"\n", "edit.review", 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			s := session(t, planBody, "# raw record formatting\n"+recordBody)
			before := snapshot(t, s.base)
			clock := &clockFact{value: editEvent}
			got := run(t, s, clock, c.input)
			report(t, got, c.code, c.exit, false)
			if clock.calls != 0 {
				t.Fatal("unaccepted/no-op answer path sampled clock")
			}
			unchanged(t, s, before)
		})
	}
}

func TestInvalidRuntimeEventNeverFallsBackOrPublishes(t *testing.T) {
	s := session(t, outcomePlan, recordBody)
	before := snapshot(t, s.base)
	clock := &clockFact{value: "invalid-event"}
	got := run(t, s, clock, confirm+"outcome access failed \"note\" \"\"\nsave\n")
	report(t, got, "edit.review", 1, false)
	if clock.calls != 1 {
		t.Fatal("invalid trusted event was normalized or resampled")
	}
	unchanged(t, s, before)
}

func TestIdentityCancellationDoesNotDecodeUnconsentedBodies(t *testing.T) {
	s := session(t, "NOT_A_PLAN_BODY", "NOT_A_RECORD_BODY")
	put(t, s.preparation, "NOT_A_PREPARATION_BODY")
	before := snapshot(t, s.base)
	clock := &clockFact{value: editEvent}
	got := run(t, s, clock, "cancel\n")
	report(t, got, "edit.cancelled", 0, false)
	if clock.calls != 0 || got.consumed != len("cancel\n") {
		t.Fatal("identity cancellation consumed unconsented flow")
	}
	unchanged(t, s, before)
}
