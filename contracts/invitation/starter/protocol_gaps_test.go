package starter_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func starterProtocolGapWriteRaw(t *testing.T, r *recovery, raw []byte) {
	t.Helper()
	r.raw = append([]byte(nil), raw...)
	r.claims.PendingSHA256 = digest(r.raw)
	put(t, filepath.Join(r.pending(), "pending.json"), r.raw)
	put(t, filepath.Join(r.pending(), "publication.json"), jsonBytes(t, r.claims))
	put(t, filepath.Join(r.pending(), "owner.json"), jsonBytes(t, owner{
		Version: 1, Transaction: recoveryToken,
		RootDevice: r.value.RootGuard.Device, RootInode: r.value.RootGuard.Inode,
		Namespace: objectGuard(t, r.pending()), PendingSHA256: digest(r.raw),
	}))
}

func starterProtocolGapReplace(t *testing.T, raw []byte, before, after string) []byte {
	t.Helper()
	if bytes.Count(raw, []byte(before)) != 1 {
		t.Fatal("independent raw protocol edit must reach exactly one specified field")
	}
	return bytes.Replace(raw, []byte(before), []byte(after), 1)
}

func starterProtocolGapRefuseBeforeConsent(t *testing.T, r recovery) {
	t.Helper()
	before := snapshot(t, r.base)
	guards := make(map[string]guard, len(before))
	for path := range before {
		guards[path] = objectGuard(t, filepath.Join(r.base, path))
	}
	clock := 0
	input := &countedReader{reader: strings.NewReader("resume\nexit\nUNCONSUMED_TAIL\n")}
	var stdout, stderr bytes.Buffer
	exit := newStarter(runtimeFacts(t, &clock)).Run(r.args(), input, &stdout, &stderr)
	got := streamResult(t, exit, stdout.String(), stderr.Bytes())
	wantReport(t, got, 1, "refused", "start.context", "Select a safe supported root, state home, and unchanged session context.")
	if got.out != "" || input.consumed != 0 || clock != 0 {
		t.Error("malformed protocol reached a recovery prompt, consumed consent or sampled time")
	}
	unchanged(t, r.sandbox, before)
	for path, expected := range guards {
		if objectGuard(t, filepath.Join(r.base, path)) != expected {
			t.Errorf("malformed protocol changed guarded native controls at %s", path)
		}
	}
}

func starterProtocolGapResumeCanonical(t *testing.T, r recovery, retain bool) {
	t.Helper()
	clock := 0
	input := &countedReader{reader: strings.NewReader("resume\nexit\nUNCONSUMED_TAIL\n")}
	var stdout, stderr bytes.Buffer
	exit := newStarter(runtimeFacts(t, &clock)).Run(r.args(), input, &stdout, &stderr)
	got := streamResult(t, exit, stdout.String(), stderr.Bytes())
	wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	want := "Pending opening: \"" + recoveryToken + "\"\nAccount: \"jules\"\nPlan: \"plan-a\"\n"
	for i, name := range []string{"local-plan", "preparation", "session-metadata", "record"} {
		action := "install"
		if i == 3 {
			action = "create"
		}
		if retain {
			action = "retain"
		}
		want += fmt.Sprintf("Role: %q Action: %q Path: %q\n", name, action, r.targets[i])
	}
	want += "Recovery: choose resume or cancel:\nFirst opened at: \"" + string(openedAt) + "\"\nS\nSession: read only. Choose exit:\n"
	if got.out != want || clock != 0 || input.consumed != len("resume\nexit\n") || got.report.Transaction != recoveryToken {
		t.Error("canonical protocol control did not replay exact frozen context/display without extra input or clock")
	}
	for i, body := range [][]byte{fixture(t, "plan.yaml"), fixture(t, "preparation.yaml"), metadataBytes(t, r.sandbox), fixture(t, "begun.yaml")} {
		assertBytes(t, r.targets[i], body)
	}
	assertIdle(t, r.sandbox)
}

func TestStarterProtocolGapCanonicalInstallAndRetainControls(t *testing.T) {
	for _, retain := range []bool{false, true} {
		t.Run(fmt.Sprint(retain), func(t *testing.T) {
			r := makeRecovery(t, retain)
			r.value.SessionDisplayBase64 = "Uwo="
			rewritePending(t, &r)
			starterProtocolGapResumeCanonical(t, r, retain)
		})
	}
}

func TestStarterProtocolGapAbsentPreimageStillRequiresEmptyBase64Key(t *testing.T) {
	r := makeRecovery(t, false)
	raw := starterProtocolGapReplace(t, r.raw,
		`"role":"local-plan","action":"install","beforePresent":false,"beforeBase64":"",`,
		`"role":"local-plan","action":"install","beforePresent":false,`)
	starterProtocolGapWriteRaw(t, &r, raw)
	starterProtocolGapRefuseBeforeConsent(t, r)
}

