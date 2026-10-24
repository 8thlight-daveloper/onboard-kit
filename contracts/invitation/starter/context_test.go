package starter_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"onboard-kit/internal/model"
)

func literalPaths(t *testing.T, s sandbox, name string) string {
	t.Helper()
	return strings.NewReplacer("@ROOT@", s.root, "@HOME@", s.home, "@LOCAL_PLAN@", filepath.Join(s.registration(), "plan.obk.yaml"), "@PREPARATION@", s.prep(), "@RECORD@", s.record(), "@METADATA@", filepath.Join(s.registration(), "session.json"), "@PENDING@", s.pending()).Replace(string(fixture(t, name)))
}

func TestExactIdentityAndCompleteReview(t *testing.T) {
	s := setup(t)
	before := snapshot(t, s.base)
	clock := 0
	identity := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "cancel\n")
	if identity.out != literalPaths(t, s, "identity.txt") {
		t.Error("exact identity/privacy block differs")
	}
	review := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\ncancel\n")
	if review.out != literalPaths(t, s, "identity.txt")+literalPaths(t, s, "fresh-review.txt") {
		t.Error("exact complete prebegin action/reconciliation/privacy review differs")
	}
	if clock != 0 {
		t.Error("review sampled event")
	}
	unchanged(t, s, before)
}

func TestConfigurationPrecedenceAndRelativeAnchoring(t *testing.T) {
	for _, mode := range []string{"explicit-equals", "environment", "platform-default", "relative"} {
		t.Run(mode, func(t *testing.T) {
			s := setup(t)
			before := snapshot(t, s.base)
			clock := 0
			facts := runtimeFacts(t, &clock)
			args := s.args()
			var env []string
			cwd, config, home := 0, 0, 0
			switch mode {
			case "explicit-equals":
				args = []string{"--home=" + s.home, "--root=" + s.root, "--invitation=" + s.invitation}
			case "environment":
				args = args[:4]
				facts.LookupEnvironment = func(name string) string { env = append(env, name); return s.home }
			case "platform-default":
				args = args[:4]
				facts.LookupEnvironment = func(name string) string {
					env = append(env, name)
					if name == "XDG_STATE_HOME" {
						return filepath.Join(s.base, "state")
					}
					return ""
				}
				facts.UserConfigDirectory = func() (string, error) { config++; return filepath.Join(s.base, "config"), nil }
				facts.UserHomeDirectory = func() (string, error) { home++; return filepath.Join(s.base, "user"), nil }
			case "relative":
				args = []string{"--invitation=invitation.zip", "--root=root", "--home=home"}
				facts.WorkingDirectory = func() (string, error) { cwd++; return s.base, nil }
			}
			got := invoke(t, newStarter(facts), args, "cancel\n")
			wantReport(t, got, 0, "cancelled", "start.cancelled", "Opening cancelled before begin; no files were changed.")
			if mode == "environment" && !reflect.DeepEqual(env, []string{"OBK_HOME"}) {
				t.Error("environment precedence/call count differs")
			}
			if mode == "platform-default" {
				if runtime.GOOS == "darwin" {
					if !reflect.DeepEqual(env, []string{"OBK_HOME"}) || config != 1 || home != 0 || !strings.Contains(got.out, filepath.Join(s.base, "config", "obk")) {
						t.Error("Darwin default precedence differs")
					}
				} else {
					if !reflect.DeepEqual(env, []string{"OBK_HOME", "XDG_STATE_HOME"}) || config != 0 || home != 0 || !strings.Contains(got.out, filepath.Join(s.base, "state", "obk")) {
						t.Error("Linux default precedence differs")
					}
				}
			}
			if mode == "relative" && cwd != 1 {
				t.Error("relative paths must share one cwd fact")
			}
			if clock != 0 {
				t.Error("configuration/identity sampled event")
			}
			unchanged(t, s, before)
		})
	}
}

