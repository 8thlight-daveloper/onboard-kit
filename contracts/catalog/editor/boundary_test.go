package editor_test

import (
	"reflect"
	"testing"

	contract "onboard-kit/contracts/catalog/editor"
	"onboard-kit/internal/model"
)

func unambiguousHistory() model.Plan {
	plan := history()
	delete(plan.Answers.Sections, "departed")
	return plan
}

func TestAmbiguousHeldDefinitionsRefuseBeforeChoosingTargets(t *testing.T) {
	for _, kind := range []string{"section", "variable", "step"} {
		plan := unambiguousHistory()
		plan.Answers = model.Answers{}
		code := ""
		message := ""
		switch kind {
		case "section":
			plan.Sections = append(plan.Sections, plan.Sections[1])
			code = "plan.duplicate-section-id"
			message = "Repair repeated held section IDs before editing."
		case "variable":
			plan.Sections[1].Variables = append(plan.Sections[1].Variables, plan.Sections[1].Variables[0])
			code = "plan.duplicate-variable-name"
			message = "Repair repeated variable declarations within their scope before editing."
		case "step":
			plan.Sections[1].Steps = append(plan.Sections[1].Steps, plan.Sections[1].Steps[0])
			code = "plan.duplicate-step-id"
			message = "Repair repeated step IDs within their section before editing."
		}
		refusal(t, plan, []contract.Operation{{Kind: contract.Retire, SectionID: "repository-old"}}, own(code, "input-plan", "plan", message))
	}
}

func TestReactivationRunsRealWholePlanValidationAndCanBeExplicitlyRepairedInBatch(t *testing.T) {
	plan := unambiguousHistory()
	plan.Sections[1].Enabled = flag(false)
	plan.Answers.Sections["repository-old"]["resource_name"] = model.Value{Kind: model.ValueKindFlag, Flag: false}
	refusal(t, plan, []contract.Operation{{Kind: contract.Reactivate, SectionID: "repository-old"}}, own("answer.wrong-kind", "final-plan", "plan", "Submit a value kind matching the declared variable type."))
	result := success(t, plan, []contract.Operation{{Kind: contract.Reactivate, SectionID: "repository-old"}, {Kind: contract.Submit, SectionID: "repository-old", Name: "resource_name", Value: text("Explicitly repaired shared resource")}})
	want := unambiguousHistory()
	want.Answers.Sections["repository-old"]["resource_name"] = text("Explicitly repaired shared resource")
	if !reflect.DeepEqual(result, want) {
		t.Error("ordered repair did not preserve same identity and unrelated state")
	}
}

func TestNewExactEmptyBindingDiffersFromExistingEmptySubmission(t *testing.T) {
	add := contract.Operation{Kind: contract.Add, SectionID: "repository-new", SourceID: "repository-access", Bindings: []contract.Binding{{Name: "repository_scope", Value: text("")}}}
	plan := history()
	result := success(t, plan, []contract.Operation{add})
	want := history()
	want.Sections = append(want.Sections, baseline(t, "repository-new", "repository-access"))
	want.Answers.Sections["repository-new"] = map[string]model.Value{"repository_scope": text("")}
	if !reflect.DeepEqual(result, want) {
		t.Error("exact new materializer binding was silently converted to saved-plan Submit semantics")
	}
	result = success(t, plan, []contract.Operation{add, {Kind: contract.Submit, SectionID: "repository-new", Name: "repository_scope", Value: text("")}})
	delete(want.Answers.Sections, "repository-new")
	if !reflect.DeepEqual(result, want) {
		t.Error("existing empty submission did not clear its key and newly emptied scope")
	}
}

func TestExplicitFalseAndOrderedListSubmissionRemainFacts(t *testing.T) {
	for _, value := range []model.Value{{Kind: model.ValueKindFlag, Flag: false}, {Kind: model.ValueKindList, List: []string{"first", "", "first"}}} {
		plan := history()
		name := "shared_flag"
		kind := model.VariableTypeBoolean
		if value.Kind == model.ValueKindList {
			name = "shared_command"
			kind = model.VariableTypeCommand
		}
		plan.Sections[1].Variables = append(plan.Sections[1].Variables, model.Variable{Name: name, Description: "Known shared fact", Type: kind, Default: model.Value{Kind: model.ValueKindDeferred}})
		result := success(t, plan, []contract.Operation{{Kind: contract.Submit, SectionID: "repository-old", Name: name, Value: value}})
		plan.Answers.Sections["repository-old"][name] = value
		if !reflect.DeepEqual(result, plan) {
			t.Error("false/list was dropped, defaulted, reordered or propagated")
		}
	}
}

func TestExplicitOperationsDoNotReachForwardOrAcceptUnusedPayloads(t *testing.T) {
	refusal(t, history(), []contract.Operation{{Kind: contract.Submit, SectionID: "future", Name: "resource_name", Value: text("not held yet")}, {Kind: contract.Add, SectionID: "future", SourceID: "repository-access"}}, own("edit.section", "operation", "operations[0].sectionId", "Select an exact section already held at this operation."))
	for _, op := range []contract.Operation{
		{Kind: "delete", SectionID: "repository-old"},
		{Kind: contract.Retire, SectionID: "repository-old", Name: "ignored"},
		{Kind: contract.Unset, SectionID: "repository-old", Name: "resource_name", Value: model.Value{Text: "ignored payload"}},
		{Kind: contract.Add, SectionID: "new-one", SourceID: "working-hours", Value: model.Value{Flag: true}},
	} {
		refusal(t, history(), []contract.Operation{op}, own("edit.operation", "operation", "operations[0]", "Use exactly the fields of one supported explicit edit operation."))
	}
	refusal(t, history(), []contract.Operation{{Kind: contract.Submit, SectionID: "repository-old", Name: "resource_name", Value: model.Value{Kind: 99}}}, own("edit.value-kind", "operation", "operations[0].value", "Choose a supported submission value kind."))
}

