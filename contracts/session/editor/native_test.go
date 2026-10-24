//go:build darwin || linux

package editor_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
	contract "onboard-kit/contracts/session/editor"
)

const editEvent = "2026-10-25T10:11:12.3400+00:00"
const confirm = "confirm \"newcomer\"\n"
const actionPrompt = "Action: choose one displayed change or exit:\n"
const savePrompt = "Review complete. Choose save or cancel:\n"
const planBody = "schemaVersion: 1\nid: plan-one\nanswers:\n  sections:\n    personal:\n      choice: org-a\nsections:\n  - id: personal\n    title: Personal choices\n    summary: Use $choice.\n    variables:\n      - name: choice\n        type: string\n        audience: onboardee\n"
const prepBody = "schemaVersion: 1\nkind: preparation\naccount: newcomer\nplanId: plan-one\n"
const recordBody = "schemaVersion: 1\nkind: record\naccount: newcomer\nstartDate: legacy-date\nfirstOpenedAt: held-event\nplan:\n  id: plan-one\n  hash: held-reference\n  url: plans/shared.obk.yaml\nfuture:\n  emptyMap: {}\n  emptyList: []\n"
const ownedRecord = "schemaVersion: 1\nkind: record\naccount: newcomer\nstartDate: legacy-date\nfirstOpenedAt: held-event\nplan:\n  id: plan-one\n  hash: held-reference\n  url: plans/shared.obk.yaml\nanswers:\n  sections:\n    personal:\n      choice: org-a\nfuture:\n  emptyList: []\n  emptyMap: {}\n"
const followerRecord = "schemaVersion: 1\nkind: record\naccount: newcomer\nstartDate: legacy-date\nfirstOpenedAt: held-event\nplan:\n  id: plan-one\n  hash: held-reference\n  url: plans/shared.obk.yaml\nfuture:\n  emptyList: []\n  emptyMap: {}\n"
const outcomePlan = "schemaVersion: 1\nid: plan-one\nsections:\n  - id: access\n    title: Access\n    summary: Request access\n"
const verifiedRecord = "schemaVersion: 1\nkind: record\naccount: newcomer\nstartDate: legacy-date\nfirstOpenedAt: held-event\nplan:\n  id: plan-one\n  hash: held-reference\n  url: plans/shared.obk.yaml\noutcomes:\n  - section: access\n    outcome: verified\n    on: \"2026-10-25T10:11:12.3400+00:00\"\n    sectionHash: \"6c0301d995d0a3b0dab95a641a087120c5266746e59c50d934c575e135ca2fdc\"\nfuture:\n  emptyList: []\n  emptyMap: {}\n"

type clockFact struct {
	calls int
	value string
}

func (c *clockFact) Now() string { c.calls++; return c.value }

type nativeObject struct {
	Mode                           os.FileMode
	Device, Inode, UID, GID, Flags uint64
	ACL                            string
	Attributes                     map[string]string
	Body                           string
}

func nativeObjectAt(t *testing.T, path string) nativeObject {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	stat := info.Sys().(*syscall.Stat_t)
	object := nativeObject{Mode: info.Mode(), Device: uint64(stat.Dev), Inode: uint64(stat.Ino), UID: uint64(stat.Uid), GID: uint64(stat.Gid)}
	if info.Mode()&os.ModeSymlink != 0 {
		object.Body, err = os.Readlink(path)
	} else if info.IsDir() || info.Mode().IsRegular() {
		object.ACL = nativeACL(t, path)
		object.Flags = nativeFlags(t, path, info)
		object.Attributes = nativeAttributes(t, path)
		if info.Mode().IsRegular() {
			var body []byte
			body, err = os.ReadFile(path)
			object.Body = string(body)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	return object
}

func nativeAttributes(t *testing.T, path string) map[string]string {
	t.Helper()
	n, err := unix.Listxattr(path, nil)
	if err != nil || n > 65536 {
		t.Fatalf("native xattr enumeration unavailable or oversized: %v", err)
	}
	buffer := make([]byte, n)
	used, err := unix.Listxattr(path, buffer)
	if err != nil || used != n {
		t.Fatal("native xattr list changed")
	}
	result := map[string]string{}
	total := 0
	for _, name := range strings.Split(string(buffer), "\x00") {
		if name == "" {
			continue
		}
		size, err := unix.Getxattr(path, name, nil)
		if err != nil || size > 65536 {
			t.Fatal("native attribute inspection unavailable or oversized")
		}
		body := make([]byte, size)
		count, err := unix.Getxattr(path, name, body)
		if err != nil || count != size {
			t.Fatal("native attribute changed")
		}
		total += size
		if len(result) >= 256 || total > 1048576 {
			t.Fatal("native attribute fixture exceeds profile")
		}
		result[name] = string(body)
	}
	return result
}

func put(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chown(path, -1, os.Getegid()); err != nil {
		t.Fatal(err)
	}
}

func sessionKey(plan, account, root string) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte("onboard-kit/session-key/v1\x00"))
	for _, value := range []string{plan, account, root} {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(value)))
		_, _ = hash.Write(length[:])
		_, _ = hash.Write([]byte(value))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