func TestBadConfiguredHomeNeverFallsBack(t *testing.T) {
	s := setup(t)
	bad := filepath.Join(s.base, "not-directory")
	put(t, bad, []byte("owner file\n"))
	before := snapshot(t, s.base)
	clock := 0
	facts := runtimeFacts(t, &clock)
	calls := 0
	facts.LookupEnvironment = func(name string) string {
		calls++
		if name != "OBK_HOME" {
			t.Error("fallback environment read")
		}
		return bad
	}
	got := invoke(t, newStarter(facts), s.args()[:4], "confirm jules\nbegin\n")
	wantReport(t, got, 1, "refused", "start.context", "Select a safe supported root, state home, and unchanged session context.")
	if calls != 1 || clock != 0 {
		t.Error("bad configuration fallback or event")
	}
	unchanged(t, s, before)
}

func TestEmptyIdleReuseAndUnknownStateRefusal(t *testing.T) {
	for _, kind := range []string{"safe-empty", "unsafe-mode", "unknown-child", "unclaimed-pending-temp"} {
		t.Run(kind, func(t *testing.T) {
			s := setup(t)
			if err := os.MkdirAll(s.pending(), 0700); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "unsafe-mode":
				if err := os.Chmod(s.pending(), 0755); err != nil {
					t.Fatal(err)
				}
			case "unknown-child":
				put(t, filepath.Join(s.pending(), "unknown"), []byte("preserve\n"))
			case "unclaimed-pending-temp":
				put(t, filepath.Join(s.pending(), ".pending.next"), nil)
			}
			before := snapshot(t, s.base)
			clock := 0
			got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\nexit\n")
			if kind == "safe-empty" {
				wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
				if clock != 1 {
					t.Error("idle fresh opening lacked event")
				}
				assertIdle(t, s)
				if snapshot(t, s.base)["home/import-pending"].Inode != before["home/import-pending"].Inode {
					t.Error("idle container replaced rather than reused")
				}
			} else {
				wantReport(t, got, 1, "refused", "start.context", "Select a safe supported root, state home, and unchanged session context.")
				if clock != 0 {
					t.Error("unknown state sampled event")
				}
				unchanged(t, s, before)
			}
		})
	}
}

func initialOpening(t *testing.T, s sandbox) {
	t.Helper()
	clock := 0
	r := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\nexit\n")
	if r.exit != 0 || clock != 1 {
		t.Fatal("initial opening failed")
	}
}

func TestRegisteredContextDoesNotReseedOrMove(t *testing.T) {
	for _, change := range []string{"missing-record", "replaced-root", "safe-dot-alias", "ordinary-repeat"} {
		t.Run(change, func(t *testing.T) {
			s := setup(t)
			initialOpening(t, s)
			args := s.args()
			switch change {
			case "missing-record":
				if err := os.Rename(s.record(), s.record()+".held"); err != nil {
					t.Fatal(err)
				}
			case "replaced-root":
				if err := os.Rename(s.root, s.root+"-held"); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(s.root, 0700); err != nil {
					t.Fatal(err)
				}
			case "safe-dot-alias":
				args[3] = s.root + "/."
			}
			before := snapshot(t, s.base)
			clock := 0
			got := invoke(t, newStarter(runtimeFacts(t, &clock)), args, "confirm jules\nbegin\nexit\n")
			if change == "ordinary-repeat" || change == "safe-dot-alias" {
				wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
				if clock != 0 {
					t.Error("repeat sampled new first event")
				}
				unchanged(t, s, before)
			} else {
				if got.exit != 1 || clock != 0 {
					t.Error("registered missing/replaced context was reseeded")
				}
				unchanged(t, s, before)
			}
		})
	}
}

