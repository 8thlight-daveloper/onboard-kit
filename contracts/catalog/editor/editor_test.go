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

func text(value string) model.Value { return model.Value{Kind: model.ValueKindText, Text: value} }

func flag(value bool) *bool { return &value }

func history() model.Plan {
	return model.Plan{
		SchemaVersion: 1, ID: "held organization identity", Provenance: model.Provenance{TemplateSet: "held set", GeneratedOn: "2026-01-01"},
		Origin: model.Origin{PlanID: "legacy-origin", Source: "held-reference", ContentHash: "held-content-hash", RefreshedOn: "unknown-date"}, URL: "held-plan-location", ResultsPath: "records/{account}/record.obk.yaml",
		Directory: model.Directory{Contacts: []model.Contact{{ID: "owner", Name: "Business owner", Email: "owner@example.invalid"}}, Roles: []model.Role{{ID: "maintainer", Name: "Maintainer", Holders: []string{"owner", "owner"}}}, Teams: []model.Team{{ID: "group", Name: "Work group", Contact: "owner", Holders: map[string][]string{"maintainer": {}}}}},
		Variables: []model.Variable{{Name: "shared_fact", Description: "Known shared fact", Example: "suggested example", Default: text("not accepted")}},
		Answers: model.Answers{Plan: map[string]model.Value{"shared_fact": text("held explicit fact")}, Sections: map[string]map[string]model.Value{
			"repository-old": {"resource_name": text("Legacy repository"), "newcomer_note": text("Held newcomer work")},
			"departed":       {"hidden_fact": model.Value{Kind: model.ValueKindList, List: []string{"held", "", "held"}}}, "empty-held": {}, "nil-held": nil,
		}},
		Sections: []model.Section{
			{ID: "landing", Title: "Landing", Summary: "Start here", Steps: []model.Step{}, UnknownFields: map[string]any{}},
			{ID: "repository-old", Title: "Locally edited repository instructions", Summary: "Keep this resource identity", Parent: "landing", Body: "Local instructions $resource_name; personal note $newcomer_note.",
				Variables: []model.Variable{{Name: "resource_name", Description: "Shared resource", Example: "example stays", Default: text("not accepted"), AskIf: &model.Condition{All: []model.Clause{{Question: "plan.shared_fact", Answer: model.AnswerTest{IsSettled: flag(true)}}}}}, {Name: "newcomer_note", Description: "Personal reflection", Audience: model.AudienceOnboardee}},
				Steps:     []model.Step{{ID: "verify", Name: "Verify locally", Run: []string{"never-execute", "", "never-execute"}, Env: map[string]string{"SAFE_VALUE": "held"}, Checks: []model.Check{{Ask: "Did the intended work succeed?"}}}},
				Source:    &model.Source{Template: "repository-access", Hash: "legacy-original-id-baseline"}, After: []string{}, Requires: []string{}, Escalation: []string{},
				UnknownFields: map[string]any{"extension": map[string]any{"ordered": []any{"keep", nil, int64(2), float64(2), map[string]any{}, []any{}}}},
			},
			{ID: "dependent", Title: "Dependent", Summary: "Retain explicit prerequisite", Requires: []string{"repository-old", "repository-old"}, After: []string{}, IncludeIf: &model.Condition{All: []model.Clause{{Question: "plan.shared_fact", Answer: model.AnswerTest{IsSettled: flag(true)}}}}, Body: "Instructions stay unresolved."},
			{ID: "retired", Title: "Reserved old activity", Summary: "Do not reuse its identity", Enabled: flag(false), Source: &model.Source{}},
		},
		UnknownFields: map[string]any{"extension": map[string]any{"items": []any{"first", nil, int(2), int64(2), float64(2), false}, "empty-map": map[string]any{}, "empty-list": []any{}, "nil-map": map[string]any(nil), "nil-list": []any(nil)}},
	}
}

func baseline(t *testing.T, fixture, source string) model.Section {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", fixture+".canonical"))
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
	for i := range section.Variables {
		section.Variables[i].Type = ""
	}
	section.Source = &model.Source{Template: source, Hash: hex.EncodeToString(sum[:])}
	return section
}

func hasErrors(diagnostics []contract.Diagnostic) bool {
	for _, d := range diagnostics {
		if d.Severity == model.SeverityError {
			return true
		}
	}
	return false
}

