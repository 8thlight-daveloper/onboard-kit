package editor_test

import (
	"reflect"
	"sync"
	"testing"

	contract "onboard-kit/contracts/catalog/editor"
	"onboard-kit/internal/model"
)

func mutateHeldParts(plan *model.Plan) {
	plan.Sections[1].Source.Hash = "CORRUPTED"
	plan.Sections[1].Variables[0].Example = "CORRUPTED"
	*plan.Sections[1].Variables[0].AskIf.All[0].Answer.IsSettled = false
	plan.Sections[1].Steps[0].Run[0] = "CORRUPTED"
	plan.Sections[1].Steps[0].Env["SAFE_VALUE"] = "CORRUPTED"
	plan.Sections[1].Steps[0].Checks[0].Ask = "CORRUPTED"
	plan.Sections[2].Requires[0] = "CORRUPTED"
	*plan.Sections[2].IncludeIf.All[0].Answer.IsSettled = false
	*plan.Sections[3].Enabled = true
	plan.Answers.Plan["shared_fact"] = text("CORRUPTED")
	plan.Answers.Sections["departed"]["hidden_fact"].List[0] = "CORRUPTED"
	plan.Directory.Roles[0].Holders[0] = "CORRUPTED"
	plan.Directory.Teams[0].Holders["maintainer"] = []string{"CORRUPTED"}
	plan.UnknownFields["extension"].(map[string]any)["items"].([]any)[0] = "CORRUPTED"
	plan.Sections[1].UnknownFields["extension"].(map[string]any)["ordered"].([]any)[0] = "CORRUPTED"
}

func TestAllMutableHeldPartsRemainCallerOwnedAcrossCalls(t *testing.T) {
	editor := newEditor()
	input := history()
	operations := []contract.Operation{{Kind: contract.Retire, SectionID: "repository-old"}}
	first, diagnostics := editor.Edit(input, operations)
	if hasErrors(diagnostics) || first.ID == "" {
		t.Fatalf("proposal required: %#v", diagnostics)
	}
	want := history()
	want.Sections[1].Enabled = flag(false)
	if !reflect.DeepEqual(first, want) || !reflect.DeepEqual(input, history()) {
		t.Fatal("edit changed input or unrelated held data")
	}
	mutateHeldParts(&first)
	if !reflect.DeepEqual(input, history()) {
		t.Error("result still aliases caller-held data")
	}
	second, nextDiagnostics := editor.Edit(input, operations)
	if !reflect.DeepEqual(second, want) || !reflect.DeepEqual(nextDiagnostics, diagnostics) {
		t.Error("result mutation changed a later independent call")
	}
	mutateHeldParts(&input)
	if !reflect.DeepEqual(second, want) {
		t.Error("input mutation changed an earlier returned proposal")
	}
}

func TestNewBindingPayloadsAndInstancesDoNotAlias(t *testing.T) {
	value := model.Value{Kind: model.ValueKindDeferred, Text: "inactive held text", List: []string{"inactive", "", "inactive"}}
	ops := []contract.Operation{{Kind: contract.Add, SectionID: "repository-new", SourceID: "repository-access", Bindings: []contract.Binding{{Name: "repository_scope", Value: value}}}, {Kind: contract.Add, SectionID: "repository-other", SourceID: "repository-access"}}
	input := history()
	result := success(t, input, ops)
	stored := result.Answers.Sections["repository-new"]["repository_scope"]
	if !reflect.DeepEqual(stored, value) {
		t.Fatal("new exact binding lost inactive payload or kind")
	}
	result.Sections[4].Variables[0].Description = "CORRUPTED"
	if result.Sections[5].Variables[0].Description == "CORRUPTED" {
		t.Error("new independent instances share declarations")
	}
	stored.List[0] = "CORRUPTED"
	if value.List[0] != "inactive" || ops[0].Bindings[0].Value.List[0] != "inactive" {
		t.Error("returned new binding aliases operation payload")
	}
	ops[0].Bindings[0].Value.List[1] = "MUTATED INPUT"
	if stored.List[1] != "" {
		t.Error("operation mutation changed returned binding")
	}
	if !reflect.DeepEqual(input, history()) {
		t.Error("addition changed held input")
	}
}

func TestSharedEditorConcurrentCallsAreIndependent(t *testing.T) {
	editor := newEditor()
	input := history()
	ops := []contract.Operation{{Kind: contract.Retire, SectionID: "repository-old"}}
	want := history()
	want.Sections[1].Enabled = flag(false)
	var done sync.WaitGroup
	for range 8 {
		done.Add(1)
		go func() {
			defer done.Done()
			result, diagnostics := editor.Edit(input, ops)
			if hasErrors(diagnostics) || !reflect.DeepEqual(result, want) {
				t.Errorf("overlapping call differed: %#v", diagnostics)
			}
		}()
	}
	done.Wait()
	if !reflect.DeepEqual(input, history()) {
		t.Error("shared calls changed held input")
	}
}

func TestRefusalPreservesEveryInputPartAndReturnedDiagnosticsAreOwned(t *testing.T) {
	editor := newEditor()
	input := history()
	ops := []contract.Operation{{Kind: contract.Add, SectionID: "retired", SourceID: "repository-access"}}
	result, diagnostics := editor.Edit(input, ops)
	if !reflect.DeepEqual(result, model.Plan{}) || len(diagnostics) != 1 || !reflect.DeepEqual(input, history()) {
		t.Fatal("refusal changed input or leaked candidate")
	}
	diagnostics[0].Message = "CORRUPTED"
	_, again := editor.Edit(input, ops)
	if len(again) != 1 || again[0].Message != "Choose a new permanent ID; held definitions and answer scopes reserve their IDs." {
		t.Error("diagnostic state shared across calls")
	}
}
