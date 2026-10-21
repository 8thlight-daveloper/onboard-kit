package tui

import (
	"fmt"
	"slices"
	"strings"

	"onboard-kit/internal/model"
)

const listedLimit = 5

type removalState struct {
	kind     directoryKind
	id       string
	name     string
	next     model.Plan
	cleared  model.Cleared
	problems []model.Problem
}

func (app App) dryRun(kind directoryKind, id string) (model.Plan, model.Cleared, []model.Problem) {
	switch kind {
	case kindRole:
		return app.backend.RemoveRole(app.plan, id)
	case kindTeam:
		return app.backend.RemoveTeam(app.plan, id)
	}
	return app.backend.RemoveContact(app.plan, id)
}

func (app App) askToRemove() App {
	entry, ok := app.selectedEntry()
	if !ok {
		return app
	}
	next, cleared, problems := app.dryRun(app.people.kind, entry.id)
	if hasError(problems) {
		app.status = line{text: plainProblem(problems[0]), tone: toneProblem}
		return app
	}
	app.people.removal = removalState{kind: app.people.kind, id: entry.id, name: entry.name, next: next, cleared: cleared}
	app.people.mode = peopleRemoving
	return app
}

func (app App) confirmRemoval() App {
	state := app.people.removal
	app.plan = state.next
	app.people.mode = peopleBrowsing
	keep := ""
	if question, ok := app.current(); ok {
		keep = question.ID
	}
	app = app.refreshed(keep).refreshedPeople()
	app.status = app.saveStatus(nil)
	return app
}

func (app App) questionText(id string) string {
	scope, name, _ := strings.Cut(id, ".")
	variables := app.plan.Variables
	if scope != "plan" {
		variables = nil
		for _, section := range app.plan.Sections {
			if section.ID == scope {
				variables = section.Variables
			}
		}
	}
	for _, variable := range variables {
		if variable.Name == name && variable.Description != "" {
			return variable.Description
		}
	}
	return id
}

func pluralWords(count int, singular string, plural string) string {
	if count == 1 {
		return fmt.Sprintf("1 %s", singular)
	}
	return fmt.Sprintf("%d %s", count, plural)
}

func holderPlaces(cleared model.Cleared) (roles int, overrides int) {
	for _, place := range cleared.Holders {
		if strings.HasPrefix(place, "role:") {
			roles++
		} else {
			overrides++
		}
	}
	return roles, overrides
}

func (state removalState) consequence() string {
	var parts []string
	if count := len(state.cleared.Answers); count > 0 {
		parts = append(parts, "clears "+pluralWords(count, "answer", "answers"))
	}
	roles, overrides := holderPlaces(state.cleared)
	var holders []string
	if roles > 0 {
		holders = append(holders, pluralWords(roles, "role", "roles"))
	}
	if overrides > 0 {
		holders = append(holders, pluralWords(overrides, "team override", "team overrides"))
	}
	switch {
	case len(holders) > 0 && state.kind == kindContact:
		parts = append(parts, "removes "+state.name+" from "+strings.Join(holders, " and "))
	case len(holders) > 0:
		parts = append(parts, "removes "+strings.Join(holders, " and ")+" for this role")
	}
	if len(parts) == 0 {
		return "Nothing else changes."
	}
	return "This also " + strings.Join(parts, " and ") + "."
}

type holderPlace struct {
	role     string
	team     string
	roleName string
	teamName string
}

func (app App) holderPlaceOf(place string) holderPlace {
	directory := app.plan.Directory
	if role, found := strings.CutPrefix(place, "role:"); found {
		return holderPlace{role: role, roleName: roleNameOf(directory, role)}
	}
	team, role, _ := strings.Cut(strings.TrimPrefix(place, "team:"), "/")
	found := holderPlace{role: role, team: team, teamName: team, roleName: roleNameOf(directory, role)}
	for _, each := range directory.Teams {
		if each.ID == team {
			found.teamName = each.Name
		}
	}
	return found
}

func (place holderPlace) text() string {
	if place.team == "" {
		return place.roleName + " role"
	}
	return place.teamName + ", as " + place.roleName
}

func (app App) vacantPlaces(state removalState) []string {
	if state.kind != kindContact {
		return nil
	}
	var vacant []string
	for _, text := range state.cleared.Holders {
		place := app.holderPlaceOf(text)
		if place.team == "" {
			index := slices.IndexFunc(state.next.Directory.Roles, func(each model.Role) bool { return each.ID == place.role })
			if index >= 0 && len(state.next.Directory.Roles[index].Holders) == 0 {
				vacant = append(vacant, place.roleName+" role has nobody left")
			}
			continue
		}
		index := slices.IndexFunc(state.next.Directory.Teams, func(each model.Team) bool { return each.ID == place.team })
		if index >= 0 && len(state.next.Directory.Teams[index].Holders[place.role]) == 0 {
			vacant = append(vacant, place.roleName+" is vacant for "+place.teamName)
		}
	}
	return vacant
}

func listedWithLimit(texts []string) []string {
	if len(texts) <= listedLimit {
		return texts
	}
	more := len(texts) - listedLimit + 1
	return append(slices.Clone(texts[:listedLimit-1]), fmt.Sprintf("and %d more", more))
}
