package editor_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	contract "onboard-kit/contracts/catalog/editor"
	"onboard-kit/internal/model"
)

func helpBaseline(t *testing.T) model.Section {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "onboarding-help-new.canonical"))
	if err != nil {
		t.Fatal(err)
	}
	canonical := bytes.TrimSuffix(data, []byte("\n"))
	sum := sha256.Sum256(canonical)
	var section model.Section
	if err := json.Unmarshal(canonical, &section); err != nil {
		t.Fatal(err)
	}
	section.Body += "\n"
	section.Variables[1].Type = ""
	section.Variables[2].Type = ""
	section.Source = &model.Source{Template: "onboarding-help", Hash: hex.EncodeToString(sum[:])}
	return section
}

func TestAddContactBindingUsesActualHeldDirectoryAndPreservesExactList(t *testing.T) {
	value := model.Value{Kind: model.ValueKindList, List: []string{"owner", "", "owner"}}
	ops := []contract.Operation{{Kind: contract.Add, SectionID: "onboarding-help-new", SourceID: "onboarding-help", Bindings: []contract.Binding{{Name: "general_contacts", Value: value}}}}
	plan := history()
	result := success(t, plan, ops)
	want := history()
	want.Sections = append(want.Sections, helpBaseline(t))
	want.Answers.Sections["onboarding-help-new"] = map[string]model.Value{"general_contacts": value}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("held directory was omitted, substituted or binding normalized: %#v", result)
	}
	if !reflect.DeepEqual(plan, history()) {
		t.Error("directory binding mutated held plan")
	}
}

func TestAddContactBindingRejectsUnknownCaseAndOtherDirectoryNamespaces(t *testing.T) {
	for _, target := range []string{"private-unknown-target", "OWNER", "maintainer", "group"} {
		refusal(t, history(), []contract.Operation{{Kind: contract.Add, SectionID: "onboarding-help-new", SourceID: "onboarding-help", Bindings: []contract.Binding{{Name: "general_contacts", Value: model.Value{Kind: model.ValueKindList, List: []string{target}}}}}}, own("directory.unknown-id", "add-answer", "operations[0].instances[0].bindings[0].value", "Submit an exact declared directory target, or leave the binding unset."))
	}
}

func warningHistory() model.Plan {
	plan := history()
	plan.Directory.Contacts[0].ID = "private-contact-id"
	plan.Directory.Contacts[0].Name = "Private Contact Display"
	plan.Directory.Contacts[0].Email = ""
	plan.Directory.Roles[0].ID = "private-role-id"
	plan.Directory.Roles[0].Name = "Private Role Display"
	plan.Directory.Roles[0].Holders = nil
	plan.Directory.Teams[0].Contact = "private-contact-id"
	plan.Directory.Teams[0].Holders = map[string][]string{"private-role-id": {}}
	return plan
}

func TestCompoundWarningsKeepOperationThenFinalOrderWithoutPrivateMetadata(t *testing.T) {
	plan := warningHistory()
	ops := []contract.Operation{{Kind: contract.Add, SectionID: "repository-new", SourceID: "repository-access"}, {Kind: contract.Add, SectionID: "document-new", SourceID: "document-reading"}}
	result, diagnostics := newEditor().Edit(plan, ops)
	data, err := os.ReadFile(filepath.Join("testdata", "compound-warnings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var expected []contract.Diagnostic
	if err := json.Unmarshal(data, &expected); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(diagnostics, expected) {
		t.Errorf("compound warning order/multiplicity/privacy projection differs: %#v", diagnostics)
	}
	want := warningHistory()
	want.Sections = append(want.Sections, baseline(t, "repository-new", "repository-access"), baseline(t, "document-new", "document-reading"))
	if !reflect.DeepEqual(result, want) || !reflect.DeepEqual(plan, warningHistory()) {
		t.Error("warnings refused, repaired or changed unrelated held data")
	}
	encoded, err := json.Marshal(diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	for _, private := range []string{"private-contact-id", "private-role-id", "Private Contact Display", "Private Role Display", "departed", "hidden_fact", "held explicit fact"} {
		if strings.Contains(string(encoded), private) {
			t.Errorf("raw dependency metadata leaked: %s", private)
		}
	}
}

func TestExactly1024IdempotentOperationsSucceedAnd1025Refuses(t *testing.T) {
	for _, count := range []int{1024, 1025} {
		plan := history()
		ops := make([]contract.Operation, count)
		for i := range ops {
			ops[i] = contract.Operation{Kind: contract.Retire, SectionID: "retired"}
		}
		if count == 1025 {
			refusal(t, plan, ops, own("edit.limit", "admission", "request", "Reduce held plan or edit data to the supported limits."))
			continue
		}
		result, diagnostics := newEditor().Edit(plan, ops)
		expected := []contract.Diagnostic{{Code: "answer.unknown-question", Stage: "final-plan", Path: "plan", Severity: model.SeverityWarning, Message: "Preserve undeclared held answers for explicit history review; submit only declared local questions."}}
		if !reflect.DeepEqual(result, history()) || !reflect.DeepEqual(diagnostics, expected) {
			t.Errorf("inclusive operation bound refused valid idempotent operations: %#v", diagnostics)
		}
		if !reflect.DeepEqual(plan, history()) {
			t.Error("boundary batch changed caller-owned data")
		}
	}
}
