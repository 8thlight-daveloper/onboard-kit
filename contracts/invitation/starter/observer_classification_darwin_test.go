package starter_test

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestObserverClassificationRoleVectors(t *testing.T) {
	for _, tc := range []struct {
		name  string
		facts [4]observedRoleFact
		wrong bool
	}{
		{"no-publication", [4]observedRoleFact{rolePreimage, rolePreimage, rolePreimage, rolePreimage}, false},
		{"ordered-prefix", [4]observedRoleFact{rolePublished, rolePublished, rolePreimage, rolePreimage}, false},
		{"all-published", [4]observedRoleFact{rolePublished, rolePublished, rolePublished, rolePublished}, false},
		{"record-before-metadata", [4]observedRoleFact{rolePublished, rolePublished, rolePreimage, rolePublished}, true},
		{"record-before-all", [4]observedRoleFact{rolePreimage, rolePreimage, rolePreimage, rolePublished}, true},
		{"retained-record-is-no-publication", [4]observedRoleFact{rolePublished, rolePublished, rolePreimage, roleRetained}, false},
		{"unknown-earlier-is-insufficient", [4]observedRoleFact{rolePublished, rolePublished, roleIndeterminate, rolePublished}, false},
		{"unrelated-unknown-cannot-hide-proved-pair", [4]observedRoleFact{roleIndeterminate, rolePublished, rolePreimage, rolePublished}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := forbiddenRoleOrder(tc.facts); got != tc.wrong {
				t.Errorf("role-vector order classification=%t, want %t", got, tc.wrong)
			}
		})
	}
}

func observerControl(t *testing.T) (recovery, observationEvidence) {
	t.Helper()
	r := makeRecovery(t, false)
	return r, captureObservation(t, r.sandbox, "fresh")
}

func publishObserverControlRole(t *testing.T, r *recovery, index int) {
	t.Helper()
	stage := filepath.Join(filepath.Dir(r.targets[index]), ".obk-start-"+recoveryToken+"-"+r.value.Roles[index].Role)
	if err := os.Mkdir(stage, 0700); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(stage, "payload")
	put(t, payload, nil)
	staging, temporary := objectGuard(t, stage), objectGuard(t, payload)
	r.claims.Roles[index].Staging, r.claims.Roles[index].Temporary = &staging, &temporary
	put(t, filepath.Join(r.pending(), "publication.json"), jsonBytes(t, r.claims))
	body, err := base64.StdEncoding.DecodeString(r.value.Roles[index].AfterBase64)
	if err != nil {
		t.Fatal(err)
	}
	put(t, payload, body)
	if err := os.Rename(payload, r.targets[index]); err != nil {
		t.Fatal(err)
	}
}

func publishAllObserverControlRoles(t *testing.T, r *recovery) {
	t.Helper()
	for i := range r.targets {
		publishObserverControlRole(t, r, i)
	}
}

func TestObserverClassificationNativeRoleFacts(t *testing.T) {
	for _, tc := range []struct {
		name      string
		published []int
		unknown   bool
		violation string
	}{
		{"initial", nil, false, ""},
		{"ordered-prefix", []int{0, 1}, false, ""},
		{"all-published", []int{0, 1, 2, 3}, false, ""},
		{"record-before-metadata", []int{0, 1, 3}, false, "forbidden-published-role-subset"},
		{"unproved-metadata", []int{0, 1, 3}, true, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, evidence := observerControl(t)
			for _, i := range tc.published {
				publishObserverControlRole(t, &r, i)
			}
			if tc.unknown {
				put(t, r.targets[2], []byte("unattributed bytes\n"))
			}
			history := observationLineage{}
			got := observeNativeInvariants(t, r.sandbox, evidence, &history)
			if got.violation != tc.violation {
				t.Errorf("native role classification=%q, want %q; gap=%q facts=%v", got.violation, tc.violation, got.gap, got.facts)
			}
			for _, i := range tc.published {
				if got.facts[i] != rolePublished {
					t.Errorf("claimed final native role %d was not proved published", i)
				}
			}
			if tc.unknown && got.facts[2] != roleIndeterminate {
				t.Error("unattributed metadata was treated as a definite preimage")
			}
		})
	}
}

