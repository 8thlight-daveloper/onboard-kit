package resolver

import (
	"strings"

	"onboard-kit/internal/model"
)

func (b book) writtenRole(id string) (string, bool) {
	for _, role := range b.directory.Roles {
		if role.ID == id {
			return b.roleText(role), true
		}
	}
	return "", false
}

func (b book) roleText(role model.Role) string {
	var holders []string
	for _, holderID := range b.holderIDs(role) {
		if text, found := b.writtenContact(holderID); found {
			holders = append(holders, text)
		}
	}
	if len(holders) == 0 {
		return role.Name + " (vacant)"
	}
	return strings.Join(holders, directorySeparator)
}

func (b book) holderIDs(role model.Role) []string {
	for _, team := range b.directory.Teams {
		if team.ID != b.team {
			continue
		}
		if override, overridden := team.Holders[role.ID]; overridden {
			return override
		}
	}
	return role.Holders
}
