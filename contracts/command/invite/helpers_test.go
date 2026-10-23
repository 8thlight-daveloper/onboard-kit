package invite_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	contract "onboard-kit/contracts/command/invite"
)

const usage = "Usage: onboard-kit invite export --plan FILE --preparation FILE --account ID --out FILE --preaccess-help-contact ID [--help-contact ID | --help-role ID ...]\n"
const success = "Invitation archive written.\n"
const baselineHash = "e943f371cfbd4173c2530caf131c1b9cf5a760ae2c8556510fc1c2463c169363"

type projectedProblem struct {
	Code     string
	Severity string
}

type report struct {
	Status   string
	Code     string
	Message  string
	Problems []projectedProblem
	Blockers []string
}

type result struct {
	exit   int
	stdout string
	stderr string
}

func run(command contract.Command, args []string) result {
	var stdout, stderr bytes.Buffer
	exit := command.Run(args, &stdout, &stderr)
	return result{exit: exit, stdout: stdout.String(), stderr: stderr.String()}
}

func literal(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

type sandbox struct {
	root        string
	plan        string
	preparation string
	out         string
}

func newSandbox(t *testing.T) sandbox {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := sandbox{root: root, plan: filepath.Join(root, "plan.yaml"), preparation: filepath.Join(root, "preparation.yaml"), out: filepath.Join(root, "out.zip")}
	writeLiteral(t, s.plan, literal(t, "plan.yaml"))
	writeLiteral(t, s.preparation, literal(t, "preparation.yaml"))
	writeLiteral(t, filepath.Join(root, "unrelated-record.yaml"), []byte("PRIVATE_UNRELATED_RECORD\n"))
	return s
}

func writeLiteral(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func (s sandbox) args() []string {
	return []string{"export", "--plan", s.plan, "--preparation", s.preparation, "--account", "newcomer", "--out", s.out, "--preaccess-help-contact", "owner"}
}

func readReport(t *testing.T, got result, code string, exit int, problems []projectedProblem, blockers []string) report {
	t.Helper()
	if got.exit != exit {
		t.Fatalf("exit got %d want %d; stdout=%q stderr=%q", got.exit, exit, got.stdout, got.stderr)
	}
	var actual report
	decoder := json.NewDecoder(strings.NewReader(got.stderr))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&actual); err != nil {
		t.Fatalf("typed report missing: %v; %q", err, got.stderr)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(got.stderr), &raw); err != nil || len(raw) != 5 {
		t.Fatal("report must be exactly one five-key JSON object")
	}
	for _, key := range []string{"status", "code", "message", "problems", "blockers"} {
		if _, exists := raw[key]; !exists {
			t.Fatalf("missing report key %s", key)
		}
	}
	if actual.Code != code || strings.TrimSpace(actual.Message) == "" || actual.Problems == nil || actual.Blockers == nil {
		t.Fatalf("incomplete report: %#v", actual)
	}
	var messages map[string]string
	if err := json.Unmarshal(literal(t, "messages.json"), &messages); err != nil || messages[code] == "" || actual.Message != messages[code] {
		t.Fatalf("static privacy-safe repair message differs: %#v", actual)
	}
	if len(actual.Problems) != len(problems) || len(actual.Blockers) != len(blockers) {
		t.Fatalf("diagnostic counts differ: %#v", actual)
	}
	for i := range problems {
		if actual.Problems[i] != problems[i] {
			t.Fatalf("dependency diagnostic %d differs: %#v", i, actual)
		}
	}
	for i := range blockers {
		if actual.Blockers[i] != blockers[i] {
			t.Fatalf("actionable blocker IDs differ: %#v", actual)
		}
	}
	if exit == 0 {
		if actual.Status != "written" || got.stdout != success || actual.Message != "Pre-access help availability has not been tested; deliver through an approved channel." {
			t.Fatalf("success reporting differs: %#v", got)
		}
	} else if code != "invite.result-output" && code != "invite.cleanup-written" && (actual.Status != "refused" || got.stdout != "") {
		t.Fatalf("failure claimed usable output: %#v", got)
	}
	if (code == "invite.result-output" || code == "invite.cleanup-written") && actual.Status != "written" {
		t.Fatal("postcommit output failure must acknowledge the published artifact")
	}
	for _, private := range []string{"PRIVATE", "private@example.invalid"} {
		if strings.Contains(got.stdout+got.stderr, private) {
			t.Fatalf("diagnostics leaked supplied private values: %q", private)
		}
	}
	return actual
}

func tree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil || relative == "." {
			return err
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		out[relative] = info.Mode().String()
		if info.Mode().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			out[relative] += string(data)
		} else if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			out[relative] += target
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func preserved(t *testing.T, s sandbox, before map[string]string, published bool) {
	t.Helper()
	after := tree(t, s.root)
	if published {
		delete(after, "out.zip")
	}
	if !reflect.DeepEqual(before, after) {
		keys := make([]string, 0, len(after))
		for key := range after {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		t.Fatalf("source/unrelated artifacts or temporary-file cleanup changed tree; entries %v", keys)
	}
}
