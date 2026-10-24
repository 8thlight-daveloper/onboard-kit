package creator_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

type creatorGapOutput struct {
	bytes.Buffer
	limit int
}

func (writer *creatorGapOutput) Write(body []byte) (int, error) {
	if len(body) > writer.limit-writer.Len() {
		return 0, errors.New("owned fixture output limit")
	}
	return writer.Buffer.Write(body)
}

func creatorGapInheritedParent(t *testing.T, root string) {
	t.Helper()
	initial := creatorGapObject(t, root)
	if initial.ACLPresent {
		t.Skip("creator privacy gap: sandbox already has an ambient ACL; fixture will not strip it")
	}
	if initial.UID != uint32(os.Geteuid()) {
		t.Fatal("chmod target is not the owned sandbox directory")
	}
	if _, err := os.Stat("/bin/chmod"); os.IsNotExist(err) {
		t.Skip("creator privacy gap: Darwin chmod ACL fixture tool unavailable")
	} else if err != nil {
		t.Fatal(err)
	}
	deadline, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(deadline, "/bin/chmod", "+a", "everyone allow read,file_inherit,directory_inherit", root)
	output := &creatorGapOutput{limit: 4096}
	command.Stdout, command.Stderr = output, output
	if err := command.Run(); err != nil {
		if strings.Contains(strings.ToLower(output.String()), "not supported") {
			t.Skip("creator privacy gap: owned filesystem explicitly reports inheritable ACLs unsupported")
		}
		t.Fatalf("owned native ACL setup failed before product invocation: %v, output bytes=%d", err, output.Len())
	}
	parent := creatorGapObject(t, root)
	if !parent.ACLPresent || !creatorGapReadGrant(t, parent.ACL, false) {
		t.Fatal("native inheritable parent read grant was not established")
	}
	control := filepath.Join(root, "owned-acl-capability-control")
	file, err := os.OpenFile(control, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = file.Write([]byte("disposable native capability fact\n")); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	child := creatorGapObject(t, control)
	if child.Mode != 0600 || child.UID != uint32(os.Geteuid()) || child.GID != uint32(os.Getegid()) || !child.ACLPresent || !creatorGapReadGrant(t, child.ACL, true) {
		t.Fatal("native disposable child did not prove inherited read despite private mode0600")
	}
	if err = os.Remove(control); err != nil {
		t.Fatal(err)
	}
	if creatorGapObject(t, root) != parent {
		t.Fatal("disposable ACL control changed parent native controls")
	}
}

func creatorGapHeldArtifacts(t *testing.T, root string) map[string]creatorGapGuard {
	t.Helper()
	guards := make(map[string]creatorGapGuard)
	for _, name := range []string{"preparation.obk.yaml", "record.obk.yaml", "organization.obk.yaml"} {
		guards[name] = creatorGapObject(t, filepath.Join(root, name))
	}
	return guards
}

func creatorGapPreserved(t *testing.T, root string, parent creatorGapGuard, held map[string]creatorGapGuard, targetPresent bool) {
	t.Helper()
	if creatorGapObject(t, root) != parent {
		t.Error("creator changed output-parent identity, ownership, mode or ACL")
	}
	for name, guard := range held {
		if creatorGapObject(t, filepath.Join(root, name)) != guard {
			t.Errorf("creator changed unrelated owned controls: %s", name)
		}
	}
	preserved(t, root)
	directory, err := os.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := directory.ReadDir(16)
	closeError := directory.Close()
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if closeError != nil {
		t.Fatal(closeError)
	}
	names := make(map[string]bool)
	for name := range held {
		names[name] = true
	}
	if targetPresent {
		names["created.obk.yaml"] = true
	}
	if len(entries) != len(names) {
		t.Errorf("creator left an unexpected artifact or temporary: entries=%d expected=%d", len(entries), len(names))
	}
	for _, entry := range entries {
		if !names[entry.Name()] {
			t.Errorf("unexpected owned sandbox entry: %s", entry.Name())
		}
	}
}

func creatorGapReport(t *testing.T, body []byte, status, code, message string) {
	t.Helper()
	var got map[string]string
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("creator report is not one JSON object: %q", body)
	}
	want := map[string]string{"status": status, "code": code, "message": message}
	if !reflect.DeepEqual(got, want) || !bytes.HasSuffix(body, []byte("\n")) {
		t.Errorf("creator report differs: %q", body)
	}
}

