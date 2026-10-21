package assembler

import (
	"slices"

	"onboard-kit/internal/model"
)

const (
	codeRoleVacant = "directory.role-vacant"
	vacantSuffix   = " (vacant)"
)

func (a Assembler) vacantRoleProblems(section model.Section, context planContext) []model.Problem {
	var problems []model.Problem
	for _, variable := range section.Variables {
		if variable.Type != model.VariableTypeRole {
			continue
		}
		if a.namesVacantRole(context.answers.Sections[section.ID][variable.Name], context) && a.readsVariable(section, variable.Name, context) {
			problems = append(problems, vacantProblem(section.ID, variable.Name))
		}
	}
	return problems
}

func vacantProblem(sectionID string, name string) model.Problem {
	return model.Problem{
		Code:     codeRoleVacant,
		Severity: model.SeverityWarning,
		Message:  "variable " + name + " names a vacant role: " + codeRoleVacant,
		Question: sectionID + "." + name,
		Section:  sectionID,
	}
}

func (a Assembler) namesVacantRole(answer model.Value, context planContext) bool {
	var identifiers []string
	switch answer.Kind {
	case model.ValueKindText:
		identifiers = []string{answer.Text}
	case model.ValueKindList:
		identifiers = answer.List
	}
	return slices.ContainsFunc(identifiers, func(identifier string) bool {
		return a.isVacant(identifier, context)
	})
}

func (a Assembler) isVacant(identifier string, context planContext) bool {
	for _, role := range context.directory.Roles {
		if role.ID == identifier {
			return a.writtenRole(identifier, context) == role.Name+vacantSuffix
		}
	}
	return false
}

func (a Assembler) writtenRole(identifier string, context planContext) string {
	probe := model.Section{
		ID:        "role",
		Body:      "$role",
		Variables: []model.Variable{{Name: "role", Type: model.VariableTypeRole}},
	}
	answers := model.Answers{Sections: map[string]map[string]model.Value{
		"role": {"role": {Kind: model.ValueKindText, Text: identifier}},
	}}
	filled, _ := a.resolver.Resolve(probe, answers, context.directory, context.team)
	return filled.Body
}
