package editor_test

import (
	"reflect"
	"testing"

	contract "onboard-kit/contracts/plan/editor"
	"onboard-kit/internal/model"
)

func TestRequiresSurvivesUnrelatedEditsAndRefusals(t *testing.T) {
	t.Parallel()
	for _, operation := range []struct {
		name    string
		refused bool
		apply   func(contract.Editor, model.Plan) (model.Plan, []model.Problem)
	}{
		{"submit", false, func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			return e.Submit(p, "plan.ready", model.Value{Kind: model.ValueKindFlag, Flag: true})
		}},
		{"refused-submit", true, func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			return e.Submit(p, "plan.missing", model.Value{Kind: model.ValueKindFlag, Flag: true})
		}},
		{"switch", false, func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			return e.SetSectionEnabled(p, "access", false)
		}},
		{"contact", false, func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			return e.SetContact(p, model.Contact{ID: "other", Name: "Other"})
		}},
		{"role", false, func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			return e.SetRole(p, model.Role{ID: "other", Name: "Other", Holders: []string{"help"}})
		}},
		{"team", false, func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			return e.SetTeam(p, model.Team{ID: "other", Name: "Other"})
		}},
		{"remove-contact", false, func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			out, _, problems := e.RemoveContact(p, "help")
			return out, problems
		}},
		{"remove-role", false, func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			out, _, problems := e.RemoveRole(p, "manager")
			return out, problems
		}},
		{"remove-team", false, func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			out, _, problems := e.RemoveTeam(p, "blue")
			return out, problems
		}},
		{"prune", false, func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			out, _, problems := e.PruneHiddenAnswers(p)
			return out, problems
		}},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()
			input := model.Plan{SchemaVersion: 1, Variables: []model.Variable{{Name: "ready", Type: model.VariableTypeBoolean}}, Directory: model.Directory{Contacts: []model.Contact{{ID: "help", Name: "Help"}}, Roles: []model.Role{{ID: "manager", Name: "Manager"}}, Teams: []model.Team{{ID: "blue", Name: "Blue"}}}, Sections: []model.Section{{ID: "access", Bootstrap: true, After: []string{"legacy"}, Requires: []string{"b", "a", "b"}}, {ID: "other"}, {ID: "empty", Requires: []string{}}}}
			got, problems := operation.apply(newEditor(), input)
			hasError := false
			for _, p := range problems {
				hasError = hasError || p.Severity == model.SeverityError
			}
			if hasError != operation.refused {
				t.Fatalf("wrong refusal state: %#v", problems)
			}
			if len(got.Sections) != 3 || !reflect.DeepEqual(got.Sections[0].Requires, []string{"b", "a", "b"}) || got.Sections[1].Requires != nil || !reflect.DeepEqual(got.Sections[2].Requires, []string{}) || !got.Sections[0].Bootstrap || !reflect.DeepEqual(got.Sections[0].After, []string{"legacy"}) {
				t.Fatalf("edit changed section definition: %#v", got.Sections)
			}
			if !reflect.DeepEqual(input.Sections[0].Requires, []string{"b", "a", "b"}) || input.Sections[0].Enabled != nil {
				t.Fatal("input changed")
			}
			if operation.refused && !reflect.DeepEqual(got, input) {
				t.Fatal("refusal changed plan")
			}
		})
	}
}