type nativeSession struct{ base, root, home, key, plan, preparation, record, metadata string }

func session(t *testing.T, plan, record string) nativeSession {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := nativeSession{base: base, root: filepath.Join(base, "root"), home: filepath.Join(base, "home")}
	s.key = sessionKey("plan-one", "newcomer", s.root)
	s.plan = filepath.Join(s.home, "sessions", s.key, "plan.obk.yaml")
	s.metadata = filepath.Join(s.home, "sessions", s.key, "session.json")
	s.record = filepath.Join(s.root, "records", "newcomer", "record.obk.yaml")
	s.preparation = filepath.Join(s.root, "records", "newcomer", "preparation.obk.yaml")
	put(t, s.plan, plan)
	put(t, s.record, record)
	put(t, s.preparation, prepBody)
	root := nativeObjectAt(t, s.root)
	digest := sha256.Sum256([]byte(plan))
	metadata := struct {
		Version     int    `json:"version"`
		SessionKey  string `json:"sessionKey"`
		PlanID      string `json:"planIdBase64"`
		Account     string `json:"account"`
		Root        string `json:"root"`
		Device      string `json:"rootDevice"`
		Inode       string `json:"rootInode"`
		Record      string `json:"record"`
		Preparation string `json:"preparation"`
		Hash        string `json:"planSHA256"`
		Transaction string `json:"transaction"`
	}{1, s.key, base64.StdEncoding.EncodeToString([]byte("plan-one")), "newcomer", s.root, strconv.FormatUint(root.Device, 10), strconv.FormatUint(root.Inode, 10), "records/newcomer/record.obk.yaml", "records/newcomer/preparation.obk.yaml", hex.EncodeToString(digest[:]), strings.Repeat("a", 64)}
	body, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	put(t, s.metadata, string(append(body, '\n')))
	for _, path := range []string{s.root, s.home, filepath.Dir(s.metadata), filepath.Join(s.home, "sessions"), filepath.Dir(s.record), filepath.Join(s.root, "records")} {
		if err := os.Chown(path, -1, os.Getegid()); err != nil {
			t.Fatal(err)
		}
		if nativeACL(t, path) != "absent" {
			t.Fatal("fresh native directory has an actual ACL; fixture cannot claim safe absence")
		}
	}
	return s
}
func (s nativeSession) args() []string {
	return []string{"--root", s.root, "--home", s.home, "--session", s.key}
}
func (s nativeSession) stage() string {
	return filepath.Join(filepath.Dir(s.record), ".obk-record-edit-"+s.key)
}

