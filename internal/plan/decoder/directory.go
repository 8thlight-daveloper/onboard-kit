package decoder

import (
	"go.yaml.in/yaml/v3"

	"onboard-kit/internal/model"
)

func (c *collector) directoryOf(node *yaml.Node, at location) model.Directory {
	var directory model.Directory
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return directory
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "contacts":
			directory.Contacts = c.contactsOf(found.value, found.at)
		case "roles":
			directory.Roles = c.rolesOf(found.value, found.at)
		case "teams":
			directory.Teams = c.teamsOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return directory
}

func (c *collector) contactsOf(node *yaml.Node, at location) []model.Contact {
	var contacts []model.Contact
	c.items(node, at, codeEmptyItem, func(item *yaml.Node, place location) {
		contacts = append(contacts, c.contactOf(item, place))
	})
	return contacts
}

func (c *collector) contactOf(node *yaml.Node, at location) model.Contact {
	var contact model.Contact
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return contact
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "id":
			contact.ID = c.stringOf(found.value, found.at)
		case "name":
			contact.Name = c.stringOf(found.value, found.at)
		case "email":
			contact.Email = c.stringOf(found.value, found.at)
		case "slack":
			contact.Slack = c.stringOf(found.value, found.at)
		case "slackId":
			contact.SlackID = c.stringOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return contact
}

func (c *collector) rolesOf(node *yaml.Node, at location) []model.Role {
	var roles []model.Role
	c.items(node, at, codeEmptyItem, func(item *yaml.Node, place location) {
		roles = append(roles, c.roleOf(item, place))
	})
	return roles
}

func (c *collector) roleOf(node *yaml.Node, at location) model.Role {
	var role model.Role
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return role
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "id":
			role.ID = c.stringOf(found.value, found.at)
		case "name":
			role.Name = c.stringOf(found.value, found.at)
		case "holders":
			role.Holders = c.stringListOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return role
}

func (c *collector) teamsOf(node *yaml.Node, at location) []model.Team {
	var teams []model.Team
	c.items(node, at, codeEmptyItem, func(item *yaml.Node, place location) {
		teams = append(teams, c.teamOf(item, place))
	})
	return teams
}

func (c *collector) teamOf(node *yaml.Node, at location) model.Team {
	var team model.Team
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return team
	}
	for _, found := range c.entries(mapping, at) {
		switch found.name {
		case "id":
			team.ID = c.stringOf(found.value, found.at)
		case "name":
			team.Name = c.stringOf(found.value, found.at)
		case "holders":
			team.Holders = c.holdersByRoleOf(found.value, found.at)
		default:
			c.rejectUnknown(found)
		}
	}
	return team
}

func (c *collector) holdersByRoleOf(node *yaml.Node, at location) map[string][]string {
	mapping := c.mappingOf(node, at)
	if mapping == nil {
		return nil
	}
	holders := map[string][]string{}
	for _, found := range c.entries(mapping, at) {
		holders[found.name] = c.stringListOf(found.value, found.at)
	}
	return holders
}
