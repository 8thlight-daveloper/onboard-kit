package invite_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestNonemptyPlanLocationSurvivesSeedAndGuideWithUnusedPreseed(t *testing.T) {
	const expectedHash = "c88a7e524717ff336fec6909dcc526b9a8173f14c6d412ed1716ae817b4f0e61"
	expectedPlan := literal(t, "plan-url.yaml")
	independentDigest := sha256.Sum256(expectedPlan)
	if hex.EncodeToString(independentDigest[:]) != expectedHash {
		t.Fatal("independently pinned literal canonical URL plan hash is wrong")
	}
	s := newSandbox(t)
	writeLiteral(t, s.plan, expectedPlan)
	before := tree(t, s.root)
	readReport(t, run(newCommand(), s.args()), "invite.written", 0, nil, nil)
	preserved(t, s, before, true)
	payloads := exportedPayloads(t, s.out)
	for name, expected := range map[string][]byte{
		"plan.obk.yaml":                         expectedPlan,
		"records/newcomer/preparation.obk.yaml": literal(t, "preparation.yaml"),
		"records/newcomer/record.obk.yaml":      literal(t, "record-url.yaml"),
		"first-steps.md":                        literal(t, "first-steps-url.md"),
	} {
		if !bytes.Equal(payloads[name], expected) {
			t.Fatalf("literal URL/preseed composition differs for %s", name)
		}
	}
	if !bytes.Contains(payloads["plan.obk.yaml"], []byte("personalanswer: orgseed\n")) {
		t.Fatal("unused onboardee preseed must remain in original plan")
	}
	if bytes.Contains(payloads["records/newcomer/record.obk.yaml"], []byte("answers:")) {
		t.Fatal("unused organization preseed must not become a personal answer")
	}
}
