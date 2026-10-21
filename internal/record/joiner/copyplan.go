package joiner

import "onboard-kit/internal/model"

func copyPlan(plan model.Plan) model.Plan {
	copied := plan
	copied.Directory = copyDirectory(plan.Directory)
	copied.Variables = copyEach(plan.Variables, copyVariable)
	copied.Sections = copyEach(plan.Sections, copySection)
	copied.UnknownFields = copyUnknownFields(plan.UnknownFields)
	return copied
}

func copyDirectory(directory model.Directory) model.Directory {
	return model.Directory{
		Contacts: copyEach(directory.Contacts, identity[model.Contact]),
		Roles:    copyEach(directory.Roles, copyRole),
		Teams:    copyEach(directory.Teams, copyTeam),
	}
}

func copyRole(role model.Role) model.Role {
	role.Holders = copyEach(role.Holders, identity[string])
	return role
}

func copyTeam(team model.Team) model.Team {
	team.Holders = copyMap(team.Holders, func(holders []string) []string {
		return copyEach(holders, identity[string])
	})
	return team
}