func snapshot(t *testing.T, base string) map[string]nativeObject {
	t.Helper()
	objects := map[string]nativeObject{}
	err := filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		objects[relative] = nativeObjectAt(t, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return objects
}
func unchanged(t *testing.T, s nativeSession, before map[string]nativeObject) {
	t.Helper()
	if !reflect.DeepEqual(snapshot(t, s.base), before) {
		t.Fatal("native bytes/identity/security/xattrs or namespace changed")
	}
}

type observedReader struct {
	text                     string
	position, hookAt, closed int
	hook                     func()
}

func (r *observedReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.hook != nil && r.position == r.hookAt {
		action := r.hook
		r.hook = nil
		action()
	}
	if r.position == len(r.text) {
		return 0, io.EOF
	}
	end := len(r.text)
	if r.hook != nil && r.position < r.hookAt && end > r.hookAt {
		end = r.hookAt
	}
	n := copy(p, r.text[r.position:end])
	r.position += n
	return n, nil
}
func (r *observedReader) Close() error { r.closed++; return nil }

type runResult struct {
	result   contract.Result
	out, err string
	consumed int
}

func run(t *testing.T, s nativeSession, clock *clockFact, input string) runResult {
	t.Helper()
	reader := &observedReader{text: input}
	var out, err bytes.Buffer
	result := newDriver(clock).Run(s.args(), contract.Streams{Input: reader, Output: &out, Error: &err})
	if reader.closed != 0 {
		t.Fatal("borrowed input closed")
	}
	return runResult{result, out.String(), err.String(), reader.position}
}
func report(t *testing.T, got runResult, code string, exit int, saved bool) {
	t.Helper()
	var decoded struct {
		Status  string `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
		Saved   bool   `json:"saved"`
	}
	if err := json.Unmarshal([]byte(got.err), &decoded); err != nil {
		t.Fatalf("nonempty canonical report required: %v; output=%q", err, got.err)
	}
	canonical, _ := json.Marshal(decoded)
	want := map[string][2]string{
		"edit.cancelled": {"cancelled", "Session change cancelled; no record was written."},
		"edit.unchanged": {"unchanged", "The record is unchanged."},
		"edit.saved":     {"saved", "The selected record change was saved."},
		"edit.usage":     {"refused", "Use an explicit root, state home and registered session key."},
		"edit.context":   {"refused", "Select a safe unchanged registered session context."},
		"edit.pending":   {"refused", "Finish or repair the pending import first."},
		"edit.review":    {"refused", "Resolve the displayed session problem before saving."},
		"edit.input":     {"refused", "Use complete valid displayed actions within the input limits."},
		"edit.output":    {"refused", "Required output could not be completed; no record was written."},
		"edit.repair":    {"saved", "The record was saved, but cleanup or reporting needs repair."},
		"edit.commit":    {"refused", "The record was not replaced; private staging may need repair."},
		"edit.source":    {"refused", "Read the bounded existing session files before editing."},
		"edit.limit":     {"refused", "The selected session or display exceeds the supported limit."},
	}[code]
	if got.err != string(append(canonical, '\n')) || got.result.Exit != exit || got.result.Saved != saved || decoded.Code != code || decoded.Saved != saved || decoded.Message != want[1] || decoded.Status != want[0] {
		t.Fatalf("report/result %#v %s", got.result, got.err)
	}
}

func companions(t *testing.T, s nativeSession, before map[string]nativeObject) {
	t.Helper()
	for _, path := range []string{s.plan, s.preparation, s.metadata} {
		rel, _ := filepath.Rel(s.base, path)
		if !reflect.DeepEqual(nativeObjectAt(t, path), before[rel]) {
			t.Fatalf("companion changed: %s", rel)
		}
	}
}
func savedRecord(t *testing.T, s nativeSession, before nativeObject, want string) {
	t.Helper()
	after := nativeObjectAt(t, s.record)
	if after.Body != want || after.Device != before.Device || after.Inode == before.Inode {
		t.Fatal("complete literal record was not atomically replaced")
	}
	after.Body, after.Inode = before.Body, before.Inode
	if !reflect.DeepEqual(after, before) {
		t.Fatal("record persistent metadata/xattrs/ACL/UID/GID/mode/flags were not preserved")
	}
	entries, err := os.ReadDir(s.stage())
	if err != nil || len(entries) != 0 {
		t.Fatal("successful stage is not proven empty")
	}
	if info, err := os.Stat(s.stage()); err != nil || info.Mode().Perm() != 0700 || nativeACL(t, s.stage()) != "absent" {
		t.Fatal("idle stage is not native private")
	}
}