func TestDistinctValidRootsRequireExplicitIndependentContextReview(t *testing.T) {
	s := setup(t)
	initialOpening(t, s)
	old := snapshot(t, s.root)
	second := s
	second.root = filepath.Join(s.base, "other-root")
	if err := os.Mkdir(second.root, 0700); err != nil {
		t.Fatal(err)
	}
	clock := 0
	facts := runtimeFacts(t, &clock)
	later := model.Timestamp("2026-10-24T09:00:00Z")
	facts.Clock = func() model.Timestamp { clock++; return later }
	got := invoke(t, newStarter(facts), second.args(), "confirm jules\nbegin\nexit\n")
	wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	if clock != 1 || !strings.Contains(got.out, "Other matching contexts: 1\n") || !strings.Contains(got.out, "This opening registers an independent local context; other contexts are not copied or moved.\n") || !strings.Contains(got.out, "Begin creates an independent record and session here; other matching work remains untouched.\n") {
		t.Error("independent context was hidden or event copied")
	}
	if !reflect.DeepEqual(snapshot(t, s.root), old) {
		t.Error("old context changed")
	}
	assertBytes(t, second.record(), []byte(strings.Replace(string(fixture(t, "begun.yaml")), string(openedAt), string(later), 1)))
	assertIdle(t, second)
}

func TestRegistryBoundAndRelevantOldRootContinuity(t *testing.T) {
	for _, mode := range []string{"entry-cap", "missing-old-root"} {
		s := setup(t)
		if mode == "entry-cap" {
			for i := 0; i < 129; i++ {
				if err := os.MkdirAll(filepath.Join(s.home, "sessions", fmt.Sprintf("entry-%03d", i)), 0700); err != nil {
					t.Fatal(err)
				}
			}
		} else {
			initialOpening(t, s)
			if err := os.Rename(s.root, s.root+"-moved"); err != nil {
				t.Fatal(err)
			}
			s.root = filepath.Join(s.base, "other-root")
			if err := os.Mkdir(s.root, 0700); err != nil {
				t.Fatal(err)
			}
		}
		before := snapshot(t, s.base)
		clock := 0
		got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\n")
		if got.exit != 1 || clock != 0 {
			t.Error("unbounded/invalid registration accepted")
		}
		unchanged(t, s, before)
	}
}

func TestRegistryInclusive128ValidEntries(t *testing.T) {
	for _, count := range []int{128, 129} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			s := setup(t)
			for i := 0; i < count; i++ {
				account := fmt.Sprintf("person%03d", i)
				root := filepath.Join(s.base, "contexts", account)
				if err := os.MkdirAll(root, 0700); err != nil {
					t.Fatal(err)
				}
				g := objectGuard(t, root)
				id := sessionKey("plan-a", account, root)
				record := "records/" + account + "/record.obk.yaml"
				preparation := "records/" + account + "/preparation.obk.yaml"
				metadata := metadata{Version: 1, SessionKey: id, PlanIDBase64: encoded([]byte("plan-a")), Account: account, Root: root, RootDevice: g.Device, RootInode: g.Inode, Record: record, Preparation: preparation, PlanSHA256: rawPlanHash, Transaction: recoveryToken}
				registration := filepath.Join(s.home, "sessions", id)
				put(t, filepath.Join(registration, "session.json"), jsonBytes(t, metadata))
				put(t, filepath.Join(registration, "plan.obk.yaml"), fixture(t, "plan.yaml"))
				put(t, filepath.Join(root, filepath.FromSlash(record)), []byte(strings.Replace(string(fixture(t, "existing.yaml")), "account: jules", "account: "+account, 1)))
				put(t, filepath.Join(root, filepath.FromSlash(preparation)), []byte(strings.Replace(string(fixture(t, "preparation.yaml")), "account: jules", "account: "+account, 1)))
			}
			before := snapshot(t, s.base)
			clock := 0
			got := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "cancel\n")
			if count == 128 {
				wantReport(t, got, 0, "cancelled", "start.cancelled", "Opening cancelled before begin; no files were changed.")
				if got.out != literalPaths(t, s, "identity.txt") {
					t.Error("valid unrelated entries blocked identity or were exposed")
				}
			} else if got.exit != 1 || got.report.Code != "start.context" {
				t.Error("129 valid entries did not isolate registry cap")
			}
			if clock != 0 {
				t.Error("registry scan sampled first-use event")
			}
			unchanged(t, s, before)
		})
	}
}
