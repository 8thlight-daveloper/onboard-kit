package editor

import (
	"slices"

	"onboard-kit/internal/model"
)

func (Editor) SetContact(plan model.Plan, contact model.Contact) (model.Plan, []model.Problem) {
	problems := entryFaults(entry{kind: model.VariableTypeContact, id: contact.ID, name: contact.Name})
	return storeEntry(plan, problems, func(directory *model.Directory) {
		directory.Contacts = stored(directory.Contacts, contact, func(held model.Contact) bool { return held.ID == contact.ID })
	})
}

func (Editor) SetRole(plan model.Plan, role model.Role) (model.Plan, []model.Problem) {
	problems := roleFaults(plan.Directory, role)
	return storeEntry(plan, problems, func(directory *model.Directory) {
		directory.Roles = stored(directory.Roles, copyRole(role), func(held model.Role) bool { return held.ID == role.ID })
	})
}

func (Editor) SetTeam(plan model.Plan, team model.Team) (model.Plan, []model.Problem) {
	problems := teamFaults(plan.Directory, team)
	return storeEntry(plan, problems, func(directory *model.Directory) {
		directory.Teams = stored(directory.Teams, copyTeam(team), func(held model.Team) bool { return held.ID == team.ID })
	})
}

func storeEntry(plan model.Plan, problems []model.Problem, store func(*model.Directory)) (model.Plan, []model.Problem) {
	plan = copyDirectoryPlan(plan)
	if len(problems) == 0 {
		store(&plan.Directory)
	}
	return plan, problems
}

func stored[Entry any](entries []Entry, entry Entry, isSame func(Entry) bool) []Entry {
	index := slices.IndexFunc(entries, isSame)
	if index < 0 {
		return append(entries, entry)
	}
	entries[index] = entry
	return entries
}
