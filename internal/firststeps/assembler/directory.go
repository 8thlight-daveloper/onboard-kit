package assembler

import (
	"strings"
	"unicode"

	contract "onboard-kit/contracts/firststeps/assembler"
	"onboard-kit/internal/model"
)

type directory struct {
	contacts map[string]model.Contact
	roles    map[string]model.Role
	holders  map[string][]string
}

func newDirectory(value model.Directory, teamID string) directory {
	directory := directory{contacts: make(map[string]model.Contact), roles: make(map[string]model.Role)}
	for _, contact := range value.Contacts {
		directory.contacts[contact.ID] = contact
	}
	for _, role := range value.Roles {
		directory.roles[role.ID] = role
	}
	for _, team := range value.Teams {
		if team.ID == teamID {
			directory.holders = team.Holders
		}
	}
	return directory
}

func (directory directory) has(reference contract.HelpReference) bool {
	switch reference.Kind {
	case "contact":
		_, exists := directory.contacts[reference.ID]
		return exists
	case "role":
		_, exists := directory.roles[reference.ID]
		return exists
	default:
		return false
	}
}

func blank(value string) bool {
	return strings.TrimFunc(value, unicode.IsSpace) == ""
}

func contactDisplay(contact model.Contact) string {
	name := "Name not supplied"
	if !blank(contact.Name) {
		name = displayText(contact.Name)
	}
	if !blank(contact.Slack) {
		return name + ", " + displayText(contact.Slack)
	}
	if !blank(contact.Email) {
		return name + ", " + displayText(contact.Email)
	}
	return name + " (contact route not supplied)"
}

func (directory directory) helpDisplay(reference contract.HelpReference) string {
	if reference.Kind == "contact" {
		return contactDisplay(directory.contacts[reference.ID])
	}
	role := directory.roles[reference.ID]
	holders := role.Holders
	if override, exists := directory.holders[role.ID]; exists {
		holders = override
	}
	if len(holders) == 0 {
		name := "Unnamed role"
		if !blank(role.Name) {
			name = displayText(role.Name)
		}
		return name + " (vacant)"
	}
	contacts := make([]string, 0, len(holders))
	for _, id := range holders {
		contacts = append(contacts, contactDisplay(directory.contacts[id]))
	}
	return strings.Join(contacts, "; ")
}

func (directory directory) helpList(references []contract.HelpReference) string {
	if len(references) == 0 {
		return "General help contacts not specified. Ask the onboarding owner to provide a help route that works before company access."
	}
	var items []string
	seen := make(map[contract.HelpReference]bool)
	for _, reference := range references {
		if !seen[reference] {
			items = append(items, "- "+directory.helpDisplay(reference))
			seen[reference] = true
		}
	}
	return strings.Join(items, "\n")
}
