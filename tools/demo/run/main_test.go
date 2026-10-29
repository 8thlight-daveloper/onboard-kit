package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

var manualTestStages = []string{"plan", "preparation", "invitation", "onboardee", "results"}

func TestStatusAndUnknownStageDoNotCreateWorkspace(t *testing.T) {
	for _, arguments := range [][]string{{"status"}, {"reset", "unknown-stage"}} {
		t.Run(strings.Join(arguments, "-"), func(t *testing.T) {
			root := t.TempDir()
			before := demoTree(t, root)
			var stdout, stderr bytes.Buffer
			err := execute(root, arguments, strings.NewReader(""), &stdout, &stderr)
			if arguments[0] == "status" && err != nil {
				t.Fatalf("status failed: %v; stderr=%q", err, stderr.String())
			}
			if arguments[0] == "reset" && err == nil {
				t.Fatal("unknown reset stage was accepted")
			}
			if after := demoTree(t, root); !reflect.DeepEqual(before, after) {
				t.Fatalf("read-only or invalid command changed the checkout: before=%v after=%v", before, after)
			}
		})
	}
}

func TestUnknownResetStagePreservesExistingWorkspace(t *testing.T) {
	root, workspace := demoOwnedWorkspace(t)
	demoWrite(t, filepath.Join(workspace, "plan", "plan.obk.yaml"), "organization plan\n", 0600)
	before := demoTree(t, root)
	demoExecuteError(t, root, "reset", "invitaton")
	if after := demoTree(t, root); !reflect.DeepEqual(before, after) {
		t.Fatalf("unknown stage changed existing files: before=%v after=%v", before, after)
	}
}

func TestStatusShowsExactImportedFilesWithoutChangingWorkspace(t *testing.T) {
	root, workspace := demoOwnedWorkspace(t)
	physicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	root = physicalRoot
	workspace = filepath.Join(root, "output", "manual-test")
	key := strings.Repeat("a", 64)
	session := filepath.Join(workspace, "onboardee", "home", "sessions", key)
	importedPlan := filepath.Join(session, "plan.obk.yaml")
	metadata := filepath.Join(session, "session.json")
	demoWrite(t, importedPlan, "known imported plan bytes\n", 0600)
	demoWrite(t, metadata, "{\"version\":1,\"fixture\":\"known session metadata\"}\n", 0600)
	demoWrite(t, filepath.Join(workspace, "preparation", "records", "newcomer", "preparation.obk.yaml"), "saved preparation\n", 0600)
	importedPreparation := filepath.Join(workspace, "onboardee", "workspace", "records", "newcomer", "preparation.obk.yaml")
	resultsPreparation := filepath.Join(workspace, "results", "records", "newcomer", "preparation.obk.yaml")
	resultsRecord := filepath.Join(workspace, "results", "records", "newcomer", "record.obk.yaml")
	demoWrite(t, importedPreparation, "imported preparation values\n", 0600)
	demoWrite(t, resultsPreparation, "results preparation snapshot\n", 0600)
	demoWrite(t, resultsRecord, "results personal record snapshot\n", 0600)
	before := demoTree(t, root)
	var stdout, stderr bytes.Buffer
	if err := execute(root, []string{"status"}, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatalf("status failed: %v; stderr=%q", err, stderr.String())
	}
	for _, path := range []string{importedPlan, metadata, importedPreparation, resultsPreparation, resultsRecord} {
		if !strings.Contains(stdout.String(), path) {
			t.Errorf("status omitted exact working-file path %q; stdout=%q", path, stdout.String())
		}
	}
	if after := demoTree(t, root); !reflect.DeepEqual(before, after) {
		t.Fatalf("status changed workspace: before=%v after=%v", before, after)
	}
}

