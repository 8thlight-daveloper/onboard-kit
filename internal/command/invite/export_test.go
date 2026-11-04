package invite_test

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"

	firststepscontract "onboard-kit/contracts/firststeps/assembler"
	"onboard-kit/internal/command/invite"
	"onboard-kit/internal/model"
	plandecoder "onboard-kit/internal/plan/decoder"
	planencoder "onboard-kit/internal/plan/encoder"
)

const readyPlan = `schemaVersion: 1
id: planone
directory:
  contacts:
    - id: owner
      name: Example owner
      email: owner@example.invalid
    - id: quiet
      name: Quiet contact
    - id: chat
      name: Chat contact
      slack: "#help"
`

const readyPreparation = `schemaVersion: 1
kind: preparation
account: newcomer
planId: planone
recipient:
  name: Example newcomer
  onboardingOwner: owner
`

type folder struct {
	root        string
	plan        string
	preparation string
}

func newFolder(t *testing.T, plan string, preparation string) folder {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	held := folder{root: root, plan: filepath.Join(root, "plan.obk.yaml"), preparation: filepath.Join(root, "preparation.obk.yaml")}
	write(t, held.plan, []byte(plan))
	write(t, held.preparation, []byte(preparation))
	return held
}

func write(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func (held folder) request(output string) invite.Request {
	return invite.Request{
		Plan: held.plan, Preparation: held.preparation, Account: "newcomer",
		Output: filepath.Join(held.root, output), Preaccess: "owner",
		Help: []firststepscontract.HelpReference{{Kind: "contact", ID: "owner"}},
	}
}

func entryNames(t *testing.T, path string) []string {
	t.Helper()
	reader, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("%s is not a zip: %v", path, err)
	}
	defer reader.Close()
	var names []string
	for _, entry := range reader.File {
		names = append(names, entry.Name)
	}
	return names
}

func TestExportWritesTheArchiveAndReturnsItsPath(t *testing.T) {
	held := newFolder(t, readyPlan, readyPreparation)
	request := held.request("obk-invitation-newcomer.zip")
	result := invite.Export(request)
	if !result.Written() || result.Code != "written" || result.Path != request.Output {
		t.Fatalf("the export is written to the requested path: %+v", result)
	}
	want := []string{"plan.obk.yaml", "records/newcomer/preparation.obk.yaml", "records/newcomer/record.obk.yaml", "first-steps.md"}
	if got := entryNames(t, request.Output); !slices.Equal(got, want) {
		t.Errorf("the archive holds the four payloads: %v", got)
	}
}

func TestExportNeverReplacesAnExistingFile(t *testing.T) {
	held := newFolder(t, readyPlan, readyPreparation)
	request := held.request("obk-invitation-newcomer.zip")
	write(t, request.Output, []byte("held by someone else\n"))
	result := invite.Export(request)
	if result.Written() || result.Code != "output-exists" || result.Path != "" {
		t.Errorf("an existing file is refused with no path: %+v", result)
	}
	if data, _ := os.ReadFile(request.Output); !bytes.Equal(data, []byte("held by someone else\n")) {
		t.Errorf("the existing file is untouched: %q", data)
	}
}

func TestExportRefusesAPreaccessContactWithNoRoute(t *testing.T) {
	held := newFolder(t, readyPlan, readyPreparation)
	request := held.request("obk-invitation-newcomer.zip")
	request.Preaccess = "quiet"
	result := invite.Export(request)
	if result.Written() || result.Code != "preaccess-help" {
		t.Errorf("a contact with no email or Slack cannot be the pre-access help: %+v", result)
	}
	if _, err := os.Lstat(request.Output); !os.IsNotExist(err) {
		t.Errorf("a refusal writes nothing: %v", err)
	}
}

