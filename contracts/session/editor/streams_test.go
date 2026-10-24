//go:build darwin || linux

package editor_test

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	contract "onboard-kit/contracts/session/editor"
)

type faultOutput struct {
	sink   bytes.Buffer
	marker string
	mode   int
	failed bool
	closes int
}

func (w *faultOutput) Write(p []byte) (int, error) {
	if w.failed {
		return 0, errors.New("fixture output already failed")
	}
	if strings.Contains(w.sink.String()+string(p), w.marker) {
		w.failed = true
		switch w.mode {
		case 0:
			return 0, errors.New("fixture output failure")
		case 1:
			n := len(p) - 1
			if n < 0 {
				n = 0
			}
			_, _ = w.sink.Write(p[:n])
			return n, nil
		default:
			_, _ = w.sink.Write(p)
			return len(p), errors.New("fixture full-count output failure")
		}
	}
	return w.sink.Write(p)
}
func (w *faultOutput) WriteString(s string) (int, error) { return w.Write([]byte(s)) }
func (w *faultOutput) Close() error                      { w.closes++; return nil }

type flushFaultOutput struct {
	sink    bytes.Buffer
	flushes int
}

func (w *flushFaultOutput) Write(p []byte) (int, error)       { return w.sink.Write(p) }
func (w *flushFaultOutput) WriteString(s string) (int, error) { return w.Write([]byte(s)) }
func (w *flushFaultOutput) Flush() error {
	w.flushes++
	return errors.New("fixture required display flush failed")
}

func TestRequiredFlushFailurePrecedesFirstIdentityByte(t *testing.T) {
	s := session(t, planBody, recordBody)
	before := snapshot(t, s.base)
	clock := &clockFact{value: editEvent}
	reader := &observedReader{text: confirm + "exit\n"}
	writer := &flushFaultOutput{}
	var stderr bytes.Buffer
	result := newDriver(clock).Run(s.args(), contract.Streams{Input: reader, Output: writer, Error: &stderr})
	report(t, runResult{result, writer.sink.String(), stderr.String(), reader.position}, "edit.output", 1, false)
	if writer.flushes != 1 || reader.position != 0 || clock.calls != 0 {
		t.Fatal("required flush failed but identity/time was consumed or flush retried")
	}
	unchanged(t, s, before)
}

func TestEveryRequiredDisplayFailureStopsQueuedConsentAcrossWriterInterfaces(t *testing.T) {
	for _, stage := range []struct {
		marker   string
		consumed int
	}{
		{"Registered account:", 0}, {actionPrompt, len(confirm)},
		{savePrompt, len(confirm + "answer personal.choice text \"org-a\"\n")},
	} {
		for mode := 0; mode < 3; mode++ {
			s := session(t, planBody, recordBody)
			before := snapshot(t, s.base)
			clock := &clockFact{value: editEvent}
			reader := &observedReader{text: confirm + "answer personal.choice text \"org-a\"\nsave\nUNREAD_TAIL\n"}
			writer := &faultOutput{marker: stage.marker, mode: mode}
			var stderr bytes.Buffer
			result := newDriver(clock).Run(s.args(), contract.Streams{Input: reader, Output: writer, Error: &stderr})
			report(t, runResult{result, writer.sink.String(), stderr.String(), reader.position}, "edit.output", 1, false)
			if !writer.failed || writer.closes != 0 || reader.closed != 0 || reader.position != stage.consumed || clock.calls != 0 {
				t.Fatal("fault bypassed via WriteString/short-write handling or queued consent was read")
			}
			unchanged(t, s, before)
		}
	}
}

func TestSavedTruthSurvivesFinalDisplayFailure(t *testing.T) {
	s := session(t, outcomePlan, recordBody)
	before, held := snapshot(t, s.base), nativeObjectAt(t, s.record)
	clock := &clockFact{value: editEvent}
	reader := &observedReader{text: confirm + "outcome access verified \"\" \"\"\nsave\n"}
	writer := &faultOutput{marker: "Record saved.\n", mode: 2}
	var stderr bytes.Buffer
	result := newDriver(clock).Run(s.args(), contract.Streams{Input: reader, Output: writer, Error: &stderr})
	report(t, runResult{result, writer.sink.String(), stderr.String(), reader.position}, "edit.repair", 1, true)
	if !writer.failed || clock.calls != 1 {
		t.Fatal("postrename failure lost publication truth or resampled")
	}
	savedRecord(t, s, held, verifiedRecord)
	companions(t, s, before)
}

func TestHelpAndRawArgumentRefusalsDoNotTouchInputOrClock(t *testing.T) {
	help, err := os.ReadFile("testdata/help.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"--help"}, {}, {"--help", "--root", "/unused"},
		{"--root", "/unused", "--root", "/duplicate", "--home", "/unused", "--session", strings.Repeat("a", 64)},
		{"--root=/unused", "--home", "/unused", "--session", strings.Repeat("a", 64)},
		{"--root", "relative", "--home", "/unused", "--session", strings.Repeat("a", 64)},
		{"--root", "/unused", "--home", "/unused", "--session", strings.Repeat("A", 64)},
	} {
		clock := &clockFact{value: editEvent}
		reader := &observedReader{text: "UNREAD\n"}
		var out, err bytes.Buffer
		result := newDriver(clock).Run(args, contract.Streams{Input: reader, Output: &out, Error: &err})
		if len(args) == 1 && args[0] == "--help" {
			if result.Exit != 0 || result.Saved || !bytes.Equal(out.Bytes(), help) || err.Len() != 0 {
				t.Fatal("static help differs")
			}
		} else {
			report(t, runResult{result, out.String(), err.String(), reader.position}, "edit.usage", 2, false)
		}
		if clock.calls != 0 || reader.position != 0 || reader.closed != 0 {
			t.Fatal("help/usage consumed runtime input")
		}
	}
}
