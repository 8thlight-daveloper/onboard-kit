package starter_test

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"syscall"
	"testing"

	contract "onboard-kit/contracts/invitation/starter"
	"onboard-kit/internal/model"
)

const openedAt model.Timestamp = "2026-10-23T12:34:56.123Z"
const rawPlanHash = "00d6473bd698e40caca271f61af7b87288f0c864924008bd24375d6b19b3c6a4"

type sandbox struct{ base, root, home, invitation string }
type report struct {
	Status      string `json:"status"`
	Code        string `json:"code"`
	Message     string `json:"message"`
	Transaction string `json:"transaction"`
}
type result struct {
	exit     int
	out, err string
	report   report
}
type object struct {
	Mode          os.FileMode
	Device, Inode uint64
	Bytes         string
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func put(t *testing.T, path string, body []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
}

func zipBytes(t *testing.T, changes map[string][]byte) []byte {
	t.Helper()
	var out bytes.Buffer
	w := zip.NewWriter(&out)
	for _, role := range []struct{ path, file string }{
		{"plan.obk.yaml", "plan.yaml"}, {"records/jules/preparation.obk.yaml", "preparation.yaml"},
		{"records/jules/record.obk.yaml", "seed.yaml"}, {"first-steps.md", "first-steps.md"},
	} {
		body, found := changes[role.path]
		if !found {
			body = fixture(t, role.file)
		}
		entry, err := w.CreateHeader(&zip.FileHeader{Name: role.path, Method: zip.Store})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = entry.Write(body); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func setup(t *testing.T) sandbox {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chown(base, -1, os.Getegid()); err != nil {
		t.Fatal("cannot establish effective group on the owned fresh sandbox")
	}
	s := sandbox{base: base, root: filepath.Join(base, "root"), home: filepath.Join(base, "home"), invitation: filepath.Join(base, "invitation.zip")}
	if err := os.Mkdir(s.root, 0700); err != nil {
		t.Fatal(err)
	}
	put(t, s.invitation, zipBytes(t, nil))
	return s
}

func (s sandbox) args() []string {
	return []string{"--invitation", s.invitation, "--root", s.root, "--home", s.home}
}
func (s sandbox) record() string  { return filepath.Join(s.root, "records/jules/record.obk.yaml") }
func (s sandbox) prep() string    { return filepath.Join(s.root, "records/jules/preparation.obk.yaml") }
func (s sandbox) pending() string { return filepath.Join(s.home, "import-pending") }

func key(root string) string {
	return sessionKey("plan-a", "jules", root)
}

func sessionKey(planID, account, root string) string {
	h := sha256.New()
	_, _ = h.Write([]byte("onboard-kit/session-key/v1\x00"))
	for _, text := range []string{planID, account, root} {
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(len(text)))
		_, _ = h.Write(size[:])
		_, _ = h.Write([]byte(text))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (s sandbox) registration() string { return filepath.Join(s.home, "sessions", key(s.root)) }

func runtimeFacts(t *testing.T, clock *int) contract.Runtime {
	t.Helper()
	bad := func() (string, error) {
		t.Error("unused runtime directory fact was read")
		return "", fmt.Errorf("unexpected fact")
	}
	return contract.Runtime{
		Clock:             func() model.Timestamp { *clock++; return openedAt },
		LookupEnvironment: func(string) string { t.Error("explicit home must not read environment"); return "" },
		WorkingDirectory:  bad, UserHomeDirectory: bad, UserConfigDirectory: bad,
	}
}

func invoke(t *testing.T, command contract.Command, args []string, input string) result {
	t.Helper()
	var stdout, stderr bytes.Buffer
	r := result{exit: command.Run(args, strings.NewReader(input), &stdout, &stderr), out: stdout.String(), err: stderr.String()}
	if err := json.Unmarshal(stderr.Bytes(), &r.report); err != nil {
		t.Fatalf("one structured report required: %v", err)
	}
	canonical, err := json.Marshal(r.report)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(append(canonical, '\n'), stderr.Bytes()) {
		t.Error("report framing/order or extra data differs")
	}
	return r
}

func wantReport(t *testing.T, got result, exit int, status, code, message string) {
	t.Helper()
	if got.exit != exit || got.report.Status != status || got.report.Code != code || got.report.Message != message {
		t.Errorf("report differs: exit=%d status=%s code=%s", got.exit, got.report.Status, got.report.Code)
	}
	if got.report.Transaction != "" && !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(got.report.Transaction) {
		t.Error("invalid transaction token")
	}
}

func snapshot(t *testing.T, root string) map[string]object {
	t.Helper()
	out := map[string]object{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, failure error) error {
		if failure != nil {
			return failure
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("native stat unavailable")
		}
		item := object{Mode: info.Mode(), Device: uint64(stat.Dev), Inode: uint64(stat.Ino)}
		if info.Mode().IsRegular() {
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			item.Bytes = string(body)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			text, err := os.Readlink(path)
			if err != nil {
				return err
			}
			item.Bytes = text
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out[rel] = item
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func unchanged(t *testing.T, s sandbox, before map[string]object) {
	t.Helper()
	if !reflect.DeepEqual(snapshot(t, s.base), before) {
		t.Error("source, inode, controls or namespace changed without publication consent")
	}
}

func assertBytes(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("literal bytes differ for %s", filepath.Base(path))
	}
}

func assertIdle(t *testing.T, s sandbox) {
	t.Helper()
	info, err := os.Lstat(s.pending())
	if err != nil {
		t.Fatal("fixed empty idle container missing")
	}
	if !info.IsDir() || info.Mode().Perm() != 0700 {
		t.Error("idle container not private directory")
	}
	g := objectGuard(t, s.pending())
	if g.Security.UID != fmt.Sprint(os.Geteuid()) || !permittedGroup(t, g.Security.GID) || g.Security.Mode != "0700" || g.Security.AccessPresent || g.Security.DefaultPresent {
		t.Error("idle container native owner or ACL privacy differs")
	}
	items, err := os.ReadDir(s.pending())
	if err != nil || len(items) != 0 {
		t.Error("personal recovery files or unknown remnants remain")
	}
	err = filepath.WalkDir(s.base, func(_ string, entry os.DirEntry, failure error) error {
		if failure != nil {
			return failure
		}
		if strings.HasPrefix(entry.Name(), ".obk-start-") {
			t.Error("role staging remains after successful cleanup")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func permittedGroup(t *testing.T, gid string) bool {
	t.Helper()
	if gid == fmt.Sprint(os.Getegid()) {
		return true
	}
	groups, err := os.Getgroups()
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range groups {
		if gid == fmt.Sprint(group) {
			return true
		}
	}
	return false
}

func assertNewPrivate(t *testing.T, path, mode string) {
	t.Helper()
	g := objectGuard(t, path)
	if g.Security.UID != fmt.Sprint(os.Geteuid()) || g.Security.GID != fmt.Sprint(os.Getegid()) || g.Security.Mode != mode || g.Security.AccessPresent || g.Security.DefaultPresent {
		t.Error("new private object does not have exact effective UID/GID and private controls")
	}
}

func TestHelpAndUsageReadNoFacts(t *testing.T) {
	for _, args := range [][]string{{"--help", "--bad"}, {"-h"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			clock := 0
			command := newStarter(runtimeFacts(t, &clock))
			var out, err bytes.Buffer
			if command.Run(args, neverReader{t}, &out, &err) != 0 || !bytes.Equal(out.Bytes(), fixture(t, "help.txt")) || err.Len() != 0 || clock != 0 {
				t.Error("help must be exact and side-effect free")
			}
		})
	}
	for _, args := range [][]string{nil, {"--invitation", "x"}, {"--root", "x"}, {"--invitation", "a", "--invitation=b", "--root=r"}, {"--resume-import", "ABC", "--root=r"}, {"--invitation=a", "--resume-import=" + strings.Repeat("a", 64), "--root=r"}, {"--invitation=", "--root=r"}, {"--invitation", "--literal", "--root=r"}, {"--invitation=a", "--root=r", "--help"}, {"--invitation=a", "--root=r", "extra"}} {
		clock := 0
		r := invoke(t, newStarter(runtimeFacts(t, &clock)), args, "")
		wantReport(t, r, 2, "refused", "start.usage", "Use one invitation or resume token, an explicit root, and an optional home.")
		if r.out != "" || clock != 0 {
			t.Error("usage must precede effects")
		}
	}
}

type neverReader struct{ t *testing.T }

func (r neverReader) Read([]byte) (int, error) { r.t.Error("unexpected input read"); return 0, io.EOF }

func TestIdentityAndReviewAreNotBegin(t *testing.T) {
	for _, input := range []string{"", "cancel\n", "confirm jules", "confirm jules\ncancel\n", "confirm jules\n", "confirm jules\nbegin"} {
		t.Run(fmt.Sprintf("%q", input), func(t *testing.T) {
			s := setup(t)
			before := snapshot(t, s.base)
			clock := 0
			r := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), input)
			wantReport(t, r, 0, "cancelled", "start.cancelled", "Opening cancelled before begin; no files were changed.")
			if clock != 0 {
				t.Error("clock sampled before begin")
			}
			unchanged(t, s, before)
			if strings.Contains(input, "confirm jules\n") && !strings.Contains(r.out, "Review complete. Choose begin or cancel:\n") {
				t.Error("full review missing before decision")
			}
		})
	}
}

func TestFreshBeginPublishesLiteralArtifactsAndActualSession(t *testing.T) {
	s := setup(t)
	clock := 0
	r := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\nexit\n")
	wantReport(t, r, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	if clock != 1 || r.report.Transaction == "" {
		t.Fatal("new event or token missing")
	}
	assertBytes(t, s.record(), fixture(t, "begun.yaml"))
	assertBytes(t, s.prep(), fixture(t, "preparation.yaml"))
	assertBytes(t, filepath.Join(s.registration(), "plan.obk.yaml"), fixture(t, "plan.yaml"))
	ending := "First opened at: \"" + string(openedAt) + "\"\n" + string(fixture(t, "session.txt")) + "Session: read only. Choose exit:\n"
	if !strings.HasSuffix(r.out, ending) {
		t.Error("literal complete record-backed session differs")
	}
	for _, literal := range []string{"Role: \"local-plan\" Action: \"install\"", "Role: \"preparation\" Action: \"install\"", "Role: \"session-metadata\" Action: \"install\"", "Role: \"record\" Action: \"create\"", "Private recovery copies may temporarily exist in the state home and selected artifact parents."} {
		if !strings.Contains(r.out, literal) {
			t.Error("review omits required action/privacy fact")
		}
	}
	info, err := os.Stat(s.root)
	if err != nil {
		t.Fatal(err)
	}
	stat := info.Sys().(*syscall.Stat_t)
	want := fmt.Sprintf("{\"version\":1,\"sessionKey\":%q,\"planIdBase64\":%q,\"account\":\"jules\",\"root\":%q,\"rootDevice\":%q,\"rootInode\":%q,\"record\":\"records/jules/record.obk.yaml\",\"preparation\":\"records/jules/preparation.obk.yaml\",\"planSHA256\":%q,\"transaction\":%q}\n", key(s.root), base64.StdEncoding.EncodeToString([]byte("plan-a")), s.root, fmt.Sprint(stat.Dev), fmt.Sprint(stat.Ino), rawPlanHash, r.report.Transaction)
	assertBytes(t, filepath.Join(s.registration(), "session.json"), []byte(want))
	assertIdle(t, s)
	for _, path := range []string{s.home, s.pending(), filepath.Join(s.home, "sessions"), s.registration(), filepath.Join(s.root, "records"), filepath.Dir(s.record())} {
		assertNewPrivate(t, path, "0700")
	}
	for _, path := range []string{s.prep(), s.record(), filepath.Join(s.registration(), "plan.obk.yaml"), filepath.Join(s.registration(), "session.json")} {
		assertNewPrivate(t, path, "0600")
	}
}

func TestExistingUnknownEventAndFormattingRemainOwned(t *testing.T) {
	s := setup(t)
	put(t, s.record(), fixture(t, "existing.yaml"))
	before := snapshot(t, s.root)
	clock := 0
	r := invoke(t, newStarter(runtimeFacts(t, &clock)), s.args(), "confirm jules\nbegin\nexit\n")
	wantReport(t, r, 0, "opened", "start.opened", "Opening completed and the read-only session was displayed.")
	if clock != 0 {
		t.Error("existing event must not sample clock")
	}
	assertBytes(t, s.record(), fixture(t, "existing.yaml"))
	if snapshot(t, s.root)["records/jules/record.obk.yaml"] != before["records/jules/record.obk.yaml"] {
		t.Error("no-op record inode or controls changed")
	}
	if !strings.Contains(r.out, "First opened at: \"\"\n") || !strings.Contains(r.out, "Role: \"record\" Action: \"retain\"") {
		t.Error("unknown event/retain truth lost")
	}
	assertIdle(t, s)
}