func TestStatusDoesNotSelectEitherOfTwoImportedSessions(t *testing.T) {
	root, _ := demoOwnedWorkspace(t)
	physicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	root = physicalRoot
	workspace := filepath.Join(root, "output", "manual-test")
	var importedPaths []string
	for _, digit := range []string{"a", "b"} {
		session := filepath.Join(workspace, "onboardee", "home", "sessions", strings.Repeat(digit, 64))
		plan := filepath.Join(session, "plan.obk.yaml")
		metadata := filepath.Join(session, "session.json")
		demoWrite(t, plan, "distinct imported plan "+digit+"\n", 0600)
		demoWrite(t, metadata, "{\"version\":1,\"fixture\":\"session "+digit+"\"}\n", 0600)
		importedPaths = append(importedPaths, plan, metadata)
	}
	before := demoTree(t, root)
	var stdout, stderr bytes.Buffer
	if err := execute(root, []string{"status"}, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatalf("ambiguous status failed instead of reporting read-only state: %v; stderr=%q", err, stderr.String())
	}
	for _, path := range importedPaths {
		if strings.Contains(stdout.String(), path) {
			t.Errorf("ambiguous status selected a concrete imported file %q; stdout=%q", path, stdout.String())
		}
	}
	if !strings.Contains(stdout.String(), "no session selected") {
		t.Errorf("ambiguous status did not explicitly state no selection; stdout=%q", stdout.String())
	}
	if after := demoTree(t, root); !reflect.DeepEqual(before, after) {
		t.Fatalf("ambiguous status changed workspace: before=%v after=%v", before, after)
	}
}

func TestStageResetIsIsolatedAndRecoverable(t *testing.T) {
	for _, selected := range []string{"invitation", "onboardee", "results"} {
		t.Run(selected, func(t *testing.T) {
			root, workspace := demoOwnedWorkspace(t)
			before := map[string]map[string]string{}
			for _, stage := range manualTestStages {
				demoWrite(t, filepath.Join(workspace, stage, "nested", "state.bin"), "private "+stage+"\x00\n", 0600)
				before[stage] = demoTree(t, filepath.Join(workspace, stage))
			}
			record := filepath.Join(workspace, "onboardee", "record.obk.yaml")
			demoWrite(t, record, "personal answers and actual recorded outcomes\n", 0600)
			before["onboardee"] = demoTree(t, filepath.Join(workspace, "onboardee"))
			demoExecuteOK(t, root, "reset", selected)
			for _, stage := range manualTestStages {
				if stage == selected {
					continue
				}
				if after := demoTree(t, filepath.Join(workspace, stage)); !reflect.DeepEqual(before[stage], after) {
					t.Errorf("reset %s changed %s: before=%v after=%v", selected, stage, before[stage], after)
				}
			}
			if _, err := os.Stat(filepath.Join(workspace, selected, "nested", "state.bin")); !os.IsNotExist(err) {
				t.Errorf("old %s payload remains active: %v", selected, err)
			}
			demoArchivedStage(t, root, selected, before[selected])
			if body, err := os.ReadFile(filepath.Join(workspace, ".owner")); err != nil || string(body) != "onboard-kit-manual-test-v1\n" {
				t.Errorf("ownership marker changed: %q %v", body, err)
			}
		})
	}
}

func TestResetAllArchivesEveryStageWithoutDeletingPreviousHistory(t *testing.T) {
	root, workspace := demoOwnedWorkspace(t)
	previous := filepath.Join(root, "output", "manual-test-history", "earlier-slot", "invitation", "old.zip")
	demoWrite(t, previous, "earlier recoverable invitation\n", 0600)
	before := map[string]map[string]string{}
	for _, stage := range manualTestStages {
		demoWrite(t, filepath.Join(workspace, stage, "state.bin"), stage+" work\n", 0600)
		before[stage] = demoTree(t, filepath.Join(workspace, stage))
	}
	demoExecuteOK(t, root, "reset", "all")
	for _, stage := range manualTestStages {
		demoArchivedStage(t, root, stage, before[stage])
		if _, err := os.Stat(filepath.Join(workspace, stage, "state.bin")); !os.IsNotExist(err) {
			t.Errorf("old %s work remains active: %v", stage, err)
		}
	}
	if body, err := os.ReadFile(previous); err != nil || string(body) != "earlier recoverable invitation\n" {
		t.Fatalf("earlier history was lost: %q %v", body, err)
	}
	if body, err := os.ReadFile(filepath.Join(workspace, ".owner")); err != nil || string(body) != "onboard-kit-manual-test-v1\n" {
		t.Fatalf("reset all removed ownership marker: %q %v", body, err)
	}
}