func TestObserverClassificationPartialTemporaryControls(t *testing.T) {
	for _, temporary := range []string{".owner.next", ".publication.next", ".complete.next", "partial-payload", "unclaimed-empty-stage", "claimed-empty-payload"} {
		t.Run(temporary, func(t *testing.T) {
			r, evidence := observerControl(t)
			publishObserverControlRole(t, &r, 0)
			if temporary == "partial-payload" || temporary == "unclaimed-empty-stage" || temporary == "claimed-empty-payload" {
				stage := filepath.Join(filepath.Dir(r.targets[1]), ".obk-start-"+recoveryToken+"-preparation")
				if err := os.Mkdir(stage, 0700); err != nil {
					t.Fatal(err)
				}
				if temporary != "unclaimed-empty-stage" {
					payload := filepath.Join(stage, "payload")
					put(t, payload, nil)
					staging, initial := objectGuard(t, stage), objectGuard(t, payload)
					r.claims.Roles[1].Staging, r.claims.Roles[1].Temporary = &staging, &initial
					put(t, filepath.Join(r.pending(), "publication.json"), jsonBytes(t, r.claims))
					if temporary == "partial-payload" {
						put(t, payload, []byte("schemaVersion:"))
					}
				}
			} else {
				put(t, filepath.Join(r.pending(), temporary), []byte("{\"version\":"))
			}
			history := observationLineage{}
			got := observeNativeInvariants(t, r.sandbox, evidence, &history)
			if got.violation != "" || got.facts[0] != rolePublished || got.facts[1] != rolePreimage {
				t.Errorf("legal partial temporary changed independent role facts: violation=%q gap=%q facts=%v", got.violation, got.gap, got.facts)
			}
		})
	}
}

func TestObserverClassificationFinalReceiptControls(t *testing.T) {
	for _, tc := range []struct {
		name, violation string
	}{
		{"valid", ""},
		{"three-roles", "attributable-final-receipt-invalid"},
		{"truncated-final", "attributable-final-receipt-malformed"},
		{"partial-temporary-only", ""},
		{"missing-lineage", ""},
		{"unproved-final-role", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, evidence := observerControl(t)
			publishAllObserverControlRoles(t, &r)
			installReceipt(t, r)
			switch tc.name {
			case "three-roles":
				var receipt complete
				raw, err := os.ReadFile(filepath.Join(r.pending(), "complete.json"))
				if err != nil || json.Unmarshal(raw, &receipt) != nil {
					t.Fatal("cannot read independently seeded receipt")
				}
				receipt.Roles = receipt.Roles[:3]
				put(t, filepath.Join(r.pending(), "complete.json"), jsonBytes(t, receipt))
			case "truncated-final":
				put(t, filepath.Join(r.pending(), "complete.json"), []byte("{\"version\":1"))
			case "partial-temporary-only":
				if err := os.Rename(filepath.Join(r.pending(), "complete.json"), filepath.Join(r.pending(), ".complete.next")); err != nil {
					t.Fatal(err)
				}
				put(t, filepath.Join(r.pending(), ".complete.next"), []byte("{\"version\":1"))
			case "missing-lineage":
				put(t, filepath.Join(r.pending(), "owner.json"), []byte("{\"version\":1"))
				put(t, filepath.Join(r.pending(), "complete.json"), []byte("{\"version\":1"))
			case "unproved-final-role":
				if err := os.Remove(filepath.Join(filepath.Dir(r.record()), ".obk-start-"+recoveryToken+"-record")); err != nil {
					t.Fatal(err)
				}
			}
			history := observationLineage{}
			got := observeNativeInvariants(t, r.sandbox, evidence, &history)
			if got.violation != tc.violation {
				t.Errorf("final receipt classification=%q, want %q; gap=%q", got.violation, tc.violation, got.gap)
			}
			if tc.name == "valid" && !history.sawReceipt {
				t.Error("independently valid receipt was not remembered")
			}
			if tc.name == "missing-lineage" && got.gap == "" {
				t.Error("malformed source lineage did not remain an explicit gap")
			}
			if tc.name == "unproved-final-role" && got.gap != "final-receipt-role-native-proof-incomplete" {
				t.Error("missing final native role proof did not remain an explicit gap")
			}
		})
	}
}

