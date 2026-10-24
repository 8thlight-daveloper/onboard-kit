package materializer_test

import (
	"bytes"
	"reflect"
	"testing"

	api "onboard-kit/contracts/catalog/materializer"
	"onboard-kit/internal/model"
)

func nestedEmptyRequest() api.Request {
	r := base()
	r.Directory.Roles = []model.Role{{ID: "helpers", Name: "Helpers", Holders: []string{}}}
	r.Directory.Teams = []model.Team{
		{ID: "team-a", Name: "Team A", Holders: map[string][]string{"helpers": []string{}}},
		{ID: "team-empty", Name: "Empty team", Holders: map[string][]string{}},
	}
	r.PlanVariables = []model.Variable{
		{Name: "command_hint", Type: model.VariableTypeCommand, Options: []string{}, Default: model.Value{Kind: model.ValueKindList, List: []string{}}},
		{Name: "explicit_empty", Options: []string{}},
	}
	r.PlanBindings = []api.Binding{{Name: "explicit_empty", Value: model.Value{Kind: model.ValueKindNone, List: []string{}}}}
	return r
}

func TestNestedAllocatedEmptyCollectionsKeepTheirHeldShapes(t *testing.T) {
	r := nestedEmptyRequest()
	before := snapshot(t, r)
	got, ds := newMaterializer().Create(r)
	noErrors(t, ds)
	oracle := nestedEmptyRequest()
	want := expectedPlan(oracle)
	want.Directory.Roles = oracle.Directory.Roles
	want.Directory.Teams = oracle.Directory.Teams
	want.Variables = oracle.PlanVariables
	want.Answers.Plan["explicit_empty"] = model.Value{Kind: model.ValueKindNone, List: []string{}}
	if !reflect.DeepEqual(got, api.Result{Plan: want, Guidance: []api.Guidance{}}) {
		t.Error("a held allocated-empty nested list/map became nil or a new outer container changed shape")
	}
	if !bytes.Equal(snapshot(t, r), before) {
		t.Error("creation normalized a caller-held nested collection")
	}
	if len(got.Plan.Directory.Teams) != 2 || got.Plan.Directory.Teams[1].Holders == nil {
		t.Fatal("allocated-empty team holders map missing")
	}
	got.Plan.Directory.Teams[1].Holders["result-marker"] = []string{"result edit"}
	if len(r.Directory.Teams[1].Holders) != 0 {
		t.Error("returned empty team map aliases the caller's map")
	}
	r.Directory.Teams[1].Holders["caller-marker"] = []string{"caller edit"}
	if _, found := got.Plan.Directory.Teams[1].Holders["caller-marker"]; found {
		t.Error("caller edit reached the returned team map")
	}
}

func defaultListRequest() api.Request {
	r := base()
	r.PlanVariables = []model.Variable{{Name: "command_hint", Type: model.VariableTypeCommand, Default: model.Value{Kind: model.ValueKindList, List: []string{"original-one", "original-two"}}}}
	return r
}

func TestVariableDefaultListsDoNotAliasCallerOrOtherResults(t *testing.T) {
	r := defaultListRequest()
	m := newMaterializer()
	first, ds := m.Create(r)
	noErrors(t, ds)
	second, ds := m.Create(r)
	noErrors(t, ds)
	oracle := defaultListRequest()
	want := expectedPlan(oracle)
	want.Variables = oracle.PlanVariables
	expected := api.Result{Plan: want, Guidance: []api.Guidance{}}
	if !reflect.DeepEqual(first, expected) || !reflect.DeepEqual(second, expected) {
		t.Fatal("nonempty literal default-list controls or unanswered new-plan shape differ")
	}
	first.Plan.Variables[0].Default.List[0] = "result edit"
	if !reflect.DeepEqual(r.PlanVariables[0].Default.List, []string{"original-one", "original-two"}) || !reflect.DeepEqual(second, expected) {
		t.Error("returned default-list edit reached the caller or another result")
	}
	r.PlanVariables[0].Default.List[1] = "caller edit"
	if !reflect.DeepEqual(first.Plan.Variables[0].Default.List, []string{"result edit", "original-two"}) || !reflect.DeepEqual(second, expected) {
		t.Error("caller default-list edit reached an earlier result")
	}
	third, ds := m.Create(r)
	noErrors(t, ds)
	thirdWant := expectedPlan(r)
	thirdWant.Variables = []model.Variable{{Name: "command_hint", Type: model.VariableTypeCommand, Default: model.Value{Kind: model.ValueKindList, List: []string{"original-one", "caller edit"}}}}
	if !reflect.DeepEqual(third, api.Result{Plan: thirdWant, Guidance: []api.Guidance{}}) {
		t.Error("later call retained another result's edit or lost the current request default")
	}
}
