package starter_test

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestMissingSharedParentsRequireExactCreationClaims(t *testing.T) {
	for _, state := range []string{"all-absent", "claimed-shared-parent", "unclaimed-parent"} {
		r := makeRecovery(t, false)
		for _, path := range []string{r.registration(), filepath.Dir(r.registration()), filepath.Dir(r.record()), filepath.Join(r.root, "records")} {
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
		}
		for i := range r.value.Roles {
			for j := range r.value.Roles[i].Parents {
				r.value.Roles[i].Parents[j].Present = false
				r.value.Roles[i].Parents[j].Guard = nil
			}
		}
		if state != "all-absent" {
			path := filepath.Join(r.home, "sessions")
			if err := os.Mkdir(path, 0700); err != nil {
				t.Fatal(err)
			}
			if state == "claimed-shared-parent" {
				r.claims.Roles[0].NewParents = []newParent{{Ordinal: "0", Guard: objectGuard(t, path)}}
			}
		}
		rewritePending(t, &r)
		before := snapshot(t, r.base)
		clock := 0
		got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "resume\nexit\n")
		if state == "unclaimed-parent" {
			if got.exit != 1 || clock != 0 {
				t.Error("unclaimed appearing parent was adopted")
			}
			unchanged(t, r.sandbox, before)
		} else {
			wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
			for i, path := range r.targets {
				body, err := base64.StdEncoding.DecodeString(r.value.Roles[i].AfterBase64)
				if err != nil {
					t.Fatal(err)
				}
				assertBytes(t, path, body)
			}
			assertIdle(t, r.sandbox)
		}
	}
}

func TestClaimedPayloadInitialAndFinalControlsAreFiniteStates(t *testing.T) {
	for _, state := range []string{"private-initial", "intended-final", "already-published", "other-mode"} {
		r := replacementRecovery(t, 0400)
		stage := filepath.Join(filepath.Dir(r.record()), ".obk-start-"+recoveryToken+"-record")
		if err := os.Mkdir(stage, 0700); err != nil {
			t.Fatal(err)
		}
		payload := filepath.Join(stage, "payload")
		body, err := base64.StdEncoding.DecodeString(r.value.Roles[3].AfterBase64)
		if err != nil {
			t.Fatal(err)
		}
		put(t, payload, nil)
		directory, initial := objectGuard(t, stage), objectGuard(t, payload)
		r.claims.Roles[3].Staging = &directory
		r.claims.Roles[3].Temporary = &initial
		put(t, filepath.Join(r.pending(), "publication.json"), jsonBytes(t, r.claims))
		put(t, payload, body)
		if state == "intended-final" || state == "already-published" {
			if err := os.Chmod(payload, 0400); err != nil {
				t.Fatal(err)
			}
		}
		if state == "other-mode" {
			if err := os.Chmod(payload, 0640); err != nil {
				t.Fatal(err)
			}
		}
		if state == "already-published" {
			if err := os.Rename(payload, r.record()); err != nil {
				t.Fatal(err)
			}
		}
		before := snapshot(t, r.base)
		clock := 0
		got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "resume\nexit\n")
		if state == "other-mode" {
			if got.exit != 1 || clock != 0 {
				t.Error("arbitrary payload security was accepted")
			}
			unchanged(t, r.sandbox, before)
		} else {
			wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
			after := objectGuard(t, r.record())
			if after.Inode != initial.Inode || after.Device != initial.Device || after.Security != r.value.Roles[3].BeforeGuard.Security {
				t.Error("claimed payload identity/final controls were not preserved")
			}
			assertBytes(t, r.record(), body)
			assertIdle(t, r.sandbox)
		}
	}
}

func TestAlternateOwnedGroupIsPreserved(t *testing.T) {
	groups, err := os.Getgroups()
	if err != nil {
		t.Fatal(err)
	}
	other := -1
	for _, group := range groups {
		if group != os.Getegid() {
			other = group
			break
		}
	}
	if other < 0 {
		t.Skip("native group-preservation gap: no secondary owned group available")
	}
	r := replacementRecovery(t, 0640)
	if err := os.Chown(r.record(), -1, other); err != nil {
		t.Skip("native group-preservation gap: host refuses owned secondary-group fixture")
	}
	g := objectGuard(t, r.record())
	r.value.Roles[3].BeforeGuard = &g
	rewritePending(t, &r)
	clock := 0
	got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "resume\nexit\n")
	wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	after := objectGuard(t, r.record())
	if after.Security != g.Security {
		t.Error("replacement reset owned UID/GID/mode/native ACL")
	}
	assertIdle(t, r.sandbox)
}

func TestExistingPrivateStatePermittedGroupIsPreserved(t *testing.T) {
	groups, err := os.Getgroups()
	if err != nil {
		t.Fatal(err)
	}
	other := -1
	for _, group := range groups {
		if group != os.Getegid() {
			other = group
			break
		}
	}
	if other < 0 {
		t.Skip("native existing-state group gap: no secondary owned group")
	}
	s := setup(t)
	if err := os.MkdirAll(s.pending(), 0700); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{s.home, s.pending()} {
		if err := os.Chown(path, -1, other); err != nil {
			t.Skip("native existing-state group gap: host refuses owned secondary-group fixture")
		}
	}
	home, idle := objectGuard(t, s.home), objectGuard(t, s.pending())
	clock := 0
	got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\nexit\n")
	wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	if objectGuard(t, s.home) != home || objectGuard(t, s.pending()) != idle {
		t.Error("existing permitted private state ownership was normalized/replaced")
	}
	assertIdle(t, s)
	assertNewPrivate(t, s.record(), "0600")
	assertNewPrivate(t, s.registration(), "0700")
}
