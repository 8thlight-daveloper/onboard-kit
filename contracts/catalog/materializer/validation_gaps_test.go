package materializer_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	api "onboard-kit/contracts/catalog/materializer"
	"onboard-kit/internal/model"
)

func TestDigitStartedPermanentSectionIDRemainsExplicitAndValid(t *testing.T) {
	r := base()
	r.Instances = []api.Instance{{SourceID: "repository-access", PermanentSectionID: "1repo"}}
	before := snapshot(t, r)
	got, ds := newMaterializer().Create(r)
	noErrors(t, ds)
	canonical := fixture(t, "repository-a.canonical")
	if bytes.Count(canonical, []byte(`"id":"repo-a"`)) != 1 || len(canonical) == 0 || canonical[len(canonical)-1] != '\n' {
		t.Fatal("independent canonical fixture no longer isolates the permanent ID and framing LF")
	}
	canonical = bytes.Replace(canonical[:len(canonical)-1], []byte(`"id":"repo-a"`), []byte(`"id":"1repo"`), 1)
	sum := sha256.Sum256(canonical)
	section := expectedRepository(t, "repo-a")
	section.ID = "1repo"
	section.Source.Hash = hex.EncodeToString(sum[:])
	want := expectedPlan(r)
	want.Sections = []model.Section{section}
	if !reflect.DeepEqual(got.Plan, want) {
		t.Error("a valid digit-started permanent ID was rejected, renamed or omitted from its unresolved baseline")
	}
	guides(t, got.Guidance, []guideOracle{{"invitation-prepare", []string{"1repo"}}, {"resource-access-arrange", []string{"1repo"}}})
	if !bytes.Equal(snapshot(t, r), before) {
		t.Error("creation changed the selected permanent identity")
	}
}

func TestAuthoringRefusalPreservesTheCompleteOrderedBatch(t *testing.T) {
	r := base()
	r.PlanVariables = []model.Variable{
		{Name: "shared_fact", Type: "invented", Audience: "unexpected"},
		{Name: "ab"},
	}
	before := snapshot(t, r)
	got, ds := newMaterializer().Create(r)
	want := []api.Diagnostic{
		{Stage: "authoring", Problem: model.Problem{Code: "authoring.variable-audience", Severity: model.SeverityError, Message: "Choose onboarder or onboardee audience.", Question: "plan.shared_fact", Detail: map[string]string{"path": "variables[0].audience"}}},
		{Stage: "authoring", Problem: model.Problem{Code: "authoring.variable-type", Severity: model.SeverityError, Message: "Choose a supported variable type.", Question: "plan.shared_fact", Detail: map[string]string{"path": "variables[0].type"}}},
		{Stage: "authoring", Problem: model.Problem{Code: "authoring.variable-name", Severity: model.SeverityError, Message: "Use a nonreserved variable name with at least three characters.", Detail: map[string]string{"path": "variables[1].name"}}},
	}
	if !reflect.DeepEqual(got, api.Result{}) || !reflect.DeepEqual(ds, want) {
		t.Error("authoring refusal truncated, reordered or rewrote its complete independent three-error batch")
	}
	if !bytes.Equal(snapshot(t, r), before) {
		t.Error("refused creation changed declaration data")
	}
}

func TestBindingRefusalPreservesEveryMissingContactItem(t *testing.T) {
	r := base()
	r.PlanVariables = []model.Variable{{Name: "helper_contacts", Type: model.VariableTypeContact, Multiple: true, Audience: model.AudienceOnboarder}}
	r.PlanBindings = []api.Binding{{Name: "helper_contacts", Value: model.Value{Kind: model.ValueKindList, List: []string{"missing-first", "", "missing-second"}}}}
	before := snapshot(t, r)
	got, ds := newMaterializer().Create(r)
	want := []api.Diagnostic{
		{Stage: "answer", Path: "planBindings[0].value", Problem: model.Problem{Code: "directory.unknown-id", Severity: model.SeverityError, Question: "plan.helper_contacts", Detail: map[string]string{"kind": "contact", "item": "0"}}},
		{Stage: "answer", Path: "planBindings[0].value", Problem: model.Problem{Code: "directory.unknown-id", Severity: model.SeverityError, Question: "plan.helper_contacts", Detail: map[string]string{"kind": "contact", "item": "2"}}},
	}
	projection := make([]api.Diagnostic, len(ds))
	for i, d := range ds {
		if d.Problem.Message == "" || strings.Contains(d.Problem.Message, "missing-first") || strings.Contains(d.Problem.Message, "missing-second") {
			t.Error("actual answer diagnostic needs useful wording without copying the rejected contact ID")
		}
		projection[i] = d
		projection[i].Problem.Message = ""
	}
	if !reflect.DeepEqual(got, api.Result{}) || !reflect.DeepEqual(projection, want) {
		t.Error("binding refusal truncated or reordered the two actual missing-contact errors")
	}
	if !bytes.Equal(snapshot(t, r), before) {
		t.Error("refused creation changed the explicit contact binding")
	}
}
