package walk

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"onboard-kit/internal/tui"
)

const readyPreparation = "schemaVersion: 1\nkind: preparation\naccount: newcomer-one\nplanId: \"6f1c2d3e-4a5b-4c6d-8e7f-90a1b2c3d4e5\"\nrecipient:\n  name: Example newcomer\n  onboardingOwner: alex-rivera\n"

func preparedRoot(t *testing.T) (string, string) {
	t.Helper()
	root, plan := copiedPlan(t)
	folder := filepath.Join(root, "records", "newcomer-one")
	if err := os.MkdirAll(folder, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "preparation.obk.yaml"), []byte(readyPreparation), 0o600); err != nil {
		t.Fatal(err)
	}
	return root, plan
}

func TestTheInvitationPortChecksEachPreparedPerson(t *testing.T) {
	root, plan := preparedRoot(t)
	people, outcome := newInvitationFiles(root, plan).People()
	if outcome != tui.InvitationDone || len(people) != 1 {
		t.Fatalf("the prepared person is listed: %+v %v", people, outcome)
	}
	person := people[0]
	if person.Account != "newcomer-one" || person.Refusal != tui.InvitationDone || !person.OwnerNamed || len(person.Blockers) != 0 || len(person.PreparationErrors) != 0 {
		t.Errorf("the person is ready: %+v", person)
	}
	if want := filepath.Join(root, "obk-invitation-newcomer-one.zip"); person.Output != want {
		t.Errorf("the default file sits beside the plan: %q, want %q", person.Output, want)
	}
	if len(person.PreaccessContacts) != 4 || person.Preparation.Recipient.Name != "Example newcomer" || person.Plan.ID == "" {
		t.Errorf("the port carries the checked values: %+v", person)
	}
}

func TestTheInvitationPortWritesOnceAndOffersTheNextFreeName(t *testing.T) {
	root, plan := preparedRoot(t)
	files := newInvitationFiles(root, plan)
	people, _ := files.People()
	request := tui.InvitationRequest{Account: "newcomer-one", Output: people[0].Output, Preaccess: "alex-rivera", Help: []tui.InvitationHelp{{Kind: "contact", ID: "alex-rivera"}, {Kind: "role", ID: "it-contact"}}}
	created := files.Create(request)
	if created.Outcome != tui.InvitationDone || created.Path != request.Output {
		t.Fatalf("the invitation is written: %+v", created)
	}
	first, err := os.ReadFile(request.Output)
	if err != nil || !bytes.HasPrefix(first, []byte("PK")) {
		t.Fatalf("the file is a zip: %v", err)
	}
	again := files.Create(request)
	if again.Outcome != tui.InvitationOutputExists || again.FreeOutput != filepath.Join(root, "obk-invitation-newcomer-one-2.zip") {
		t.Errorf("a second write to the same name is refused with the next free name: %+v", again)
	}
	if held, _ := os.ReadFile(request.Output); !bytes.Equal(held, first) {
		t.Error("the first file is untouched")
	}
	if people, _ := files.People(); people[0].Output != again.FreeOutput {
		t.Errorf("the list proposes the next free name: %q", people[0].Output)
	}
}

func TestTheInvitationPortReportsWhatBlocksAPerson(t *testing.T) {
	root, plan := preparedRoot(t)
	held := []byte("schemaVersion: 1\nkind: preparation\naccount: newcomer-one\nplanId: \"6f1c2d3e-4a5b-4c6d-8e7f-90a1b2c3d4e5\"\n")
	if err := os.WriteFile(filepath.Join(root, "records", "newcomer-one", "preparation.obk.yaml"), held, 0o600); err != nil {
		t.Fatal(err)
	}
	files := newInvitationFiles(root, plan)
	people, _ := files.People()
	if people[0].OwnerNamed {
		t.Errorf("a preparation with no owner is not ready: %+v", people[0])
	}
	refused := files.Create(tui.InvitationRequest{Account: "newcomer-one", Output: people[0].Output, Preaccess: "alex-rivera"})
	if refused.Outcome != tui.InvitationOwnerMissing {
		t.Errorf("the export refuses for the same reason: %+v", refused)
	}
	if _, err := os.Lstat(people[0].Output); !os.IsNotExist(err) {
		t.Errorf("a refusal writes nothing: %v", err)
	}
}