func TestCallerAdmissionIsBoundedBeforeExpansion(t *testing.T) {
	plan := history()
	ops := make([]contract.Operation, 1025)
	refusal(t, plan, ops, own("edit.limit", "admission", "request", "Reduce held plan or edit data to the supported limits."))
	plan = history()
	plan.UnknownFields["private-key"] = []byte("private raw unsupported object")
	refusal(t, plan, nil, own("edit.held-shape", "admission", "plan", "Supply ordinary decoded metadata values without unsupported objects."))
	plan = history()
	plan.UnknownFields["invalid"] = string([]byte{0xff})
	refusal(t, plan, nil, own("edit.text", "admission", "request", "Use valid UTF8 held plan and edit data."))
	plan = history()
	cycle := map[string]any{}
	cycle["again"] = cycle
	plan.UnknownFields["cycle"] = cycle
	refusal(t, plan, nil, own("edit.limit", "admission", "request", "Reduce held plan or edit data to the supported limits."))
}

func TestLaterFailureCannotCommitEarlierSharedFactOrSwitch(t *testing.T) {
	refusal(t, history(), []contract.Operation{{Kind: contract.Submit, SectionID: "repository-old", Name: "resource_name", Value: text("new candidate fact")}, {Kind: contract.Retire, SectionID: "repository-old"}, {Kind: contract.Reactivate, SectionID: "UNKNOWN_PRIVATE_TARGET"}}, own("edit.section", "operation", "operations[2].sectionId", "Select an exact section already held at this operation."))
}

func TestNewPermanentSlugValidationAndCarrierChecksStayDelegated(t *testing.T) {
	for _, id := range []string{"plan", "true", "Bad-ID", "bad space", "-first", "ends\n"} {
		refusal(t, history(), []contract.Operation{{Kind: contract.Add, SectionID: id, SourceID: "repository-access"}}, own("materialize.instance-id", "add-request", "operations[0].instances[0].permanentSectionId", "Supply a nonreserved lowercase permanent section ID."))
	}
	refusal(t, history(), []contract.Operation{{Kind: contract.Add, SectionID: "not-an-activity", SourceID: "invitation-prepare"}}, own("materialize.carrier", "add-source", "operations[0].instances[0].sourceId", "Select a section-template source for an activity."))
}

func TestEmptyHeldShapeAndUnrelatedPlanFactsAreNotReplacedByAddSubproposal(t *testing.T) {
	plan := model.Plan{SchemaVersion: 1, ID: "minimal held identity"}
	result := success(t, plan, nil)
	if !reflect.DeepEqual(result, plan) {
		t.Error("no-op allocated absent containers")
	}
	result = success(t, plan, []contract.Operation{{Kind: contract.Add, SectionID: "document-new", SourceID: "document-reading"}})
	want := model.Plan{SchemaVersion: 1, ID: "minimal held identity", Sections: []model.Section{baseline(t, "document-new", "document-reading")}}
	if !reflect.DeepEqual(result, want) {
		t.Errorf("isolated materializer proposal overwrote held root shape: %#v", result)
	}
}

func TestInactiveUnsetAndUnknownBindingsCannotAlterHeldFacts(t *testing.T) {
	plan := history()
	plan.Sections[1].Enabled = flag(false)
	refusal(t, plan, []contract.Operation{{Kind: contract.Unset, SectionID: "repository-old", Name: "resource_name"}}, own("answer.inactive-question", "submit", "operations[0]", "Reactivate or satisfy this question's declared applicability before submitting or unsetting."))
	refusal(t, history(), []contract.Operation{{Kind: contract.Submit, SectionID: "repository-old", Name: "PRIVATE_UNKNOWN_NAME", Value: text("PRIVATE_REJECTED_VALUE")}}, own("edit.binding", "operation", "operations[0].name", "Select an exact local variable declared in this section."))
	refusal(t, history(), []contract.Operation{{Kind: contract.Submit, SectionID: "plan", Name: "shared_fact", Value: text("not authorized here")}}, own("edit.section", "operation", "operations[0].sectionId", "Select an exact section already held at this operation."))
}

func TestInvalidInactiveRelationshipsCannotBeLaunderedByRetirement(t *testing.T) {
	plan := unambiguousHistory()
	plan.Sections[3].Requires = []string{"PRIVATE_MISSING_TARGET"}
	refusal(t, plan, []contract.Operation{{Kind: contract.Retire, SectionID: "retired"}}, own("plan.requires-unknown", "input-plan", "plan", "Repair the plan's reported structure or stored facts before editing."))
}

func TestMissingHeldPlanIdentityIsExplicitRefusal(t *testing.T) {
	refusal(t, model.Plan{}, nil, own("edit.plan-id", "admission", "plan.id", "Select an existing plan with an explicit identity."))
}