func success(t *testing.T, plan model.Plan, ops []contract.Operation) model.Plan {
	t.Helper()
	result, diagnostics := newEditor().Edit(plan, ops)
	if hasErrors(diagnostics) {
		t.Fatalf("unexpected refusal: %#v", diagnostics)
	}
	if result.ID != plan.ID {
		t.Fatalf("successful proposal lost held identity: %#v", result)
	}
	return result
}

func refusal(t *testing.T, plan model.Plan, ops []contract.Operation, want contract.Diagnostic) {
	t.Helper()
	result, diagnostics := newEditor().Edit(plan, ops)
	if !reflect.DeepEqual(result, model.Plan{}) {
		t.Error("failed batch leaked a plan or accepted prefix")
	}
	if len(diagnostics) != 1 || !reflect.DeepEqual(diagnostics[0], want) {
		t.Errorf("refusal=%#v want=%#v", diagnostics, want)
	}
}

func own(code, stage, path, message string) contract.Diagnostic {
	return contract.Diagnostic{Code: code, Stage: stage, Path: path, Severity: model.SeverityError, Message: message}
}

func TestNoOperationPreservesCompleteHeldHistoryAndShape(t *testing.T) {
	for _, ops := range [][]contract.Operation{nil, {}} {
		plan := history()
		result := success(t, plan, ops)
		if !reflect.DeepEqual(result, plan) {
			t.Errorf("omission changed held state: %#v", result)
		}
	}
}

func TestPreservedHiddenAnswerWarningAppearsOnceWithoutPrivateMetadata(t *testing.T) {
	result, diagnostics := newEditor().Edit(history(), nil)
	want := []contract.Diagnostic{{Code: "answer.unknown-question", Stage: "final-plan", Path: "plan", Severity: model.SeverityWarning, Message: "Preserve undeclared held answers for explicit history review; submit only declared local questions."}}
	if !reflect.DeepEqual(result, history()) || !reflect.DeepEqual(diagnostics, want) {
		t.Errorf("held warning was duplicated, suppressed or privately expanded: %#v", diagnostics)
	}
}

func TestReorderAddRetirePreservesOldBaselinesAndScopedFacts(t *testing.T) {
	for _, reordered := range []bool{false, true} {
		plan := history()
		if reordered {
			plan.Sections[0], plan.Sections[2] = plan.Sections[2], plan.Sections[0]
		}
		ops := []contract.Operation{
			{Kind: contract.Add, SectionID: "repository-new", SourceID: "repository-access", Bindings: []contract.Binding{{Name: "repository_scope", Value: text("Independent new repository")}}},
			{Kind: contract.Retire, SectionID: "repository-old"},
			{Kind: contract.Add, SectionID: "document-new", SourceID: "document-reading", Bindings: []contract.Binding{{Name: "reading_material", Value: model.Value{Kind: model.ValueKindDeferred}}}},
		}
		result := success(t, plan, ops)
		want := history()
		if reordered {
			want.Sections[0], want.Sections[2] = want.Sections[2], want.Sections[0]
		}
		for i := range want.Sections {
			if want.Sections[i].ID == "repository-old" {
				want.Sections[i].Enabled = flag(false)
			}
		}
		want.Sections = append(want.Sections, baseline(t, "repository-new", "repository-access"), baseline(t, "document-new", "document-reading"))
		want.Answers.Sections["repository-new"] = map[string]model.Value{"repository_scope": text("Independent new repository")}
		want.Answers.Sections["document-new"] = map[string]model.Value{"reading_material": {Kind: model.ValueKindDeferred}}
		if !reflect.DeepEqual(result, want) {
			t.Errorf("reorder/add/retire changed independent history: %#v", result)
		}
	}
}

func TestHeldAndBatchIDsStayReserved(t *testing.T) {
	for _, id := range []string{"repository-old", "retired", "departed", "empty-held", "nil-held"} {
		refusal(t, history(), []contract.Operation{{Kind: contract.Retire, SectionID: "retired"}, {Kind: contract.Add, SectionID: id, SourceID: "PRIVATE_UNKNOWN_SOURCE"}}, own("edit.reserved-id", "reservation", "operations[1].sectionId", "Choose a new permanent ID; held definitions and answer scopes reserve their IDs."))
	}
	refusal(t, history(), []contract.Operation{{Kind: contract.Add, SectionID: "new-one", SourceID: "repository-access"}, {Kind: contract.Retire, SectionID: "new-one"}, {Kind: contract.Add, SectionID: "new-one", SourceID: "document-reading"}}, own("edit.reserved-id", "reservation", "operations[2].sectionId", "Choose a new permanent ID; held definitions and answer scopes reserve their IDs."))
}

