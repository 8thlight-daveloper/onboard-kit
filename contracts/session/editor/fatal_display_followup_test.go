//go:build darwin || linux

package editor_test

import (
	"strconv"
	"strings"
	"testing"
)

func TestFatalInspectionDisplaysProblemsWithoutInventingAnActionableSession(t *testing.T) {
	for _, c := range []struct {
		name        string
		plan        string
		record      string
		preparation string
		problems    []string
	}{
		{
			name:        "preparation-codec",
			plan:        planBody,
			record:      recordBody,
			preparation: prepBody + "recipient:\n  name: PRIVATE_PREPARATION_NAME\n  buddy: [PRIVATE_INVALID_BUDDY]\n",
			problems:    []string{"  - Stage: \"preparation.decode\" Code: \"format.wrong-type\" Severity: \"error\"\n"},
		},
		{
			name:        "preparation-identity",
			plan:        planBody,
			record:      recordBody,
			preparation: strings.Replace(prepBody, "account: newcomer", "account: PRIVATE_OTHER_ACCOUNT", 1) + "recipient:\n  name: PRIVATE_PREPARATION_NAME\n",
			problems:    []string{"  - Stage: \"preparation.validate\" Code: \"preparation.account-mismatch\" Severity: \"error\"\n"},
		},
		{
			name:        "record-ownership",
			plan:        strings.Replace(planBody, "audience: onboardee", "audience: onboarder", 1),
			record:      recordBody + "answers:\n  sections:\n    personal:\n      choice: PRIVATE_RECORD_ANSWER\n",
			preparation: prepBody,
			problems:    []string{"  - Stage: \"record.ownership\" Code: \"session.record-owned-answer\" Severity: \"error\"\n"},
		},
		{
			name:        "effective-answer-validation",
			plan:        planBody,
			record:      recordBody + "answers:\n  sections:\n    personal:\n      choice: false\nprivateRecordData: PRIVATE_UNRELATED_RECORD_TEXT\n",
			preparation: prepBody,
			problems:    []string{"  - Stage: \"effective.validate\" Code: \"answer.wrong-kind\" Severity: \"error\"\n"},
		},
		{
			name:        "mixed-invalid-and-stale-preparation",
			plan:        planBody,
			record:      recordBody,
			preparation: prepBody + "recipient:\n  name: PRIVATE_PREPARATION_NAME\n  plannedStartDate: \"2026-02-30\"\n  buddy: missing-contact\n",
			problems: []string{
				"  - Stage: \"preparation.validate\" Code: \"preparation.date\" Severity: \"error\"\n",
				"  - Stage: \"preparation.validate\" Code: \"preparation.unknown-contact\" Severity: \"error\"\n",
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			s := session(t, c.plan, c.record)
			put(t, s.preparation, c.preparation)
			before := snapshot(t, s.base)
			clock := &clockFact{value: editEvent}
			got := run(t, s, clock, confirm+"answer personal.choice text \"UNREAD_ANSWER\"\nsave\nUNREAD_TAIL\n")
			report(t, got, "edit.review", 1, false)
			if got.consumed != len(confirm) || clock.calls != 0 {
				t.Fatal("fatal inspection consumed an action/save byte or sampled an event")
			}
			wantProblems := "Session problems: " + strconv.Itoa(len(c.problems)) + "\n" + strings.Join(c.problems, "") + "Review the selected source files to resolve any reported errors.\n"
			if !strings.Contains(got.out, "Identity: type confirm \"newcomer\" or cancel:\n") || !strings.HasSuffix(got.out, wantProblems) || strings.Count(got.out, "Session problems:") != 1 {
				t.Fatalf("fatal inspection did not display its complete safe problem batch: %q", got.out)
			}
			for _, forbidden := range []string{
				"Current record-backed session.\n", "Activities:", "Questions:", "Recorded outcomes:",
				"Support available:", "Current support is unavailable;", "Plan reference matches current bytes:",
				"Choose one change; nothing here runs a command or check.\n", actionPrompt,
				"Selected record change", savePrompt, "Record saved.\n", "Record unchanged.\n", "PRIVATE_", "UNREAD_",
			} {
				if strings.Contains(got.out+got.err, forbidden) {
					t.Fatalf("fatal inspection displayed a fabricated session, action or private payload: %q", forbidden)
				}
			}
			unchanged(t, s, before)
		})
	}
}

