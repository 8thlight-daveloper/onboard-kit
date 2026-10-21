package validator

import (
	"slices"

	"onboard-kit/internal/model"
)

func holderProblems(directory model.Directory) []model.Problem {
	contacts := contactIDs(directory)
	roles := roleIDs(directory)
	problems := []model.Problem{}
	for _, role := range directory.Roles {
		problems = append(problems, holdersNotContacts(contacts, role.Holders, map[string]string{"role": role.ID})...)
		if len(role.Holders) == 0 {
			problems = append(problems, vacantRole(role))
		}
	}
	for _, team := range directory.Teams {
		problems = append(problems, teamContactProblems(contacts, team)...)
		problems = append(problems, teamOverrideProblems(contacts, roles, team)...)
	}
	return problems
}

func teamContactProblems(contacts []string, team model.Team) []model.Problem {
	if team.Contact == "" || slices.Contains(contacts, team.Contact) {
		return []model.Problem{}
	}
	return []model.Problem{{
		Code:     "directory.holder-not-a-contact",
		Severity: model.SeverityError,
		Message:  "A team contact is not a contact.",
		Detail:   map[string]string{"team": team.ID, "team-contact": team.Contact},
	}}
}

func teamOverrideProblems(contacts []string, roles []string, team model.Team) []model.Problem {
	problems := []model.Problem{}
	for role, holders := range team.Holders {
		if !slices.Contains(roles, role) {
			problems = append(problems, model.Problem{
				Code:     "directory.team-role-unknown",
				Severity: model.SeverityError,
				Message:  "A team overrides a role that does not exist.",
				Detail:   map[string]string{"team": team.ID, "role": role},
			})
		}
		problems = append(problems, holdersNotContacts(contacts, holders, map[string]string{"role": role, "team": team.ID})...)
	}
	return problems
}

func holdersNotContacts(contacts []string, holders []string, owner map[string]string) []model.Problem {
	problems := []model.Problem{}
	for _, holder := range holders {
		if slices.Contains(contacts, holder) {
			continue
		}
		detail := map[string]string{"holder": holder}
		for key, value := range owner {
			detail[key] = value
		}
		problems = append(problems, model.Problem{
			Code:     "directory.holder-not-a-contact",
			Severity: model.SeverityError,
			Message:  "A holder is not a contact.",
			Detail:   detail,
		})
	}
	return problems
}

func vacantRole(role model.Role) model.Problem {
	return model.Problem{
		Code:     "directory.role-vacant",
		Severity: model.SeverityWarning,
		Message:  "This role has no holder.",
		Detail:   map[string]string{"role": role.ID},
	}
}
