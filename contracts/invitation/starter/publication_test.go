package starter_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func replacementRecovery(t *testing.T, mode os.FileMode) recovery {
	t.Helper()
	r := makeRecovery(t, true)
	if err := os.Chmod(r.record(), mode); err != nil {
		t.Fatal(err)
	}
	g := objectGuard(t, r.record())
	r.value.Roles[3].BeforeGuard = &g
	r.value.Roles[3].Action = "reconcile"
	after := append(fixture(t, "begun.yaml"), []byte("# Frozen reviewed publication bytes.\n")...)
	r.value.Roles[3].AfterBase64 = encoded(after)
	r.value.Roles[3].AfterSHA256 = digest(after)
	rewritePending(t, &r)
	return r
}

func TestExistingPublicationControlsAndHardlinkAnchor(t *testing.T) {
	for _, mode := range []os.FileMode{0400, 0600, 0644} {
		t.Run(fmt.Sprintf("%04o", mode), func(t *testing.T) {
			r := replacementRecovery(t, mode)
			anchor := filepath.Join(r.base, "held-owner-inode")
			if err := os.Link(r.record(), anchor); err != nil {
				t.Fatal(err)
			}
			held := objectGuard(t, r.record())
			clock := 0
			got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "resume\nexit\n")
			wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
			after := objectGuard(t, r.record())
			if after.Device == held.Device && after.Inode == held.Inode {
				t.Error("replacement did not replace pathname inode")
			}
			if after.Security != held.Security || clock != 0 {
				t.Error("publication escalated controls or sampled event")
			}
			assertBytes(t, anchor, fixture(t, "begun.yaml"))
			if objectGuard(t, anchor) != held {
				t.Error("hardlink anchor ownership changed")
			}
			assertBytes(t, r.record(), append(fixture(t, "begun.yaml"), []byte("# Frozen reviewed publication bytes.\n")...))
			assertIdle(t, r.sandbox)
		})
	}
}

func TestRawSourceBoundsRefuseBeforePublication(t *testing.T) {
	for _, kind := range []string{"archive", "record", "preparation"} {
		s := setup(t)
		switch kind {
		case "archive":
			put(t, s.invitation, bytes.Repeat([]byte{'x'}, 2097153))
		case "record":
			put(t, s.record(), bytes.Repeat([]byte{'x'}, 1048577))
		case "preparation":
			put(t, s.prep(), bytes.Repeat([]byte{'x'}, 1048577))
		}
		before := snapshot(t, s.base)
		clock := 0
		got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\n")
		wantReport(t, got, 1, "refused", "start.limit", "The opening or frozen display exceeds the supported limit.")
		if clock != 0 {
			t.Error("oversize body sampled event")
		}
		unchanged(t, s, before)
	}
}

func TestPendingRawAndDecodedDisplayBounds(t *testing.T) {
	for _, kind := range []string{"raw-json", "display"} {
		r := makeRecovery(t, true)
		if kind == "raw-json" {
			put(t, filepath.Join(r.pending(), "pending.json"), bytes.Repeat([]byte{' '}, 33554433))
		} else {
			r.value.SessionDisplayBase64 = encoded(bytes.Repeat([]byte{'S'}, 8388609))
			rewritePending(t, &r)
		}
		before := snapshot(t, r.base)
		clock := 0
		got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "resume\nexit\n")
		if got.exit != 1 || clock != 0 {
			t.Error("unbounded pending/display accepted")
		}
		unchanged(t, r.sandbox, before)
	}
}

func TestReceiptOnlyCleanupBindsNativeRootAndHome(t *testing.T) {
	for _, change := range []string{"root", "home", "namespace", "root-safe-mode"} {
		r := makeRecovery(t, true)
		installReceipt(t, r)
		for _, name := range []string{"owner.json", "pending.json", "publication.json"} {
			if err := os.Remove(filepath.Join(r.pending(), name)); err != nil {
				t.Fatal(err)
			}
		}
		switch change {
		case "root":
			if err := os.Rename(r.root, r.root+"-held"); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(r.root, 0700); err != nil {
				t.Fatal(err)
			}
		case "home":
			if err := os.Rename(r.home, r.home+"-held"); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(r.pending(), 0700); err != nil {
				t.Fatal(err)
			}
			body, err := os.ReadFile(filepath.Join(r.home+"-held", "import-pending", "complete.json"))
			if err != nil {
				t.Fatal(err)
			}
			put(t, filepath.Join(r.pending(), "complete.json"), body)
		case "namespace":
			if err := os.Rename(r.pending(), r.pending()+"-held"); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(r.pending(), 0700); err != nil {
				t.Fatal(err)
			}
			body, err := os.ReadFile(filepath.Join(r.pending()+"-held", "complete.json"))
			if err != nil {
				t.Fatal(err)
			}
			put(t, filepath.Join(r.pending(), "complete.json"), body)
		case "root-safe-mode":
			if err := os.Chmod(r.root, 0750); err != nil {
				t.Fatal(err)
			}
		}
		before := snapshot(t, r.base)
		clock := 0
		got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "clean\n")
		if change == "root-safe-mode" {
			wantReport(t, got, 0, "cleaned", "start.cleaned", "Completed recovery data was cleaned; no session was displayed.")
			assertIdle(t, r.sandbox)
		} else {
			if got.exit != 1 || got.report.Code != "start.context" {
				t.Error("receipt adopted replaced native context")
			}
			unchanged(t, r.sandbox, before)
		}
		if clock != 0 {
			t.Error("receipt-only cleanup sampled event")
		}
	}
}

func TestNoFilesystemReadsAfterReadOnlySessionBegins(t *testing.T) {
	s := setup(t)
	clock := 0
	prefix := "confirm jules\nbegin\n"
	in := &countedReader{reader: strings.NewReader(prefix + "exit\n"), at: len(prefix)}
	in.hook = func() {
		put(t, s.record(), []byte("ordinary owner edit after opening\n"))
		put(t, filepath.Join(s.registration(), "session.json"), []byte("ordinary owner metadata edit\n"))
	}
	var out, stderr bytes.Buffer
	exit := newStarter(runtimeFacts(t, &clock)).Run(s.args(), in, &out, &stderr)
	if exit != 0 || clock != 1 || !strings.Contains(stderr.String(), `"code":"start.opened"`) {
		t.Error("session loop reread/reinterpreted completed work")
	}
	assertBytes(t, s.record(), []byte("ordinary owner edit after opening\n"))
	assertBytes(t, filepath.Join(s.registration(), "session.json"), []byte("ordinary owner metadata edit\n"))
	assertIdle(t, s)
}

func TestAllRetainedRolesStillRequireAnExplicitReceipt(t *testing.T) {
	r := makeRecovery(t, true)
	before := snapshot(t, r.root)
	clock := 0
	got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "cancel\n")
	wantReport(t, got, 0, "retained", "start.recovery-retained", "Existing recovery state was retained.")
	if strings.Contains(got.out, "Completed opening:") {
		t.Error("byte-equal record was mistaken for group completion")
	}
	if !reflect.DeepEqual(snapshot(t, r.root), before) || clock != 0 {
		t.Error("cancelled retained-role transaction changed state")
	}
}
