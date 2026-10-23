package reader_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"sync"
	"testing"

	contract "onboard-kit/contracts/catalog/reader"
	"onboard-kit/internal/model"
)

type inventoryFixture struct {
	Source  string             `json:"source"`
	Entries []contract.Summary `json:"entries"`
}

type digestRow struct {
	ID     string `json:"id"`
	Size   int    `json:"size"`
	SHA256 string `json:"sha256"`
}

type digestFixture struct {
	Source  string      `json:"source"`
	Entries []digestRow `json:"entries"`
}

func loadFixture(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile("testdata/" + path)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		t.Fatal(err)
	}
}

func inventory(t *testing.T) []contract.Summary {
	t.Helper()
	var fixture inventoryFixture
	loadFixture(t, "inventory.json", &fixture)
	if fixture.Source == "" || len(fixture.Entries) != 80 {
		t.Fatal("invalid independent inventory fixture")
	}
	counts := map[contract.Carrier]int{}
	for i, row := range fixture.Entries {
		if row.ID == "" || row.Title == "" || row.Category == "" || row.Goal == "" || row.Result == "" || row.ContentPath == "" {
			t.Fatal("incomplete inventory oracle")
		}
		if i > 0 && fixture.Entries[i-1].ID >= row.ID {
			t.Fatal("inventory oracle not unique and ID sorted")
		}
		counts[row.Carrier]++
	}
	if counts[contract.CarrierSectionTemplate] != 59 || counts[contract.CarrierPreparationGuidance] != 19 || counts[contract.CarrierToolHelp] != 2 {
		t.Fatal("wrong independent carrier split")
	}
	return fixture.Entries
}

func digestOracles(t *testing.T) map[string]digestRow {
	t.Helper()
	var fixture digestFixture
	loadFixture(t, "content-digests.json", &fixture)
	if fixture.Source == "" || len(fixture.Entries) != 80 {
		t.Fatal("content digests must freeze all eighty normative assets")
	}
	rows := make(map[string]digestRow, 80)
	for _, row := range fixture.Entries {
		if _, duplicate := rows[row.ID]; duplicate {
			t.Fatal("duplicate digest oracle")
		}
		if row.Size <= 0 || len(row.SHA256) != 64 {
			t.Fatal("unfrozen or invalid digest oracle")
		}
		rows[row.ID] = row
	}
	return rows
}

func assertKnown(t *testing.T, reader contract.Reader, summary contract.Summary, oracle digestRow) {
	t.Helper()
	entry, problems := reader.Read(summary.ID)
	if len(problems) != 0 || !reflect.DeepEqual(entry.Summary, summary) {
		t.Fatalf("Read %s metadata: got %#v, problems %#v", summary.ID, entry.Summary, problems)
	}
	if len(entry.Content) != oracle.Size {
		t.Fatalf("Read %s content size: got %d want %d", summary.ID, len(entry.Content), oracle.Size)
	}
	hash := sha256.Sum256(entry.Content)
	if hex.EncodeToString(hash[:]) != oracle.SHA256 {
		t.Fatalf("Read %s differs from independently frozen content digest", summary.ID)
	}
}

func TestExactFullInventory(t *testing.T) {
	t.Parallel()
	want := inventory(t)
	shared := newReader()
	for range 2 {
		if got := shared.List(); !reflect.DeepEqual(got, want) {
			t.Fatalf("catalog List got %#v want %#v", got, want)
		}
	}
}

func TestEveryExactEntryAndPinnedContent(t *testing.T) {
	t.Parallel()
	rows := inventory(t)
	digests := digestOracles(t)
	shared := newReader()
	for _, row := range rows {
		t.Run(row.ID, func(t *testing.T) { assertKnown(t, shared, row, digests[row.ID]) })
	}
}

var unknownIDs = []string{"", "unknown-private-selector", "Company-email", " company-email", "company-email ", "Company email", "sections/company-email.yaml", "../company-email", "./company-email", "company-email/", "\x00company-email", "\xff", "invitation-prepare.md", "onboarding-resume.yaml"}

func assertUnknown(t *testing.T, reader contract.Reader, id string) {
	t.Helper()
	entry, problems := reader.Read(id)
	if entry.Summary != (contract.Summary{}) || len(entry.Content) != 0 {
		t.Fatalf("unknown Read returned partial entry: %#v", entry)
	}
	want := []model.Problem{{Code: "catalog.unknown-id", Severity: model.SeverityError, Message: "Choose an exact ID from catalog list.", Detail: map[string]string{"field": "id"}}}
	if !reflect.DeepEqual(problems, want) {
		t.Fatalf("unknown Read problems: got %#v want %#v", problems, want)
	}
}

func TestUnknownExactIdentitiesReturnNoPartialEntry(t *testing.T) {
	t.Parallel()
	shared := newReader()
	for _, id := range unknownIDs {
		t.Run(id, func(t *testing.T) { assertUnknown(t, shared, id) })
	}
}

func TestOwnedListsContentAndProblemMaps(t *testing.T) {
	t.Parallel()
	want := inventory(t)
	shared := newReader()
	first := shared.List()
	held := shared.List()
	if !reflect.DeepEqual(first, want) || !reflect.DeepEqual(held, want) {
		t.Fatal("missing literal inventory before ownership check")
	}
	first[0].Title = "caller edit"
	slices.Reverse(first)
	if !reflect.DeepEqual(held, want) || !reflect.DeepEqual(shared.List(), want) {
		t.Fatal("catalog list storage aliases another result or compiled data")
	}
	entry, problems := shared.Read("company-email")
	other, more := shared.Read("company-email")
	if len(problems)+len(more) != 0 || len(entry.Content) == 0 || !bytes.Equal(entry.Content, other.Content) {
		t.Fatal("missing known content ownership control")
	}
	snapshot := bytes.Clone(other.Content)
	entry.Content[0] ^= 1
	entry.Summary.ID = "caller edit"
	again, _ := shared.Read("company-email")
	if !bytes.Equal(other.Content, snapshot) || !bytes.Equal(again.Content, snapshot) || again.Summary.ID != "company-email" {
		t.Fatal("returned content or metadata edits changed another call")
	}
	_, firstProblems := shared.Read("unknown")
	_, heldProblems := shared.Read("unknown")
	if len(firstProblems) != 1 || len(heldProblems) != 1 {
		t.Fatal("missing unknown control")
	}
	firstProblems[0].Detail["field"] = "caller edit"
	firstProblems[0].Code = "caller.edit"
	if heldProblems[0].Detail["field"] != "id" {
		t.Fatal("diagnostic detail aliases")
	}
	assertUnknown(t, shared, "unknown")
}

func TestConcurrentKnownUnknownAndListCalls(t *testing.T) {
	t.Parallel()
	rows := inventory(t)
	digests := digestOracles(t)
	shared := newReader()
	var group sync.WaitGroup
	for round := range 3 {
		for _, row := range rows {
			group.Add(1)
			go func() {
				defer group.Done()
				assertKnown(t, shared, row, digests[row.ID])
				assertUnknown(t, shared, unknownIDs[round])
				if !reflect.DeepEqual(shared.List(), rows) {
					t.Error("concurrent catalog state changed")
				}
			}()
		}
	}
	group.Wait()
}
