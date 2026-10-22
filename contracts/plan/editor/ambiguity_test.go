package editor_test

import (
	"reflect"
	"testing"

	"onboard-kit/internal/model"
)

func TestAmbiguousSubmissionPreservesStoredAnswers(t *testing.T) {
	t.Parallel()
	for _, scope := range []string{"plan", "setup"} {
		t.Run(scope, func(t *testing.T) {
			t.Parallel()
			newPlan := func() model.Plan {
				variables := []model.Variable{{Name: "label"}, {Name: "label", Type: model.VariableTypeBoolean}}
				values := map[string]model.Value{"label": {Kind: model.ValueKindText, Text: "original"}}
				if scope == "plan" {
					return model.Plan{Variables: variables, Answers: model.Answers{Plan: values}}
				}
				return model.Plan{
					Sections: []model.Section{{ID: scope, Body: "$label", Variables: variables}},
					Answers:  model.Answers{Sections: map[string]map[string]model.Value{scope: values}},
				}
			}
			before, want := newPlan(), newPlan()
			got, problems := newEditor().Submit(before, scope+".label", model.Value{Kind: model.ValueKindText, Text: "replacement"})
			if len(problems) != 1 || problems[0].Code != "answer.ambiguous-question" || problems[0].Severity != model.SeverityError {
				t.Fatalf("ambiguous submission: got %+v", problems)
			}
			if !reflect.DeepEqual(before, want) || !reflect.DeepEqual(got, want) {
				t.Fatal("refused ambiguous submission changed the input or returned plan")
			}
		})
	}
}
