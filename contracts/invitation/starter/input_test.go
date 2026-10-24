package starter_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

type countedReader struct {
	reader   io.Reader
	consumed int
	at       int
	hook     func()
}

func (r *countedReader) Read(p []byte) (int, error) {
	if r.hook != nil && r.consumed >= r.at {
		hook := r.hook
		r.hook = nil
		hook()
	}
	n, err := r.reader.Read(p)
	r.consumed += n
	return n, err
}

type failedWriter struct {
	accepted bytes.Buffer
	marker   string
	mode     int
	failed   bool
}

func (w *failedWriter) Write(p []byte) (int, error) {
	if w.failed {
		return 0, errors.New("invented writer failure")
	}
	if strings.Contains(w.accepted.String()+string(p), w.marker) {
		w.failed = true
		switch w.mode {
		case 0:
			return 0, errors.New("invented writer failure")
		case 1:
			if len(p) > 0 {
				return len(p) - 1, nil
			}
			return 0, nil
		default:
			w.accepted.Write(p)
			return len(p), errors.New("invented writer failure")
		}
	}
	return w.accepted.Write(p)
}

func TestRequiredDisplayFailureDoesNotConsumeQueuedConsent(t *testing.T) {
	for _, stage := range []struct {
		marker, input string
		consumed      int
	}{
		{"Identity: choose", "confirm jules\nbegin\nexit\n", 0},
		{"Review complete.", "confirm jules\nbegin\nexit\n", len("confirm jules\n")},
	} {
		for mode := 0; mode < 3; mode++ {
			s := setup(t)
			before := snapshot(t, s.base)
			clock := 0
			in := &countedReader{reader: strings.NewReader(stage.input)}
			out := &failedWriter{marker: stage.marker, mode: mode}
			var stderr bytes.Buffer
			exit := newStarter(runtimeFacts(t, &clock)).Run(s.args(), in, out, &stderr)
			if exit != 1 || in.consumed != stage.consumed || clock != 0 {
				t.Error("failed display consumed queued consent or sampled event")
			}
			if !strings.Contains(stderr.String(), `"code":"start.output"`) {
				t.Error("missing prebegin output refusal")
			}
			unchanged(t, s, before)
		}
	}
}

func TestPartialAndInvalidInputBoundaries(t *testing.T) {
	for _, tc := range []struct {
		input string
		code  string
		exit  int
	}{
		{"confirm jules\r\ncancel\r\n", "start.cancelled", 0},
		{"confirm jules\r", "start.cancelled", 0},
		{string([]byte{0xff}), "start.input", 1},
		{strings.Repeat("x", 65537), "start.input", 1},
		{strings.Repeat("Ü", 32769), "start.input", 1},
		{strings.Repeat("Ü", 32768) + "\r\ncancel\n", "start.cancelled", 0},
		{strings.Repeat("x", 65536) + "\r", "start.input", 1},
	} {
		s := setup(t)
		before := snapshot(t, s.base)
		clock := 0
		r := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), tc.input)
		if r.exit != tc.exit || r.report.Code != tc.code || clock != 0 {
			t.Error("input framing/UTF8/byte boundary differs")
		}
		unchanged(t, s, before)
	}
}

func fillBudget(prefix, action string) string {
	remaining := 1048576 - len(prefix) - len(action)
	var out strings.Builder
	out.WriteString(prefix)
	for remaining > 0 {
		n := remaining
		if n > 65537 {
			n = 65537
		}
		out.WriteString(strings.Repeat("x", n-1))
		out.WriteByte('\n')
		remaining -= n
	}
	out.WriteString(action)
	return out.String()
}

func TestTotalInputBudgetAndUnusedTrailingBytes(t *testing.T) {
	for _, tc := range []struct {
		input, code string
		exit, clock int
	}{
		{fillBudget("", "cancel\n"), "start.cancelled", 0, 0},
		{strings.Repeat("\n", 1048576) + "cancel\n", "start.input", 1, 0},
		{fillBudget("confirm jules\n", "begin\n"), "start.limit-opened", 1, 1},
		{"confirm jules\nbegin\nexit\n" + strings.Repeat("x", 1048577), "start.opened", 0, 1},
	} {
		s := setup(t)
		clock := 0
		var stderr bytes.Buffer
		in := &countedReader{reader: strings.NewReader(tc.input)}
		exit := newStarter(runtimeFacts(t, &clock)).Run(s.args(), in, io.Discard, &stderr)
		if exit != tc.exit || clock != tc.clock || !strings.Contains(stderr.String(), `"code":"`+tc.code+`"`) {
			t.Error("total input accounting/completion truth differs")
		}
		if in.consumed > 1048576 {
			t.Error("input was read past total budget")
		}
		if tc.clock == 0 {
			if _, err := os.Stat(s.home); !os.IsNotExist(err) {
				t.Error("prebegin budget refusal created home")
			}
		}
	}
}