func TestStarterProtocolGapUnsupportedIntegerVersionRefuses(t *testing.T) {
	r := makeRecovery(t, false)
	raw := starterProtocolGapReplace(t, r.raw, `{"version":1,`, `{"version":2,`)
	starterProtocolGapWriteRaw(t, &r, raw)
	starterProtocolGapRefuseBeforeConsent(t, r)
}

func TestStarterProtocolGapLoneSurrogatesRefuseBeforeConsent(t *testing.T) {
	for _, escape := range []string{`\ud800`, `\udfff`} {
		t.Run(escape, func(t *testing.T) {
			r := makeRecovery(t, true)
			raw := starterProtocolGapReplace(t, r.raw,
				`"firstOpenedAt":"`+string(openedAt)+`"`,
				`"firstOpenedAt":"`+escape+`"`)
			starterProtocolGapWriteRaw(t, &r, raw)
			starterProtocolGapRefuseBeforeConsent(t, r)
		})
	}
}

func TestStarterProtocolGapValidSupplementaryRootSurrogatePair(t *testing.T) {
	r := makeRecovery(t, false)
	oldRoot, oldRegistration := r.root, r.registration()
	r.root += "-😀"
	if err := os.Rename(oldRoot, r.root); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldRegistration, r.registration()); err != nil {
		t.Fatal(err)
	}
	r.targets = []string{filepath.Join(r.registration(), "plan.obk.yaml"), r.prep(), filepath.Join(r.registration(), "session.json"), r.record()}
	r.value.Root = r.root
	r.value.SessionKey = key(r.root)
	r.value.SessionDisplayBase64 = "Uwo="
	metadataBody := metadataBytes(t, r.sandbox)
	r.value.Roles[2].AfterBase64 = encoded(metadataBody)
	r.value.Roles[2].AfterSHA256 = digest(metadataBody)
	raw := starterProtocolGapReplace(t, jsonBytes(t, r.value), "😀", `\ud83d\ude00`)
	starterProtocolGapWriteRaw(t, &r, raw)
	starterProtocolGapResumeCanonical(t, r, false)
}

func TestStarterProtocolGapBase64WhitespaceRefuses(t *testing.T) {
	for _, value := range []string{" Uwo=", "Uwo= ", "\tUwo=\n"} {
		t.Run(fmt.Sprintf("%q", value), func(t *testing.T) {
			r := makeRecovery(t, true)
			r.value.SessionDisplayBase64 = value
			rewritePending(t, &r)
			starterProtocolGapRefuseBeforeConsent(t, r)
		})
	}
}

func TestStarterProtocolGapPresentPreimageDigestMustMatchHeldBytes(t *testing.T) {
	r := makeRecovery(t, true)
	r.value.Roles[0].BeforeSHA256 = strings.Repeat("0", 64)
	rewritePending(t, &r)
	starterProtocolGapRefuseBeforeConsent(t, r)
}

func TestStarterProtocolGapRetainCannotCarryDifferentAfterProposal(t *testing.T) {
	r := makeRecovery(t, true)
	different := append(append([]byte(nil), fixture(t, "plan.yaml")...), []byte("# distinct frozen proposal\n")...)
	r.value.Roles[0].AfterBase64 = encoded(different)
	r.value.Roles[0].AfterSHA256 = digest(different)
	rewritePending(t, &r)
	starterProtocolGapRefuseBeforeConsent(t, r)
}

func TestStarterProtocolGapParentOrdinalsMustMatchTheirPositions(t *testing.T) {
	for _, kind := range []string{"wrong-first", "duplicate", "reversed"} {
		t.Run(kind, func(t *testing.T) {
			r := makeRecovery(t, false)
			switch kind {
			case "wrong-first":
				r.value.Roles[0].Parents[0].Ordinal = "9"
			case "duplicate":
				r.value.Roles[0].Parents[1].Ordinal = "0"
			case "reversed":
				r.value.Roles[0].Parents[0].Ordinal = "1"
				r.value.Roles[0].Parents[1].Ordinal = "0"
			}
			rewritePending(t, &r)
			starterProtocolGapRefuseBeforeConsent(t, r)
		})
	}
}

func starterProtocolGapAbsentParents(t *testing.T) (recovery, guard) {
	t.Helper()
	r := makeRecovery(t, false)
	formerChild := *r.value.Roles[0].Parents[1].Guard
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
	r.value.SessionDisplayBase64 = "Uwo="
	return r, formerChild
}

func TestStarterProtocolGapAllAbsentParentTailControl(t *testing.T) {
	r, _ := starterProtocolGapAbsentParents(t)
	rewritePending(t, &r)
	starterProtocolGapResumeCanonical(t, r, false)
}

func TestStarterProtocolGapMissingParentPresentChildRefusesBeforeCreation(t *testing.T) {
	r, formerChild := starterProtocolGapAbsentParents(t)
	r.value.Roles[0].Parents[1].Present = true
	r.value.Roles[0].Parents[1].Guard = &formerChild
	rewritePending(t, &r)
	starterProtocolGapRefuseBeforeConsent(t, r)
}
