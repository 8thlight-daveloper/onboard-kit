package starter_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

const recoveryToken = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

type security struct {
	UID            string `json:"uid"`
	GID            string `json:"gid"`
	Mode           string `json:"mode"`
	ACLFormat      string `json:"aclFormat"`
	AccessPresent  bool   `json:"accessPresent"`
	AccessBase64   string `json:"accessBase64"`
	DefaultPresent bool   `json:"defaultPresent"`
	DefaultBase64  string `json:"defaultBase64"`
}
type guard struct {
	Kind     string   `json:"kind"`
	Device   string   `json:"device"`
	Inode    string   `json:"inode"`
	Security security `json:"security"`
}
type parent struct {
	Ordinal string `json:"ordinal"`
	Present bool   `json:"present"`
	Guard   *guard `json:"guard"`
}
type role struct {
	Role          string   `json:"role"`
	Action        string   `json:"action"`
	BeforePresent bool     `json:"beforePresent"`
	BeforeBase64  string   `json:"beforeBase64"`
	BeforeSHA256  string   `json:"beforeSHA256"`
	BeforeGuard   *guard   `json:"beforeGuard"`
	AfterBase64   string   `json:"afterBase64"`
	AfterSHA256   string   `json:"afterSHA256"`
	Parents       []parent `json:"parents"`
}
type pending struct {
	Version              int    `json:"version"`
	Transaction          string `json:"transaction"`
	SessionKey           string `json:"sessionKey"`
	PlanIDBase64         string `json:"planIdBase64"`
	Account              string `json:"account"`
	Root                 string `json:"root"`
	RootGuard            guard  `json:"rootGuard"`
	Home                 string `json:"home"`
	HomeGuard            guard  `json:"homeGuard"`
	Record               string `json:"record"`
	Preparation          string `json:"preparation"`
	FirstOpenedAt        string `json:"firstOpenedAt"`
	SessionDisplayBase64 string `json:"sessionDisplayBase64"`
	Roles                []role `json:"roles"`
}
type owner struct {
	Version          int    `json:"version"`
	Transaction      string `json:"transaction"`
	RootDevice       string `json:"rootDevice"`
	RootInode        string `json:"rootInode"`
	Namespace        guard  `json:"namespace"`
	PendingSHA256    string `json:"pendingSHA256"`
	PendingTemporary *guard `json:"pendingTemporary"`
}
type newParent struct {
	Ordinal string `json:"ordinal"`
	Guard   guard  `json:"guard"`
}
type claimRole struct {
	Role       string      `json:"role"`
	NewParents []newParent `json:"newParents"`
	Staging    *guard      `json:"staging"`
	Temporary  *guard      `json:"temporary"`
}
type publication struct {
	Version       int         `json:"version"`
	Transaction   string      `json:"transaction"`
	PendingSHA256 string      `json:"pendingSHA256"`
	Roles         []claimRole `json:"roles"`
}
type completeRole struct {
	Role           string `json:"role"`
	SHA256         string `json:"sha256"`
	Device         string `json:"device"`
	Inode          string `json:"inode"`
	SecuritySHA256 string `json:"securitySHA256"`
}
type complete struct {
	Version                 int            `json:"version"`
	Transaction             string         `json:"transaction"`
	SessionKey              string         `json:"sessionKey"`
	PendingSHA256           string         `json:"pendingSHA256"`
	RootDevice              string         `json:"rootDevice"`
	RootInode               string         `json:"rootInode"`
	HomeDevice              string         `json:"homeDevice"`
	HomeInode               string         `json:"homeInode"`
	NamespaceDevice         string         `json:"namespaceDevice"`
	NamespaceInode          string         `json:"namespaceInode"`
	NamespaceSecuritySHA256 string         `json:"namespaceSecuritySHA256"`
	Roles                   []completeRole `json:"roles"`
}
type metadata struct {
	Version      int    `json:"version"`
	SessionKey   string `json:"sessionKey"`
	PlanIDBase64 string `json:"planIdBase64"`
	Account      string `json:"account"`
	Root         string `json:"root"`
	RootDevice   string `json:"rootDevice"`
	RootInode    string `json:"rootInode"`
	Record       string `json:"record"`
	Preparation  string `json:"preparation"`
	PlanSHA256   string `json:"planSHA256"`
	Transaction  string `json:"transaction"`
}
type recovery struct {
	sandbox
	value   pending
	raw     []byte
	claims  publication
	targets []string
}