func TestMissingStageResetDoesNotCreateOtherStages(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "output", "manual-test")
	demoWrite(t, filepath.Join(workspace, ".owner"), "onboard-kit-manual-test-v1\n", 0600)
	demoExecuteOK(t, root, "reset", "results")
	for _, stage := range manualTestStages {
		if _, err := os.Lstat(filepath.Join(workspace, stage)); !os.IsNotExist(err) {
			t.Errorf("missing-stage reset created %s: %v", stage, err)
		}
	}
}

func TestActiveStageLockRefusesResetBeforeAnyStageMoves(t *testing.T) {
	for _, selected := range []string{"onboardee", "all"} {
		t.Run(selected, func(t *testing.T) {
			root, workspace := demoOwnedWorkspace(t)
			before := map[string]map[string]string{}
			for _, stage := range manualTestStages {
				demoWrite(t, filepath.Join(workspace, stage, "state.bin"), "owned "+stage+" work\x00\n", 0600)
				before[stage] = demoTree(t, filepath.Join(workspace, stage))
			}
			history := filepath.Join(root, "output", "manual-test-history")
			demoWrite(t, filepath.Join(history, "prior-slot", "results", "prior.bin"), "prior recoverable results\n", 0600)
			historyBefore := demoTree(t, history)
			lock, err := os.OpenFile(filepath.Join(workspace, ".lock-onboardee"), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
			if err != nil {
				t.Fatal(err)
			}
			defer lock.Close()
			if err := unix.Flock(int(lock.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
				t.Fatalf("could not establish independent active-stage lock: %v", err)
			}
			defer unix.Flock(int(lock.Fd()), unix.LOCK_UN)
			demoExecuteError(t, root, "reset", selected)
			for _, stage := range manualTestStages {
				if after := demoTree(t, filepath.Join(workspace, stage)); !reflect.DeepEqual(before[stage], after) {
					t.Errorf("active onboardee lock allowed reset %s to change %s: before=%v after=%v", selected, stage, before[stage], after)
				}
			}
			if after := demoTree(t, history); !reflect.DeepEqual(historyBefore, after) {
				t.Fatalf("refused reset %s changed archive history: before=%v after=%v", selected, historyBefore, after)
			}
			if err := unix.Flock(int(lock.Fd()), unix.LOCK_UN); err != nil {
				t.Fatal(err)
			}
			demoExecuteOK(t, root, "reset", selected)
			for _, stage := range manualTestStages {
				if selected != "all" && stage != selected {
					if after := demoTree(t, filepath.Join(workspace, stage)); !reflect.DeepEqual(before[stage], after) {
						t.Errorf("unlocked reset %s changed unrelated %s", selected, stage)
					}
					continue
				}
				demoArchivedStage(t, root, stage, before[stage])
				if _, err := os.Stat(filepath.Join(workspace, stage, "state.bin")); !os.IsNotExist(err) {
					t.Errorf("unlocked reset left old %s payload active: %v", stage, err)
				}
			}
			if body, err := os.ReadFile(filepath.Join(history, "prior-slot", "results", "prior.bin")); err != nil || string(body) != "prior recoverable results\n" {
				t.Fatalf("successful reset lost prior archive: %q %v", body, err)
			}
		})
	}
}

func TestUnsafeWorkspaceRefusesWithoutChangingExistingObjects(t *testing.T) {
	for _, kind := range []string{"unowned", "wrong-marker", "workspace-symlink", "stage-symlink", "marker-symlink", "marker-directory"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			workspace := filepath.Join(root, "output", "manual-test")
			target := filepath.Join(root, "untouched-target")
			demoWrite(t, filepath.Join(target, "owner-data"), "must remain byte exact\x00\n", 0600)
			switch kind {
			case "workspace-symlink":
				demoMkdir(t, filepath.Dir(workspace))
				demoSymlink(t, target, workspace)
			case "unowned":
				demoWrite(t, filepath.Join(workspace, "existing-work"), "not owned by helper\n", 0600)
			case "wrong-marker":
				demoWrite(t, filepath.Join(workspace, ".owner"), "someone else's workspace\n", 0600)
			case "marker-symlink":
				demoWrite(t, filepath.Join(target, "marker"), "onboard-kit-manual-test-v1\n", 0600)
				demoMkdir(t, workspace)
				demoSymlink(t, filepath.Join(target, "marker"), filepath.Join(workspace, ".owner"))
			case "marker-directory":
				demoMkdir(t, filepath.Join(workspace, ".owner"))
			case "stage-symlink":
				demoWrite(t, filepath.Join(workspace, ".owner"), "onboard-kit-manual-test-v1\n", 0600)
				demoSymlink(t, target, filepath.Join(workspace, "plan"))
			}
			before := demoTree(t, root)
			demoExecuteError(t, root, "reset", "plan")
			if after := demoTree(t, root); !reflect.DeepEqual(before, after) {
				t.Fatalf("unsafe %s changed objects: before=%v after=%v", kind, before, after)
			}
		})
	}
}

