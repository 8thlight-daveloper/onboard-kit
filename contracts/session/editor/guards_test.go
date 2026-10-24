//go:build darwin || linux

package editor_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	contract "onboard-kit/contracts/session/editor"
)

func TestPendingProofPrecedesAnyUsableMixedSession(t *testing.T) {
	for _, child := range []string{"owner.json", "pending.json", "complete.json", ".pending.next", "unknown"} {
		t.Run(child, func(t *testing.T) {
			s := session(t, "NOT_A_PLAN_BODY", recordBody)
			put(t, filepath.Join(s.home, "import-pending", child), "preserve this pending state")
			before := snapshot(t, s.base)
			clock := &clockFact{value: editEvent}
			got := run(t, s, clock, confirm+"answer personal.choice text \"org-a\"\nsave\n")
			report(t, got, "edit.pending", 1, false)
			if clock.calls != 0 || got.consumed != 0 {
				t.Fatal("pending session was used or consent consumed")
			}
			unchanged(t, s, before)
		})
	}
}

func TestActualSymlinkHardlinkAndSpecialParticipantRefusals(t *testing.T) {
	for _, kind := range []string{"symlink", "hardlink", "fifo", "metadata-duplicate-key"} {
		t.Run(kind, func(t *testing.T) {
			s := session(t, planBody, recordBody)
			switch kind {
			case "symlink":
				outside := filepath.Join(s.base, "outside-record")
				put(t, outside, recordBody)
				if err := os.Remove(s.record); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, s.record); err != nil {
					t.Fatal(err)
				}
			case "hardlink":
				if err := os.Remove(s.preparation); err != nil {
					t.Fatal(err)
				}
				if err := os.Link(s.record, s.preparation); err != nil {
					t.Fatal(err)
				}
			case "fifo":
				if err := os.Remove(s.record); err != nil {
					t.Fatal(err)
				}
				if err := syscall.Mkfifo(s.record, 0600); err != nil {
					t.Fatal(err)
				}
			case "metadata-duplicate-key":
				body, err := os.ReadFile(s.metadata)
				if err != nil {
					t.Fatal(err)
				}
				put(t, s.metadata, strings.Replace(string(body), "\"version\":1", "\"version\":1,\"version\":1", 1))
			}
			before := snapshot(t, s.base)
			clock := &clockFact{value: editEvent}
			got := run(t, s, clock, "cancel\n")
			report(t, got, "edit.context", 1, false)
			if got.consumed != 0 || clock.calls != 0 {
				t.Fatal("unsafe native registration consumed identity")
			}
			unchanged(t, s, before)
		})
	}
}

func TestEverySelectedSourceAndNativeGuardIsFreshAfterSave(t *testing.T) {
	for _, kind := range []string{"plan-bytes", "preparation-bytes", "record-bytes", "metadata-bytes", "same-bytes-new-inode", "record-mode", "parent-inode"} {
		t.Run(kind, func(t *testing.T) {
			s := session(t, outcomePlan, recordBody)
			before := snapshot(t, s.base)
			clock := &clockFact{value: editEvent}
			prefix := confirm + "outcome access failed \"\" \"\"\n"
			var ownerState map[string]nativeObject
			reader := &observedReader{text: prefix + "save\nUNREAD_TAIL\n", hookAt: len(prefix)}
			reader.hook = func() {
				unchanged(t, s, before)
				switch kind {
				case "plan-bytes":
					put(t, s.plan, outcomePlan+"# owner plan edit\n")
				case "preparation-bytes":
					put(t, s.preparation, prepBody+"# owner preparation edit\n")
				case "record-bytes":
					put(t, s.record, recordBody+"# owner personal edit\n")
				case "metadata-bytes":
					body, err := os.ReadFile(s.metadata)
					if err != nil {
						t.Fatal(err)
					}
					put(t, s.metadata, " "+string(body))
				case "same-bytes-new-inode":
					replacement := filepath.Join(s.base, "owner-replacement")
					put(t, replacement, recordBody)
					if err := os.Rename(replacement, s.record); err != nil {
						t.Fatal(err)
					}
				case "record-mode":
					if err := os.Chmod(s.record, 0644); err != nil {
						t.Fatal(err)
					}
				case "parent-inode":
					old := filepath.Join(s.root, "records", "held-old-parent")
					if err := os.Rename(filepath.Dir(s.record), old); err != nil {
						t.Fatal(err)
					}
					put(t, s.record, recordBody)
					put(t, s.preparation, prepBody)
				}
				ownerState = snapshot(t, s.base)
			}
			var stdout, stderr bytes.Buffer
			result := newDriver(clock).Run(s.args(), contract.Streams{Input: reader, Output: &stdout, Error: &stderr})
			got := runResult{result, stdout.String(), stderr.String(), reader.position}
			report(t, got, "edit.context", 1, false)
			if reader.hook != nil || ownerState == nil || clock.calls != 0 || reader.position != len(prefix+"save\n") {
				t.Fatal("native owner change was not reached or clock/reader advanced incorrectly")
			}
			unchanged(t, s, ownerState)
		})
	}
}

func TestAbandonedNonemptyStageIsNeverAdoptedOrReplayed(t *testing.T) {
	s := session(t, planBody, recordBody)
	put(t, filepath.Join(s.stage(), "record.new"), ownedRecord)
	before := snapshot(t, s.base)
	clock := &clockFact{value: editEvent}
	got := run(t, s, clock, confirm+"answer personal.choice text \"org-a\"\nsave\n")
	report(t, got, "edit.context", 1, false)
	if clock.calls != 0 {
		t.Fatal("abandoned staging caused a fresh event")
	}
	unchanged(t, s, before)
}

func TestInteractiveWaitsReleaseActualNativeRootAndHomeLocks(t *testing.T) {
	for _, c := range []struct {
		input string
		at    int
	}{
		{"cancel\n", 0}, {confirm + "exit\n", len(confirm)},
		{confirm + "answer personal.choice text \"org-a\"\ncancel\n", len(confirm + "answer personal.choice text \"org-a\"\n")},
	} {
		s := session(t, planBody, recordBody)
		clock := &clockFact{value: editEvent}
		before := snapshot(t, s.base)
		reached := false
		reader := &observedReader{text: c.input, hookAt: c.at}
		reader.hook = func() {
			reached = true
			for _, path := range []string{s.root, s.home} {
				file, err := os.Open(path)
				if err != nil {
					t.Fatal(err)
				}
				if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
					file.Close()
					t.Fatal("driver retains native lock while waiting for user")
				}
				if err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
					t.Fatal(err)
				}
				if err := file.Close(); err != nil {
					t.Fatal(err)
				}
			}
		}
		var out, err bytes.Buffer
		result := newDriver(clock).Run(s.args(), contract.Streams{Input: reader, Output: &out, Error: &err})
		report(t, runResult{result, out.String(), err.String(), reader.position}, "edit.cancelled", 0, false)
		if !reached || clock.calls != 0 {
			t.Fatal("native input boundary not reached")
		}
		unchanged(t, s, before)
	}
}
