package starter_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func rewritePending(t *testing.T, r *recovery) {
	t.Helper()
	r.raw = jsonBytes(t, r.value)
	r.claims.PendingSHA256 = digest(r.raw)
	put(t, filepath.Join(r.pending(), "pending.json"), r.raw)
	put(t, filepath.Join(r.pending(), "publication.json"), jsonBytes(t, r.claims))
	put(t, filepath.Join(r.pending(), "owner.json"), jsonBytes(t, owner{Version: 1, Transaction: recoveryToken, RootDevice: r.value.RootGuard.Device, RootInode: r.value.RootGuard.Inode, Namespace: objectGuard(t, r.pending()), PendingSHA256: digest(r.raw)}))
}

func TestCanonicalBase64DisplayControl(t *testing.T) {
	r := makeRecovery(t, true)
	r.value.SessionDisplayBase64 = "Uwo="
	rewritePending(t, &r)
	clock := 0
	got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "resume\nexit\n")
	wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	if !strings.HasSuffix(got.out, "First opened at: \""+string(openedAt)+"\"\nS\nSession: read only. Choose exit:\n") || clock != 0 {
		t.Error("canonical otherwise safe frozen display was rerendered/rejected")
	}
	assertIdle(t, r.sandbox)
}

func makeIncomplete(t *testing.T) sandbox {
	t.Helper()
	s := setup(t)
	if err := os.MkdirAll(s.pending(), 0700); err != nil {
		t.Fatal(err)
	}
	temporary := filepath.Join(s.pending(), ".pending.next")
	put(t, temporary, nil)
	claim := objectGuard(t, temporary)
	root := objectGuard(t, s.root)
	put(t, filepath.Join(s.pending(), "owner.json"), jsonBytes(t, owner{Version: 1, Transaction: recoveryToken, RootDevice: root.Device, RootInode: root.Inode, Namespace: objectGuard(t, s.pending()), PendingTemporary: &claim}))
	put(t, temporary, []byte(`{"version":1,"account":"jules",`))
	return s
}

func resumeArgs(s sandbox) []string {
	return []string{"--resume-import", recoveryToken, "--root", s.root, "--home", s.home}
}

func TestAttributableIncompleteStagingOnlyCleansAfterConsent(t *testing.T) {
	for _, input := range []string{"cancel\n", "clean\n", "resume\nclean\n"} {
		s := makeIncomplete(t)
		before := snapshot(t, s.base)
		clock := 0
		got := invoke(t, newStarter(runtimeFacts(t, &clock)), resumeArgs(s), input)
		if input == "cancel\n" {
			wantReport(t, got, 0, "retained", "start.recovery-retained", "Existing recovery state was retained.")
			unchanged(t, s, before)
		} else {
			wantReport(t, got, 0, "cleaned", "start.staging-cleaned", "Owned incomplete staging was cleaned; reopen the invitation to begin.")
			assertIdle(t, s)
		}
		if clock != 0 || strings.Contains(got.out, "Read-only onboarding session.") {
			t.Error("incomplete staging was treated as completed opening")
		}
		if !strings.Contains(got.out, "Incomplete staging: \""+recoveryToken+"\"\nReopen the invitation after cleanup.\nStaging: choose clean or cancel:\n") {
			t.Error("incomplete staging prompt differs")
		}
		if _, err := os.Stat(s.record()); !os.IsNotExist(err) {
			t.Error("incomplete staging seeded a record")
		}
	}
}

func TestRecoveryDisplayFailureRetainsStateAndDoesNotReadConsent(t *testing.T) {
	for _, state := range []string{"pending", "completed", "incomplete"} {
		s := sandbox{}
		var args []string
		if state == "incomplete" {
			s = makeIncomplete(t)
			args = resumeArgs(s)
		} else {
			r := makeRecovery(t, true)
			if state == "completed" {
				installReceipt(t, r)
			}
			s = r.sandbox
			args = r.args()
		}
		before := snapshot(t, s.base)
		clock := 0
		in := &countedReader{reader: strings.NewReader("resume\nclean\n")}
		out := &failedWriter{marker: "opening:", mode: 2}
		if state == "incomplete" {
			out.marker = "Incomplete staging:"
		}
		var stderr bytes.Buffer
		exit := newStarter(runtimeFacts(t, &clock)).Run(args, in, out, &stderr)
		var got report
		if err := json.Unmarshal(stderr.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		status := state
		if state == "completed" {
			status = "opened"
		}
		if state == "incomplete" {
			status = "retained"
		}
		if exit != 1 || got.Status != status || got.Code != "start.recovery-output" || got.Message != "Existing recovery state was retained; required output could not be completed." || in.consumed != 0 || clock != 0 {
			t.Error("recovery writer failure lost prior truth or consumed consent")
		}
		unchanged(t, s, before)
	}
}

func TestRecoveryInputBudgetUsesPriorState(t *testing.T) {
	for _, completed := range []bool{false, true} {
		r := makeRecovery(t, true)
		if completed {
			installReceipt(t, r)
		}
		before := snapshot(t, r.base)
		clock := 0
		var stderr bytes.Buffer
		exit := newStarter(runtimeFacts(t, &clock)).Run(r.args(), strings.NewReader(strings.Repeat("\n", 1048576)+"clean\n"), io.Discard, &stderr)
		var got report
		if err := json.Unmarshal(stderr.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		status := "pending"
		if completed {
			status = "opened"
		}
		if exit != 1 || got.Code != "start.recovery-limit" || got.Status != status || clock != 0 {
			t.Error("recovery budget reported fresh cancellation/refusal")
		}
		unchanged(t, r.sandbox, before)
	}
}

func TestRecoveryRechecksRetainedRolesAndContextIdentity(t *testing.T) {
	for _, change := range []string{"retained-bytes", "retained-inode", "root-object", "parent-object", "claimed-temp-object"} {
		t.Run(change, func(t *testing.T) {
			r := makeRecovery(t, true)
			switch change {
			case "retained-bytes":
				put(t, r.targets[0], []byte("owner edit\n"))
			case "retained-inode":
				if err := os.Rename(r.targets[0], r.targets[0]+".held"); err != nil {
					t.Fatal(err)
				}
				put(t, r.targets[0], fixture(t, "plan.yaml"))
			case "root-object":
				if err := os.Rename(r.root, r.root+"-held"); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(r.root, 0700); err != nil {
					t.Fatal(err)
				}
			case "parent-object":
				parent := filepath.Dir(r.record())
				if err := os.Rename(parent, parent+"-held"); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(parent, 0700); err != nil {
					t.Fatal(err)
				}
			case "claimed-temp-object":
				stage := filepath.Join(filepath.Dir(r.targets[0]), ".obk-start-"+recoveryToken+"-local-plan")
				if err := os.Mkdir(stage, 0700); err != nil {
					t.Fatal(err)
				}
				payload := filepath.Join(stage, "payload")
				put(t, payload, fixture(t, "plan.yaml"))
				g, p := objectGuard(t, stage), objectGuard(t, payload)
				r.claims.Roles[0].Staging = &g
				r.claims.Roles[0].Temporary = &p
				put(t, filepath.Join(r.pending(), "publication.json"), jsonBytes(t, r.claims))
				if err := os.Rename(payload, payload+".held"); err != nil {
					t.Fatal(err)
				}
				put(t, payload, fixture(t, "plan.yaml"))
			}
			before := snapshot(t, r.base)
			clock := 0
			got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "resume\n")
			if got.exit != 1 || clock != 0 {
				t.Error("changed source/namespace/ownership was adopted")
			}
			unchanged(t, r.sandbox, before)
		})
	}
}