func TestInitCreatesPlanOnceAndDoesNotReplaceExistingPlan(t *testing.T) {
	root := t.TempDir()
	physicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	root = physicalRoot
	binary := filepath.Join(root, "output", "demo-bin", "onboard-kit")
	demoWrite(t, binary, "#!/bin/sh\nset -eu\nprintf '%s\\n' \"$@\" > \"${0%/*}/calls\"\n[ \"$1\" = new ]\nshift\nout=\nwhile [ \"$#\" -gt 0 ]; do\n case \"$1\" in\n  --templates) shift 2 ;;\n  --out) out=$2; shift 2 ;;\n  *) exit 9 ;;\n esac\ndone\n[ -n \"$out\" ]\nprintf 'fake blank Acme plan\\n' > \"$out\"\n", 0700)
	demoMkdir(t, filepath.Join(root, "docs", "reference", "examples", "acme", "templates"))
	demoExecuteOK(t, root, "init")
	plan := filepath.Join(root, "output", "manual-test", "plan", "plan.obk.yaml")
	if body, err := os.ReadFile(plan); err != nil || string(body) != "fake blank Acme plan\n" {
		t.Fatalf("init did not create the app's plan: %q %v", body, err)
	}
	calls := filepath.Join(filepath.Dir(binary), "calls")
	arguments, err := os.ReadFile(calls)
	if err != nil {
		t.Fatal(err)
	}
	want := "new\n--templates\n" + filepath.Join(root, "docs", "reference", "examples", "acme", "templates") + "\n--out\n" + plan + "\n"
	if string(arguments) != want {
		t.Fatalf("wrong init app arguments: got %q want %q", arguments, want)
	}
	demoWrite(t, plan, "manually customized organization plan\n", 0600)
	if err := os.Remove(calls); err != nil {
		t.Fatal(err)
	}
	demoExecuteOK(t, root, "init")
	if body, err := os.ReadFile(plan); err != nil || string(body) != "manually customized organization plan\n" {
		t.Fatalf("ordinary init replaced existing plan: %q %v", body, err)
	}
	if _, err := os.Stat(calls); !os.IsNotExist(err) {
		t.Fatalf("ordinary init invoked the app for an existing plan: %v", err)
	}
}

func demoOwnedWorkspace(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	workspace := filepath.Join(root, "output", "manual-test")
	demoWrite(t, filepath.Join(workspace, ".owner"), "onboard-kit-manual-test-v1\n", 0600)
	for _, stage := range manualTestStages {
		demoMkdir(t, filepath.Join(workspace, stage))
	}
	return root, workspace
}