func digest(b []byte) string  { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }
func encoded(b []byte) string { return base64.StdEncoding.EncodeToString(b) }
func jsonBytes(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return append(b, '\n')
}

func objectGuard(t *testing.T, path string) guard {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatal("native stat unavailable")
	}
	kind := "regular"
	if info.IsDir() {
		kind = "directory"
	} else if !info.Mode().IsRegular() {
		t.Fatal("guard fixture must be regular or directory")
	}
	format := "linux-posix-v1"
	if runtime.GOOS == "darwin" {
		format = "darwin-filesec-v1"
	}
	sec := security{UID: fmt.Sprint(stat.Uid), GID: fmt.Sprint(stat.Gid), Mode: fmt.Sprintf("%04o", uint32(stat.Mode)&07777), ACLFormat: format}
	sec.AccessPresent, sec.AccessBase64, sec.DefaultPresent, sec.DefaultBase64 = nativeACL(t, path)
	return guard{Kind: kind, Device: fmt.Sprint(stat.Dev), Inode: fmt.Sprint(stat.Ino), Security: sec}
}

func metadataBytes(t *testing.T, s sandbox) []byte {
	g := objectGuard(t, s.root)
	return jsonBytes(t, metadata{Version: 1, SessionKey: key(s.root), PlanIDBase64: encoded([]byte("plan-a")), Account: "jules", Root: s.root, RootDevice: g.Device, RootInode: g.Inode, Record: "records/jules/record.obk.yaml", Preparation: "records/jules/preparation.obk.yaml", PlanSHA256: rawPlanHash, Transaction: recoveryToken})
}

func makeRecovery(t *testing.T, retain bool) recovery {
	t.Helper()
	s := setup(t)
	for _, path := range []string{s.pending(), s.registration(), filepath.Dir(s.record())} {
		if err := os.MkdirAll(path, 0700); err != nil {
			t.Fatal(err)
		}
	}
	targets := []string{filepath.Join(s.registration(), "plan.obk.yaml"), s.prep(), filepath.Join(s.registration(), "session.json"), s.record()}
	bodies := [][]byte{fixture(t, "plan.yaml"), fixture(t, "preparation.yaml"), metadataBytes(t, s), fixture(t, "begun.yaml")}
	if retain {
		for i, path := range targets {
			put(t, path, bodies[i])
		}
	}
	value := pending{Version: 1, Transaction: recoveryToken, SessionKey: key(s.root), PlanIDBase64: encoded([]byte("plan-a")), Account: "jules", Root: s.root, RootGuard: objectGuard(t, s.root), Home: s.home, HomeGuard: objectGuard(t, s.home), Record: "records/jules/record.obk.yaml", Preparation: "records/jules/preparation.obk.yaml", FirstOpenedAt: string(openedAt), SessionDisplayBase64: encoded(fixture(t, "session.txt"))}
	claims := publication{Version: 1, Transaction: recoveryToken}
	for i, name := range []string{"local-plan", "preparation", "session-metadata", "record"} {
		base := s.home
		rel := "sessions/" + key(s.root)
		if i == 1 || i == 3 {
			base = s.root
			rel = "records/jules"
		}
		parents := []parent{}
		for j, part := range strings.Split(rel, "/") {
			base = filepath.Join(base, part)
			g := objectGuard(t, base)
			parents = append(parents, parent{Ordinal: fmt.Sprint(j), Present: true, Guard: &g})
		}
		action := "install"
		if i == 3 {
			action = "create"
		}
		item := role{Role: name, Action: action, AfterBase64: encoded(bodies[i]), AfterSHA256: digest(bodies[i]), Parents: parents}
		if retain {
			g := objectGuard(t, targets[i])
			item.Action = "retain"
			item.BeforePresent = true
			item.BeforeBase64 = encoded(bodies[i])
			item.BeforeSHA256 = digest(bodies[i])
			item.BeforeGuard = &g
		}
		value.Roles = append(value.Roles, item)
		claims.Roles = append(claims.Roles, claimRole{Role: name, NewParents: []newParent{}})
	}
	raw := jsonBytes(t, value)
	claims.PendingSHA256 = digest(raw)
	put(t, filepath.Join(s.pending(), "owner.json"), jsonBytes(t, owner{Version: 1, Transaction: recoveryToken, RootDevice: value.RootGuard.Device, RootInode: value.RootGuard.Inode, Namespace: objectGuard(t, s.pending()), PendingSHA256: digest(raw)}))
	put(t, filepath.Join(s.pending(), "pending.json"), raw)
	put(t, filepath.Join(s.pending(), "publication.json"), jsonBytes(t, claims))
	return recovery{sandbox: s, value: value, raw: raw, claims: claims, targets: targets}
}

