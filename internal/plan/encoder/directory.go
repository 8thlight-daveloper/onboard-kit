package encoder

import "onboard-kit/internal/model"

func (c *collector) directoryNode(directory model.Directory) node {
	builder := &mappingBuilder{}
	builder.add("contacts", c.contactSequence(directory.Contacts))
	builder.add("roles", c.roleSequence(directory.Roles))
	builder.add("teams", c.teamSequence(directory.Teams))
	return builder.build()
}

func (c *collector) contactSequence(contacts []model.Contact) node {
	return sequenceOf(contacts, func(index int, contact model.Contact) node {
		builder := &mappingBuilder{}
		builder.add("id", stringField(contact.ID))
		builder.add("name", stringField(contact.Name))
		builder.add("email", stringField(contact.Email))
		builder.add("slack", stringField(contact.Slack))
		builder.add("slackId", stringField(contact.SlackID))
		return c.requireKeys(builder.build(), indexed("directory.contacts", index))
	})
}

func (c *collector) roleSequence(roles []model.Role) node {
	return sequenceOf(roles, func(index int, role model.Role) node {
		builder := &mappingBuilder{}
		builder.add("id", stringField(role.ID))
		builder.add("name", stringField(role.Name))
		builder.add("holders", listField(role.Holders))
		return c.requireKeys(builder.build(), indexed("directory.roles", index))
	})
}

func (c *collector) teamSequence(teams []model.Team) node {
	return sequenceOf(teams, func(index int, team model.Team) node {
		builder := &mappingBuilder{}
		builder.add("id", stringField(team.ID))
		builder.add("name", stringField(team.Name))
		builder.add("holders", holdersByRoleNode(team.Holders))
		return c.requireKeys(builder.build(), indexed("directory.teams", index))
	})
}

func holdersByRoleNode(holders map[string][]string) node {
	builder := &mappingBuilder{}
	for _, roleID := range sortedKeys(holders) {
		builder.add(roleID, valueList(holders[roleID]))
	}
	return builder.build()
}
