package starter_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

type pairedTerminalReader struct {
	text        string
	at          int
	terminal    error
	position    int
	paired      bool
	pairedCount int
}

func (r *pairedTerminalReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.paired {
		return 0, r.terminal
	}
	end := r.at
	if end == 0 {
		end = len(r.text)
	}
	if r.position == end {
		return 0, r.terminal
	}
	n := copy(p, r.text[r.position:end])
	r.position += n
	if r.position == end {
		r.paired = true
		r.pairedCount = n
		return n, r.terminal
	}
	return n, nil
}

type borrowedInput struct {
	source     io.Reader
	consumed   int
	closeCalls int
	closed     bool
}

func (r *borrowedInput) Read(p []byte) (int, error) {
	if r.closed {
		return 0, errors.New("invented borrowed input is closed")
	}
	n, err := r.source.Read(p)
	r.consumed += n
	return n, err
}

func (r *borrowedInput) Close() error {
	r.closeCalls++
	r.closed = true
	return errors.New("invented borrowed input close error")
}

type borrowedOutput struct {
	sink       io.Writer
	closeCalls int
	closed     bool
}

func (w *borrowedOutput) Write(p []byte) (int, error) {
	if w.closed {
		return 0, errors.New("invented borrowed output is closed")
	}
	return w.sink.Write(p)
}

func (w *borrowedOutput) Close() error {
	w.closeCalls++
	w.closed = true
	return errors.New("invented borrowed output close error")
}

func streamResult(t *testing.T, exit int, out string, stderr []byte) result {
	t.Helper()
	r := result{exit: exit, out: out, err: string(stderr)}
	if err := json.Unmarshal(stderr, &r.report); err != nil {
		t.Fatal("stream control did not produce one structured report")
	}
	canonical, err := json.Marshal(r.report)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(append(canonical, '\n'), stderr) {
		t.Error("stream control report framing/order differs")
	}
	return r
}

func TestBorrowedStreamClosersAreNeverCalled(t *testing.T) {
	for _, mode := range []string{"help", "cancel", "opened", "display-failure", "reader-failure"} {
		t.Run(mode, func(t *testing.T) {
			s := setup(t)
			before := snapshot(t, s.base)
			clock := 0
			prefix := "confirm jules\nbegin\nexit\n"
			tail := "UNCONSUMED_PRIVATE_TAIL\n"
			args := s.args()
			consumed := len(prefix)
			var source io.Reader = strings.NewReader(prefix + tail)
			if mode == "help" {
				args = []string{"--help"}
				consumed = 0
			}
			if mode == "cancel" {
				prefix = "cancel\n"
				source = strings.NewReader(prefix + tail)
				consumed = len(prefix)
			}
			if mode == "display-failure" {
				consumed = len("confirm jules\n")
			}
			if mode == "reader-failure" {
				consumed = len("confirm jules\n")
				source = &pairedTerminalReader{text: prefix + tail, at: consumed, terminal: errors.New("invented input failure")}
			}
			input := &borrowedInput{source: source}
			var stdout, stderr bytes.Buffer
			var sink io.Writer = &stdout
			failed := &failedWriter{marker: "Review complete.", mode: 2}
			if mode == "display-failure" {
				sink = failed
			}
			out := &borrowedOutput{sink: sink}
			errout := &borrowedOutput{sink: &stderr}
			exit := newStarter(runtimeFacts(t, &clock)).Run(args, input, out, errout)
			if input.closeCalls != 0 || out.closeCalls != 0 || errout.closeCalls != 0 || input.closed || out.closed || errout.closed {
				t.Error("Run closed a caller-owned stream")
			}
			if input.consumed != consumed {
				t.Errorf("input consumed %d bytes, expected exactly %d", input.consumed, consumed)
			}
			if mode == "help" {
				if exit != 0 || !bytes.Equal(stdout.Bytes(), fixture(t, "help.txt")) || stderr.Len() != 0 || clock != 0 {
					t.Error("help borrowed-stream control differs")
				}
				unchanged(t, s, before)
				return
			}
			got := streamResult(t, exit, stdout.String(), stderr.Bytes())
			switch mode {
			case "cancel":
				wantReport(t, got, 0, "cancelled", "start.cancelled", "Opening cancelled before begin; no files were changed.")
			case "opened":
				wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
				assertBytes(t, s.record(), fixture(t, "begun.yaml"))
				assertIdle(t, s)
			case "display-failure":
				wantReport(t, got, 1, "refused", "start.output", "Required output could not be completed; opening did not begin.")
				if !failed.failed {
					t.Error("required display failure was not reached")
				}
			case "reader-failure":
				wantReport(t, got, 1, "refused", "start.input", "Use complete valid input lines within the displayed limits.")
			}
			if mode != "opened" {
				if clock != 0 {
					t.Error("borrowed-stream failure/cancellation sampled event")
				}
				unchanged(t, s, before)
			} else if clock != 1 {
				t.Error("new event was not sampled exactly once")
			}
			if mode == "cancel" || mode == "opened" {
				rest, err := io.ReadAll(input)
				if err != nil || string(rest) != tail {
					t.Error("borrowed input did not retain the exact unused tail")
				}
			}
			if mode != "display-failure" {
				if _, err := out.Write([]byte("caller-owned-output-probe")); err != nil {
					t.Error("borrowed stdout unusable after Run")
				}
			}
			if _, err := errout.Write([]byte("caller-owned-report-probe")); err != nil {
				t.Error("borrowed stderr unusable after Run")
			}
		})
	}
}