func (r recovery) args() []string {
	return []string{"--resume-import", recoveryToken, "--root", r.root, "--home", r.home}
}

func installReceipt(t *testing.T, r recovery) {
	t.Helper()
	namespace := objectGuard(t, r.pending())
	c := complete{Version: 1, Transaction: recoveryToken, SessionKey: key(r.root), PendingSHA256: digest(r.raw), RootDevice: r.value.RootGuard.Device, RootInode: r.value.RootGuard.Inode, HomeDevice: r.value.HomeGuard.Device, HomeInode: r.value.HomeGuard.Inode, NamespaceDevice: namespace.Device, NamespaceInode: namespace.Inode}
	securityJSON, err := json.Marshal(namespace.Security)
	if err != nil {
		t.Fatal(err)
	}
	c.NamespaceSecuritySHA256 = digest(securityJSON)
	for i, path := range r.targets {
		g := objectGuard(t, path)
		b, err := json.Marshal(g.Security)
		if err != nil {
			t.Fatal(err)
		}
		c.Roles = append(c.Roles, completeRole{Role: r.value.Roles[i].Role, SHA256: r.value.Roles[i].AfterSHA256, Device: g.Device, Inode: g.Inode, SecuritySHA256: digest(b)})
	}
	put(t, filepath.Join(r.pending(), "complete.json"), jsonBytes(t, c))
}

func TestPendingRecoveryUsesFrozenBytesWithoutArchiveOrClock(t *testing.T) {
	for _, retain := range []bool{false, true} {
		r := makeRecovery(t, retain)
		if err := os.Remove(r.invitation); err != nil {
			t.Fatal(err)
		}
		clock := 0
		got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "resume\nexit\n")
		wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
		if clock != 0 || got.report.Transaction != recoveryToken {
			t.Error("recovery resampled event or changed token")
		}
		for i, path := range r.targets {
			b, err := base64.StdEncoding.DecodeString(r.value.Roles[i].AfterBase64)
			if err != nil {
				t.Fatal(err)
			}
			assertBytes(t, path, b)
		}
		if !strings.HasSuffix(got.out, "First opened at: \""+string(openedAt)+"\"\n"+string(fixture(t, "session.txt"))+"Session: read only. Choose exit:\n") {
			t.Error("ZIP-free recovery did not replay exact frozen session")
		}
		assertIdle(t, r.sandbox)
	}
}

func TestCompletedCleanupNeverReplaysAgainstLaterOwnerEdits(t *testing.T) {
	for _, pendingRemains := range []bool{false, true} {
		r := makeRecovery(t, true)
		installReceipt(t, r)
		if !pendingRemains {
			for _, name := range []string{"pending.json", "owner.json", "publication.json"} {
				if err := os.Remove(filepath.Join(r.pending(), name)); err != nil {
					t.Fatal(err)
				}
			}
		}
		put(t, r.record(), []byte("later owner record: do not decode or overwrite\n"))
		put(t, filepath.Join(r.registration(), "session.json"), []byte("later owner metadata\n"))
		put(t, r.targets[0], []byte("later owner plan\n"))
		before := snapshot(t, r.root)
		clock := 0
		got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "clean\nexit\n")
		if pendingRemains {
			wantReport(t, got, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
		} else {
			wantReport(t, got, 0, "cleaned", "start.cleaned", "Completed recovery data was cleaned; no session was displayed.")
			if strings.Contains(got.out, "Read-only onboarding session.") {
				t.Error("lost display was reconstructed")
			}
		}
		if !reflectObjects(snapshot(t, r.root), before) || clock != 0 {
			t.Error("completed cleanup reinterpreted later owner work")
		}
		assertBytes(t, r.targets[0], []byte("later owner plan\n"))
		assertBytes(t, r.targets[2], []byte("later owner metadata\n"))
		assertIdle(t, r.sandbox)
	}
}

func reflectObjects(a, b map[string]object) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func TestRecoveryCancelAndInputFailureRetainPriorTruth(t *testing.T) {
	for _, completed := range []bool{false, true} {
		for _, input := range []string{"cancel\n", "", string([]byte{0xff}) + "\n"} {
			r := makeRecovery(t, true)
			if completed {
				installReceipt(t, r)
			}
			before := snapshot(t, r.base)
			clock := 0
			got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), input)
			if strings.Contains(input, string([]byte{0xff})) {
				status := "pending"
				if completed {
					status = "opened"
				}
				wantReport(t, got, 1, status, "start.recovery-input", "Existing recovery state was retained after an input error.")
			} else {
				wantReport(t, got, 0, "retained", "start.recovery-retained", "Existing recovery state was retained.")
			}
			if clock != 0 {
				t.Error("retained recovery sampled clock")
			}
			unchanged(t, r.sandbox, before)
		}
	}
}