func TestExistingSubmitAndUnsetUseClearingAndOwnershipRules(t *testing.T) {
	for _, c := range []struct {
		name   string
		value  model.Value
		stored bool
	}{
		{"none", model.Value{}, false}, {"empty-text", text(""), false}, {"empty-list", model.Value{Kind: model.ValueKindList, List: []string{}}, false}, {"empty-items", model.Value{Kind: model.ValueKindList, List: []string{"", ""}}, false},
		{"spaces", text("  "), true}, {"deferred", model.Value{Kind: model.ValueKindDeferred}, true}, {"literal-deferred", text("deferred"), true},
	} {
		t.Run(c.name, func(t *testing.T) {
			plan := history()
			result := success(t, plan, []contract.Operation{{Kind: contract.Submit, SectionID: "repository-old", Name: "resource_name", Value: c.value}})
			want := history()
			if c.stored {
				want.Answers.Sections["repository-old"]["resource_name"] = c.value
			} else {
				delete(want.Answers.Sections["repository-old"], "resource_name")
			}
			if !reflect.DeepEqual(result, want) {
				t.Error("submission did not preserve actual editor semantics and unrelated newcomer facts")
			}
		})
	}
	result := success(t, history(), []contract.Operation{{Kind: contract.Unset, SectionID: "repository-old", Name: "resource_name"}})
	want := history()
	delete(want.Answers.Sections["repository-old"], "resource_name")
	if !reflect.DeepEqual(result, want) {
		t.Error("explicit unset changed unrelated values")
	}
	refusal(t, history(), []contract.Operation{{Kind: contract.Unset, SectionID: "repository-old", Name: "newcomer_note"}}, own("edit.owner", "operation", "operations[0].name", "Change only onboarder shared facts; preserve newcomer-owned answers."))
}

func TestRetireReactivateOrderingIsExplicitAndAtomic(t *testing.T) {
	refusal(t, history(), []contract.Operation{{Kind: contract.Retire, SectionID: "repository-old"}, {Kind: contract.Submit, SectionID: "repository-old", Name: "resource_name", Value: text("must not install")}}, own("answer.inactive-question", "submit", "operations[1]", "Reactivate or satisfy this question's declared applicability before submitting or unsetting."))
	plan := history()
	plan.Sections[1].Enabled = flag(false)
	result := success(t, plan, []contract.Operation{{Kind: contract.Reactivate, SectionID: "repository-old"}, {Kind: contract.Submit, SectionID: "repository-old", Name: "resource_name", Value: text("Deliberately same resource")}})
	want := history()
	want.Answers.Sections["repository-old"]["resource_name"] = text("Deliberately same resource")
	if !reflect.DeepEqual(result, want) {
		t.Error("same-identity reactivation lost definition, baseline, dependent or held newcomer work")
	}
	for _, ops := range [][]contract.Operation{{{Kind: contract.Retire, SectionID: "repository-old"}, {Kind: contract.Reactivate, SectionID: "repository-old"}}, {{Kind: contract.Reactivate, SectionID: "repository-old"}, {Kind: contract.Retire, SectionID: "repository-old"}}} {
		result := success(t, history(), ops)
		want := history()
		if ops[1].Kind == contract.Retire {
			want.Sections[1].Enabled = flag(false)
		}
		if !reflect.DeepEqual(result, want) {
			t.Error("switch order was collapsed or cascaded")
		}
	}
}

func TestPrivateFailureHasUsefulFixedProjectionAndNoAcceptedPrefix(t *testing.T) {
	ops := []contract.Operation{{Kind: contract.Submit, SectionID: "repository-old", Name: "resource_name", Value: text("accepted only in candidate")}, {Kind: contract.Add, SectionID: "new-one", SourceID: "PRIVATE_REJECTED_SOURCE"}}
	result, diagnostics := newEditor().Edit(history(), ops)
	if !reflect.DeepEqual(result, model.Plan{}) || len(diagnostics) != 1 {
		t.Fatalf("atomic refusal missing: %#v %#v", result, diagnostics)
	}
	want := own("catalog.unknown-id", "add-source", "operations[1].instances[0].sourceId", "Choose an exact section-template ID from catalog list.")
	if !reflect.DeepEqual(diagnostics[0], want) {
		t.Errorf("fixed safe lookup projection differs: %#v", diagnostics)
	}
	data, err := json.Marshal(diagnostics)
	if err != nil || strings.Contains(string(data), "PRIVATE_REJECTED_SOURCE") {
		t.Error("rejected selector leaked")
	}
}