func TestStaleSupportAloneStillDisplaysTheRealEditableSession(t *testing.T) {
	s := session(t, planBody, recordBody)
	put(t, s.preparation, prepBody+"recipient:\n  buddy: missing-contact\n")
	before := snapshot(t, s.base)
	clock := &clockFact{value: editEvent}
	got := run(t, s, clock, confirm+"exit\nUNREAD_TAIL\n")
	report(t, got, "edit.cancelled", 0, false)
	if got.consumed != len(confirm+"exit\n") || clock.calls != 0 {
		t.Fatal("nonfatal support consumed beyond its explicit exit or sampled an event")
	}
	for _, required := range []string{
		"Current record-backed session.\n", "Activities: 1\n", "Questions: 1\n",
		"Support available: false\n", "Current support is unavailable; ask the onboarding owner to repair preparation.\n",
		"Support problems: 1\n  - Stage: \"preparation.validate\" Code: \"preparation.unknown-contact\" Severity: \"error\"\n",
		"Session problems: 0\n", actionPrompt,
	} {
		if !strings.Contains(got.out, required) {
			t.Fatalf("nonfatal support lost the actual editable session or safe support problem: %q", required)
		}
	}
	if strings.Contains(got.out, "Activities: 0\n") || strings.Contains(got.out+got.err, "UNREAD_TAIL") {
		t.Fatal("nonfatal support fabricated an empty session or displayed unread input")
	}
	unchanged(t, s, before)
}

func TestFatalFinalizationDisplaysActualSafeDiagnosticsWithoutAnotherSession(t *testing.T) {
	s := session(t, outcomePlan, recordBody)
	before := snapshot(t, s.base)
	clock := &clockFact{value: "PRIVATE_INVALID_EVENT"}
	prefix := confirm + "outcome access failed \"\" \"\"\nsave\n"
	got := run(t, s, clock, prefix+"UNREAD_TAIL\n")
	report(t, got, "edit.review", 1, false)
	if got.consumed != len(prefix) || clock.calls != 1 {
		t.Fatal("fatal finalization resampled an event or consumed beyond explicit save")
	}
	wantProblems := "Session problems: 1\n  - Stage: \"outcome\" Code: \"outcome.timestamp-invalid\" Severity: \"error\"\nReview the selected source files to resolve any reported errors.\n"
	if !strings.HasSuffix(got.out, wantProblems) {
		t.Fatalf("fatal finalization omitted or altered actual safe diagnostics: %q", got.out)
	}
	for _, once := range []string{
		"Current record-backed session.\n", actionPrompt,
		"Selected record change — not saved.\n", savePrompt,
	} {
		if strings.Count(got.out, once) != 1 {
			t.Fatalf("the actual review was missing or a fatal result reopened a session: %q", once)
		}
	}
	if !strings.Contains(got.out, "New outcome time: sampled after save\n") || !strings.Contains(got.out, "Activities: 1\n") {
		t.Fatal("finalization did not follow an actual outcome/session review")
	}
	for _, forbidden := range []string{
		"Activities: 0\n", "Record saved.\n", "Record unchanged.\n",
		"PRIVATE_INVALID_EVENT", "UNREAD_TAIL", "2000-01-01T00:00:00Z",
	} {
		if strings.Contains(got.out+got.err, forbidden) {
			t.Fatalf("fatal finalization displayed fabricated success, an empty session or private event: %q", forbidden)
		}
	}
	unchanged(t, s, before)
}