func TestDataAndTerminalErrorAreReturnedTogether(t *testing.T) {
	for _, tc := range []struct {
		name, text            string
		at                    int
		terminal              error
		exit, clock           int
		status, code, message string
		fullReview            bool
	}{
		{"eof-complete-identity", "confirm jules\n", 0, io.EOF, 0, 0, "cancelled", "start.cancelled", "Opening cancelled before begin; no files were changed.", true},
		{"eof-partial-identity", "confirm jules", 0, io.EOF, 0, 0, "cancelled", "start.cancelled", "Opening cancelled before begin; no files were changed.", false},
		{"eof-complete-begin", "confirm jules\nbegin\n", 0, io.EOF, 0, 1, "opened", "start.opened", "Opening completed and the read-only session was displayed.", true},
		{"eof-partial-begin", "confirm jules\nbegin", 0, io.EOF, 0, 0, "cancelled", "start.cancelled", "Opening cancelled before begin; no files were changed.", true},
		{"eof-complete-exit", "confirm jules\nbegin\nexit\n", 0, io.EOF, 0, 1, "opened", "start.opened", "Opening completed and the read-only session was displayed.", true},
		{"error-complete-identity", "confirm jules\nbegin\nexit\n", len("confirm jules\n"), errors.New("invented identity read failure"), 1, 0, "refused", "start.input", "Use complete valid input lines within the displayed limits.", false},
		{"error-partial-identity", "confirm jules", 0, errors.New("invented partial read failure"), 1, 0, "refused", "start.input", "Use complete valid input lines within the displayed limits.", false},
		{"error-complete-begin", "confirm jules\nbegin\nexit\n", len("confirm jules\nbegin\n"), errors.New("invented begin read failure"), 1, 0, "refused", "start.input", "Use complete valid input lines within the displayed limits.", true},
		{"error-complete-exit", "confirm jules\nbegin\nexit\n", 0, errors.New("invented completed-session read failure"), 1, 1, "opened", "start.input-opened", "Opening is complete; the read-only session ended with an input error.", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := setup(t)
			before := snapshot(t, s.base)
			clock := 0
			input := &pairedTerminalReader{text: tc.text, at: tc.at, terminal: tc.terminal}
			var stdout, stderr bytes.Buffer
			exit := newStarter(runtimeFacts(t, &clock)).Run(s.args(), input, &stdout, &stderr)
			got := streamResult(t, exit, stdout.String(), stderr.Bytes())
			wantReport(t, got, tc.exit, tc.status, tc.code, tc.message)
			wantConsumed := tc.at
			if wantConsumed == 0 {
				wantConsumed = len(tc.text)
			}
			if !input.paired || input.pairedCount <= 0 || input.position != wantConsumed {
				t.Error("target n>0 plus terminal error was not independently reached")
			}
			if clock != tc.clock {
				t.Error("read error/EOF lost event or publication boundary")
			}
			if strings.Contains(got.out, "Review complete. Choose begin or cancel:\n") != tc.fullReview {
				t.Error("reader condition was masked by earlier grammar or interpreted an errored confirmation")
			}
			if tc.clock == 0 {
				unchanged(t, s, before)
				want := literalPaths(t, s, "identity.txt")
				if tc.fullReview {
					want += literalPaths(t, s, "fresh-review.txt")
				}
				if got.out != want {
					t.Error("prebegin data/error framing changed the exact reached display state")
				}
			} else {
				assertBytes(t, s.record(), fixture(t, "begun.yaml"))
				assertIdle(t, s)
				ending := "First opened at: \"" + string(openedAt) + "\"\n" + string(fixture(t, "session.txt")) + "Session: read only. Choose exit:\n"
				if !strings.HasSuffix(got.out, ending) {
					t.Error("postcompletion stream control did not reach the complete actual session")
				}
			}
		})
	}
}
