//go:build darwin || linux

package editor_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"golang.org/x/sys/unix"
	contract "onboard-kit/contracts/session/editor"
)

func TestGenuineNativeCaseAliasKeepsOriginalKeyAndSpelling(t *testing.T) {
	s := session(t, planBody, recordBody)
	alias := filepath.Join(s.base, "ROOT")
	info, err := os.Lstat(alias)
	if os.IsNotExist(err) {
		t.Skip("case-sensitive native host: same-object case alias not available; no symlink substitution")
	}
	if err != nil || info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("native case alias fixture unavailable or is a symlink")
	}
	if !os.SameFile(info, mustInfo(t, s.root)) {
		t.Fatal("case spelling does not name the actual same native object")
	}
	before, held := snapshot(t, s.base), nativeObjectAt(t, s.record)
	clock := &clockFact{value: editEvent}
	args := s.args()
	args[1] = alias
	reader := &observedReader{text: confirm + "answer personal.choice text \"org-a\"\nsave\n"}
	out, errout := &plainOutput{}, &plainOutput{}
	result := newDriver(clock).Run(args, contract.Streams{Input: reader, Output: out, Error: errout})
	report(t, runResult{result, out.text, errout.text, reader.position}, "edit.saved", 0, true)
	savedRecord(t, s, held, ownedRecord)
	companions(t, s, before)
	if clock.calls != 0 {
		t.Fatal("case alias resampled event")
	}
}
func mustInfo(t *testing.T, path string) os.FileInfo {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info
}

type plainOutput struct{ text string }

func (w *plainOutput) Write(p []byte) (int, error) { w.text += string(p); return len(p), nil }

func TestNestedRootHomeWithDisjointActualParticipantsSucceeds(t *testing.T) {
	s := session(t, planBody, recordBody)
	newHome := filepath.Join(s.root, "state")
	if err := os.Rename(s.home, newHome); err != nil {
		t.Fatal(err)
	}
	s.home = newHome
	s.plan = filepath.Join(newHome, "sessions", s.key, "plan.obk.yaml")
	s.metadata = filepath.Join(newHome, "sessions", s.key, "session.json")
	before, held := snapshot(t, s.base), nativeObjectAt(t, s.record)
	clock := &clockFact{value: editEvent}
	got := run(t, s, clock, confirm+"answer personal.choice text \"org-a\"\nsave\n")
	report(t, got, "edit.saved", 0, true)
	savedRecord(t, s, held, ownedRecord)
	companions(t, s, before)
}

func TestAllNativeRecordXattrsArePreservedOrExplicitlyRefused(t *testing.T) {
	s := session(t, planBody, recordBody)
	if err := unix.Setxattr(s.record, "user.obk_fixture", []byte("keep-exact"), 0); err != nil {
		t.Skipf("actual native custom attribute fixture unavailable: %v", err)
	}
	before, held := snapshot(t, s.base), nativeObjectAt(t, s.record)
	if held.Attributes["user.obk_fixture"] != "keep-exact" {
		t.Fatal("actual native attribute not established")
	}
	clock := &clockFact{value: editEvent}
	got := run(t, s, clock, confirm+"answer personal.choice text \"org-a\"\nsave\n")
	if got.result.Saved {
		report(t, got, "edit.saved", 0, true)
		savedRecord(t, s, held, ownedRecord)
	} else {
		if got.result.Exit != 1 {
			t.Fatal("attribute refusal was reported as success")
		}
		if !reflect.DeepEqual(nativeObjectAt(t, s.record), held) {
			t.Fatal("refused replacement dropped selected metadata or changed record")
		}
		var diagnostic struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal([]byte(got.err), &diagnostic); err != nil {
			t.Fatal("attribute refusal is not a canonical diagnostic")
		}
		if diagnostic.Code != "edit.context" && diagnostic.Code != "edit.commit" {
			t.Fatal("attribute refusal used an unrelated failure")
		}
		report(t, got, diagnostic.Code, 1, false)
	}
	if clock.calls != 0 {
		t.Fatal("answer attribute path sampled event")
	}
	companions(t, s, before)
}
