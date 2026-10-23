package editor_test

import (
	"testing"

	contract "onboard-kit/contracts/plan/editor"
	"onboard-kit/internal/model"
)

func TestBootstrapSurvivesUnrelatedEdits(t *testing.T) {
	t.Parallel()
	for _, operation := range []struct {
		name  string
		apply func(contract.Editor, model.Plan) (model.Plan, []model.Problem)
	}{
		{"submit", func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			return e.Submit(p, "plan.ready", model.Value{Kind: model.ValueKindFlag, Flag: true})
		}},
		{"switch", func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			return e.SetSectionEnabled(p, "access", false)
		}},
		{"contact", func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			return e.SetContact(p, model.Contact{ID: "other", Name: "Other"})
		}},
		{"role", func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			return e.SetRole(p, model.Role{ID: "other", Name: "Other", Holders: []string{"help"}})
		}},
		{"team", func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			return e.SetTeam(p, model.Team{ID: "other", Name: "Other"})
		}},
		{"remove-contact", func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			out, _, problems := e.RemoveContact(p, "help")
			return out, problems
		}},
		{"remove-role", func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			out, _, problems := e.RemoveRole(p, "manager")
			return out, problems
		}},
		{"remove-team", func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			out, _, problems := e.RemoveTeam(p, "blue")
			return out, problems
		}},
		{"prune", func(e contract.Editor, p model.Plan) (model.Plan, []model.Problem) {
			out, _, problems := e.PruneHiddenAnswers(p)
			return out, problems
		}},
	} {
		t.Run(operation.name, func(t *testing.T) {
			t.Parallel()
			input := model.Plan{SchemaVersion: 1, Variables: []model.Variable{{Name: "ready", Type: model.VariableTypeBoolean}}, Directory: model.Directory{Contacts: []model.Contact{{ID: "help", Name: "Help"}}, Roles: []model.Role{{ID: "manager", Name: "Manager"}}, Teams: []model.Team{{ID: "blue", Name: "Blue"}}}, Sections: []model.Section{{ID: "access", Title: "Access", Summary: "Request access", Bootstrap: true}, {ID: "other", Title: "Other", Summary: "Other work", Bootstrap: false}}}
			got, problems := operation.apply(newEditor(), input)
			for _, problem := range problems {
				if problem.Severity == model.SeverityError {
					t.Fatalf("unrelated edit failed: %#v", problems)
				}
			}
			if len(got.Sections) != 2 || !got.Sections[0].Bootstrap || got.Sections[1].Bootstrap {
				t.Fatalf("edit lost or manufactured marker: %#v", got.Sections)
			}
			got.Sections[0].Bootstrap = false
			got.Sections[1].Bootstrap = true
			if !input.Sections[0].Bootstrap || input.Sections[1].Bootstrap {
				t.Fatal("edit output aliases input section marker")
			}
		})
	}
}
