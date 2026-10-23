package invite_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func aclCommand(path string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "/bin/chmod", append(args, path)...).Run()
}

func removeKnownACLFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
		return
	}
	if err := aclCommand(path, "-N"); err != nil {
		t.Errorf("owned child ACL cleanup failed: %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Errorf("owned child removal failed: %v", err)
	}
}

func inheritedDeleteDeniedParent(t *testing.T, root string) string {
	t.Helper()
	if runtime.GOOS != "darwin" || os.Geteuid() == 0 {
		t.Skip("unsupported capability: requires unprivileged Darwin; no cleanup fault verified")
	}
	identity, err := user.LookupId(strconv.Itoa(os.Geteuid()))
	if err != nil || identity.Username == "" {
		t.Skip("unsupported capability: current ACL identity lookup failed; no cleanup fault verified")
	}
	parent := filepath.Join(root, "owned-output")
	if err := os.Mkdir(parent, 0o700); err != nil {
		t.Fatal(err)
	}
	aclInstalled := false
	t.Cleanup(func() {
		entries, err := os.ReadDir(parent)
		if err != nil {
			t.Errorf("owned parent cleanup enumeration failed: %v", err)
			return
		}
		for _, entry := range entries {
			path := filepath.Join(parent, entry.Name())
			info, err := os.Lstat(path)
			if err != nil || !info.Mode().IsRegular() {
				t.Error("unexpected nonregular entry in exclusive owned output sandbox; not removing it")
				continue
			}
			removeKnownACLFile(t, path)
		}
		if aclInstalled {
			if err := aclCommand(parent, "-N"); err != nil {
				t.Errorf("owned parent ACL cleanup failed: %v", err)
			}
		}
		if err := os.Remove(parent); err != nil {
			t.Errorf("owned empty parent removal failed: %v", err)
		}
	})
	if err := aclCommand(parent, "+a", "user:"+identity.Username+" deny delete,file_inherit"); err != nil {
		t.Skip("unsupported capability: inheritable ACL unavailable; no cleanup fault verified")
	}
	aclInstalled = true
	temporary := filepath.Join(parent, "capability-temporary")
	published := filepath.Join(parent, "capability-published")
	f, err := os.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		t.Skip("unsupported capability: ACL control creation failed; no cleanup fault verified")
	}
	modeErr := f.Chmod(0o600)
	n, writeErr := f.Write([]byte("independent owned cleanup capability\n"))
	closeErr := f.Close()
	if modeErr != nil || writeErr != nil || closeErr != nil || n != len("independent owned cleanup capability\n") {
		t.Skip("unsupported capability: ACL control write/mode/close failed; no cleanup fault verified")
	}
	if err := os.Link(temporary, published); err != nil {
		t.Skip("unsupported capability: ACL control publication denied; no cleanup fault verified")
	}
	a, aErr := os.Stat(temporary)
	b, bErr := os.Stat(published)
	payload, readErr := os.ReadFile(published)
	if aErr != nil || bErr != nil || !os.SameFile(a, b) || readErr != nil || string(payload) != "independent owned cleanup capability\n" {
		t.Fatal("independent ACL control could not establish complete same-inode publication")
	}
	err = os.Remove(temporary)
	if !errors.Is(err, syscall.EACCES) && !errors.Is(err, syscall.EPERM) {
		t.Skip("unsupported capability: control unlink not denied; no cleanup fault verified")
	}
	if _, err := os.Stat(temporary); err != nil {
		t.Fatal("ACL control did not retain denied temporary entry")
	}
	removeKnownACLFile(t, temporary)
	removeKnownACLFile(t, published)
	if t.Failed() {
		t.Fatal("independent ACL control cleanup failed; no product invocation")
	}
	return parent
}

func cleanupReport(t *testing.T, got result) {
	t.Helper()
	readReport(t, got, "invite.cleanup-written", 1, nil, nil)
	var actual report
	var fields map[string]json.RawMessage
	if got.exit != 1 || json.Unmarshal([]byte(got.stderr), &actual) != nil || json.Unmarshal([]byte(got.stderr), &fields) != nil || len(fields) != 5 {
		t.Fatalf("cleanup failure needs exit1 and one typed five-key report: %#v", got)
	}
	for _, key := range []string{"status", "code", "message", "problems", "blockers"} {
		if _, exists := fields[key]; !exists {
			t.Fatalf("missing cleanup report key %s", key)
		}
	}
	if actual.Status != "written" || actual.Code != "invite.cleanup-written" || actual.Message != "The archive was written, but temporary invitation data may remain. Keep the archive; restore access to the selected output directory and check exporter-owned temporary data." || actual.Problems == nil || actual.Blockers == nil || len(actual.Problems) != 0 || len(actual.Blockers) != 0 {
		t.Fatalf("cleanup repair need or committed-state truth lost: %#v", actual)
	}
	for _, forbidden := range []string{"PRIVATE", "owned-output", "capability-", "newcomer", "permission denied", "operation not permitted"} {
		if strings.Contains(got.stderr, forbidden) {
			t.Fatalf("cleanup report discloses private/OS detail %q", forbidden)
		}
	}
}

func TestACLPostcommitCleanupFailureIsWrittenError(t *testing.T) {
	for _, failStdout := range []bool{false, true} {
		name := "ordinary-stdout"
		if failStdout {
			name = "stdout-failure-does-not-erase-cleanup"
		}
		t.Run(name, func(t *testing.T) {
			s := newSandbox(t)
			parent := inheritedDeleteDeniedParent(t, s.root)
			s.out = filepath.Join(parent, "out.zip")
			before := tree(t, s.root)
			var stdout, stderr bytes.Buffer
			var writer io.Writer = &stdout
			if failStdout {
				writer = failingWriter{}
			}
			exit := newCommand().Run(s.args(), writer, &stderr)
			payloads := exportedPayloads(t, s.out)
			for entry, fixture := range map[string]string{"plan.obk.yaml": "plan.yaml", "records/newcomer/preparation.obk.yaml": "preparation.yaml", "records/newcomer/record.obk.yaml": "record.yaml", "first-steps.md": "first-steps.md"} {
				if !bytes.Equal(payloads[entry], literal(t, fixture)) {
					t.Fatalf("cleanup failure altered literal payload %s", entry)
				}
			}
			finalInfo, err := os.Stat(s.out)
			if err != nil {
				t.Fatal(err)
			}
			entries, err := os.ReadDir(parent)
			if err != nil || len(entries) != 2 {
				t.Fatal("controlled denied cleanup must retain exactly final plus owned temporary")
			}
			after := tree(t, s.root)
			for _, entry := range entries {
				path := filepath.Join(parent, entry.Name())
				info, err := os.Lstat(path)
				if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 || !os.SameFile(finalInfo, info) {
					t.Fatal("retained artifact is not a private complete publication hard link")
				}
				relative, err := filepath.Rel(s.root, path)
				if err != nil {
					t.Fatal(err)
				}
				delete(after, relative)
			}
			if !reflect.DeepEqual(before, after) {
				t.Fatal("cleanup failure changed inputs, unrelated state or output parent")
			}
			cleanupReport(t, result{exit: exit, stdout: stdout.String(), stderr: stderr.String()})
			if (!failStdout && stdout.String() != success) || (failStdout && stdout.Len() != 0) {
				t.Fatal("postcommit stdout did not truthfully follow writer capability")
			}
		})
	}
}