func demoExecuteOK(t *testing.T, root string, arguments ...string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if err := execute(root, arguments, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatalf("execute %v failed: %v; stdout=%q stderr=%q", arguments, err, stdout.String(), stderr.String())
	}
}

func demoExecuteError(t *testing.T, root string, arguments ...string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if err := execute(root, arguments, strings.NewReader(""), &stdout, &stderr); err == nil {
		t.Fatalf("execute %v unexpectedly succeeded; stdout=%q stderr=%q", arguments, stdout.String(), stderr.String())
	}
}

func demoArchivedStage(t *testing.T, root, stage string, want map[string]string) {
	t.Helper()
	history := filepath.Join(root, "output", "manual-test-history")
	slots, err := os.ReadDir(history)
	if err != nil {
		t.Fatalf("no recoverable history for %s: %v", stage, err)
	}
	for _, slot := range slots {
		if slot.IsDir() && reflect.DeepEqual(demoTree(t, filepath.Join(history, slot.Name(), stage)), want) {
			return
		}
	}
	t.Fatalf("history has no byte-exact recoverable %s stage", stage)
}

func demoTree(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if path == root && os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		if path == root {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		switch {
		case entry.Type()&os.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			result[relative] = "symlink:" + target
		case entry.IsDir():
			result[relative] = "directory"
		default:
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			result[relative] = "file:" + string(body)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func demoWrite(t *testing.T, path, body string, mode fs.FileMode) {
	t.Helper()
	demoMkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatal(err)
	}
}

func demoMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
}

func demoSymlink(t *testing.T, target, path string) {
	t.Helper()
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
}

func TestPrepareNeedsAnAccountThatIsOneFolderName(t *testing.T) {
	root, _ := demoOwnedWorkspace(t)
	demoExecuteError(t, root, "prepare")
	for _, account := range []string{"", ".", "..", "a/b", `a\b`} {
		demoExecuteError(t, root, "prepare", account)
	}
}

func TestLaterStagesNeedAPreparedAccount(t *testing.T) {
	root, workspace := demoOwnedWorkspace(t)
	demoWrite(t, filepath.Join(workspace, "plan", "plan.obk.yaml"), "plan\n", 0600)
	demoExecuteError(t, root, "invite", "helper")
	demoExecuteError(t, root, "results")
}

func TestPrepareRefusesASecondAccountUntilPreparationIsReset(t *testing.T) {
	root, workspace := demoOwnedWorkspace(t)
	demoWrite(t, filepath.Join(workspace, "plan", "plan.obk.yaml"), "plan\n", 0600)
	demoWrite(t, filepath.Join(workspace, "preparation", "records", "first", "preparation.obk.yaml"), "saved\n", 0600)
	before := demoTree(t, filepath.Join(workspace, "preparation"))
	demoExecuteError(t, root, "prepare", "second")
	if after := demoTree(t, filepath.Join(workspace, "preparation")); !reflect.DeepEqual(before, after) {
		t.Fatalf("a refused second account changed the preparation: %v -> %v", before, after)
	}
}

func TestSecondAccountRefusalNamesTheExistingAccount(t *testing.T) {
	root, workspace := demoOwnedWorkspace(t)
	demoWrite(t, filepath.Join(workspace, "plan", "plan.obk.yaml"), "plan\n", 0600)
	demoWrite(t, filepath.Join(workspace, "preparation", "records", "first", "preparation.obk.yaml"), "saved\n", 0600)
	var stdout, stderr bytes.Buffer
	err := execute(root, []string{"prepare", "second"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), `preparation already exists for "first"`) {
		t.Fatalf("wrong refusal: %v", err)
	}
}

func TestPrepareWithoutAnAccountGivesAnExample(t *testing.T) {
	root, _ := demoOwnedWorkspace(t)
	var stdout, stderr bytes.Buffer
	err := execute(root, []string{"prepare", ""}, strings.NewReader(""), &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "for example: just demo-prepare jules") {
		t.Fatalf("wrong refusal: %v", err)
	}
}