func TestFreshGuardsRecheckAfterCompleteBegin(t *testing.T) {
	for _, change := range []string{"bytes", "inode", "security", "new-target", "archive"} {
		t.Run(change, func(t *testing.T) {
			s := setup(t)
			put(t, s.record(), fixture(t, "existing.yaml"))
			clock := 0
			in := &countedReader{reader: strings.NewReader("confirm jules\nbegin\nexit\n"), at: len("confirm jules\n")}
			in.hook = func() {
				switch change {
				case "bytes":
					put(t, s.record(), append(fixture(t, "existing.yaml"), []byte("# independent owner edit\n")...))
				case "inode":
					old := s.record() + ".held"
					if err := os.Rename(s.record(), old); err != nil {
						t.Fatal(err)
					}
					put(t, s.record(), fixture(t, "existing.yaml"))
				case "security":
					if err := os.Chmod(s.record(), 0644); err != nil {
						t.Fatal(err)
					}
				case "new-target":
					put(t, s.prep(), []byte("owner-created preparation\n"))
				case "archive":
					put(t, s.invitation, []byte("owner changed archive\n"))
				}
			}
			var out, stderr bytes.Buffer
			exit := newStarter(runtimeFacts(t, &clock)).Run(s.args(), in, &out, &stderr)
			if exit != 1 || clock != 0 || !strings.Contains(stderr.String(), `"code":"start.context"`) {
				t.Error("stale source guard allowed publication/event")
			}
			if _, err := os.Stat(s.home); !os.IsNotExist(err) {
				t.Error("guard failure bootstrapped home")
			}
		})
	}
}

func TestInvalidClockIsRefusalNotReplacement(t *testing.T) {
	s := setup(t)
	clock := 0
	facts := runtimeFacts(t, &clock)
	facts.Clock = func() model.Timestamp { clock++; return "invented-invalid-event" }
	r := invoke(t, newStarter(facts), s.args(), "confirm jules\nbegin\n")
	wantReport(t, r, 1, "refused", "start.review", "Resolve the displayed invitation review problem before beginning.")
	if clock != 1 {
		t.Error("invalid clock was resampled")
	}
	if _, err := os.Stat(s.record()); !os.IsNotExist(err) {
		t.Error("invalid event published record")
	}
}

func TestPostCompletionWriterAndInputFailuresRemainOpened(t *testing.T) {
	for _, mode := range []string{"display", "input", "report"} {
		t.Run(mode, func(t *testing.T) {
			s := setup(t)
			clock := 0
			command := newStarter(runtimeFacts(t, &clock))
			var stderr bytes.Buffer
			var out io.Writer = io.Discard
			var input io.Reader = strings.NewReader("confirm jules\nbegin\nexit\n")
			var errorOutput io.Writer = &stderr
			if mode == "display" {
				out = &failedWriter{marker: "Read-only onboarding session.", mode: 2}
			}
			if mode == "input" {
				input = strings.NewReader("confirm jules\nbegin\n" + string([]byte{0xff}) + "\n")
			}
			if mode == "report" {
				errorOutput = &failedWriter{marker: `"code":"start.opened"`, mode: 2}
			}
			exit := command.Run(s.args(), input, out, errorOutput)
			if exit != 1 || clock != 1 {
				t.Error("postcompletion error lost completion or resampled")
			}
			assertBytes(t, s.record(), fixture(t, "begun.yaml"))
			assertIdle(t, s)
			if mode != "report" && !strings.Contains(stderr.String(), `"code":"start.`+map[string]string{"display": "display-opened", "input": "input-opened"}[mode]+`"`) {
				t.Error("postcompletion report truth differs")
			}
		})
	}
}

func TestLexicalAndNativeAliasRefusal(t *testing.T) {
	for _, kind := range []string{"dotdot", "nul", "root-symlink", "archive-record-hardlink", "parent-symlink"} {
		t.Run(kind, func(t *testing.T) {
			s := setup(t)
			args := s.args()
			switch kind {
			case "dotdot":
				args[3] = filepath.Join(s.root, "child") + "/../.."
			case "nul":
				args[3] = s.root + "\x00"
			case "root-symlink":
				alias := filepath.Join(s.base, "alias")
				if err := os.Symlink(s.root, alias); err != nil {
					t.Fatal(err)
				}
				args[3] = alias
			case "archive-record-hardlink":
				if err := os.MkdirAll(filepath.Dir(s.record()), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Link(s.invitation, s.record()); err != nil {
					t.Fatal(err)
				}
			case "parent-symlink":
				elsewhere := filepath.Join(s.base, "elsewhere")
				if err := os.Mkdir(elsewhere, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(elsewhere, filepath.Join(s.root, "records")); err != nil {
					t.Fatal(err)
				}
			}
			before := snapshot(t, s.base)
			clock := 0
			r := invoke(t, newStarter(runtimeFacts(t, &clock)), args, "confirm jules\nbegin\n")
			if r.exit != 1 || r.report.Code != "start.context" || clock != 0 {
				t.Error("unsafe lexical/native alias accepted")
			}
			unchanged(t, s, before)
		})
	}
}