func TestObserverClassificationCompletionHistoryControls(t *testing.T) {
	for _, tc := range []struct {
		name, violation string
	}{
		{"pending-without-prior-receipt", ""},
		{"receipt-removed-while-pending", "completion-proof-lost-with-personal-pending"},
		{"pending-and-receipt-absent-with-owned-remnants", "published-cleanup-lost-pending-and-receipt"},
		{"empty-idle-after-completion", ""},
		{"replacement-namespace", ""},
		{"changed-pending-is-unproved", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, evidence := observerControl(t)
			publishAllObserverControlRoles(t, &r)
			if tc.name != "pending-without-prior-receipt" && tc.name != "pending-and-receipt-absent-with-owned-remnants" {
				installReceipt(t, r)
			}
			history := observationLineage{}
			first := observeNativeInvariants(t, r.sandbox, evidence, &history)
			if first.violation != "" || !history.allPublished {
				t.Fatalf("cannot establish independent legal history: %q gap=%q", first.violation, first.gap)
			}
			switch tc.name {
			case "receipt-removed-while-pending":
				if err := os.Remove(filepath.Join(r.pending(), "complete.json")); err != nil {
					t.Fatal(err)
				}
			case "pending-and-receipt-absent-with-owned-remnants":
				if err := os.Remove(filepath.Join(r.pending(), "pending.json")); err != nil {
					t.Fatal(err)
				}
			case "empty-idle-after-completion":
				for _, name := range []string{"pending.json", "publication.json", "owner.json", "complete.json"} {
					if err := os.Remove(filepath.Join(r.pending(), name)); err != nil {
						t.Fatal(err)
					}
				}
				for i, path := range r.targets {
					if err := os.Remove(filepath.Join(filepath.Dir(path), ".obk-start-"+recoveryToken+"-"+r.value.Roles[i].Role)); err != nil {
						t.Fatal(err)
					}
				}
			case "replacement-namespace":
				if err := os.Rename(r.pending(), r.pending()+"-held"); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(r.pending(), 0700); err != nil {
					t.Fatal(err)
				}
				put(t, filepath.Join(r.pending(), "complete.json"), []byte("{\"version\":1"))
			case "changed-pending-is-unproved":
				if err := os.Remove(filepath.Join(r.pending(), "complete.json")); err != nil {
					t.Fatal(err)
				}
				put(t, filepath.Join(r.pending(), "pending.json"), []byte("changed bytes\n"))
			}
			got := observeNativeInvariants(t, r.sandbox, evidence, &history)
			if got.violation != tc.violation {
				t.Errorf("history classification=%q, want %q; gap=%q", got.violation, tc.violation, got.gap)
			}
		})
	}
}

func TestObserverClassificationInitialStagingHasNoCompletionHistory(t *testing.T) {
	r, evidence := observerControl(t)
	if err := os.Remove(filepath.Join(r.pending(), "pending.json")); err != nil {
		t.Fatal(err)
	}
	history := observationLineage{}
	got := observeNativeInvariants(t, r.sandbox, evidence, &history)
	if got.violation != "" || got.gap == "" {
		t.Error("initial staging without a proved publication history was treated as completion loss")
	}
}
