package starter_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

type starterGapBoundedOutput struct {
	bytes.Buffer
	failed bool
}

func (w *starterGapBoundedOutput) Write(p []byte) (int, error) {
	if w.failed || len(p) > 65536-w.Len() {
		w.failed = true
		return 0, errors.New("independent bounded fixture output exhausted")
	}
	return w.Buffer.Write(p)
}

type starterGapZeroReader struct {
	prefix, tail string
	position     int
	zeroReturned bool
	readsAfter   int
}

func (r *starterGapZeroReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.position < len(r.prefix) {
		n := copy(p, r.prefix[r.position:])
		r.position += n
		return n, nil
	}
	if !r.zeroReturned {
		r.zeroReturned = true
		return 0, nil
	}
	r.readsAfter++
	if r.tail == "" {
		return 0, io.EOF
	}
	n := copy(p, r.tail)
	r.tail = r.tail[n:]
	return n, nil
}

func starterGapExpectedOpening(t *testing.T, s sandbox, stage string) string {
	t.Helper()
	want := literalPaths(t, s, "identity.txt")
	if stage == "identity" {
		return want
	}
	want += literalPaths(t, s, "fresh-review.txt")
	if stage == "review" {
		return want
	}
	return want + "First opened at: \"" + string(openedAt) + "\"\n" + string(fixture(t, "session.txt")) + "Session: read only. Choose exit:\n"
}

func TestStarterGapZeroNilReadIsInputFailureAtEachState(t *testing.T) {
	for _, tc := range []struct {
		stage, prefix, code, status, message string
		clock                                int
	}{
		{"identity", "", "start.input", "refused", "Use complete valid input lines within the displayed limits.", 0},
		{"review", "confirm jules\n", "start.input", "refused", "Use complete valid input lines within the displayed limits.", 0},
		{"session", "confirm jules\nbegin\n", "start.input-opened", "opened", "Opening is complete; the read-only session ended with an input error.", 1},
	} {
		t.Run(tc.stage, func(t *testing.T) {
			s := setup(t)
			before := snapshot(t, s.base)
			clock := 0
			input := &starterGapZeroReader{prefix: tc.prefix, tail: "cancel\nexit\n"}
			var stdout starterGapBoundedOutput
			var stderr bytes.Buffer
			exit := newStarter(runtimeFacts(t, &clock)).Run(s.args(), input, &stdout, &stderr)
			got := streamResult(t, exit, stdout.String(), stderr.Bytes())
			wantReport(t, got, 1, tc.status, tc.code, tc.message)
			if !input.zeroReturned || input.readsAfter != 0 || input.position != len(tc.prefix) || input.tail != "cancel\nexit\n" || clock != tc.clock {
				t.Error("zero-progress read was not an immediate error at the exact consumed boundary")
			}
			if stdout.failed || got.out != starterGapExpectedOpening(t, s, tc.stage) {
				t.Error("zero-progress read changed the exact completed display or repeated prompts")
			}
			if tc.clock == 0 {
				unchanged(t, s, before)
			} else {
				assertBytes(t, s.record(), fixture(t, "begun.yaml"))
				assertIdle(t, s)
			}
		})
	}
}

func TestStarterGapActionsPreserveWhitespaceAndCase(t *testing.T) {
	for _, stage := range []string{"identity", "review", "session"} {
		for _, variant := range []string{"leading-space", "trailing-space", "title-case", "upper-case"} {
			t.Run(stage+"/"+variant, func(t *testing.T) {
				s := setup(t)
				before := snapshot(t, s.base)
				clock := 0
				action, prefix, finish, prompt := "confirm jules", "", "cancel\n", "Identity: choose confirm jules or cancel:\n"
				if stage == "review" {
					action, prefix, prompt = "begin", "confirm jules\n", "Review complete. Choose begin or cancel:\n"
				}
				if stage == "session" {
					action, prefix, finish, prompt = "exit", "confirm jules\nbegin\n", "exit\n", "Session: read only. Choose exit:\n"
				}
				switch variant {
				case "leading-space":
					action = " " + action
				case "trailing-space":
					action += " "
				case "title-case":
					action = strings.ToUpper(action[:1]) + action[1:]
				case "upper-case":
					action = strings.ToUpper(action)
				}
				acceptedInput := prefix + action + "\n" + finish
				input := &countedReader{reader: strings.NewReader(acceptedInput + "exit\nUNCONSUMED_TAIL\n")}
				var stdout starterGapBoundedOutput
				var stderr bytes.Buffer
				exit := newStarter(runtimeFacts(t, &clock)).Run(s.args(), input, &stdout, &stderr)
				got := streamResult(t, exit, stdout.String(), stderr.Bytes())
				if stage == "session" {
					wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
					if clock != 1 {
						t.Error("completed exit control changed first-use event count")
					}
					assertBytes(t, s.record(), fixture(t, "begun.yaml"))
					assertIdle(t, s)
				} else {
					wantReport(t, got, 0, "cancelled", "start.cancelled", "Opening cancelled before begin; no files were changed.")
					if clock != 0 {
						t.Error("nonexact action authorized begin")
					}
					unchanged(t, s, before)
				}
				want := starterGapExpectedOpening(t, s, stage) + "Choose one of the displayed actions.\n" + prompt
				if stdout.failed || got.out != want || input.consumed != len(acceptedInput) {
					t.Error("nonexact action was accepted, wrong prompt repeated, or unused tail consumed")
				}
			})
		}
	}
}

func TestStarterGapIncompleteEOFNeverBecomesAnAction(t *testing.T) {
	for _, tc := range []struct {
		stage, input string
		clock        int
	}{
		{"identity", "confirm jules", 0},
		{"review", "confirm jules\nbegin", 0},
		{"session", "confirm jules\nbegin\nexit", 1},
	} {
		t.Run(tc.stage, func(t *testing.T) {
			s := setup(t)
			before := snapshot(t, s.base)
			clock := 0
			input := &countedReader{reader: strings.NewReader(tc.input)}
			var stdout starterGapBoundedOutput
			var stderr bytes.Buffer
			exit := newStarter(runtimeFacts(t, &clock)).Run(s.args(), input, &stdout, &stderr)
			got := streamResult(t, exit, stdout.String(), stderr.Bytes())
			if tc.clock == 0 {
				wantReport(t, got, 0, "cancelled", "start.cancelled", "Opening cancelled before begin; no files were changed.")
				unchanged(t, s, before)
			} else {
				wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
				assertBytes(t, s.record(), fixture(t, "begun.yaml"))
				assertIdle(t, s)
			}
			if stdout.failed || got.out != starterGapExpectedOpening(t, s, tc.stage) || input.consumed != len(tc.input) || clock != tc.clock {
				t.Error("partial EOF accepted consent, repeated prompts, or changed completed truth")
			}
		})
	}
}