func TestExportReportsTheBlockersOfAnUnansweredPlan(t *testing.T) {
	held := newFolder(t, string(unansweredExamplePlan(t)), examplePreparation)
	request := held.request("obk-invitation-newcomer.zip")
	request.Preaccess, request.Help = "alex-rivera", nil
	result := invite.Export(request)
	if result.Code != "onboarder-blocked" || len(result.Blockers) == 0 {
		t.Errorf("an unanswered plan blocks the export and names its questions: %+v", result)
	}
}

func TestCheckFindsAReadyPersonAndTheContactsThatCanHelpBeforeAccess(t *testing.T) {
	held := newFolder(t, readyPlan, readyPreparation)
	readiness := invite.Check(held.request("unused.zip"))
	if !readiness.Ready() || readiness.Code != "" {
		t.Errorf("the person is ready: %+v", readiness)
	}
	if !slices.Equal(readiness.PreaccessContacts, []string{"owner", "chat"}) {
		t.Errorf("only named contacts with an email or Slack route can help before access: %v", readiness.PreaccessContacts)
	}
	if readiness.Plan.ID != "planone" || readiness.Preparation.Recipient.Name != "Example newcomer" {
		t.Errorf("the checked plan and preparation come back: %+v %+v", readiness.Plan.ID, readiness.Preparation.Recipient)
	}
	if _, err := os.Lstat(filepath.Join(held.root, "unused.zip")); !os.IsNotExist(err) {
		t.Errorf("a check writes nothing: %v", err)
	}
}

func TestCheckCountsEveryKindOfBlockerAtOnce(t *testing.T) {
	preparation := "schemaVersion: 1\nkind: preparation\naccount: someone-else\nplanId: \"6f1c2d3e-4a5b-4c6d-8e7f-90a1b2c3d4e5\"\n"
	held := newFolder(t, string(unansweredExamplePlan(t)), preparation)
	readiness := invite.Check(held.request("unused.zip"))
	if readiness.Ready() || readiness.Code != "" {
		t.Fatalf("the person is not ready, and the plan is readable: %+v", readiness.Code)
	}
	if len(readiness.PreparationErrors) == 0 || len(readiness.Blockers) == 0 || readiness.OwnerNamed {
		t.Errorf("the account mismatch, the blockers, and the missing owner are all found: %d errors, %d blockers, owner %v", len(readiness.PreparationErrors), len(readiness.Blockers), readiness.OwnerNamed)
	}
}

func TestCheckStopsAtAPlanItCannotUse(t *testing.T) {
	for name, held := range map[string]folder{
		"plan-read":          {plan: "missing.obk.yaml", preparation: "missing.obk.yaml"},
		"plan-identity":      newFolder(t, "schemaVersion: 1\n", readyPreparation),
		"preparation-decode": newFolder(t, readyPlan, "not: [a preparation\n"),
	} {
		if readiness := invite.Check(held.request("unused.zip")); readiness.Code != name || readiness.Ready() {
			t.Errorf("%s: got %+v", name, readiness)
		}
	}
	if readiness := invite.Check(invite.Request{Account: "../escape"}); readiness.Code != "account" {
		t.Errorf("an unsafe account is refused before any read: %+v", readiness)
	}
}

const examplePreparation = `schemaVersion: 1
kind: preparation
account: newcomer
planId: "6f1c2d3e-4a5b-4c6d-8e7f-90a1b2c3d4e5"
recipient:
  onboardingOwner: alex-rivera
`

func unansweredExamplePlan(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../../docs/reference/examples/acme/acme.obk.yaml")
	if err != nil {
		t.Fatal(err)
	}
	plan, problems := plandecoder.New().Decode(data)
	if len(problems) > 0 {
		t.Fatalf("the example plan decodes: %v", problems)
	}
	plan.Answers = model.Answers{}
	encoded, problems := planencoder.New().Encode(plan)
	if len(problems) > 0 {
		t.Fatalf("the unanswered plan encodes: %v", problems)
	}
	return encoded
}
