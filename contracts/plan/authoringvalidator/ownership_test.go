package authoringvalidator_test

import (
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestDeepShapesAndCallerHeldReferences(t *testing.T) {
	plan := deepOwnershipFixture()
	before := deepOwnershipFixture()
	heldEnabled := plan.Sections[0].Enabled
	heldCondition := plan.Sections[0].Variables[0].AskIf
	heldSettled := heldCondition.All[0].Answer.IsSettled
	heldSource := plan.Sections[0].Source
	heldEnv := plan.Sections[0].Steps[0].Env
	heldHolders := plan.Directory.Teams[0].Holders
	heldAnswers := plan.Answers.Plan
	heldUnknown := plan.UnknownFields
	heldDefaultItem := &plan.Variables[0].Default.List[0]
	heldRunItem := &plan.Sections[0].Steps[0].Run[0]
	want := []expectation{{Code: "authoring.pattern", Path: "variables[1].pattern", Question: "plan.invalid_pattern"}}
	validator := newValidator()
	first := validator.Validate(plan)
	second := validator.Validate(plan)
	assertProblems(t, first, want)
	assertProblems(t, second, want)
	if !reflect.DeepEqual(plan, before) {
		t.Fatal("validation changed deep values, concrete unknown-data types or nil/empty shapes")
	}
	if plan.Sections[0].Enabled != heldEnabled || plan.Sections[0].Variables[0].AskIf != heldCondition || heldCondition.All[0].Answer.IsSettled != heldSettled || plan.Sections[0].Source != heldSource {
		t.Fatal("validation replaced a caller-held pointer")
	}
	if &plan.Variables[0].Default.List[0] != heldDefaultItem || &plan.Sections[0].Steps[0].Run[0] != heldRunItem {
		t.Fatal("validation replaced a caller-held backing list")
	}
	if len(first) == 1 && first[0].Detail != nil {
		first[0].Detail["path"] = "changed"
		first[0].Detail["extra"] = "changed"
	}
	assertProblems(t, second, want)
	assertProblems(t, validator.Validate(plan), want)
	if !reflect.DeepEqual(plan, before) {
		t.Fatal("returned diagnostics share mutable data with the input")
	}
	heldEnv["OUTSIDE"] = "caller edit"
	heldHolders["outside"] = []string{"caller edit"}
	heldAnswers["outside"] = model.Value{Kind: model.ValueKindFlag, Flag: true}
	heldUnknown["outside"] = "caller edit"
	if plan.Sections[0].Steps[0].Env["OUTSIDE"] != "caller edit" || !reflect.DeepEqual(plan.Directory.Teams[0].Holders["outside"], []string{"caller edit"}) || plan.Answers.Plan["outside"].Kind != model.ValueKindFlag || plan.UnknownFields["outside"] != "caller edit" {
		t.Fatal("validation detached caller-held maps from the input graph")
	}
}

func deepOwnershipFixture() model.Plan {
	enabled := false
	settled := false
	return model.Plan{
		Variables: []model.Variable{
			{Name: "contact_list", Type: model.VariableTypeContact, Multiple: true, Default: model.Value{Kind: model.ValueKindList, List: []string{"help", "help"}}},
			{Name: "invalid_pattern", Pattern: "[", Options: []string{}},
		},
		Directory: model.Directory{
			Contacts: []model.Contact{{ID: "help", Name: "Help"}},
			Roles:    []model.Role{{ID: "helper", Name: "Helper", Holders: []string{"help"}}, {ID: "empty-role", Name: "Empty", Holders: []string{}}},
			Teams:    []model.Team{{ID: "team-a", Name: "Team A", Holders: map[string][]string{"helper": {"help"}, "empty-role": {}}}},
		},
		Answers: model.Answers{
			Plan:     map[string]model.Value{"hidden": {Kind: model.ValueKindList, List: []string{}}},
			Sections: map[string]map[string]model.Value{"access": {}, "hidden": {"nil_list": {Kind: model.ValueKindList}}},
		},
		Sections: []model.Section{{
			ID: "access", Title: "Access", Summary: "Ask $helper", Enabled: &enabled,
			Source: &model.Source{Template: "held", Hash: "held"},
			After:  []string{}, Requires: nil, Escalation: []string{"$helper"},
			Variables:     []model.Variable{{Name: "helper", Type: model.VariableTypeContact, AskIf: &model.Condition{All: []model.Clause{{Question: "plan.invalid_pattern", Answer: model.AnswerTest{IsSettled: &settled}}}}}},
			Steps:         []model.Step{{Run: []string{"echo", ""}, Env: map[string]string{"APP_ENV": ""}, Escalation: []string{}, Checks: []model.Check{{Run: []string{}, Ask: "Continue?"}, {Env: "TOKEN"}}}},
			UnknownFields: map[string]any{"nil_map": map[string]string(nil), "empty_map": map[string]string{}},
		}},
		UnknownFields: map[string]any{
			"integer": int(7), "floating": float64(7), "nil_list": []string(nil), "empty_list": []string{},
			"typed_map": map[string][]int{"nil": nil, "empty": {}},
		},
	}
}
