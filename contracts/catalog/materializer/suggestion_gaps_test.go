package materializer_test

import (
	"bytes"
	"reflect"
	"testing"

	api "onboard-kit/contracts/catalog/materializer"
	"onboard-kit/internal/model"
)

func exampleDeclarations(audience model.Audience) []model.Variable {
	return []model.Variable{
		{Name: "example_only", Example: "  illustrative shared fact Ω  ", Required: true, Audience: audience},
		{Name: "example_with_default", Example: "  illustrated choice  ", Default: model.Value{Kind: model.ValueKindText, Text: "suggested choice"}, Required: true, Audience: audience},
	}
}

func TestPlanExamplesRemainUnacceptedWithoutBindings(t *testing.T) {
	for _, audience := range []model.Audience{"", model.AudienceOnboarder} {
		for _, selection := range []string{"empty", "repository"} {
			t.Run(string(audience)+"/"+selection, func(t *testing.T) {
				r := base()
				r.PlanVariables = exampleDeclarations(audience)
				if selection == "repository" {
					r.Instances = []api.Instance{{SourceID: "repository-access", PermanentSectionID: "repo-a"}}
				}
				before := snapshot(t, r)
				got, ds := newMaterializer().Create(r)
				noErrors(t, ds)
				want := expectedPlan(r)
				want.Variables = exampleDeclarations(audience)
				if selection == "repository" {
					want.Sections = []model.Section{expectedRepository(t, "repo-a")}
				}
				if !reflect.DeepEqual(got.Plan, want) {
					t.Error("unbound examples/defaults became answers or the copied declarations/new-plan shape changed")
				}
				if selection == "repository" {
					guides(t, got.Guidance, []guideOracle{{"invitation-prepare", []string{"repo-a"}}, {"resource-access-arrange", []string{"repo-a"}}})
				} else if !reflect.DeepEqual(got.Guidance, []api.Guidance{}) {
					t.Error("empty selection invented guidance or changed its allocated-empty shape")
				}
				if !bytes.Equal(snapshot(t, r), before) {
					t.Error("creation changed the held example/default declarations or request")
				}
			})
		}
	}
}

func TestExplicitPlanBindingsRemainDistinctFromExamplesAndDefaults(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value model.Value
	}{
		{"accept-example", model.Value{Kind: model.ValueKindText, Text: "  illustrated choice  "}},
		{"accept-default", model.Value{Kind: model.ValueKindText, Text: "suggested choice"}},
		{"different-answer", model.Value{Kind: model.ValueKindText, Text: "accepted different choice"}},
		{"explicit-empty-text", model.Value{Kind: model.ValueKindText, Text: ""}},
		{"explicit-none", model.Value{Kind: model.ValueKindNone, Text: "held inactive text", List: []string{"held inactive item"}}},
		{"explicit-deferred", model.Value{Kind: model.ValueKindDeferred}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := base()
			r.PlanVariables = exampleDeclarations(model.AudienceOnboarder)
			r.PlanBindings = []api.Binding{{Name: "example_with_default", Value: tc.value}}
			before := snapshot(t, r)
			got, ds := newMaterializer().Create(r)
			noErrors(t, ds)
			want := expectedPlan(r)
			want.Variables = exampleDeclarations(model.AudienceOnboarder)
			want.Answers.Plan["example_with_default"] = tc.value
			if !reflect.DeepEqual(got, api.Result{Plan: want, Guidance: []api.Guidance{}}) {
				t.Error("only the exact explicit binding should become an answer; companion examples remain unaccepted")
			}
			if !bytes.Equal(snapshot(t, r), before) {
				t.Error("creation changed the explicit binding or held suggestions")
			}
		})
	}
}
