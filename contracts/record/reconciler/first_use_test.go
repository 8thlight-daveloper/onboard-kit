package reconciler_test

import (
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestFirstOpenedAtSurvivesReconciliationAndRefusal(t *testing.T) {
	for _, held := range []model.Timestamp{"", "2026-10-23T12:34:56Z", " held\nnot a time\u2028\u2029 "} {
		for _, id := range []string{"same", "other", ""} {
			t.Run(string(held)+"/"+id, func(t *testing.T) {
				record := model.Record{SchemaVersion: 1, StartDate: "legacy date", FirstOpenedAt: held, Plan: model.PlanReference{ID: "same", Hash: "old"}, Answers: model.Answers{Plan: map[string]model.Value{"q": {Kind: model.ValueKindText, Text: "gone"}, "hidden": {Kind: model.ValueKindText, Text: "kept"}}}}
				plan := model.Plan{ID: id, Variables: []model.Variable{{Name: "q", Audience: model.AudienceOnboardee, Type: model.VariableTypeChoice, Options: []string{"kept"}}}}
				got, report, problems := newReconciler().Reconcile(record, plan, plan, model.PlanReference{Hash: "new"})
				want := record
				if id == "same" {
					want.Plan = model.PlanReference{ID: "same", Hash: "new"}
					want.Answers.Plan = map[string]model.Value{"hidden": {Kind: model.ValueKindText, Text: "kept"}}
					if len(problems) != 0 || !reflect.DeepEqual(report.AnswersToReview, []string{"plan.q"}) {
						t.Fatalf("want removed answer listed, got %+v and %+v", report, problems)
					}
				} else {
					code := "reconcile.different-plan"
					if id == "" {
						code = "reconcile.missing-plan-id"
					}
					if len(problems) != 1 || problems[0].Code != code {
						t.Fatalf("want %s, got %+v", code, problems)
					}
				}
				if !reflect.DeepEqual(got, want) || record.FirstOpenedAt != held {
					t.Fatalf("got %#v, want %#v; input %#v", got, want, record)
				}
			})
		}
	}
}
