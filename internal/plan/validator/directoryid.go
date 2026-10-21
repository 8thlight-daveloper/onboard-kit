package validator

import (
	"strconv"

	"onboard-kit/internal/model"
)

func directoryIDProblems(directory model.Directory, variable model.Variable, value model.Value, severity model.Severity) []model.Problem {
	kind := typeOf(variable)
	known := knownIDs(directory, kind)
	switch value.Kind {
	case model.ValueKindText:
		if known[value.Text] {
			return nil
		}
		return []model.Problem{unknownID(kind, severity, nil)}
	case model.ValueKindList:
		return unknownItemProblems(kind, known, value.List, severity)
	default:
		return nil
	}
}

func unknownItemProblems(kind model.VariableType, known map[string]bool, items []string, severity model.Severity) []model.Problem {
	problems := []model.Problem{}
	for position, item := range items {
		if item == "" || known[item] {
			continue
		}
		problems = append(problems, unknownID(kind, severity, map[string]string{"item": strconv.Itoa(position)}))
	}
	return problems
}

func unknownID(kind model.VariableType, severity model.Severity, detail map[string]string) model.Problem {
	if detail == nil {
		detail = map[string]string{}
	}
	detail["kind"] = string(kind)
	return model.Problem{
		Code:     "directory.unknown-id",
		Severity: severity,
		Message:  "This answer names an id that is not in the directory.",
		Detail:   detail,
	}
}

func knownIDs(directory model.Directory, kind model.VariableType) map[string]bool {
	ids := map[string]bool{}
	switch kind {
	case model.VariableTypeContact:
		for _, contact := range directory.Contacts {
			ids[contact.ID] = true
		}
	case model.VariableTypeRole:
		for _, role := range directory.Roles {
			ids[role.ID] = true
		}
	case model.VariableTypeTeam:
		for _, team := range directory.Teams {
			ids[team.ID] = true
		}
	}
	return ids
}
