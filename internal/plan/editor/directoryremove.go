package editor

import (
	"slices"

	"onboard-kit/internal/model"
)

func (Editor) RemoveContact(plan model.Plan, id string) (model.Plan, model.Cleared, []model.Problem) {
	plan = copyDirectoryPlan(plan)
	index := slices.IndexFunc(plan.Directory.Contacts, func(contact model.Contact) bool { return contact.ID == id })
	if index < 0 {
		return refusedRemoval(plan, model.VariableTypeContact, id)
	}
	plan.Directory.Contacts = slices.Delete(plan.Directory.Contacts, index, index+1)
	places := removeHolder(&plan.Directory, id)
	return clearedRemoval(plan, model.VariableTypeContact, id, places)
}

func (Editor) RemoveRole(plan model.Plan, id string) (model.Plan, model.Cleared, []model.Problem) {
	plan = copyDirectoryPlan(plan)
	index := slices.IndexFunc(plan.Directory.Roles, func(role model.Role) bool { return role.ID == id })
	if index < 0 {
		return refusedRemoval(plan, model.VariableTypeRole, id)
	}
	plan.Directory.Roles = slices.Delete(plan.Directory.Roles, index, index+1)
	places := removeOverrides(&plan.Directory, id)
	return clearedRemoval(plan, model.VariableTypeRole, id, places)
}

func (Editor) RemoveTeam(plan model.Plan, id string) (model.Plan, model.Cleared, []model.Problem) {
	plan = copyDirectoryPlan(plan)
	index := slices.IndexFunc(plan.Directory.Teams, func(team model.Team) bool { return team.ID == id })
	if index < 0 {
		return refusedRemoval(plan, model.VariableTypeTeam, id)
	}
	plan.Directory.Teams = slices.Delete(plan.Directory.Teams, index, index+1)
	return clearedRemoval(plan, model.VariableTypeTeam, id, nil)
}

func clearedRemoval(plan model.Plan, kind model.VariableType, id string, places []string) (model.Plan, model.Cleared, []model.Problem) {
	plan, answers := clearAnswers(plan, kind, id)
	return plan, model.Cleared{Answers: answers, Holders: sortedUnique(places)}, []model.Problem{}
}

func refusedRemoval(plan model.Plan, kind model.VariableType, id string) (model.Plan, model.Cleared, []model.Problem) {
	problem := model.Problem{
		Code:     "directory.entry-not-found",
		Severity: model.SeverityError,
		Message:  "The directory has no entry of that kind with that id.",
		Detail:   map[string]string{"kind": string(kind), "id": id},
	}
	return plan, model.Cleared{Answers: []string{}, Holders: []string{}}, []model.Problem{problem}
}

func sortedUnique(places []string) []string {
	sorted := slices.Clone(places)
	slices.Sort(sorted)
	return append([]string{}, slices.Compact(sorted)...)
}
