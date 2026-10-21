package editor

import (
	"slices"

	"onboard-kit/internal/model"
)

func removeHolder(directory *model.Directory, contact string) []string {
	places := []string{}
	for index := range directory.Roles {
		role := &directory.Roles[index]
		if without, removed := withoutItem(role.Holders, contact); removed {
			role.Holders = without
			places = append(places, "role:"+role.ID)
		}
	}
	for _, team := range directory.Teams {
		for role, holders := range team.Holders {
			if without, removed := withoutItem(holders, contact); removed {
				team.Holders[role] = without
				places = append(places, "team:"+team.ID+"/"+role)
			}
		}
	}
	return append(places, clearTeamContacts(directory, contact)...)
}

func clearTeamContacts(directory *model.Directory, contact string) []string {
	places := []string{}
	for index := range directory.Teams {
		team := &directory.Teams[index]
		if team.Contact == contact {
			team.Contact = ""
			places = append(places, "team-contact:"+team.ID)
		}
	}
	return places
}

func removeOverrides(directory *model.Directory, role string) []string {
	places := []string{}
	for _, team := range directory.Teams {
		if _, overridden := team.Holders[role]; overridden {
			delete(team.Holders, role)
			places = append(places, "team:"+team.ID+"/"+role)
		}
	}
	return places
}

func withoutItem(items []string, removed string) ([]string, bool) {
	kept := slices.DeleteFunc(slices.Clone(items), func(item string) bool { return item == removed })
	return kept, len(kept) != len(items)
}
