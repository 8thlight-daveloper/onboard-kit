//go:build darwin || linux

package editor_test

import (
	"strings"
	"testing"
)

const richDisplayPlan = `schemaVersion: 1
id: plan-one
answers:
  sections:
    personal:
      choice: org-a
sections:
  - id: personal
    title: Personal choices
    summary: Use $choice.
    variables:
      - name: choice
        type: string
        audience: onboardee
    steps:
      - id: display-example
        name: Read the example instructions
        owner: Example teammate
        timing: week-one
        open: EXAMPLE_GUIDE.md
        directory: EXAMPLE_WORKSPACE
        notes: |
          First example line.
          Second example line.
        run:
          - EXAMPLE_COMMAND
          - argument with spaces
          - ""
          - repeated-example
          - repeated-example
        env:
          Z_EXAMPLE: last example value
          A_EXAMPLE: 'first "quoted" example value'
        checks:
          - run:
              - EXAMPLE_CHECK
              - --example-only
          - env: EXAMPLE_READY
          - ask: Has the example instruction been reviewed?
      - id: empty-environment
        name: Empty environment control
        env: {}
`

const richDisplaySteps = `Steps: 2
Step: 1
ID: "display-example"
Name: "Read the example instructions"
Owner: "Example teammate"
Timing: "week-one"
Open: "EXAMPLE_GUIDE.md"
Directory: "EXAMPLE_WORKSPACE"
Notes:
  | "First example line."
  | "Second example line."
  | ""
Run: 5
  - "EXAMPLE_COMMAND"
  - "argument with spaces"
  - ""
  - "repeated-example"
  - "repeated-example"
Escalation: 0
Environment: 2
  - Key: "A_EXAMPLE" Value: "first \"quoted\" example value"
  - Key: "Z_EXAMPLE" Value: "last example value"
Checks: 3
Check: 1
Run: 2
  - "EXAMPLE_CHECK"
  - "--example-only"
Ask:
  | ""
Environment question: ""
Check: 2
Run: 0
Ask:
  | ""
Environment question: "EXAMPLE_READY"
Check: 3
Run: 0
Ask:
  | "Has the example instruction been reviewed?"
Environment question: ""
Step: 2
ID: "empty-environment"
Name: "Empty environment control"
Owner: ""
Timing: ""
Open: ""
Directory: ""
Notes:
  | ""
Run: 0
Escalation: 0
Environment: 0
Checks: 0
`

func TestActualStepDisplayPreservesSortedEnvironmentAndEmptyControl(t *testing.T) {
	for _, c := range []struct {
		name   string
		input  string
		blocks int
	}{
		{"current-session", confirm + "exit\n", 1},
		{"current-and-proposed-session", confirm + "answer personal.choice text \"org-b\"\ncancel\n", 2},
	} {
		t.Run(c.name, func(t *testing.T) {
			s := session(t, richDisplayPlan, recordBody)
			before := snapshot(t, s.base)
			clock := &clockFact{value: editEvent}
			got := run(t, s, clock, c.input+"UNREAD_TAIL\n")
			report(t, got, "edit.cancelled", 0, false)
			if got.consumed != len(c.input) || clock.calls != 0 {
				t.Fatal("display-only instructions consumed unread input or sampled an event")
			}
			if strings.Count(got.out, "Steps: 2\n") != c.blocks || strings.Count(got.out, richDisplaySteps+"Questions: 1\n") != c.blocks {
				t.Fatalf("actual step/check/environment block differs from the independent literal grammar: %q", got.out)
			}
			if strings.Count(got.out, "Environment: 2\n") != c.blocks || strings.Count(got.out, "Environment: 0\n") != c.blocks || strings.Count(got.out, "  - Key:") != 2*c.blocks {
				t.Fatal("environment entries were dropped, invented, duplicated or miscounted")
			}
			if strings.Count(got.out, "Current record-backed session.\n") != 1 || strings.Count(got.out, actionPrompt) != 1 || !strings.Contains(got.out, "Instructions are displayed, not executed. Verified is your explicit statement.\n") {
				t.Fatal("the actual current-session nonexecution display was not reached exactly once")
			}
			if c.blocks == 2 {
				if strings.Count(got.out, savePrompt) != 1 || !strings.Contains(got.out, "Before answer: text \"org-a\"\nAfter answer: text \"org-b\"\n") || !strings.Contains(got.out, "Stored override after save: true\n") {
					t.Fatal("the complete proposed session display did not accompany the actual explicit answer review")
				}
			} else if strings.Contains(got.out, savePrompt) {
				t.Fatal("display-only exit unexpectedly entered a save review")
			}
			for _, forbidden := range []string{"  - Key: \"\"", "UNREAD_TAIL", "Record saved.\n", "Record unchanged.\n", "\x1b"} {
				if strings.Contains(got.out+got.err, forbidden) {
					t.Fatalf("literal display emitted an invented entry, unread input or unexpected action: %q", forbidden)
				}
			}
			unchanged(t, s, before)
		})
	}
}