func TestStrictPendingProtocolRefusesWithoutRepairing(t *testing.T) {
	for _, mutation := range []string{"duplicate", "unknown", "version-number", "digest", "roles-reordered", "base64-extra-padding", "base64-padbits", "terminal-control", "display-utf8", "after-absent", "trailing-json", "owner-inode", "extra-child"} {
		t.Run(mutation, func(t *testing.T) {
			r := makeRecovery(t, false)
			raw := append([]byte(nil), r.raw...)
			switch mutation {
			case "duplicate":
				raw = bytes.Replace(raw, []byte(`"version":1`), []byte(`"version":1,"version":1`), 1)
			case "unknown":
				raw = bytes.Replace(raw, []byte(`"version":1`), []byte(`"version":1,"extra":false`), 1)
			case "version-number":
				raw = bytes.Replace(raw, []byte(`"version":1`), []byte(`"version":1.0`), 1)
			case "digest":
				r.value.Roles[0].AfterSHA256 = strings.Repeat("0", 64)
				raw = jsonBytes(t, r.value)
			case "roles-reordered":
				r.value.Roles[0], r.value.Roles[1] = r.value.Roles[1], r.value.Roles[0]
				raw = jsonBytes(t, r.value)
			case "base64-extra-padding":
				r.value.PlanIDBase64 = "cGxhbi1h=="
				raw = jsonBytes(t, r.value)
			case "base64-padbits":
				r.value.SessionDisplayBase64 = "Uwp="
				raw = jsonBytes(t, r.value)
			case "terminal-control":
				r.value.SessionDisplayBase64 = encoded([]byte("unsafe\x1b[2J\n"))
				raw = jsonBytes(t, r.value)
			case "display-utf8":
				r.value.SessionDisplayBase64 = encoded([]byte{0xff})
				raw = jsonBytes(t, r.value)
			case "after-absent":
				raw = bytes.Replace(raw, []byte(`"afterBase64":"`+r.value.Roles[0].AfterBase64+`",`), nil, 1)
			case "trailing-json":
				raw = append(raw, []byte("{}\n")...)
			case "owner-inode":
				body, err := os.ReadFile(filepath.Join(r.pending(), "owner.json"))
				if err != nil {
					t.Fatal(err)
				}
				var o owner
				if err = json.Unmarshal(body, &o); err != nil {
					t.Fatal(err)
				}
				o.Namespace.Inode = "0"
				put(t, filepath.Join(r.pending(), "owner.json"), jsonBytes(t, o))
			case "extra-child":
				put(t, filepath.Join(r.pending(), "owner-unattributed"), []byte("do not delete\n"))
			}
			put(t, filepath.Join(r.pending(), "pending.json"), raw)
			if mutation != "owner-inode" {
				o := owner{Version: 1, Transaction: recoveryToken, RootDevice: r.value.RootGuard.Device, RootInode: r.value.RootGuard.Inode, Namespace: objectGuard(t, r.pending()), PendingSHA256: digest(raw)}
				put(t, filepath.Join(r.pending(), "owner.json"), jsonBytes(t, o))
				r.claims.PendingSHA256 = digest(raw)
				put(t, filepath.Join(r.pending(), "publication.json"), jsonBytes(t, r.claims))
			}
			before := snapshot(t, r.base)
			clock := 0
			got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "resume\nexit\n")
			if got.exit != 1 || got.report.Code != "start.context" || clock != 0 {
				t.Error("malformed recovery protocol was trusted")
			}
			unchanged(t, r.sandbox, before)
		})
	}
}

func TestPendingConflictPreservesEveryParticipant(t *testing.T) {
	r := makeRecovery(t, false)
	put(t, r.record(), []byte("unrelated owner work\n"))
	before := snapshot(t, r.base)
	clock := 0
	got := invoke(t, newStarter(runtimeFacts(t, &clock)), r.args(), "resume\n")
	if got.exit != 1 || clock != 0 {
		t.Error("unexpected target must refuse")
	}
	for _, path := range r.targets[:3] {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Error("preflight conflict allowed earlier role publication")
		}
	}
	unchanged(t, r.sandbox, before)
}
