package validator_test

import (
	"slices"
	"testing"

	"onboard-kit/internal/model"
)

func TestDuplicateDeclarationsHaveNoWinner(t *testing.T) {
	t.Parallel()
	secret := "ghp_" + "FAKEFAKEFAKEFAKEFAKEFAKEFAKEFAKEFAKE"
	hidden := &model.Condition{All: []model.Clause{{Question: "plan.missing", Answer: model.AnswerTest{Is: model.Value{Kind: model.ValueKindText, Text: "yes"}}}}}
	off := false
	for _, scope := range []string{"plan", "main", "off", "hidden", "hidden-question", "child"} {
		for _, identical := range []bool{false, true} {
			for _, reverse := range []bool{false, true} {
				name := scope
				if identical {
					name += "/identical"
				}
				if reverse {
					name += "/reversed"
				}
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					variables := []model.Variable{
						{Name: "owner", Type: model.VariableTypeBoolean},
						{Name: "owner", Type: model.VariableTypeChoice, Options: []string{"known"}, Audience: model.AudienceOnboardee},
						{Name: "owner", Type: model.VariableTypeContact},
						{Name: "owner", Type: model.VariableTypeString, Pattern: "^known$"},
					}
					if identical {
						for i := range variables {
							variables[i] = variables[0]
						}
					}
					if reverse {
						slices.Reverse(variables)
					}
					if scope == "hidden-question" {
						for i := range variables {
							variables[i].AskIf = hidden
						}
					}
					plan := model.Plan{Answers: model.Answers{Plan: map[string]model.Value{}, Sections: map[string]map[string]model.Value{}}}
					section := ""
					if scope == "plan" {
						plan.Variables = variables
						plan.Answers.Plan["owner"] = model.Value{Kind: model.ValueKindText, Text: secret}
					} else {
						section = scope
						entry := model.Section{ID: scope, Variables: variables}
						switch scope {
						case "off":
							entry.Enabled = &off
						case "hidden":
							entry.IncludeIf = hidden
						case "child":
							entry.Parent = "parent"
							plan.Sections = append(plan.Sections, model.Section{ID: "parent"})
						}
						plan.Sections = append(plan.Sections, entry)
						plan.Answers.Sections[scope] = map[string]model.Value{"owner": {Kind: model.ValueKindText, Text: secret}}
					}
					question := scope + ".owner"
					want := []expectedProblem{
						{Code: "plan.duplicate-variable-name", Severity: "error", Section: section, Detail: map[string]string{"scope": scope, "name": "owner"}},
						{Code: "secret.token-prefix", Severity: "error", Question: question, Section: section},
					}
					assertProblems(t, newValidator().Validate(plan), want, []string{"FAKEFAKE"})
					for _, value := range []model.Value{{Kind: model.ValueKindText, Text: secret}, {Kind: model.ValueKindFlag, Flag: true}, {Kind: model.ValueKindDeferred}, {}} {
						assertProblems(t, newValidator().CheckValue(plan, question, value), []expectedProblem{{Code: "answer.ambiguous-question", Severity: "error", Question: question, Section: section}}, []string{"FAKEFAKE"})
					}
				})
			}
		}
	}
}

func TestStepIdentityNamespaces(t *testing.T) {
	t.Parallel()
	off := false
	plan := model.Plan{
		Variables: []model.Variable{{Name: "shared", Type: model.VariableTypeChoice, Options: []string{"same", "same"}}},
		Answers:   model.Answers{Plan: map[string]model.Value{"shared": {Kind: model.ValueKindText, Text: "same"}}},
		Sections: []model.Section{
			{ID: "zulu", Steps: []model.Step{{ID: "repeat"}, {ID: "repeat"}}, Variables: []model.Variable{{Name: "shared"}}},
			{ID: "alpha", Enabled: &off, Steps: []model.Step{{ID: "zebra"}, {ID: "apple"}, {ID: "zebra"}, {ID: "apple"}, {ID: "apple"}, {}, {ID: ""}}},
			{ID: "control", Steps: []model.Step{{ID: "repeat", Run: []string{"echo", "same", "same"}}, {}, {}}, Variables: []model.Variable{{Name: "shared"}}},
		},
	}
	want := []expectedProblem{
		{Code: "plan.duplicate-step-id", Severity: "error", Section: "alpha", Detail: map[string]string{"step": "apple"}},
		{Code: "plan.duplicate-step-id", Severity: "error", Section: "alpha", Detail: map[string]string{"step": "zebra"}},
		{Code: "plan.duplicate-step-id", Severity: "error", Section: "zulu", Detail: map[string]string{"step": "repeat"}},
	}
	assertProblems(t, newValidator().Validate(plan), want, nil)
	for i := range plan.Sections {
		slices.Reverse(plan.Sections[i].Steps)
	}
	slices.Reverse(plan.Sections)
	assertProblems(t, newValidator().Validate(plan), want, nil)
	assertProblems(t, newValidator().CheckValue(plan, "plan.shared", model.Value{Kind: model.ValueKindText, Text: "same"}), nil, nil)
	assertProblems(t, newValidator().CheckValue(plan, "control.shared", model.Value{Kind: model.ValueKindText, Text: "anything"}), nil, nil)
}

func TestDuplicateVariableNamesSortWithinScopes(t *testing.T) {
	t.Parallel()
	plan := model.Plan{
		Variables: []model.Variable{{Name: "zebra"}, {Name: "apple"}, {Name: "zebra"}, {Name: "apple"}},
		Sections: []model.Section{
			{ID: "zulu", Variables: []model.Variable{{Name: "shared"}, {Name: "shared"}}},
			{ID: "alpha", Variables: []model.Variable{{Name: "zebra"}, {Name: "apple"}, {Name: "apple"}, {Name: "zebra"}}},
		},
	}
	want := []expectedProblem{
		{Code: "plan.duplicate-variable-name", Severity: "error", Section: "alpha", Detail: map[string]string{"scope": "alpha", "name": "apple"}},
		{Code: "plan.duplicate-variable-name", Severity: "error", Section: "alpha", Detail: map[string]string{"scope": "alpha", "name": "zebra"}},
		{Code: "plan.duplicate-variable-name", Severity: "error", Detail: map[string]string{"scope": "plan", "name": "apple"}},
		{Code: "plan.duplicate-variable-name", Severity: "error", Detail: map[string]string{"scope": "plan", "name": "zebra"}},
		{Code: "plan.duplicate-variable-name", Severity: "error", Section: "zulu", Detail: map[string]string{"scope": "zulu", "name": "shared"}},
	}
	assertProblems(t, newValidator().Validate(plan), want, nil)
	slices.Reverse(plan.Variables)
	for i := range plan.Sections {
		slices.Reverse(plan.Sections[i].Variables)
	}
	slices.Reverse(plan.Sections)
	assertProblems(t, newValidator().Validate(plan), want, nil)
}
