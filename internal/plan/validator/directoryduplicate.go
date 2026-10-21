package validator

import "onboard-kit/internal/model"

func duplicateIDProblems(directory model.Directory) []model.Problem {
	problems := duplicatesOf("contact", contactIDs(directory))
	problems = append(problems, duplicatesOf("role", roleIDs(directory))...)
	return append(problems, duplicatesOf("team", teamIDs(directory))...)
}

func duplicatesOf(kind string, ids []string) []model.Problem {
	counts := map[string]int{}
	problems := []model.Problem{}
	for _, id := range ids {
		counts[id]++
		if counts[id] == 2 {
			problems = append(problems, model.Problem{
				Code:     "directory.duplicate-id",
				Severity: model.SeverityError,
				Message:  "Two or more entries of one directory list share an id.",
				Detail:   map[string]string{"kind": kind, "id": id},
			})
		}
	}
	return problems
}

func contactIDs(directory model.Directory) []string {
	ids := []string{}
	for _, contact := range directory.Contacts {
		ids = append(ids, contact.ID)
	}
	return ids
}

func roleIDs(directory model.Directory) []string {
	ids := []string{}
	for _, role := range directory.Roles {
		ids = append(ids, role.ID)
	}
	return ids
}

func teamIDs(directory model.Directory) []string {
	ids := []string{}
	for _, team := range directory.Teams {
		ids = append(ids, team.ID)
	}
	return ids
}
