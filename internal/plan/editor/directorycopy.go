package editor

import (
	"slices"

	"onboard-kit/internal/model"
)

func copyDirectoryPlan(plan model.Plan) model.Plan {
	plan = copyEditableParts(plan)
	plan.Directory = copyDirectory(plan.Directory)
	return plan
}

func copyDirectory(directory model.Directory) model.Directory {
	return model.Directory{
		Contacts: slices.Clone(directory.Contacts),
		Roles:    copyRoles(directory.Roles),
		Teams:    copyTeams(directory.Teams),
	}
}

func copyRoles(roles []model.Role) []model.Role {
	copied := slices.Clone(roles)
	for index := range copied {
		copied[index] = copyRole(copied[index])
	}
	return copied
}

func copyRole(role model.Role) model.Role {
	role.Holders = slices.Clone(role.Holders)
	return role
}

func copyTeams(teams []model.Team) []model.Team {
	copied := slices.Clone(teams)
	for index := range copied {
		copied[index] = copyTeam(copied[index])
	}
	return copied
}

func copyTeam(team model.Team) model.Team {
	team.Holders = copyOverrides(team.Holders)
	return team
}

func copyOverrides(overrides map[string][]string) map[string][]string {
	if overrides == nil {
		return nil
	}
	copied := make(map[string][]string, len(overrides))
	for role, holders := range overrides {
		copied[role] = slices.Clone(holders)
	}
	return copied
}
