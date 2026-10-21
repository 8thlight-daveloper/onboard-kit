package editor

import (
	"maps"
	"slices"

	"onboard-kit/internal/model"
)

const (
	idMalformedCode     = "directory.id-malformed"
	nameEmptyCode       = "directory.name-empty"
	notContactCode      = "directory.holder-not-a-contact"
	teamRoleUnknownCode = "directory.team-role-unknown"
)

type entry struct {
	kind model.VariableType
	id   string
	name string
}

func entryFaults(subject entry) []model.Problem {
	problems := []model.Problem{}
	if !isWellFormedID(subject.id) {
		problems = append(problems, entryProblem(idMalformedCode, "This id is empty or has characters other than a to z, 0 to 9 and hyphens.", subject, nil))
	}
	if subject.name == "" {
		problems = append(problems, entryProblem(nameEmptyCode, "This entry has no name.", subject, nil))
	}
	return problems
}

func roleFaults(directory model.Directory, role model.Role) []model.Problem {
	subject := entry{kind: model.VariableTypeRole, id: role.ID, name: role.Name}
	problems := entryFaults(subject)
	return append(problems, holderFaults(directory, subject, role.Holders, nil)...)
}

func teamFaults(directory model.Directory, team model.Team) []model.Problem {
	subject := entry{kind: model.VariableTypeTeam, id: team.ID, name: team.Name}
	problems := entryFaults(subject)
	if team.Contact != "" && !hasContact(directory, team.Contact) {
		problems = append(problems, entryProblem(notContactCode, "This team has a contact that is not a contact in the directory.", subject, map[string]string{"contact": team.Contact}))
	}
	for _, role := range slices.Sorted(maps.Keys(team.Holders)) {
		if !hasRole(directory, role) {
			problems = append(problems, entryProblem(teamRoleUnknownCode, "This team overrides a role that is not in the directory.", subject, map[string]string{"role": role}))
		}
		problems = append(problems, holderFaults(directory, subject, team.Holders[role], map[string]string{"role": role})...)
	}
	return problems
}

func holderFaults(directory model.Directory, subject entry, holders []string, extra map[string]string) []model.Problem {
	problems := []model.Problem{}
	for _, holder := range holders {
		if hasContact(directory, holder) {
			continue
		}
		detail := map[string]string{"holder": holder}
		maps.Copy(detail, extra)
		problems = append(problems, entryProblem(notContactCode, "This holder is not a contact in the directory.", subject, detail))
	}
	return problems
}

func entryProblem(code, message string, subject entry, extra map[string]string) model.Problem {
	detail := map[string]string{"kind": string(subject.kind), "id": subject.id}
	maps.Copy(detail, extra)
	return model.Problem{Code: code, Severity: model.SeverityError, Message: message, Detail: detail}
}

func hasContact(directory model.Directory, id string) bool {
	return slices.ContainsFunc(directory.Contacts, func(contact model.Contact) bool { return contact.ID == id })
}

func hasRole(directory model.Directory, id string) bool {
	return slices.ContainsFunc(directory.Roles, func(role model.Role) bool { return role.ID == id })
}

func isWellFormedID(id string) bool {
	if id == "" || id[0] == '-' {
		return false
	}
	for index := 0; index < len(id); index++ {
		character := id[index]
		isLower := character >= 'a' && character <= 'z'
		isDigit := character >= '0' && character <= '9'
		if !isLower && !isDigit && character != '-' {
			return false
		}
	}
	return true
}