func creatorGapPlanBytes(t *testing.T, path string) []byte {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		file.Close()
		t.Fatal("created plan must be a regular file")
	}
	body, err := io.ReadAll(io.LimitReader(file, 1048577))
	closeError := file.Close()
	if err != nil || closeError != nil || len(body) > 1048576 {
		t.Fatal("created plan read failed or exceeded its independently bounded size")
	}
	return body
}

func TestDarwinCreatorInheritedACLRequiresConsentThenRefusesUnsafePrivateCreation(t *testing.T) {
	for _, test := range []struct {
		name, input, consumed string
		save                  bool
	}{
		{"draft-cancel", "cancel\nUNREAD", "cancel\n", false},
		{"review-cancel", "review\ncancel\nUNREAD", "review\ncancel\n", false},
		{"review-back-cancel", "review\nback\ncancel\nUNREAD", "review\nback\ncancel\n", false},
		{"partial-save", "review\nsave", "review\nsave", false},
		{"consented-save", "review\nsave\nUNREAD", "review\nsave\n", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, target := sandbox(t)
			creatorGapInheritedParent(t, root)
			watch := creatorGapNamespaceWatch(t, root)
			parent, held := creatorGapObject(t, root), creatorGapHeldArtifacts(t, root)
			input := &countedInput{source: []byte(test.input)}
			stdout, stderr := &creatorGapOutput{limit: 65536}, &creatorGapOutput{limit: 4096}
			exit := newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, stdout, stderr)
			if test.save {
				if exit != 1 {
					t.Errorf("unsafe inherited-ACL save exit=%d, expected1", exit)
				}
				creatorGapReport(t, stderr.Bytes(), "refused", "create.commit", "The plan was not published; owned temporary data may need cleanup.")
			} else {
				if exit != 0 {
					t.Errorf("no-save action exit=%d, expected0", exit)
				}
				creatorGapReport(t, stderr.Bytes(), "cancelled", "create.cancelled", "Cancelled. No files changed.")
				if watch.changed(t) {
					t.Error("creator changed the parent namespace before a complete save, including a transient create/remove")
				}
			}
			if strings.HasPrefix(test.input, "review\n") && !strings.Contains(stdout.String(), "Review complete. Choose save, back, or cancel:\n") {
				t.Error("inherited parent ACL improperly prevented the read-only complete review")
			}
			if input.position != len(test.consumed) {
				t.Errorf("consumed input=%d expected=%d", input.position, len(test.consumed))
			}
			if _, err := os.Lstat(target); !os.IsNotExist(err) {
				t.Error("unsafe or unconsented creation published a target")
			}
			creatorGapPreserved(t, root, parent, held, false)
		})
	}
}

func TestDarwinCreatorActualNoACLControlPublishesExactPrivatePlan(t *testing.T) {
	root, target := sandbox(t)
	parent := creatorGapObject(t, root)
	if parent.ACLPresent {
		t.Skip("creator privacy gap: no-ACL positive control unavailable without stripping ambient ACL")
	}
	watch := creatorGapNamespaceWatch(t, root)
	held := creatorGapHeldArtifacts(t, root)
	input := &countedInput{source: []byte("review\nsave\nUNREAD")}
	stdout, stderr := &creatorGapOutput{limit: 65536}, &creatorGapOutput{limit: 4096}
	if exit := newCommand().Run([]string{"--id=plan-one", "--out=" + target}, input, stdout, stderr); exit != 0 {
		t.Errorf("actual no-ACL save exit=%d expected0", exit)
	}
	creatorGapReport(t, stderr.Bytes(), "written", "create.written", "The reviewed organization plan was created.")
	file := creatorGapObject(t, target)
	body := creatorGapPlanBytes(t, target)
	if !bytes.Equal(body, fixture(t, "empty-plan.yaml")) {
		t.Error("actual no-ACL control did not publish the independently specified reviewed bytes")
	}
	if file.Mode != 0600 || file.UID != uint32(os.Geteuid()) || file.GID != uint32(os.Getegid()) || file.ACLPresent {
		t.Error("published private plan has wrong effective ownership, mode or native ACL")
	}
	if input.position != len("review\nsave\n") {
		t.Error("successful private publication consumed trailing input")
	}
	if !watch.changed(t) {
		t.Error("native watcher did not observe the actual successful publication")
	}
	creatorGapPreserved(t, root, parent, held, true)
}
