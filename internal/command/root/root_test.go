package root

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionFlagPrintsTheBuildVersionAndExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"--version"}, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d, stderr %q", code, stderr.String())
	}
	if want := "onboard-kit " + version + "\n"; stdout.String() != want {
		t.Fatalf("got %q, want %q", stdout.String(), want)
	}
}

func TestVersionFlagDefaultsToDevelopmentBuild(t *testing.T) {
	if version != "development" {
		t.Fatalf("version is %q, want development until a build sets it", version)
	}
}

func runIn(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(args, strings.NewReader(""), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func isolated(t *testing.T) (string, string) {
	t.Helper()
	home := filepath.Join(t.TempDir(), "state")
	working := t.TempDir()
	t.Setenv("OBK_HOME", home)
	t.Chdir(working)
	return home, working
}

var (
	acmePlan, _      = filepath.Abs("../../../docs/reference/examples/acme/acme.obk.yaml")
	acmeTemplates, _ = filepath.Abs("../../../docs/reference/examples/acme/templates")
)

func copiedPlan(t *testing.T, folder string) string {
	t.Helper()
	data, err := os.ReadFile(acmePlan)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(folder, "acme.obk.yaml")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestNoArgumentWithoutATerminalPrintsUsageAndTheScriptingCommands(t *testing.T) {
	isolated(t)
	code, stdout, stderr := runIn(t)
	want := usageText + "\nThe interactive tool needs a terminal. For scripts, use new --out FILE, assemble, catalog, prepare, invite, or stdio.\n"
	if code != exitUsage || stdout != "" || stderr != want {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

func TestAFileOfTheWrongKindIsRefusedInOneSentenceByEitherSpelling(t *testing.T) {
	_, working := isolated(t)
	if err := os.WriteFile(filepath.Join(working, "notes.txt"), []byte("hello\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"notes.txt"}, {"open", "notes.txt"}} {
		code, stdout, stderr := runIn(t, args...)
		if code != 1 || stdout != "" || stderr != "onboard-kit: notes.txt is neither a plan nor an invitation.\n" {
			t.Errorf("%v: exit %d, stdout %q, stderr %q", args, code, stdout, stderr)
		}
	}
}

func TestAPlanWithoutATerminalNamesTheScriptingCommands(t *testing.T) {
	home, working := isolated(t)
	copiedPlan(t, working)
	for _, args := range [][]string{{"acme.obk.yaml"}, {"open", "acme.obk.yaml"}} {
		code, _, stderr := runIn(t, args...)
		if code != exitUsage || stderr != "onboard-kit: the plan editor needs a terminal. For scripts, use assemble --plan FILE or stdio.\n" {
			t.Errorf("%v: exit %d, stderr %q", args, code, stderr)
		}
	}
	if _, err := os.Stat(home); !os.IsNotExist(err) {
		t.Fatalf("a refused open touched the state home: %v", err)
	}
}

func TestAnInvitationWithoutATerminalNamesTheStartCommand(t *testing.T) {
	_, working := isolated(t)
	if err := os.WriteFile(filepath.Join(working, "obk-invitation-a.zip"), []byte("PK\x03\x04"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runIn(t, "obk-invitation-a.zip")
	if code != exitUsage || stderr != "onboard-kit: beginning an onboarding needs a terminal. For scripts, use start --invitation FILE --root DIR.\n" {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
}

func TestOpenNeedsExactlyOneFileThatExists(t *testing.T) {
	isolated(t)
	if code, _, stderr := runIn(t, "open"); code != exitUsage || stderr != "usage: onboard-kit open FILE\n" {
		t.Errorf("exit %d, stderr %q", code, stderr)
	}
	if code, _, stderr := runIn(t, "open", "gone.obk.yaml"); code != 1 || stderr != "onboard-kit: gone.obk.yaml does not exist.\n" {
		t.Errorf("exit %d, stderr %q", code, stderr)
	}
}

func TestAWordThatIsNeitherACommandNorAFileIsAnUnknownCommand(t *testing.T) {
	isolated(t)
	if code, _, stderr := runIn(t, "gone.obk.yaml"); code != exitUsage || stderr != "onboard-kit: unknown command \"gone.obk.yaml\"\n" {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
}

func TestACommandNameWinsOverAFileOfThatName(t *testing.T) {
	_, working := isolated(t)
	if err := os.WriteFile(filepath.Join(working, "walk"), []byte("hello\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := runIn(t, "walk"); code != exitUsage || stderr != "walk: --plan is required\n" {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
}

func TestNewForScriptsNamesAMissingFolderInPlainWords(t *testing.T) {
	_, working := isolated(t)
	missing := filepath.Join(working, "plans", "acme")
	code, _, stderr := runIn(t, "new", "--out", filepath.Join(missing, "core.obk.yaml"))
	if code != 1 || stderr != "new: the folder "+missing+" does not exist. Create it first, or give a path in a folder that exists.\n" {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if _, err := os.Stat(missing); err == nil {
		t.Fatal("a script's new created the folder")
	}
}

func TestNewForScriptsWritesTheCorePlanOnceAndRecordsNothing(t *testing.T) {
	home, working := isolated(t)
	code, stdout, stderr := runIn(t, "new", "--out", "core.obk.yaml", "--id", "plan-one", "--generated-on", "2026-11-04")
	if code != 0 || stdout != "" || stderr != "" {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	written, err := os.ReadFile(filepath.Join(working, "core.obk.yaml"))
	if err != nil || !strings.Contains(string(written), "id: plan-one\n") || !strings.Contains(string(written), "templateSet: "+version+"\n") || !strings.Contains(string(written), "- id: ticket-pairing-navigator\n") {
		t.Fatalf("plan %q, %v", written, err)
	}
	code, _, stderr = runIn(t, "new", "--out", "core.obk.yaml")
	if code != 1 || stderr != "new: core.obk.yaml already exists, and a new plan never replaces a file.\n" {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if again, _ := os.ReadFile(filepath.Join(working, "core.obk.yaml")); string(again) != string(written) {
		t.Fatal("the plan was replaced")
	}
	code, stdout, _ = runIn(t, "new", "--id", "plan-one", "--generated-on", "2026-11-04")
	if code != 0 || stdout != string(written) {
		t.Fatalf("standard output differs from the file: exit %d", code)
	}
	if _, err := os.Stat(home); !os.IsNotExist(err) {
		t.Fatalf("a script run touched the state home: %v", err)
	}
}

func TestNewWithTemplatesKeepsItsOldBehavior(t *testing.T) {
	_, working := isolated(t)
	if code, _, stderr := runIn(t, "new", "--templates", acmeTemplates, "--template-set", "example", "--id", "x", "--out", "plan.obk.yaml"); code != 0 {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if data, _ := os.ReadFile(filepath.Join(working, "plan.obk.yaml")); !strings.Contains(string(data), "templateSet: example\n") {
		t.Fatalf("got %q", data)
	}
	if code, _, stderr := runIn(t, "new", "--template-set", "example"); code != exitUsage || !strings.Contains(stderr, "--template-set applies only with --templates") {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
}
